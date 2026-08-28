package web

import "testing"

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
