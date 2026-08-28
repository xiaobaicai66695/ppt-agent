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
				response: `{"intent":"chat","reason":"模型判断为能力咨询","clarification_question":"请说明要制作的 PPT 主题。","confidence":0.91}`,
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
