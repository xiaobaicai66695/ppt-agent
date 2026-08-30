package web

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/assets/unsplash"
	"github.com/cloudwego/ppt-agent/pkg/tools"
	"github.com/cloudwego/ppt-agent/pkg/tools/search"
)

type chatImageResult struct {
	PreviewURL   string `json:"preview_url"`
	ImageURL     string `json:"image_url"`
	SourceURL    string `json:"source_url"`
	Photographer string `json:"photographer"`
	Attribution  string `json:"attribution"`
}

type chatImageSearchResponse struct {
	Photos []chatImageResult `json:"photos"`
}

type chatAugmentations struct {
	promptParts []string
	webResults  []search.SearchResult
	images      []chatImageResult
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
		if fallback != "" {
			return appendChatSupplement(fallback+augmentationNote(augmentations.promptParts), augmentations)
		}
		return appendChatSupplement("我可以先按普通对话回答；如果你要做 PPT，请说明主题、受众、页数和使用场景。"+augmentationNote(augmentations.promptParts), augmentations)
	}
	model, err := modelFactory(ctx)
	if err != nil || model == nil {
		if fallback != "" {
			return appendChatSupplement(fallback+augmentationNote(augmentations.promptParts), augmentations)
		}
		return appendChatSupplement("普通对话模型暂时不可用。你仍可以明确输入 PPT 创建、规划或修复需求。"+augmentationNote(augmentations.promptParts), augmentations)
	}

	chatCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	prompt := chatReplyPrompt(message, augmentations.promptParts)

	resp, err := model.Generate(chatCtx, []*schema.Message{schema.UserMessage(prompt)})
	if err != nil || resp == nil || strings.TrimSpace(resp.Content) == "" {
		if fallback != "" {
			return appendChatSupplement(fallback+augmentationNote(augmentations.promptParts), augmentations)
		}
		return appendChatSupplement("我可以先按普通对话回答；如果你要做 PPT，请说明主题、受众、页数和使用场景。"+augmentationNote(augmentations.promptParts), augmentations)
	}
	return appendChatSupplement(strings.TrimSpace(resp.Content), augmentations)
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
		reply = strings.TrimSpace(fallback)
		if reply == "" {
			reply = "普通对话暂时响应较慢，请稍后重试；你也可以继续描述 PPT 主题、受众和页数。"
		}
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
	query := chatSearchQuery(message, conversationContext)
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
			out, runErr := tools.NewImageSearchTool(client, "").InvokableRun(imageCtx, mustJSON(map[string]any{
				"query":         query,
				"asset_purpose": "scene",
				"per_page":      2,
				"download":      false,
				"reason":        "闲聊回答需要图片候选作为参考",
			}))
			if runErr != nil {
				augmentations.promptParts = append(augmentations.promptParts, "image_search_error: "+runErr.Error())
			} else if strings.TrimSpace(out) != "" {
				augmentations.promptParts = append(augmentations.promptParts, "image_search_result:\n"+out)
				var response chatImageSearchResponse
				if json.Unmarshal([]byte(out), &response) == nil {
					augmentations.images = append(augmentations.images, response.Photos...)
				}
			}
		} else {
			augmentations.promptParts = append(augmentations.promptParts, "image_search_error: "+err.Error())
		}
	}
	return augmentations
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

func augmentationNote(augmentations []string) string {
	if len(augmentations) == 0 {
		return ""
	}
	return "\n\n补充：已尝试使用可用的搜索/图片能力组织答案；如果能力未配置，结果会在服务端降级。"
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
		for _, result := range augmentations.webResults[:min(len(augmentations.webResults), 4)] {
			if strings.TrimSpace(result.URL) == "" || strings.TrimSpace(result.Title) == "" {
				continue
			}
			items = append(items, fmt.Sprintf("- [%s](%s)", result.Title, result.URL))
		}
		if len(items) > 0 {
			sections = append(sections, "### 补充资料来源\n"+strings.Join(items, "\n"))
		}
	}
	if len(augmentations.images) > 0 {
		items := make([]string, 0, min(len(augmentations.images), 2))
		for _, image := range augmentations.images[:min(len(augmentations.images), 2)] {
			preview := strings.TrimSpace(image.PreviewURL)
			if preview == "" {
				preview = strings.TrimSpace(image.ImageURL)
			}
			if preview == "" {
				continue
			}
			label := strings.TrimSpace(image.Attribution)
			if label == "" {
				label = strings.TrimSpace(image.Photographer)
			}
			if label == "" {
				label = "图片参考"
			}
			items = append(items, fmt.Sprintf("![%s](%s)", label, preview))
		}
		if len(items) > 0 {
			sections = append(sections, "### 图片参考\n"+strings.Join(items, "\n"))
		}
	}
	return strings.Join(sections, "\n\n")
}
