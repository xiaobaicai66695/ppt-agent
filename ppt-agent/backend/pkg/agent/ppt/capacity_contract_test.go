package ppt

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestReviewReportsStructuredCapacityDiagnostics(t *testing.T) {
	manifest := &TasksManifest{Tasks: []*TaskItem{{
		TaskID: "slide-1", PageIndex: 1, Title: "指标", ContentType: "kpi_dashboard",
		OutputFile: "slide_01.pptx", Status: StatusPending,
		ContentPlan: &ContentPlan{Summary: "用指标解释经营判断。", SlideIntent: "建立数据基线。", Components: []PlanComponent{
			{ID: "1", Type: "stat", Body: "指标一"},
			{ID: "2", Type: "stat", Body: "指标二"},
			{ID: "3", Type: "stat", Body: "指标三"},
			{ID: "4", Type: "stat", Body: "指标四"},
			{ID: "5", Type: "stat", Body: "指标五"},
		}},
	}}}

	report := ReviewTasksManifest(manifest, "test", 1)
	var capacity *PlanReviewIssue
	for i := range report.Issues {
		if report.Issues[i].Code == "overload_capacity" {
			capacity = &report.Issues[i]
			break
		}
	}
	if capacity == nil || capacity.Severity != "error" || capacity.ActualComponents != 5 || capacity.MaxComponents != 4 || capacity.OverflowComponents != 1 {
		t.Fatalf("unexpected capacity issue: %#v", capacity)
	}
	if capacity.RecommendedMin != 3 || capacity.RecommendedMax != 4 || capacity.ContractVersion == "" {
		t.Fatalf("capacity issue omitted contract facts: %#v", capacity)
	}
}

func TestPlannerManifestUsesLoadedContractForHardLimit(t *testing.T) {
	workDir := t.TempDir()
	planner := newPlannerManifestTool(workDir, nil, "容量测试")
	result, err := planner.InvokableRun(context.Background(), `{"mode":"initialize","tasks":[{"page_index":1,"title":"内容","content_type":"content_slide","content_plan":{"summary":"页面核心判断足够具体。","slide_intent":"说明容量硬上限。","visual_intent":{"asset_purpose":"background","asset_query":"technology"},"components":[{"id":"1","type":"insight","body":"具体判断一"},{"id":"2","type":"insight","body":"具体判断二"},{"id":"3","type":"insight","body":"具体判断三"},{"id":"4","type":"insight","body":"具体判断四"},{"id":"5","type":"insight","body":"具体判断五"},{"id":"6","type":"insight","body":"具体判断六"},{"id":"7","type":"insight","body":"具体判断七"},{"id":"8","type":"insight","body":"具体判断八"},{"id":"9","type":"insight","body":"具体判断九"}]}}]}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"quality_gate_passed":false`) || !strings.Contains(result, `"max_components":8`) {
		t.Fatalf("planner did not expose hard capacity diagnostics: %s", result)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(result), &payload); err != nil {
		t.Fatal(err)
	}
	issues, ok := payload["issues"].([]any)
	if !ok || len(issues) == 0 {
		t.Fatalf("planner result has no structured issues: %#v", payload)
	}
}

func TestInspectManifestCapacityReportsRecommendedAndHardLimits(t *testing.T) {
	manifest := &TasksManifest{Tasks: []*TaskItem{
		{PageIndex: 1, ContentType: "content_slide", ContentPlan: &ContentPlan{Components: make([]PlanComponent, 7)}},
		{PageIndex: 2, ContentType: "kpi_dashboard", ContentPlan: &ContentPlan{Components: make([]PlanComponent, 5)}},
	}}
	report := InspectManifestCapacity(manifest, "")
	if len(report.AboveRecommendedPages) != 2 || len(report.OverLimitPages) != 1 || report.OverLimitPages[0] != 2 {
		t.Fatalf("unexpected benchmark capacity report: %#v", report)
	}
}
