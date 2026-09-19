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
	var capacityIssue *PlanReviewIssue
	for index := range report.Issues {
		issue := &report.Issues[index]
		if issue.Code == "overload_capacity" {
			capacityIssue = issue
			break
		}
	}
	if capacityIssue == nil || capacityIssue.PageIndex != 10 {
		t.Fatalf("capacity issue = %#v, want page_index 10", capacityIssue)
	}

	payload := buildPlanReviewRevisionPayload(manifest, 1, report)
	if len(payload.Scope.AllowedPageIndexes) != 1 || payload.Scope.AllowedPageIndexes[0] != 10 {
		t.Fatalf("allowed review scope = %#v, want [10]", payload.Scope)
	}
	if len(payload.IncludedTasks) != 1 || payload.IncludedTasks[0].PageIndex != 10 {
		t.Fatalf("included tasks = %#v, want slide 10", payload.IncludedTasks)
	}
}

func TestBuildReviewAdviceCarriesCapacityFacts(t *testing.T) {
	manifest := &TasksManifest{Tasks: []*TaskItem{{
		PageIndex: 4, ContentType: "card_grid",
		ContentPlan: &ContentPlan{Components: []PlanComponent{{ID: "headline", Type: "headline"}}},
	}}}
	issues := []PlanReviewIssue{{Code: "overload_capacity", PageIndex: 4, ActualComponents: 7, RecommendedMin: 3, RecommendedMax: 5, MaxComponents: 6, OverflowComponents: 1, Message: "too many"}}
	advice := buildReviewAdvice(manifest.Tasks, issues)
	if len(advice) != 1 {
		t.Fatalf("advice length = %d", len(advice))
	}
	got := advice[0]
	if got.PageIndex != 4 || got.ContentType != "card_grid" || got.CurrentComponents != 7 || got.RecommendedMax != 5 || got.MaxComponents != 6 {
		t.Fatalf("advice = %#v", got)
	}
	if got.RecommendedChange == "" {
		t.Fatal("advice should include recommended change")
	}
}
