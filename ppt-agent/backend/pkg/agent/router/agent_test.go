package router

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestAgentRoutesExplicitCreateWithoutOptionalFields(t *testing.T) {
	agent := NewAgent(func(context.Context, []*schema.Message) (*schema.Message, error) {
		return schema.AssistantMessage(`{"intent":"create","mode":"pptagent","confidence":0.92,"normalized_request":"做一份产品复盘 PPT","action":"prepare_create","reason":"明确新建演示"}`, nil), nil
	})
	got, err := agent.Route(context.Background(), Input{Query: "做一份产品复盘 PPT"})
	if err != nil || got.Intent != IntentCreate || got.Action != "prepare_create" {
		t.Fatalf("route = %#v, err = %v", got, err)
	}
}

func TestAgentRejectsUnknownIntent(t *testing.T) {
	agent := NewAgent(func(context.Context, []*schema.Message) (*schema.Message, error) {
		return schema.AssistantMessage(`{"intent":"delete_everything"}`, nil), nil
	})
	if _, err := agent.Route(context.Background(), Input{Query: "测试"}); err == nil {
		t.Fatal("unknown intent must fail so web fallback can take over")
	}
}

func TestAgentRoutesBoundPageFix(t *testing.T) {
	agent := NewAgent(func(context.Context, []*schema.Message) (*schema.Message, error) {
		return schema.AssistantMessage(`{"intent":"fix","reason":"用户明确要求修改第 2 页标题","target_pages":[2],"fix_details":{"aspect":"text_content","detail":"缩短标题","target_elements":"标题"}}`, nil), nil
	})
	got, err := agent.RouteContinuation(context.Background(), ContinuationInput{Message: "把第 2 页标题缩短", TasksSummary: "当前 PPT 共 4 页"})
	if err != nil || got.Intent != "fix" || len(got.TargetPages) != 1 || got.TargetPages[0] != 2 || got.FixDetails == nil || got.FixDetails.TargetElements != "标题" {
		t.Fatalf("route = %#v, err = %v", got, err)
	}
}

func TestAgentDoesNotUpgradeExplicitChartReplacementToRegeneration(t *testing.T) {
	agent := NewAgent(func(context.Context, []*schema.Message) (*schema.Message, error) {
		return schema.AssistantMessage(`{"intent":"regenerate","reason":"页面结构变化","target_pages":[5],"fix_details":{"aspect":"layout","detail":"把折线图换成按城市对比的柱状图","target_elements":"图表"},"regenerate_scope":[5]}`, nil), nil
	})
	got, err := agent.RouteContinuation(context.Background(), ContinuationInput{
		Message:      "第 5 页不要折线图，换成按城市对比的柱状图；其它页和第 5 页数据都保持不变。",
		TasksSummary: "当前 PPT 共 6 页",
	})
	if err != nil || got.Intent != "fix" || len(got.TargetPages) != 1 || got.TargetPages[0] != 5 || len(got.RegenerateScope) != 0 {
		t.Fatalf("route = %#v, err = %v", got, err)
	}
}

func TestAgentKeepsExplicitPageRegeneration(t *testing.T) {
	agent := NewAgent(func(context.Context, []*schema.Message) (*schema.Message, error) {
		return schema.AssistantMessage(`{"intent":"regenerate","reason":"用户要求重做第 3 页","target_pages":[3],"regenerate_scope":[3]}`, nil), nil
	})
	got, err := agent.RouteContinuation(context.Background(), ContinuationInput{Message: "第 3 页重新生成一版", TasksSummary: "当前 PPT 共 6 页"})
	if err != nil || got.Intent != "regenerate" {
		t.Fatalf("route = %#v, err = %v", got, err)
	}
}
