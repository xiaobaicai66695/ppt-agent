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
)

func (s *Server) buildChatReply(ctx context.Context, message, fallback string) string {
	fallback = strings.TrimSpace(fallback)
	augmentations := s.collectChatAugmentations(ctx, message)

	modelFactory := s.textModelFactory
	if modelFactory == nil {
		modelFactory = s.aiModelFactory
	}
	if modelFactory == nil {
		if fallback != "" {
			return fallback + augmentationNote(augmentations)
		}
		return "我可以先按普通对话回答；如果你要做 PPT，请说明主题、受众、页数和使用场景。" + augmentationNote(augmentations)
	}
	model, err := modelFactory(ctx)
	if err != nil || model == nil {
		if fallback != "" {
			return fallback + augmentationNote(augmentations)
		}
		return "普通对话模型暂时不可用。你仍可以明确输入 PPT 创建、规划或修复需求。" + augmentationNote(augmentations)
	}

	chatCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	prompt := fmt.Sprintf(`你是 PPT Agent 工作台里的闲聊助手。当前消息已经被后端判定为 chat，不得创建 PPT 任务，不得承诺已经开始生成文件。

回答要求：
- 直接回答用户问题。
- 如果补充材料里有 web search 或 image search 结果，整合它们组织答案，并保留简短来源说明。
- 如果补充材料显示某能力未配置，只说明当前无法使用该能力，不要编造搜索结果。
- 如果用户实际想做 PPT，引导其补充主题、受众、页数、风格或选择已有任务修复。

用户消息：
%s

补充材料：
%s`, message, strings.Join(augmentations, "\n\n"))

	resp, err := model.Generate(chatCtx, []*schema.Message{schema.UserMessage(prompt)})
	if err != nil || resp == nil || strings.TrimSpace(resp.Content) == "" {
		if fallback != "" {
			return fallback + augmentationNote(augmentations)
		}
		return "我可以先按普通对话回答；如果你要做 PPT，请说明主题、受众、页数和使用场景。" + augmentationNote(augmentations)
	}
	return strings.TrimSpace(resp.Content)
}

func (s *Server) collectChatAugmentations(ctx context.Context, message string) []string {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil
	}
	var augmentations []string
	if chatNeedsWebSearch(message) {
		searchCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		out, err := tools.NewSearchTool().InvokableRun(searchCtx, mustJSON(map[string]string{
			"query":  compactSearchQuery(message),
			"reason": "用户在闲聊中请求最新或需要核实的信息",
		}))
		if err != nil {
			augmentations = append(augmentations, "web_search_error: "+err.Error())
		} else if strings.TrimSpace(out) != "" {
			augmentations = append(augmentations, "web_search_result:\n"+out)
		}
	}
	if chatNeedsImageSearch(message) {
		if !unsplash.IsConfigured() {
			augmentations = append(augmentations, "image_search_error: 未配置 UNSPLASH_ACCESS_KEY，当前无法检索图片候选。")
		} else if client, err := unsplash.NewClientFromEnv(); err == nil {
			imageCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
			defer cancel()
			out, runErr := tools.NewImageSearchTool(client, "").InvokableRun(imageCtx, mustJSON(map[string]any{
				"query":         compactSearchQuery(message),
				"asset_purpose": "scene",
				"per_page":      2,
				"download":      false,
				"reason":        "闲聊回答需要图片候选作为参考",
			}))
			if runErr != nil {
				augmentations = append(augmentations, "image_search_error: "+runErr.Error())
			} else if strings.TrimSpace(out) != "" {
				augmentations = append(augmentations, "image_search_result:\n"+out)
			}
		} else {
			augmentations = append(augmentations, "image_search_error: "+err.Error())
		}
	}
	return augmentations
}

func chatNeedsWebSearch(message string) bool {
	normalized := strings.ToLower(message)
	for _, keyword := range []string{"最新", "今天", "现在", "近期", "新闻", "搜索", "查一下", "websearch", "web search", "联网", "资料来源", "source"} {
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
