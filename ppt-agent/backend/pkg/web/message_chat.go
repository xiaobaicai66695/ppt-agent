package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/assets/unsplash"
	"github.com/cloudwego/ppt-agent/pkg/tools"
	"github.com/cloudwego/ppt-agent/pkg/tools/search"
)

type chatImageResult struct {
	PreviewURL      string `json:"preview_url"`
	ImageURL        string `json:"image_url"`
	SourceURL       string `json:"source_url"`
	Photographer    string `json:"photographer"`
	PhotographerURL string `json:"photographer_url"`
	Attribution     string `json:"attribution"`
}

type chatAugmentations struct {
	promptParts []string
	webResults  []search.SearchResult
	images      []chatImageResult
	query       string
}

func (s *Server) buildChatReply(ctx context.Context, message, fallback, conversationContext string, forceWebSearch, forceImageSearch bool) string {
	fallback = strings.TrimSpace(fallback)
	augmentations := s.collectChatAugmentations(ctx, message, conversationContext, forceWebSearch, forceImageSearch)
	return s.buildChatReplyWithAugmentations(ctx, message, fallback, augmentations)
}

func (s *Server) buildChatReplyWithAugmentations(ctx context.Context, message, fallback string, augmentations chatAugmentations) string {
	fallback = strings.TrimSpace(fallback)

	modelFactory := s.textModelFactory
	if modelFactory == nil {
		modelFactory = s.aiModelFactory
	}
	if modelFactory == nil {
		return appendChatSupplement(chatFallbackReply(message, fallback, augmentations), augmentations)
	}
	model, err := modelFactory(ctx)
	if err != nil || model == nil {
		return appendChatSupplement(chatFallbackReply(message, fallback, augmentations), augmentations)
	}

	chatCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	prompt := chatReplyPrompt(message, augmentations.promptParts)

	resp, err := model.Generate(chatCtx, []*schema.Message{schema.UserMessage(prompt)})
	if err != nil || resp == nil || strings.TrimSpace(resp.Content) == "" {
		return appendChatSupplement(chatFallbackReply(message, fallback, augmentations), augmentations)
	}
	reply := strings.TrimSpace(resp.Content)
	if len(augmentations.webResults) > 0 && isGenericChatReply(reply) {
		reply = chatFallbackReply(message, fallback, augmentations)
	}
	return appendChatSupplement(reply, augmentations)
}

// streamChatReply runs the existing chat model asynchronously, then delivers
// the result in SSE chunks. Its outer deadline protects the composer from a
// model provider that ignores cancellation on a streaming request.
func (s *Server) streamChatReply(ctx context.Context, message, fallback, conversationContext string, forceWebSearch, forceImageSearch bool, emit func(string)) {
	augmentations := s.collectChatAugmentations(ctx, message, conversationContext, forceWebSearch, forceImageSearch)
	result := make(chan string, 1)
	go func() {
		result <- s.buildChatReplyWithAugmentations(ctx, message, fallback, augmentations)
	}()

	var reply string
	select {
	case reply = <-result:
	case <-time.After(35 * time.Second):
		// Preserve the same retrieval-first and capability-degradation contract
		// as the synchronous path when an upstream model ignores cancellation.
		reply = appendChatSupplement(chatFallbackReply(message, fallback, augmentations), augmentations)
	}
	for _, chunk := range splitChatReplyForSSE(reply) {
		emit(chunk)
	}
}

func chatReplyPrompt(message string, augmentations []string) string {
	return fmt.Sprintf(`你是 PPT Agent 工作台里的闲聊助手。当前消息已经被后端判定为 chat，不得创建 PPT 任务，不得承诺已经开始生成文件。

回答要求：
- 直接回答用户问题。
- 如果补充材料里有 web search 或 image search 结果，整合它们组织答案，并保留简短来源说明。
- 如果补充材料显示某能力未配置，只说明当前无法使用该能力，不要编造搜索结果。
- 如果用户实际想做 PPT，引导其补充主题、受众、页数、风格或选择已有任务修复。

用户消息：
%s

补充材料：
%s`, message, strings.Join(augmentations, "\n\n"))
}

func splitChatReplyForSSE(reply string) []string {
	runes := []rune(strings.TrimSpace(reply))
	if len(runes) == 0 {
		return nil
	}
	const chunkSize = 28
	chunks := make([]string, 0, (len(runes)+chunkSize-1)/chunkSize)
	for len(runes) > 0 {
		size := chunkSize
		if len(runes) < size {
			size = len(runes)
		}
		chunks = append(chunks, string(runes[:size]))
		runes = runes[size:]
	}
	return chunks
}

func (s *Server) collectChatAugmentations(ctx context.Context, message, conversationContext string, forceWebSearch, forceImageSearch bool) chatAugmentations {
	message = strings.TrimSpace(message)
	if message == "" {
		return chatAugmentations{}
	}
	augmentations := chatAugmentations{}
	if conversation := compactChatConversationContext(conversationContext); conversation != "" {
		// Follow-up chat requests often omit the subject (for example "那预算"
		// or "再补充一条"). Keep a bounded recent context in the text-model
		// prompt so direct answers remain grounded even without web retrieval.
		augmentations.promptParts = append(augmentations.promptParts, "conversation_context:\n"+conversation)
	}
	query := chatSearchQuery(message, conversationContext)
	augmentations.query = query
	if forceWebSearch || chatNeedsWebSearch(message) {
		searchCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		out, err := tools.NewSearchTool().InvokableRun(searchCtx, mustJSON(map[string]string{
			"query":  query,
			"reason": "用户在闲聊中请求最新或需要核实的信息",
		}))
		if err != nil {
			augmentations.promptParts = append(augmentations.promptParts, "web_search_error: "+err.Error())
		} else if strings.TrimSpace(out) != "" {
			augmentations.promptParts = append(augmentations.promptParts, "web_search_result:\n"+out)
			var response search.SearchResponse
			if json.Unmarshal([]byte(out), &response) == nil {
				augmentations.webResults = append(augmentations.webResults, response.Results...)
			}
		}
	}
	if forceImageSearch || chatNeedsImageSearch(message) {
		if !unsplash.IsConfigured() {
			augmentations.promptParts = append(augmentations.promptParts, "image_search_error: 未配置 UNSPLASH_ACCESS_KEY，当前无法检索图片候选。")
		} else if client, err := unsplash.NewClientFromEnv(); err == nil {
			imageCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
			defer cancel()
			response, searchErr := client.Search(imageCtx, unsplash.SearchOptions{
				Query:         query,
				ContentFilter: "high",
				OrderBy:       "relevant",
				PerPage:       2,
			})
			if searchErr != nil {
				augmentations.promptParts = append(augmentations.promptParts, "image_search_error: "+searchErr.Error())
			} else {
				augmentations.images = append(augmentations.images, chatImageResults(response)...)
			}
		} else {
			augmentations.promptParts = append(augmentations.promptParts, "image_search_error: "+err.Error())
		}
	}
	return augmentations
}

func chatImageResults(response *unsplash.SearchResponse) []chatImageResult {
	if response == nil || len(response.Results) == 0 {
		return nil
	}
	results := make([]chatImageResult, 0, min(len(response.Results), 2))
	for _, photo := range response.Results[:min(len(response.Results), 2)] {
		photographer := strings.TrimSpace(photo.User.Name)
		if photographer == "" {
			photographer = strings.TrimSpace(photo.User.Username)
		}
		results = append(results, chatImageResult{
			PreviewURL:      firstNonEmpty(photo.URLs.Small, photo.URLs.Regular, photo.URLs.Thumb),
			ImageURL:        firstNonEmpty(photo.URLs.Regular, photo.URLs.Full, photo.URLs.Small),
			SourceURL:       unsplash.AttributionURL(photo.Links.HTML),
			Photographer:    photographer,
			PhotographerURL: unsplash.AttributionURL(photo.User.Links.HTML),
			Attribution:     chatImageAttribution(photographer),
		})
	}
	return results
}

func chatImageAttribution(photographer string) string {
	if photographer == "" {
		return "Photo on Unsplash"
	}
	return "Photo by " + photographer + " on Unsplash"
}

func chatNeedsWebSearch(message string) bool {
	normalized := strings.ToLower(message)
	for _, keyword := range []string{"最新", "今天", "现在", "近期", "新闻", "搜索", "查一下", "websearch", "web search", "联网", "资料来源", "补充材料", "补充些材料", "补充资料", "攻略", "source"} {
		if strings.Contains(normalized, keyword) {
			return true
		}
	}
	return false
}

func chatNeedsImageSearch(message string) bool {
	normalized := strings.ToLower(message)
	for _, keyword := range []string{"图片", "配图", "找图", "照片", "素材图", "image", "photo", "visual"} {
		if strings.Contains(normalized, keyword) {
			return true
		}
	}
	return false
}

func compactSearchQuery(message string) string {
	message = strings.Join(strings.Fields(message), " ")
	runes := []rune(message)
	if len(runes) > 80 {
		return string(runes[:80])
	}
	return message
}

func compactChatConversationContext(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	runes := []rune(value)
	const maxRunes = 2400
	if len(runes) <= maxRunes {
		return value
	}
	return string(runes[len(runes)-maxRunes:])
}

func chatSearchQuery(message, conversationContext string) string {
	message = compactSearchQuery(message)
	if !chatNeedsContextualQuery(message) || strings.TrimSpace(conversationContext) == "" {
		return message
	}
	for _, line := range strings.Split(conversationContext, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "用户：") {
			continue
		}
		prior := strings.TrimSpace(strings.TrimPrefix(line, "用户："))
		if prior == "" || prior == message {
			continue
		}
		return compactSearchQuery(prior + " " + message)
	}
	return message
}

func chatNeedsContextualQuery(message string) bool {
	for _, keyword := range []string{"补充材料", "补充些材料", "补充资料", "找图", "配图", "照片", "素材图"} {
		if strings.Contains(message, keyword) {
			return true
		}
	}
	return false
}

func mustJSON(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func chatFallbackReply(message, fallback string, augmentations chatAugmentations) string {
	if len(augmentations.webResults) > 0 {
		topic := strings.TrimSpace(augmentations.query)
		if topic == "" {
			topic = message
		}
		return chatSearchFallback(topic, augmentations.webResults)
	}
	if strings.TrimSpace(fallback) != "" {
		return strings.TrimSpace(fallback)
	}
	if hasImageSearchConfigurationError(augmentations.promptParts) {
		return "当前无法检索图片参考：服务端尚未配置图片搜索。你可以先继续描述想要的画面、地点或风格，我会据此组织文字内容。"
	}
	return "暂时无法调用普通对话模型。请稍后重试，或直接说明 PPT 的主题、受众、页数和使用场景。"
}

func chatSearchFallback(message string, results []search.SearchResult) string {
	items := make([]string, 0, min(len(results), 3))
	for _, result := range results[:min(len(results), 3)] {
		title := strings.TrimSpace(result.Title)
		description := compactChatDescription(result.Description)
		if title == "" && description == "" {
			continue
		}
		if description == "" {
			items = append(items, "- "+title)
			continue
		}
		if title == "" {
			items = append(items, "- "+description)
			continue
		}
		items = append(items, fmt.Sprintf("- **%s**：%s", title, description))
	}
	if len(items) == 0 {
		return "已找到可供核对的补充资料，但摘要为空。请直接打开下方来源查看原文。"
	}
	query := strings.TrimSpace(message)
	if query == "" {
		query = "这个主题"
	}
	return fmt.Sprintf("围绕“%s”，我已整理到以下可直接用于进一步了解和核对的资料：\n\n%s\n\n下方附有可点击的原始来源；如需，我可以继续把这些资料整理成路线、要点或 PPT 大纲。", query, strings.Join(items, "\n"))
}

// chatSafeURL only permits complete HTTP(S) URLs in assistant Markdown. Search
// providers occasionally return truncated/whitespace-corrupted redirect URLs;
// emitting those as links produces the broken sources shown in the workbench.
func chatSafeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.ContainsAny(raw, "\r\n\t ") {
		return ""
	}
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return ""
	}
	return parsed.String()
}

func chatSafeLabel(value, fallback string) string {
	value = strings.TrimSpace(value)
	value = strings.NewReplacer("[", "（", "]", "）", "\r", " ", "\n", " ").Replace(value)
	if value == "" {
		return fallback
	}
	return value
}

func compactChatDescription(value string) string {
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	runes := []rune(value)
	if len(runes) > 120 {
		return string(runes[:120]) + "…"
	}
	return value
}

func hasImageSearchConfigurationError(parts []string) bool {
	for _, part := range parts {
		if strings.Contains(part, "未配置 UNSPLASH_ACCESS_KEY") {
			return true
		}
	}
	return false
}

func isGenericChatReply(reply string) bool {
	reply = strings.TrimSpace(reply)
	for _, generic := range []string{
		"我可以先按普通对话回答",
		"请说明主题、受众、页数和使用场景",
		"你仍可以明确输入 PPT 创建、规划或修复需求",
	} {
		if strings.Contains(reply, generic) {
			return true
		}
	}
	return false
}

func appendChatSupplement(reply string, augmentations chatAugmentations) string {
	supplement := chatSupplement(augmentations)
	if supplement == "" {
		return strings.TrimSpace(reply)
	}
	return strings.TrimSpace(reply) + "\n\n" + supplement
}

func chatSupplement(augmentations chatAugmentations) string {
	sections := make([]string, 0, 2)
	if len(augmentations.webResults) > 0 {
		items := make([]string, 0, min(len(augmentations.webResults), 4))
		seenURLs := make(map[string]struct{})
		for _, result := range augmentations.webResults[:min(len(augmentations.webResults), 4)] {
			link := chatSafeURL(result.URL)
			if link == "" {
				continue
			}
			if _, duplicate := seenURLs[link]; duplicate {
				continue
			}
			seenURLs[link] = struct{}{}
			items = append(items, fmt.Sprintf("- [%s](%s)", chatSafeLabel(result.Title, "资料来源"), link))
		}
		if len(items) > 0 {
			sections = append(sections, "### 补充资料来源\n"+strings.Join(items, "\n"))
		}
	}
	if len(augmentations.images) > 0 {
		items := make([]string, 0, min(len(augmentations.images), 2))
		seenImages := make(map[string]struct{})
		for _, image := range augmentations.images[:min(len(augmentations.images), 2)] {
			preview := strings.TrimSpace(image.PreviewURL)
			if preview == "" {
				preview = strings.TrimSpace(image.ImageURL)
			}
			preview = chatSafeURL(preview)
			if preview == "" {
				continue
			}
			if _, duplicate := seenImages[preview]; duplicate {
				continue
			}
			seenImages[preview] = struct{}{}
			label := strings.TrimSpace(image.Attribution)
			if label == "" {
				label = strings.TrimSpace(image.Photographer)
			}
			label = chatSafeLabel(label, "图片参考")
			credit := ""
			if photographerURL := chatSafeURL(image.PhotographerURL); photographerURL != "" {
				photographer := chatSafeLabel(image.Photographer, label)
				credit = fmt.Sprintf("\n摄影：[%s](%s) · [Unsplash](https://unsplash.com/?utm_source=ppt_agent&utm_medium=referral)", photographer, photographerURL)
			}
			items = append(items, fmt.Sprintf("![%s](%s)%s", label, preview, credit))
		}
		if len(items) > 0 {
			sections = append(sections, "### 图片参考\n"+strings.Join(items, "\n"))
		}
	}
	if len(augmentations.images) == 0 && hasImageSearchError(augmentations.promptParts) {
		sections = append(sections, "### 图片参考\n当前未检索到图片候选：图片搜索能力未配置或暂不可用。")
	}
	return strings.Join(sections, "\n\n")
}

func hasImageSearchError(parts []string) bool {
	for _, part := range parts {
		if strings.HasPrefix(strings.TrimSpace(part), "image_search_error:") {
			return true
		}
	}
	return false
}
