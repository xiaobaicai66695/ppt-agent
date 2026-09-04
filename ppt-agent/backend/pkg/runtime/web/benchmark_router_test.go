package web

import (
	"context"
	"testing"
)

func TestBenchmarkCreateRouteUsesStableEvaluationVocabulary(t *testing.T) {
	got := benchmarkCreateRoute(createRequestRoute{Intent: createIntentDeck, Reason: "明确新建"}, "帮我做一份市场分析 PPT")
	if got.Intent != "create_deck" || got.TargetAgent != "PPTPlanner" {
		t.Fatalf("benchmark route = %#v", got)
	}
	if got.NormalizedRequest != "帮我做一份市场分析 PPT" {
		t.Fatalf("normalized request = %q", got.NormalizedRequest)
	}
}

func TestBenchmarkCreateRouteKeepsExistingTaskGuard(t *testing.T) {
	got := benchmarkCreateRoute(createRequestRoute{Intent: createIntentFixExisting}, "把第 2 页标题缩短")
	if got.Intent != "fix_existing" || got.TargetAgent != "PPTFixer after selecting an existing task" {
		t.Fatalf("benchmark route = %#v", got)
	}
}

func TestClassifyTaskMessageForBenchmarkKeepsConversationTask(t *testing.T) {
	got := ClassifyTaskMessageForBenchmark(
		context.Background(),
		"你决定主题和风格吧",
		"task-qinggan",
		"用户：我想做个青甘大环线的旅游项目介绍\n助手：请补充主题和风格，或授权我决定。",
		"",
	)
	if got.Intent != "create_deck" || got.TargetAgent != "PPTPlanner" || got.TaskID != "task-qinggan" || got.Action != messageActionPrepareCreate {
		t.Fatalf("benchmark contextual route = %#v", got)
	}
}

func TestClassifyTaskMessageForBenchmarkKeepsUnboundFixOutOfFixer(t *testing.T) {
	got := ClassifyTaskMessageForBenchmark(
		context.Background(),
		"把第 2 页标题缩短",
		"",
		"",
		"",
	)
	if got.Intent != messageIntentFix || got.Action != messageActionAskClarification || got.TargetAgent != "PPTFixer after selecting an existing task" {
		t.Fatalf("benchmark unbound fix = %#v", got)
	}
}
