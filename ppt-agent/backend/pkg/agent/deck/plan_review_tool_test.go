package deck

import "testing"

func TestPlanReviewAllowsTextOnlySlideWithoutBackground(t *testing.T) {
	workDir := t.TempDir()
	manifest := &TasksManifest{
		Title:    "架构清理验证",
		Theme:    "simple_gray",
		Template: "dynamic",
		Tasks: []*TaskItem{{
			TaskID:      "slide-1",
			PageIndex:   1,
			Title:       "架构清理验证",
			ContentType: "content_slide",
			Description: "说明固定模板和旧自由工具已清理，当前链路以动态 DeckSpec、Task Reviewer 和确定性渲染为核心。",
			OutputFile:  "1_architecture_cleanup.pptx",
			Status:      StatusPending,
			ContentPlan: &ContentPlan{
				SlideIntent: "确认架构清理范围、运行契约和验证结果。",
				Components: []PlanComponent{
					{
						ID:    "point-1",
						Type:  "key_point",
						Title: "动态 DeckSpec",
						Body:  "Planner 不再依赖固定整套 preset，而是根据用户输入和页面能力契约生成可执行组件计划。",
					},
					{
						ID:    "point-2",
						Type:  "key_point",
						Title: "确定性工具边界",
						Body:  "Reviewer 只修正 Go 审查报告指出的问题，渲染和文件交付由代码路径负责闭环。",
					},
				},
			},
		}},
	}
	if err := WriteTasksDraftManifest(workDir, manifest); err != nil {
		t.Fatal(err)
	}

	report, err := ReviewTasksDraftManifest(workDir, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Passed {
		t.Fatalf("text-only slide should pass with non-blocking background warning: %#v", report)
	}
	if report.IssueCount != 1 || report.Issues[0].Code != "missing_background_image" || report.Issues[0].Severity != "warning" {
		t.Fatalf("expected one background warning, got %#v", report.Issues)
	}
	if _, ok, err := CommitReviewedTasksDraftManifestIfPresent(workDir); err != nil || !ok {
		t.Fatalf("reviewed text-only draft was not committed: ok=%v err=%v", ok, err)
	}
}
