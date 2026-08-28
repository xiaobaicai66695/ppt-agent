package deck

import (
	"fmt"
	"strings"
	"testing"
)

func TestPlanReviewAllowsTextOnlySlideWithoutBackground(t *testing.T) {
	workDir := t.TempDir()
	manifest := &TasksManifest{
		Title: "架构清理验证",
		Tasks: []*TaskItem{{
			TaskID:      "slide-1",
			PageIndex:   1,
			Title:       "架构清理验证",
			ContentType: "content_slide",
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
	foundBackgroundWarning := false
	for _, issue := range report.Issues {
		if issue.Code == "missing_background_image" && issue.Severity == "warning" {
			foundBackgroundWarning = true
		}
		if issue.Severity != "warning" {
			t.Fatalf("expected only non-blocking warnings, got %#v", report.Issues)
		}
	}
	if !foundBackgroundWarning {
		t.Fatalf("expected background warning, got %#v", report.Issues)
	}
	if _, ok, err := CommitReviewedTasksDraftManifestIfPresent(workDir); err != nil || !ok {
		t.Fatalf("reviewed text-only draft was not committed: ok=%v err=%v", ok, err)
	}
}

func TestPlanReviewUsesContractArgumentBlockMinimum(t *testing.T) {
	shortBody := strings.Repeat("论据与解释仍然不足。", 30)
	manifest := &TasksManifest{
		Title: "AI 产业分析",
		Tasks: []*TaskItem{{
			TaskID: "slide-1", PageIndex: 1, Title: "产业判断", ContentType: "deep_dive",
			OutputFile: "1_analysis.pptx", Status: StatusPending,
			ContentPlan: &ContentPlan{
				SlideIntent:  "用完整论述串联技术、市场和组织证据。",
				VisualIntent: &VisualIntent{AssetPurpose: "background", AssetQuery: "AI data center wide landscape clean negative space"},
				Components:   []PlanComponent{{ID: "argument-1", Type: "argument_block", Body: shortBody}},
			},
		}},
	}

	report := ReviewTasksManifest(manifest, "tasks.draft.json", 1)
	found := false
	for _, issue := range report.Issues {
		if issue.Code == "low_information_density" && issue.ComponentID == "argument-1" {
			found = true
			if !strings.Contains(issue.Message, "440") {
				t.Fatalf("expected threshold in issue message, got %q", issue.Message)
			}
		}
	}
	if !found {
		t.Fatalf("expected short argument_block issue, got %#v", report.Issues)
	}
}

func TestPlannerPreflightBlocksOutlinePlaceholders(t *testing.T) {
	manifest := &TasksManifest{
		Title: "制造强国复盘",
		Tasks: []*TaskItem{{
			TaskID: "slide-12", PageIndex: 12, Title: "成就与短板：客观审视", ContentType: "two_column",
			OutputFile: "12_gap.pptx", Status: StatusPending,
			ContentPlan: &ContentPlan{
				Summary:      "从事实、风险和洞察三个层面对制造强国建设进行客观审视。",
				SlideIntent:  "用左右对比说明已有积累和仍需补齐的关键短板。",
				VisualIntent: &VisualIntent{AssetPurpose: "background", AssetQuery: "industry"},
				Components: []PlanComponent{
					{ID: "left-fact-1", Type: "fact_card", Title: "事实", Body: "左栏成就"},
					{ID: "right-risk-1", Type: "risk_item", Title: "风险", Body: "右栏短板"},
					{ID: "insight-1", Type: "insight", Title: "洞察", Body: "对比后的判断"},
				},
			},
		}},
	}

	report := ReviewTasksManifest(manifest, "tasks.draft.json", 1)
	if report.Passed {
		t.Fatalf("placeholder scaffold text must block review: %#v", report)
	}
	found := 0
	for _, issue := range report.Issues {
		if issue.Code == "low_information_density" && issue.Severity == "error" {
			found++
		}
	}
	if found < 3 {
		t.Fatalf("expected blocking density errors for placeholder bodies, got %#v", report.Issues)
	}
}

func TestPlannerPreflightBlocksShortOutlineBodies(t *testing.T) {
	manifest := &TasksManifest{
		Title: "制造强国复盘",
		Tasks: []*TaskItem{{
			TaskID: "slide-14", PageIndex: 14, Title: "未来规划", ContentType: "content_slide",
			OutputFile: "14_future.pptx", Status: StatusPending,
			ContentPlan: &ContentPlan{
				Summary:      "梳理自主可控、智能制造、绿色低碳和开放合作四个方向。",
				SlideIntent:  "把纲要式方向扩写为可执行判断，避免只有标题词。",
				VisualIntent: &VisualIntent{AssetPurpose: "background", AssetQuery: "industry"},
				Components: []PlanComponent{
					{ID: "future-list", Type: "numbered_list", Items: []string{"自主可控", "智能制造", "绿色低碳", "开放合作"}},
					{ID: "insight-1", Type: "insight", Title: "洞察", Body: "未来展望"},
				},
			},
		}},
	}

	report := ReviewTasksManifest(manifest, "tasks.draft.json", 1)
	if report.Passed {
		t.Fatalf("outline-only bodies must block review: %#v", report)
	}
	found := false
	for _, issue := range report.Issues {
		if issue.Code == "low_information_density" && issue.Severity == "error" && issue.ComponentID == "future-list" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected short list item density error, got %#v", report.Issues)
	}
}

func TestPlanReviewAcceptsExplicitCleanTextOnlyPolicy(t *testing.T) {
	manifest := &TasksManifest{
		Title: "纯文字汇报",
		Tasks: []*TaskItem{{
			TaskID: "slide-1", PageIndex: 1, Title: "核心结论", ContentType: "content_slide",
			OutputFile: "1_summary.pptx", Status: StatusPending,
			ContentPlan: &ContentPlan{
				SlideIntent:  "在无图片约束下清晰呈现核心判断。",
				VisualIntent: &VisualIntent{Role: "clean_text_only"},
				Components:   []PlanComponent{{ID: "point-1", Type: "key_point", Body: "纯文字模式仍保留清晰观点与论据层级。"}},
			},
		}},
	}

	issues := plannerPreflightIssues(manifest)
	for _, issue := range issues {
		if issue.Code == "missing_background_image" {
			t.Fatalf("explicit clean text-only policy must not require a background: %#v", issues)
		}
	}
}

func TestPlannerPreflightFlagsMultipleBackgroundKeywordsForSameContentType(t *testing.T) {
	manifest := &TasksManifest{
		Title: "AI 产业分析",
		Tasks: []*TaskItem{
			backgroundQueryTask(1, "diplomacy"),
			backgroundQueryTask(2, "energy"),
		},
	}
	issues := plannerPreflightIssues(manifest)
	found := false
	for _, issue := range issues {
		if issue.Code == "layout_mismatch" && strings.Contains(issue.Message, "同一页面类型") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected same content_type background warning, got %#v", issues)
	}
}

func TestPlannerPreflightFlagsVerboseBackgroundQuery(t *testing.T) {
	manifest := &TasksManifest{
		Title: "AI 产业分析",
		Tasks: []*TaskItem{
			backgroundQueryTask(1, "light global diplomacy meeting wide landscape clean negative space"),
		},
	}
	issues := plannerPreflightIssues(manifest)
	found := false
	for _, issue := range issues {
		if issue.Code == "layout_mismatch" && strings.Contains(issue.Message, "asset_query 过长") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected verbose background query warning, got %#v", issues)
	}
}

func TestPlannerPreflightFlagsLowVisualMixForNarrativeDeck(t *testing.T) {
	manifest := &TasksManifest{
		Title: "AI 产业分析",
		Tasks: []*TaskItem{
			backgroundQueryTask(1, "light AI office collaboration wide landscape clean negative space"),
			backgroundQueryTask(2, "light AI office collaboration wide landscape clean negative space"),
			backgroundQueryTask(3, "light AI office collaboration wide landscape clean negative space"),
			backgroundQueryTask(4, "light AI office collaboration wide landscape clean negative space"),
		},
	}
	issues := plannerPreflightIssues(manifest)
	found := false
	for _, issue := range issues {
		if issue.Code == "layout_mismatch" && strings.Contains(issue.Message, "图文混排不足") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected visual mix warning, got %#v", issues)
	}
}

func TestPlannerPreflightFlagsImageTextWithoutForegroundImage(t *testing.T) {
	task := imageTextTask("slide-1", 1)
	task.ContentPlan.Components = []PlanComponent{
		{ID: "point-1", Type: "key_point", Body: "本页用观点和解释说明场景，但没有把真实示例图片作为页内素材。"},
		{ID: "paragraph-1", Type: "paragraph", Body: "正文负责解释场景、问题和结论，形成可读的图文混排页面。"},
	}
	manifest := &TasksManifest{
		Title: "AI 产业分析",
		Tasks: []*TaskItem{task},
	}
	issues := plannerPreflightIssues(manifest)
	found := false
	for _, issue := range issues {
		if issue.Code == "layout_mismatch" && strings.Contains(issue.Message, "缺少 scene/evidence 图片组件") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected foreground image warning, got %#v", issues)
	}
}

func TestPlanReviewFlagsSingleImageTextVariantAcrossDeck(t *testing.T) {
	manifest := &TasksManifest{
		Title: "AI 产业分析",
		Tasks: []*TaskItem{
			imageTextTask("slide-1", 1),
			imageTextTask("slide-2", 2),
			imageTextTask("slide-3", 3),
		},
	}
	for _, task := range manifest.Tasks {
		task.LayoutVariant = "image_left"
	}
	report := ReviewTasksManifest(manifest, "tasks.draft.json", 1)
	found := false
	for _, issue := range report.Issues {
		if issue.Code == "layout_mismatch" && strings.Contains(issue.Message, "layout_variant 过于单一") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected image_text variant warning, got %#v", report.Issues)
	}
}

func TestNormalizeManifestLayoutVariantsRotatesImageTextVariants(t *testing.T) {
	manifest := &TasksManifest{Tasks: []*TaskItem{
		imageTextTask("slide-1", 1),
		imageTextTask("slide-2", 2),
		imageTextTask("slide-3", 3),
		imageTextTask("slide-4", 4),
	}}
	normalizeManifestLayoutVariants(manifest)
	got := []string{
		manifest.Tasks[0].LayoutVariant,
		manifest.Tasks[1].LayoutVariant,
		manifest.Tasks[2].LayoutVariant,
		manifest.Tasks[3].LayoutVariant,
	}
	want := []string{"image_left", "image_right", "image_top_band", "image_bottom_band"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("variant[%d]=%q, want %q; all=%#v", i, got[i], want[i], got)
		}
		if manifest.Tasks[i].ContentPlan.VisualIntent.PreferredVariant != want[i] {
			t.Fatalf("preferred variant[%d]=%q, want %q", i, manifest.Tasks[i].ContentPlan.VisualIntent.PreferredVariant, want[i])
		}
	}
}

func backgroundQueryTask(page int, query string) *TaskItem {
	return &TaskItem{
		TaskID: fmt.Sprintf("slide-%d", page), PageIndex: page, Title: fmt.Sprintf("页面 %d", page), ContentType: "content_slide",
		OutputFile: fmt.Sprintf("%d_slide.pptx", page), Status: StatusPending,
		ContentPlan: &ContentPlan{
			SlideIntent:  "用观点和证据支撑整套叙事。",
			VisualIntent: &VisualIntent{AssetPurpose: "background", AssetQuery: query},
			Components:   []PlanComponent{{ID: "point-1", Type: "key_point", Body: "本页用明确观点串联事实证据，避免只有短句罗列。"}},
		},
	}
}

func imageTextTask(id string, page int) *TaskItem {
	return &TaskItem{
		TaskID: id, PageIndex: page, Title: fmt.Sprintf("图文页 %d", page), ContentType: "image_text",
		OutputFile: fmt.Sprintf("%d_image_text.pptx", page), Status: StatusPending,
		ContentPlan: &ContentPlan{
			SlideIntent:  "通过图文混排说明场景证据。",
			VisualIntent: &VisualIntent{AssetPurpose: "background", AssetQuery: "light AI office collaboration wide landscape clean negative space"},
			Components: []PlanComponent{
				{ID: "image-1", Type: "image", AssetPurpose: "scene", AssetQuery: "AI workplace team reviewing dashboard", Caption: "真实场景图"},
				{ID: "point-1", Type: "key_point", Body: "本页用真实场景图承载示例素材，正文负责解释业务语境、关键问题和判断。"},
				{ID: "paragraph-1", Type: "paragraph", Body: "图像用于承载真实业务环境，正文负责解释场景、问题和结论，形成可读的图文混排页面。"},
			},
		},
	}
}
