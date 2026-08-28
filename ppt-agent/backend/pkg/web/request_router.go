package web

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
)

const (
	createIntentDeck         = "create_deck"
	createIntentFixExisting  = "fix_existing"
	createIntentClarifyTopic = "clarify_topic"
	createIntentChat         = "chat"
)

type createRequestRoute struct {
	Intent                string  `json:"intent"`
	Reason                string  `json:"reason"`
	ClarificationQuestion string  `json:"clarification_question,omitempty"`
	Confidence            float64 `json:"confidence,omitempty"`
}

var existingPageRefPattern = regexp.MustCompile(`(?i)(第?\s*\d+\s*(页|张|slide)|封面|目录|标题页|最后一页)`)

func (s *Server) routeCreateRequest(ctx context.Context, query string, hasOutline bool, credentials ...modelCredential) createRequestRoute {
	query = strings.TrimSpace(query)
	if hasOutline {
		return createRequestRoute{Intent: createIntentDeck, Reason: "用户已提供结构化大纲"}
	}
	if query == "" {
		return createRequestRoute{Intent: createIntentClarifyTopic, Reason: "缺少用户输入", ClarificationQuestion: "请补充要制作的 PPT 主题、受众或使用场景。"}
	}
	if route, ok := s.classifyCreateRequestByLLM(ctx, query, credentials...); ok {
		return route
	}
	return fallbackCreateRequestRoute(query)
}

func fallbackCreateRequestRoute(query string) createRequestRoute {
	if looksLikeExistingTaskEdit(query) {
		return createRequestRoute{
			Intent:                createIntentFixExisting,
			Reason:                "用户输入像是在修改已有页面",
			ClarificationQuestion: "请先选择要修改的任务，再在该任务的继续修改框里发送这条要求。",
		}
	}
	if looksLikeTopicSetup(query) {
		return createRequestRoute{
			Intent:                createIntentClarifyTopic,
			Reason:                "用户仍在确定主题，暂不创建生成任务",
			ClarificationQuestion: "请先说明候选方向、受众或使用场景，我可以据此帮你收敛成可执行的 PPT 主题。",
		}
	}
	if looksLikeSmallTalk(query) {
		return createRequestRoute{Intent: createIntentChat, Reason: "用户输入不是 PPT 生成任务", ClarificationQuestion: "请描述要制作的 PPT 主题、受众、页数和交付场景。"}
	}
	if looksLikeVagueDeckRequest(query) {
		return createRequestRoute{
			Intent:                createIntentClarifyTopic,
			Reason:                "用户想创建 PPT，但主题信息不足",
			ClarificationQuestion: "请补充 PPT 主题、受众、页数或你想讲清楚的核心结论。",
		}
	}
	return createRequestRoute{Intent: createIntentDeck, Reason: "按明确生成请求创建 PPT"}
}

func looksLikeExistingTaskEdit(query string) bool {
	lower := strings.ToLower(strings.TrimSpace(query))
	if !existingPageRefPattern.MatchString(lower) {
		return false
	}
	for _, keyword := range []string{
		"修改", "调整", "改成", "换成", "重做", "重新生成", "删掉", "删除", "修复",
		"太小", "太大", "不好看", "错了", "对齐", "字体", "颜色", "配色", "间距",
		"fix", "change", "edit", "redo", "regenerate", "delete",
	} {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

func looksLikeTopicSetup(query string) bool {
	normalized := strings.TrimSpace(strings.ToLower(query))
	for _, keyword := range []string{
		"确定主题", "选主题", "想主题", "主题怎么定", "还没想好主题", "不知道做什么主题",
		"help me choose a topic", "choose a topic", "pick a topic",
	} {
		if strings.Contains(normalized, keyword) {
			return true
		}
	}
	return false
}

func looksLikeSmallTalk(query string) bool {
	normalized := strings.TrimSpace(strings.ToLower(query))
	normalized = strings.Trim(normalized, " ，。！？!?~")
	if normalized == "" {
		return false
	}
	for _, exact := range []string{"你好", "您好", "在吗", "谢谢", "thank you", "thanks", "hi", "hello"} {
		if normalized == exact {
			return true
		}
	}
	return strings.Contains(normalized, "你是谁") ||
		strings.Contains(normalized, "你能做什么") ||
		strings.Contains(normalized, "怎么使用")
}

func looksLikeVagueDeckRequest(query string) bool {
	normalized := strings.TrimSpace(strings.ToLower(query))
	if !mentionsDeck(normalized) {
		return false
	}
	runes := []rune(normalized)
	if len(runes) > 24 {
		return false
	}
	for _, vague := range []string{
		"做个ppt", "做一个ppt", "帮我做ppt", "帮我做个ppt", "生成ppt", "生成一个ppt",
		"做份演示", "做一个演示", "帮我做演示", "做幻灯片", "帮我做幻灯片",
	} {
		if strings.Contains(normalized, vague) {
			return true
		}
	}
	return false
}

func firstCreateRouteText(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func mentionsDeck(query string) bool {
	for _, keyword := range []string{"ppt", "演示", "幻灯片", "汇报", "路演", "presentation", "deck"} {
		if strings.Contains(query, keyword) {
			return true
		}
	}
	return false
}

func (s *Server) classifyCreateRequestByLLM(ctx context.Context, query string, credentials ...modelCredential) (createRequestRoute, bool) {
	if len(credentials) > 0 && strings.TrimSpace(credentials[0].APIKey) != "" {
		model, err := agentutils.NewFallbackToolCallingChatModel(ctx,
			agentutils.WithTextModel(),
			agentutils.WithAPIKeyForProvider(credentials[0].Provider, credentials[0].APIKey),
		)
		if err == nil && model != nil {
			if route, ok := classifyCreateRequestWithToolModel(ctx, model, query); ok {
				return route, true
			}
		}
	}

	modelFactory := s.textModelFactory
	if modelFactory == nil {
		modelFactory = s.aiModelFactory
	}
	if modelFactory == nil {
		return createRequestRoute{}, false
	}
	model, err := modelFactory(ctx)
	if err != nil || model == nil {
		return createRequestRoute{}, false
	}
	return classifyCreateRequestWithModel(ctx, model, query)
}

func classifyCreateRequestWithToolModel(ctx context.Context, model einomodel.ToolCallingChatModel, query string) (createRequestRoute, bool) {
	routeCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	resp, err := model.Generate(routeCtx, []*schema.Message{schema.UserMessage(createRequestRoutePrompt(query))})
	if err != nil || resp == nil {
		return createRequestRoute{}, false
	}
	return parseCreateRequestRoute(resp.Content)
}

func classifyCreateRequestWithModel(ctx context.Context, model interface {
	Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
}, query string) (createRequestRoute, bool) {
	routeCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	resp, err := model.Generate(routeCtx, []*schema.Message{schema.UserMessage(createRequestRoutePrompt(query))})
	if err != nil || resp == nil {
		return createRequestRoute{}, false
	}
	return parseCreateRequestRoute(resp.Content)
}

func createRequestRoutePrompt(query string) string {
	return fmt.Sprintf(`你是 PPT 创建入口的意图分类路由器。根据当前用户输入判断要进入哪个流程。

可选 intent：
- create_deck：用户要新建 PPT，且主题/对象/场景至少有一个明确内容。
- fix_existing：用户像是在修改已有 PPT 或已有页面，例如调整第几页、改颜色、重新生成某页。
- clarify_topic：用户想做 PPT，但主题、对象或场景不足，需要先追问。
- chat：闲聊、问候、咨询能力说明，当前不应启动 PPT 生成。

用户输入：
%q

严格输出 JSON：
{"intent":"create_deck|fix_existing|clarify_topic|chat","reason":"一句话理由","clarification_question":"需要追问时填写中文问题，否则空字符串","confidence":0.0}`, query)
}

func parseCreateRequestRoute(content string) (createRequestRoute, bool) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var route createRequestRoute
	if err := json.Unmarshal([]byte(content), &route); err != nil {
		return createRequestRoute{}, false
	}
	route.Intent = strings.TrimSpace(route.Intent)
	route.Reason = strings.TrimSpace(route.Reason)
	route.ClarificationQuestion = strings.TrimSpace(route.ClarificationQuestion)
	switch route.Intent {
	case createIntentDeck, createIntentFixExisting, createIntentClarifyTopic, createIntentChat:
		return route, true
	default:
		return createRequestRoute{}, false
	}
}
