package web

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
)

type fakeCreateRouteModel struct {
	response string
	called   *bool
}

func TestNormalizeMessageRouteUsesConfigurableConfidenceThreshold(t *testing.T) {
	t.Setenv("PPT_INTENT_LOW_CONFIDENCE_THRESHOLD", "0.7")
	got := normalizeMessageRoute(MessageRouteResult{
		Intent:     messageIntentChat,
		Mode:       messageModeChat,
		Action:     messageActionReply,
		Confidence: 0.6,
	}, "你好", "")
	if got.Action != messageActionAskClarification || !got.NeedsConfirmation {
		t.Fatalf("low confidence should ask clarification: %#v", got)
	}
}

func TestNormalizeMessageRouteKeepsExplicitCreateDespiteOptionalMissingFields(t *testing.T) {
	got := normalizeMessageRoute(MessageRouteResult{
		Intent:            messageIntentCreate,
		Mode:              messageModePPTAgent,
		Action:            messageActionPrepareCreate,
		Confidence:        0.5,
		NormalizedRequest: "帮我做个 PPT",
		MissingFields:     []string{"page_count", "audience", "style"},
	}, "帮我做一份产品复盘 PPT", "")
	if got.Action != messageActionPrepareCreate || got.NeedsConfirmation {
		t.Fatalf("explicit create must remain directly startable: %#v", got)
	}
}

func (f fakeCreateRouteModel) Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error) {
	if f.called != nil {
		*f.called = true
	}
	return schema.AssistantMessage(f.response, nil), nil
}

func TestRouteCreateRequestUsesLLMClassification(t *testing.T) {
	called := false
	server := &Server{
		textModelFactory: func(context.Context) (interface {
			Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
		}, error) {
			return fakeCreateRouteModel{
				called:   &called,
				response: `{"intent":"chat","mode":"chat","reason":"模型判断为能力咨询","action":"reply","reply":"请说明要制作的 PPT 主题。","confidence":0.91}`,
			}, nil
		},
	}

	got := server.routeCreateRequest(context.Background(), "为产品委员会做一份 8 页季度复盘 PPT", false)
	if !called {
		t.Fatal("routeCreateRequest did not call the LLM classifier")
	}
	if got.Intent != createIntentChat || got.Reason != "模型判断为能力咨询" {
		t.Fatalf("route = %#v, want LLM classification result", got)
	}
}

func TestRouteCreateRequestDoesNotClarifyOptionalCreateFields(t *testing.T) {
	server := &Server{
		textModelFactory: func(context.Context) (interface {
			Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
		}, error) {
			return fakeCreateRouteModel{response: `{"intent":"create","mode":"pptagent","reason":"主题已明确","action":"prepare_create","confidence":0.95,"missing_fields":["page_count","audience","style"]}`}, nil
		},
	}

	got := server.routeCreateRequest(context.Background(), "帮我做一份新能源汽车出海趋势分析报告", false)
	if got.Intent != createIntentDeck {
		t.Fatalf("intent = %q, want %q; route=%#v", got.Intent, createIntentDeck, got)
	}
	if got.ClarificationQuestion != "" {
		t.Fatalf("successful create must not carry a clarification question: %#v", got)
	}
}

func TestRouteMessageRequestRuleFallback(t *testing.T) {
	server := &Server{}
	cases := []struct {
		name       string
		query      string
		selectedID string
		wantIntent string
		wantMode   string
		wantAction string
		wantNeeds  bool
	}{
		{name: "small talk", query: "你好", wantIntent: messageIntentChat, wantMode: messageModeChat, wantAction: messageActionReply},
		{name: "clear create", query: "为研发负责人做一份 10 页产品评审 PPT，风格务实", wantIntent: messageIntentCreate, wantMode: messageModePPTAgent, wantAction: messageActionPrepareCreate},
		{name: "plan only", query: "先规划一下 AI 技术分享的结构，不要生成", wantIntent: messageIntentPlan, wantMode: messageModePPTAgent, wantAction: messageActionSavePlan},
		{name: "fix without task", query: "修复上一版 PPT 的第三页标题溢出", wantIntent: messageIntentFix, wantMode: messageModePPTAgent, wantAction: messageActionAskClarification, wantNeeds: true},
		{name: "fix with task", query: "把第2页标题字体调大一点", selectedID: "task-1", wantIntent: messageIntentFix, wantMode: messageModePPTAgent, wantAction: messageActionUpdateTask},
		{name: "vague create", query: "帮我做个PPT", wantIntent: messageIntentCreate, wantMode: messageModePPTAgent, wantAction: messageActionAskClarification, wantNeeds: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := server.routeMessageRequest(context.Background(), tc.query, tc.selectedID)
			if got.Intent != tc.wantIntent || got.Mode != tc.wantMode || got.Action != tc.wantAction || got.NeedsConfirmation != tc.wantNeeds {
				t.Fatalf("route = %#v", got)
			}
		})
	}
}

func TestRouteTaskMessageRequestUsesConversationForDelegatedCreation(t *testing.T) {
	server := &Server{}
	got := server.routeTaskMessageRequest(
		context.Background(),
		"你决定主题和风格吧",
		"task-qinggan",
		"用户：我想做个青甘大环线的旅游项目介绍\n助手：我可以先帮你收敛主题、受众和呈现风格。",
	)
	if got.Intent != messageIntentCreate || got.Action != messageActionPrepareCreate || got.TaskID != "task-qinggan" {
		t.Fatalf("route = %#v, want contextual create on the current task", got)
	}
}

func TestFinalizeContextualDelegatedCreateRemovesClarificationGate(t *testing.T) {
	got := finalizeContextualDelegatedCreate(MessageRouteResult{
		Intent:            messageIntentCreate,
		Action:            messageActionAskClarification,
		NeedsConfirmation: true,
		MissingFields:     []string{"topic", "style"},
	}, "你决定主题和风格吧", "用户：我想做个青甘大环线的旅游项目介绍")
	if got.Action != messageActionPrepareCreate || got.NeedsConfirmation || len(got.MissingFields) != 0 {
		t.Fatalf("delegated create should start without clarification: %#v", got)
	}
}

func TestNormalizeMessageRouteRequiresTaskForFix(t *testing.T) {
	got := normalizeMessageRoute(MessageRouteResult{
		Intent:     messageIntentFix,
		Mode:       messageModePPTAgent,
		Action:     messageActionUpdateTask,
		Confidence: 0.92,
	}, "修复上一版第三页", "")
	if got.Action != messageActionAskClarification || !got.NeedsConfirmation {
		t.Fatalf("fix without task should ask clarification: %#v", got)
	}
	if len(got.MissingFields) != 1 || got.MissingFields[0] != "task_id" {
		t.Fatalf("missing fields = %#v, want task_id", got.MissingFields)
	}
}

func TestRouteCreateRequestRuleFallback(t *testing.T) {
	server := &Server{}
	cases := []struct {
		name    string
		query   string
		outline bool
		want    string
	}{
		{name: "outline always creates", query: "随便", outline: true, want: createIntentDeck},
		{name: "clear deck request", query: "为产品委员会做一份 8 页季度复盘 PPT", want: createIntentDeck},
		{name: "existing page edit", query: "把第2页标题字体调大一点", want: createIntentFixExisting},
		{name: "vague deck request", query: "帮我做个PPT", want: createIntentClarifyTopic},
		{name: "topic setup", query: "我还没想好主题，帮我确定主题", want: createIntentClarifyTopic},
		{name: "small talk", query: "你好", want: createIntentChat},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := server.routeCreateRequest(context.Background(), tc.query, tc.outline)
			if got.Intent != tc.want {
				t.Fatalf("intent = %q, want %q; route=%#v", got.Intent, tc.want, got)
			}
		})
	}
}
