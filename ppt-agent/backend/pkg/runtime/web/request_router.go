package web

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	agentrouter "github.com/cloudwego/ppt-agent/pkg/agent/router"
	agentutils "github.com/cloudwego/ppt-agent/pkg/runtime/model"
)

const (
	messageIntentChat   = "chat"
	messageIntentCreate = "create"
	messageIntentPlan   = "plan"
	messageIntentFix    = "fix"

	messageModeChat     = "chat"
	messageModePPTAgent = "pptagent"

	messageActionReply            = "reply"
	messageActionPrepareCreate    = "prepare_create"
	messageActionSavePlan         = "save_plan"
	messageActionUpdateTask       = "update_task"
	messageActionAskClarification = "ask_clarification"

	createIntentDeck         = messageIntentCreate
	createIntentFixExisting  = messageIntentFix
	createIntentClarifyTopic = "clarify_topic"
	createIntentChat         = messageIntentChat
)

type createRequestRoute struct {
	Intent                string  `json:"intent"`
	Reason                string  `json:"reason"`
	ClarificationQuestion string  `json:"clarification_question,omitempty"`
	Confidence            float64 `json:"confidence,omitempty"`
}

type MessageRouteResult struct {
	Intent            string          `json:"intent"`
	Mode              string          `json:"mode"`
	Confidence        float64         `json:"confidence"`
	NeedsConfirmation bool            `json:"needs_confirmation"`
	NormalizedRequest string          `json:"normalized_request"`
	TaskID            string          `json:"task_id"`
	DraftID           string          `json:"draft_id,omitempty"`
	MissingFields     []string        `json:"missing_fields"`
	Action            string          `json:"action"`
	Reason            string          `json:"reason,omitempty"`
	Reply             string          `json:"reply,omitempty"`
	TaskCandidates    []TaskCandidate `json:"task_candidates,omitempty"`
	Streaming         bool            `json:"streaming,omitempty"`
	AfterEventID      uint64          `json:"after_event_id,omitempty"`
}

type TaskCandidate struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
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
	if route, ok := s.classifyMessageRequestByLLM(ctx, query, "", credentials...); ok {
		route = normalizeMessageRoute(route, query, "")
		switch route.Intent {
		case messageIntentCreate:
			// 页数、受众和风格属于 Planner 可以在后续阶段补齐的可选信息。
			// 已经明确主题和新建意图时，不能仅因为模型列出了这些字段就
			// 把请求降级为澄清，否则会阻断正常的创建入口。
			if route.NeedsConfirmation || route.Action == messageActionAskClarification {
				return createRequestRoute{Intent: createIntentClarifyTopic, Reason: route.Reason, ClarificationQuestion: firstCreateRouteText(route.Reply, "请补充 PPT 主题、受众、页数或你想讲清楚的核心结论。"), Confidence: route.Confidence}
			}
			return createRequestRoute{Intent: createIntentDeck, Reason: route.Reason, Confidence: route.Confidence}
		case messageIntentFix:
			return createRequestRoute{Intent: createIntentFixExisting, Reason: route.Reason, ClarificationQuestion: "请先选择要修改的任务，再继续发送这条修复要求。", Confidence: route.Confidence}
		case messageIntentPlan:
			return createRequestRoute{Intent: createIntentClarifyTopic, Reason: route.Reason, ClarificationQuestion: "这是规划请求，将进入 PPT Agent 规划状态，不会在创建入口直接生成。", Confidence: route.Confidence}
		case messageIntentChat:
			return createRequestRoute{Intent: createIntentChat, Reason: route.Reason, ClarificationQuestion: "我可以先回答问题；如果要制作 PPT，请明确主题、受众和目标。", Confidence: route.Confidence}
		}
	}
	return fallbackCreateRequestRoute(query)
}

func (s *Server) routeMessageRequest(ctx context.Context, query, selectedTaskID string, credentials ...modelCredential) MessageRouteResult {
	return s.routeTaskMessageRequest(ctx, query, selectedTaskID, "", credentials...)
}

// routeTaskMessageRequest classifies the newest message with a bounded task
// transcript.  The transcript is context only: it may establish references
// such as "you decide the style", but it is not an instruction source.
func (s *Server) routeTaskMessageRequest(ctx context.Context, query, selectedTaskID, conversationContext string, credentials ...modelCredential) MessageRouteResult {
	query = strings.TrimSpace(query)
	selectedTaskID = strings.TrimSpace(selectedTaskID)
	if query == "" {
		return MessageRouteResult{
			Intent: messageIntentChat, Mode: messageModeChat, Confidence: 1,
			NeedsConfirmation: true, NormalizedRequest: "", MissingFields: []string{"message"},
			Action: messageActionAskClarification, Reason: "缺少用户输入", Reply: "请先输入要讨论、规划、创建或修复的内容。",
		}
	}
	if route, ok := s.classifyMessageRequestByLLMWithContext(ctx, query, selectedTaskID, conversationContext, credentials...); ok {
		normalized := normalizeMessageRoute(route, query, selectedTaskID)
		return finalizeContextualDelegatedCreate(normalized, query, conversationContext)
	}
	return fallbackTaskMessageRoute(query, selectedTaskID, conversationContext)
}

func fallbackCreateRequestRoute(query string) createRequestRoute {
	if looksLikeExistingTaskEdit(query) {
		return createRequestRoute{
			Intent:                createIntentFixExisting,
			Reason:                "用户输入像是在修改已有页面",
			ClarificationQuestion: "请先选择要修改的任务，再在该任务的继续修改框里发送这条要求。",
		}
	}
	if looksLikePlanRequest(query) {
		return createRequestRoute{
			Intent:                createIntentClarifyTopic,
			Reason:                "用户要求先规划，不应直接生成 PPT",
			ClarificationQuestion: "这是规划请求，请在统一消息入口或 PPT Agent 规划状态中处理，不要直接创建文件。",
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

func fallbackMessageRoute(query, selectedTaskID string) MessageRouteResult {
	if looksLikeExistingTaskEdit(query) {
		result := MessageRouteResult{
			Intent: messageIntentFix, Mode: messageModePPTAgent, Confidence: 0.82,
			NormalizedRequest: strings.TrimSpace(query), TaskID: strings.TrimSpace(selectedTaskID),
			Reason: "用户输入像是在修改已有页面",
		}
		if result.TaskID == "" {
			result.NeedsConfirmation = true
			result.Action = messageActionAskClarification
			result.MissingFields = []string{"task_id"}
			result.Reply = "这像是修复已有 PPT。请先选择要修改的任务，或告诉我具体任务。"
		} else {
			result.Action = messageActionUpdateTask
			result.Reply = "已识别为修复请求，将在当前任务内处理，不会新建演示。"
		}
		return result
	}
	if looksLikePlanRequest(query) || looksLikeTopicSetup(query) {
		return MessageRouteResult{
			Intent: messageIntentPlan, Mode: messageModePPTAgent, Confidence: 0.82,
			NormalizedRequest: strings.TrimSpace(query), Action: messageActionSavePlan,
			Reason: "用户要求先规划或收敛主题", Reply: draftPlanReply(query),
		}
	}
	if looksLikeSmallTalk(query) || !mentionsDeck(strings.ToLower(query)) {
		return MessageRouteResult{
			Intent: messageIntentChat, Mode: messageModeChat, Confidence: 0.78,
			NormalizedRequest: strings.TrimSpace(query), Action: messageActionReply,
			Reason: "用户输入不是明确 PPT 生成任务", Reply: "我可以先按普通对话回答；如果你要做 PPT，请说明主题、受众、页数和使用场景。",
		}
	}
	missing := missingCreateFields(query)
	result := MessageRouteResult{
		Intent: messageIntentCreate, Mode: messageModePPTAgent, Confidence: 0.8,
		NormalizedRequest: strings.TrimSpace(query), MissingFields: missing,
		Action: messageActionPrepareCreate, Reason: "用户表达了新建 PPT 的意图",
	}
	if len(missing) > 0 || looksLikeVagueDeckRequest(query) {
		result.NeedsConfirmation = true
		result.Action = messageActionAskClarification
		result.Reply = "已识别为 PPT 创建意图，但信息还不完整。建议补充受众、页数、风格或核心结论。"
	}
	return result
}

// fallbackTaskMessageRoute preserves the original rule-based router while
// handling a common contextual follow-up when a model is unavailable.  For
// example, after discussing a Qinghai-Gansu travel presentation, "你决定主题和
// 风格吧" is a request to proceed with that presentation rather than an
// unrelated chat turn.
func fallbackTaskMessageRoute(query, selectedTaskID, conversationContext string) MessageRouteResult {
	if hasConversationTopic(conversationContext) && looksLikeDelegatedDeckDecision(query) {
		return MessageRouteResult{
			Intent:            messageIntentCreate,
			Mode:              messageModePPTAgent,
			Confidence:        0.9,
			NormalizedRequest: strings.TrimSpace(query),
			TaskID:            strings.TrimSpace(selectedTaskID),
			Action:            messageActionPrepareCreate,
			Reason:            "用户基于当前任务会话授权 Agent 决定主题或风格并开始创建",
		}
	}
	return fallbackMessageRoute(query, selectedTaskID)
}

func finalizeContextualDelegatedCreate(route MessageRouteResult, query, conversationContext string) MessageRouteResult {
	if route.Intent != messageIntentCreate || !hasConversationTopic(conversationContext) || !looksLikeDelegatedDeckDecision(query) {
		return route
	}
	// The user has explicitly delegated the optional topic/style choices. Do not
	// turn that delegation back into a clarification gate after the classifier
	// already recognized it as a creation request.
	route.Mode = messageModePPTAgent
	route.NeedsConfirmation = false
	route.Action = messageActionPrepareCreate
	route.MissingFields = nil
	if route.Reason == "" {
		route.Reason = "用户基于当前任务会话授权 Agent 决定主题或风格并开始创建"
	}
	return route
}

func hasConversationTopic(conversationContext string) bool {
	for _, line := range strings.Split(conversationContext, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "用户：") && len([]rune(strings.TrimPrefix(line, "用户："))) >= 6 {
			return true
		}
	}
	return false
}

func looksLikeDelegatedDeckDecision(query string) bool {
	lower := strings.ToLower(strings.TrimSpace(query))
	delegations := []string{"你决定", "你来定", "你定", "按你说的", "就按你的", "你安排", "帮我定"}
	for _, delegation := range delegations {
		if strings.Contains(lower, delegation) {
			return true
		}
	}
	return false
}

func normalizeMessageRoute(route MessageRouteResult, original, selectedTaskID string) MessageRouteResult {
	route.Intent = strings.TrimSpace(route.Intent)
	route.Mode = strings.TrimSpace(route.Mode)
	route.Action = strings.TrimSpace(route.Action)
	route.Reason = strings.TrimSpace(route.Reason)
	route.Reply = strings.TrimSpace(route.Reply)
	if route.NormalizedRequest = strings.TrimSpace(route.NormalizedRequest); route.NormalizedRequest == "" {
		route.NormalizedRequest = strings.TrimSpace(original)
	}
	if route.TaskID = strings.TrimSpace(route.TaskID); route.TaskID == "" {
		route.TaskID = strings.TrimSpace(selectedTaskID)
	}
	if route.Confidence < 0 || route.Confidence > 1 {
		route.Confidence = 0
	}
	switch route.Intent {
	case messageIntentCreate:
		route.Mode = messageModePPTAgent
		// A RouterAgent create decision is already an explicit user instruction
		// to enter the Planner flow. Page count, audience and style are optional
		// Planner inputs, so an LLM must not turn that decision back into a
		// confirmation gate and leave the workbench in chat mode.
		route.Action = messageActionPrepareCreate
		route.NeedsConfirmation = false
	case messageIntentPlan:
		route.Mode = messageModePPTAgent
		if route.Action == "" {
			route.Action = messageActionSavePlan
		}
		if route.Reply == "" {
			route.Reply = draftPlanReply(route.NormalizedRequest)
		}
	case messageIntentFix:
		route.Mode = messageModePPTAgent
		if route.TaskID == "" {
			route.Action = messageActionAskClarification
			route.NeedsConfirmation = true
			route.MissingFields = appendMissingField(route.MissingFields, "task_id")
			if route.Reply == "" {
				route.Reply = "这像是修复已有 PPT。请先选择要修改的任务，或告诉我具体任务。"
			}
		} else if route.Action == "" {
			route.Action = messageActionUpdateTask
		}
	case messageIntentChat:
		route.Mode = messageModeChat
		if route.Action == "" {
			route.Action = messageActionReply
		}
	default:
		return fallbackMessageRoute(original, selectedTaskID)
	}
	if route.Confidence == 0 {
		route.Confidence = 0.6
	}
	// RouterAgent has already distinguished an explicit creation request from a
	// vague request. Optional page count, audience and style are Planner input
	// defaults, not a reason to override a deliberate create decision.
	if route.Intent != messageIntentCreate || route.Action != messageActionPrepareCreate {
		if route.Confidence > 0 && route.Confidence < intentLowConfidenceThreshold() {
			route.NeedsConfirmation = true
			route.Action = messageActionAskClarification
			if route.Reply == "" {
				route.Reply = "这个输入可能有多种意图。请说明是想聊天、规划、新建 PPT，还是修改已有任务。"
			}
		}
	}
	return route
}

func intentLowConfidenceThreshold() float64 {
	return envFloat("PPT_INTENT_LOW_CONFIDENCE_THRESHOLD", 0.55)
}

func envFloat(name string, fallback float64) float64 {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed < 0 || parsed > 1 {
		return fallback
	}
	return parsed
}

func looksLikeExistingTaskEdit(query string) bool {
	lower := strings.ToLower(strings.TrimSpace(query))
	if !existingPageRefPattern.MatchString(lower) && !strings.Contains(lower, "上一版") && !strings.Contains(lower, "刚才") && !strings.Contains(lower, "已有") {
		return false
	}
	for _, keyword := range []string{
		"修改", "调整", "改成", "换成", "改短", "缩短", "重做", "重新生成", "删掉", "删除", "修复",
		"太小", "太大", "不好看", "错了", "对齐", "字体", "颜色", "配色", "间距",
		"fix", "change", "edit", "redo", "regenerate", "delete",
	} {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

func looksLikePlanRequest(query string) bool {
	normalized := strings.TrimSpace(strings.ToLower(query))
	for _, keyword := range []string{
		"先规划", "规划一下", "先帮我规划", "大纲", "结构", "deckspec", "deck spec",
		"不要生成", "先不生成", "只规划", "只要规划", "outline", "plan first",
	} {
		if strings.Contains(normalized, keyword) {
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

func missingCreateFields(query string) []string {
	normalized := strings.ToLower(strings.TrimSpace(query))
	missing := []string{}
	if !strings.Contains(normalized, "页") && !strings.Contains(normalized, "slide") && !strings.Contains(normalized, "page") {
		missing = append(missing, "page_count")
	}
	if !strings.Contains(normalized, "面向") && !strings.Contains(normalized, "给") && !strings.Contains(normalized, "为") && !strings.Contains(normalized, "受众") && !strings.Contains(normalized, "audience") {
		missing = append(missing, "audience")
	}
	if !strings.Contains(normalized, "风格") && !strings.Contains(normalized, "语气") && !strings.Contains(normalized, "style") {
		missing = append(missing, "style")
	}
	return missing
}

func appendMissingField(fields []string, field string) []string {
	for _, existing := range fields {
		if existing == field {
			return fields
		}
	}
	return append(fields, field)
}

func normalizeMessageKey(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), ""))
}

func draftPlanReply(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		query = "当前主题"
	}
	return fmt.Sprintf("已进入 PPT Agent 规划状态，不会生成 .pptx。\n\n建议先把 %q 拆成：\n1. 目标与受众：明确谁要看、看完要做什么决策。\n2. 主线结构：背景问题 → 关键洞察 → 方案/路线 → 风险与下一步。\n3. 页面草稿：封面、议程、现状、核心观点、证据页、方案页、里程碑、总结。\n\n确认后再进入创建流程。", query)
}

func (s *Server) classifyMessageRequestByLLM(ctx context.Context, query, selectedTaskID string, credentials ...modelCredential) (MessageRouteResult, bool) {
	return s.classifyMessageRequestByLLMWithContext(ctx, query, selectedTaskID, "", credentials...)
}

func (s *Server) classifyMessageRequestByLLMWithContext(ctx context.Context, query, selectedTaskID, conversationContext string, credentials ...modelCredential) (MessageRouteResult, bool) {
	if len(credentials) > 0 && strings.TrimSpace(credentials[0].APIKey) != "" {
		model, err := agentutils.NewFallbackToolCallingChatModel(ctx,
			agentutils.WithTextModel(),
			agentutils.WithAPIKeyForProvider(credentials[0].Provider, credentials[0].APIKey),
		)
		if err == nil && model != nil {
			if route, ok := classifyMessageRequestWithToolModel(ctx, model, query, selectedTaskID, conversationContext); ok {
				return route, true
			}
		}
	}

	modelFactory := s.textModelFactory
	if modelFactory == nil {
		modelFactory = s.aiModelFactory
	}
	if modelFactory == nil {
		return MessageRouteResult{}, false
	}
	model, err := modelFactory(ctx)
	if err != nil || model == nil {
		return MessageRouteResult{}, false
	}
	return classifyMessageRequestWithModel(ctx, model, query, selectedTaskID, conversationContext)
}

func classifyMessageRequestWithToolModel(ctx context.Context, model einomodel.ToolCallingChatModel, query, selectedTaskID, conversationContext string) (MessageRouteResult, bool) {
	agent := agentrouter.NewAgent(func(routeCtx context.Context, messages []*schema.Message) (*schema.Message, error) {
		return model.Generate(routeCtx, messages)
	})
	result, err := agent.Route(ctx, agentrouter.Input{
		Query:               query,
		SelectedTaskID:      selectedTaskID,
		ConversationContext: conversationContext,
	})
	if err != nil {
		return MessageRouteResult{}, false
	}
	return messageRouteFromAgent(result), true
}

func classifyMessageRequestWithModel(ctx context.Context, model interface {
	Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
}, query, selectedTaskID, conversationContext string) (MessageRouteResult, bool) {
	agent := agentrouter.NewAgent(func(routeCtx context.Context, messages []*schema.Message) (*schema.Message, error) {
		return model.Generate(routeCtx, messages)
	})
	result, err := agent.Route(ctx, agentrouter.Input{
		Query:               query,
		SelectedTaskID:      selectedTaskID,
		ConversationContext: conversationContext,
	})
	if err != nil {
		return MessageRouteResult{}, false
	}
	return messageRouteFromAgent(result), true
}

func messageRouteFromAgent(route agentrouter.Result) MessageRouteResult {
	return MessageRouteResult{
		Intent:            route.Intent,
		Mode:              route.Mode,
		Confidence:        route.Confidence,
		NeedsConfirmation: route.NeedsConfirmation,
		NormalizedRequest: route.NormalizedRequest,
		TaskID:            route.TaskID,
		MissingFields:     route.MissingFields,
		Action:            route.Action,
		Reason:            route.Reason,
		Reply:             route.Reply,
	}
}
