package ppt

import "testing"

func TestPlanReviewScopesManifestValidationErrorToItsTaskPage(t *testing.T) {
	manifest := &TasksManifest{Tasks: []*TaskItem{{
		TaskID:      "slide-10",
		PageIndex:   10,
		Title:       "关键指标",
		ContentType: "kpi_dashboard",
		OutputFile:  "10_kpi.pptx",
		Status:      StatusPending,
		ContentPlan: &ContentPlan{
			Summary:     "用关键指标说明液流电池储能的增长和经济性。",
			SlideIntent: "给出核心数据结论。",
			Components: []PlanComponent{
				{ID: "one", Type: "kpi_metric", Title: "一"},
				{ID: "two", Type: "kpi_metric", Title: "二"},
				{ID: "three", Type: "kpi_metric", Title: "三"},
				{ID: "four", Type: "kpi_metric", Title: "四"},
				{ID: "five", Type: "kpi_metric", Title: "五"},
			},
		},
	}}}

	report := ReviewTasksManifest(manifest, "tasks.draft.json", 1)
	if report.Passed {
		t.Fatal("over-capacity KPI page should not pass review")
	}
	var schemaIssue *PlanReviewIssue
	for index := range report.Issues {
		issue := &report.Issues[index]
		if issue.Code == "invalid_component_schema" {
			schemaIssue = issue
			break
		}
	}
	if schemaIssue == nil || schemaIssue.PageIndex != 10 {
		t.Fatalf("schema issue = %#v, want page_index 10", schemaIssue)
	}

	payload := buildPlanReviewRevisionPayload(manifest, 1, report)
	if len(payload.Scope.AllowedPageIndexes) != 1 || payload.Scope.AllowedPageIndexes[0] != 10 {
		t.Fatalf("allowed review scope = %#v, want [10]", payload.Scope)
	}
	if len(payload.IncludedTasks) != 1 || payload.IncludedTasks[0].PageIndex != 10 {
		t.Fatalf("included tasks = %#v, want slide 10", payload.IncludedTasks)
	}
}
