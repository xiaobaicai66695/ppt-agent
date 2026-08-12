package deck

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestManifestToolPatchesMultipleTasksAtomically(t *testing.T) {
	workDir := t.TempDir()
	manifest := &TasksManifest{
		Title: "示例", Theme: "charcoal_light", Template: "personal-summary",
		Tasks: []*TaskItem{
			{TaskID: "1", PageIndex: 1, Title: "旧标题", ContentType: "title_slide", Description: "旧描述", OutputFile: "1_cover.pptx", Status: StatusPending},
			{TaskID: "2", PageIndex: 2, Title: "旧目录", ContentType: "agenda", Description: "旧描述", OutputFile: "2_agenda.pptx", Status: StatusPending},
		},
	}
	if err := WriteTasksManifest(workDir, manifest); err != nil {
		t.Fatal(err)
	}

	tool := newManifestTool(workDir)
	result, err := tool.InvokableRun(context.Background(), `{
		"mode":"patch",
		"tasks":[
			{"task_id":"1","title":"两会工作报告","description":"概括年度目标","layout_variant":"statement_cards","content_plan":{"summary":"年度目标","visual_intent":{"role":"cards","preferred_variant":"statement_cards"},"elements":[{"type":"bullet_list","items":["目标一","目标二"]}]}},
			{"task_id":"2","title":"核心议程","description":"列出四项议程"}
		]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if result == "" {
		t.Fatal("expected structured result")
	}
	got, err := ReadTasksManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Tasks[0].Title != "两会工作报告" || got.Tasks[1].Title != "核心议程" {
		t.Fatalf("patch not applied: %#v", got.Tasks)
	}
	if got.Tasks[0].ContentPlan == nil || len(got.Tasks[0].ContentPlan.Elements) != 1 {
		t.Fatalf("content plan missing: %#v", got.Tasks[0].ContentPlan)
	}
	if got.Tasks[0].LayoutVariant != "statement_cards" {
		t.Fatalf("layout_variant = %q", got.Tasks[0].LayoutVariant)
	}
	if got.Tasks[0].ContentPlan.VisualIntent == nil || got.Tasks[0].ContentPlan.VisualIntent.Role != "cards" {
		t.Fatalf("visual intent missing: %#v", got.Tasks[0].ContentPlan.VisualIntent)
	}
	if got.Tasks[0].Status != StatusPending || got.Tasks[1].OutputFile != "2_agenda.pptx" {
		t.Fatalf("runtime fields were not preserved: %#v", got.Tasks)
	}
}

func TestManifestToolLeavesFileUnchangedOnInvalidPatch(t *testing.T) {
	workDir := t.TempDir()
	manifest := &TasksManifest{Title: "示例", Tasks: []*TaskItem{{
		TaskID: "1", PageIndex: 1, Title: "封面", ContentType: "title_slide",
		Description: "描述", OutputFile: "1_cover.pptx", Status: StatusPending,
	}}}
	if err := WriteTasksManifest(workDir, manifest); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(workDir, "tasks.json")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	tool := newManifestTool(workDir)
	if _, err := tool.InvokableRun(context.Background(), `{"mode":"patch","tasks":[{"task_id":"missing","title":"不应写入"}]}`); err == nil {
		t.Fatal("expected missing task error")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("manifest changed after invalid patch\nbefore=%s\nafter=%s", before, after)
	}
}

func TestManifestToolPatchesOrderedTasksWithoutIDs(t *testing.T) {
	workDir := t.TempDir()
	manifest := &TasksManifest{
		Title: "示例",
		Tasks: []*TaskItem{
			{TaskID: "slide-1", PageIndex: 1, Title: "封面", ContentType: "title_slide", Description: "旧描述", OutputFile: "1_cover.pptx", Status: StatusPending},
			{TaskID: "slide-2", PageIndex: 2, Title: "内容", ContentType: "content_slide", Description: "旧描述", OutputFile: "2_content.pptx", Status: StatusPending},
		},
	}
	if err := WriteTasksManifest(workDir, manifest); err != nil {
		t.Fatal(err)
	}
	tool := newManifestTool(workDir)
	tasks := `[
		{"background":"","content_plan":{"summary":"封面新规划"}},
		{"background":"","content_plan":{"summary":"内容新规划"}}
	]`
	args := `{"mode":"patch","tasks":` + strconv.Quote(tasks) + `}`

	if _, err := tool.InvokableRun(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	got, err := ReadTasksManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Tasks[0].ContentPlan == nil || got.Tasks[0].ContentPlan.Summary != "封面新规划" {
		t.Fatalf("first patch not applied: %#v", got.Tasks[0].ContentPlan)
	}
	if got.Tasks[1].ContentPlan == nil || got.Tasks[1].ContentPlan.Summary != "内容新规划" {
		t.Fatalf("second patch not applied: %#v", got.Tasks[1].ContentPlan)
	}
}

func TestManifestToolAcceptsTasksAsJSONArrayString(t *testing.T) {
	workDir := t.TempDir()
	tool := newManifestTool(workDir)
	tasks := `[
		{"task_id":"1","page_index":1,"title":"封面","content_type":"title_slide","description":"封面描述","output_file":"1_cover.pptx","status":"pending"},
		{"task_id":"2","page_index":2,"title":"目录","content_type":"agenda","description":"目录描述","output_file":"2_agenda.pptx","status":"pending"}
	]`
	args := `{"mode":"initialize","title":"示例","theme":"ocean_soft","template":"generic","tasks":` + strconv.Quote(tasks) + `}`

	if _, err := tool.InvokableRun(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	got, err := ReadTasksManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tasks) != 2 || got.Tasks[0].Title != "封面" || got.Tasks[1].ContentType != "agenda" {
		t.Fatalf("unexpected manifest: %#v", got.Tasks)
	}
}

func TestManifestToolAcceptsLooseTaskObjectString(t *testing.T) {
	workDir := t.TempDir()
	tool := newManifestTool(workDir)
	tasks := `{"task_id":"1","page_index":1,"title":"封面","content_type":"title_slide","description":"封面描述","output_file":"1_cover.pptx","status":"pending"},
		{"task_id":"2","page_index":2,"title":"目录","content_type":"agenda","description":"目录描述","output_file":"2_agenda.pptx","status":"pending"}`
	args := `{"mode":"initialize","title":"示例","theme":"ocean_soft","template":"generic","tasks":` + strconv.Quote(tasks) + `}`

	if _, err := tool.InvokableRun(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	got, err := ReadTasksManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tasks) != 2 || got.Tasks[0].TaskID != "1" || got.Tasks[1].TaskID != "2" {
		t.Fatalf("unexpected manifest: %#v", got.Tasks)
	}
}

func TestManifestToolAcceptsTaskStringWithExtraDelimiters(t *testing.T) {
	workDir := t.TempDir()
	tool := newManifestTool(workDir)
	tasks := `[
		{"task_id":"1","page_index":1,"title":"封面","content_type":"title_slide","description":"封面描述","output_file":"1_cover.pptx","status":"pending"}
	]},
	[
		{"task_id":"2","page_index":2,"title":"目录","content_type":"agenda","description":"目录描述","output_file":"2_agenda.pptx","status":"pending"}
	]}`
	args := `{"mode":"initialize","title":"示例","theme":"ocean_soft","template":"generic","tasks":` + strconv.Quote(tasks) + `}`

	if _, err := tool.InvokableRun(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	got, err := ReadTasksManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tasks) != 2 || got.Tasks[0].TaskID != "1" || got.Tasks[1].TaskID != "2" {
		t.Fatalf("unexpected manifest: %#v", got.Tasks)
	}
}

func TestManifestToolRecoversTasksFromWrappedArgumentString(t *testing.T) {
	workDir := t.TempDir()
	tool := newManifestTool(workDir)
	tasks := `{
		"mode":"initialize",
		"title":"被错误嵌套的参数",
		"tasks":[
			{"task_id":"1","page_index":1,"title":"封面","content_type":"title_slide","description":"封面描述","output_file":"1_cover.pptx","status":"pending"},
			{"task_id":"2","page_index":2,"title":"目录","content_type":"agenda","description":"目录描述","output_file":"2_agenda.pptx","status":"pending"}
		]
	}}`
	args := `{"mode":"initialize","title":"示例","theme":"ocean_soft","template":"generic","tasks":` + strconv.Quote(tasks) + `}`

	if _, err := tool.InvokableRun(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	got, err := ReadTasksManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tasks) != 2 || got.Tasks[0].TaskID != "1" || got.Tasks[1].TaskID != "2" {
		t.Fatalf("unexpected manifest: %#v", got.Tasks)
	}
}

func TestManifestToolRecoversTaskStringWithTopLevelFieldFragments(t *testing.T) {
	workDir := t.TempDir()
	tool := newManifestTool(workDir)
	tasks := `[
		{"task_id":"1","page_index":1,"title":"封面","content_type":"title_slide","description":"封面描述","output_file":"1_cover.pptx","status":"pending"}
	],
	"background":"minimalist_blue",
	"content_plan":{"summary":"这个对象不是任务"},
	[
		{"task_id":"2","page_index":2,"title":"目录","content_type":"agenda","description":"目录描述","output_file":"2_agenda.pptx","status":"pending"}
	]`
	args := `{"mode":"initialize","title":"示例","theme":"ocean_soft","template":"generic","tasks":` + strconv.Quote(tasks) + `}`

	if _, err := tool.InvokableRun(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	got, err := ReadTasksManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tasks) != 2 || got.Tasks[0].TaskID != "1" || got.Tasks[1].TaskID != "2" {
		t.Fatalf("unexpected manifest: %#v", got.Tasks)
	}
}

func TestManifestToolRecoversObservedMalformedTaskString(t *testing.T) {
	workDir := t.TempDir()
	tool := newManifestTool(workDir)
	tasks := `[
		{"task_id":"1","page_index":1,"title":"封面","content_type":"title_slide","description":"封面描述","output_file":"1_cover.pptx","status":"pending","content_plan":{"summary":"核心信息"}}},
		{"task_id":"2","page_index":2,"title":"未来展望","content_type":"content_slide","description":"介绍新疆在"一带一路"倡议下的发展机遇。","output_file":"2_future.pptx","status":"pending","content_plan":{"summary":"新疆是"一带一路"核心区。","elements":[{"type":"point","title":"发展机遇","text":["国际物流枢纽","清洁能源基地"],"description":["深化区域合作"]}]}}}
	]}`
	args := `{"mode":"initialize","title":"介绍新疆","theme":"ocean_soft","template":"product-intro","tasks":` + strconv.Quote(tasks) + `}`

	if _, err := tool.InvokableRun(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	got, err := ReadTasksManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tasks) != 2 {
		t.Fatalf("task count = %d, want 2", len(got.Tasks))
	}
	if got.Tasks[1].Description != `介绍新疆在"一带一路"倡议下的发展机遇。` {
		t.Fatalf("description = %q", got.Tasks[1].Description)
	}
	element := got.Tasks[1].ContentPlan.Elements[0]
	if element.Text != "国际物流枢纽\n清洁能源基地" || element.Description != "深化区域合作" {
		t.Fatalf("element was not normalized: %#v", element)
	}
}

func TestManifestToolDoesNotRecoverPartialTaskString(t *testing.T) {
	workDir := t.TempDir()
	tool := newManifestTool(workDir)
	tasks := `[{"task_id":"1","page_index":1,"title":"封面","content_type":"title_slide","description":"封面描述","output_file":"1_cover.pptx","status":"pending"},{"task_id":"2"`
	args := `{"mode":"initialize","title":"示例","tasks":` + strconv.Quote(tasks) + `}`

	if _, err := tool.InvokableRun(context.Background(), args); err == nil {
		t.Fatal("expected truncated task string error")
	}
	if _, err := os.Stat(filepath.Join(workDir, "tasks.json")); !os.IsNotExist(err) {
		t.Fatalf("partial manifest should not be written, stat err=%v", err)
	}
}

func TestManifestToolRejectsInvalidTasksString(t *testing.T) {
	workDir := t.TempDir()
	tool := newManifestTool(workDir)

	if _, err := tool.InvokableRun(context.Background(), `{"mode":"initialize","title":"示例","tasks":"not-json"}`); err == nil {
		t.Fatal("expected invalid tasks string error")
	}
	if _, err := os.Stat(filepath.Join(workDir, "tasks.json")); !os.IsNotExist(err) {
		t.Fatalf("manifest should not be written, stat err=%v", err)
	}
}

func TestRecommendedManifestNormalizesEveryTaskToSameBackgroundTheme(t *testing.T) {
	backgroundRoot := filepath.Join(t.TempDir(), "background_templates")
	imageDir := filepath.Join(backgroundRoot, "minimalist_blue", "images")
	if err := os.MkdirAll(imageDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"1.jpg", "2.jpg", "3.jpg"} {
		if err := os.WriteFile(filepath.Join(imageDir, name), []byte("image"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	manifest := &TasksManifest{Title: "推荐背景", Tasks: []*TaskItem{
		{TaskID: "1", PageIndex: 1, ContentType: "title_slide"},
		{TaskID: "2", PageIndex: 2, ContentType: "agenda"},
		{TaskID: "3", PageIndex: 3, ContentType: "content_slide"},
		{TaskID: "4", PageIndex: 4, ContentType: "section_divider"},
		{TaskID: "5", PageIndex: 5, ContentType: "card_grid"},
		{TaskID: "6", PageIndex: 6, ContentType: "image_text"},
		{TaskID: "7", PageIndex: 7, ContentType: "chart_slide"},
		{TaskID: "8", PageIndex: 8, ContentType: "kpi_dashboard"},
		{TaskID: "9", PageIndex: 9, ContentType: "summary_slide"},
		{TaskID: "10", PageIndex: 10, ContentType: "comparison_table", Background: "missing/images/1.jpg"},
		{TaskID: "11", PageIndex: 11, ContentType: "content_slide", Background: "minimalist_blue/images/2.jpg"},
	}}
	tool := &manifestTool{
		backgroundRoot: backgroundRoot, recommendedBackground: "minimalist_blue", normalizeBackgrounds: true,
	}
	if err := tool.normalizeManifestBackgrounds(manifest); err != nil {
		t.Fatal(err)
	}
	previous := ""
	for _, item := range manifest.Tasks {
		if item.Background == "" {
			t.Fatalf("task %q missing background", item.TaskID)
		}
		if !tool.validBackgroundReference(item.Background) {
			t.Fatalf("invalid background remained: %q", item.Background)
		}
		if backgroundTheme(item.Background) != "minimalist_blue" {
			t.Fatalf("background = %q, want minimalist_blue theme", item.Background)
		}
		if previous == item.Background {
			t.Fatalf("adjacent assigned backgrounds repeated: %q", item.Background)
		}
		previous = item.Background
	}
}

func TestManifestToolBackfillsLayoutVariantsAndUsesOneSectionVariant(t *testing.T) {
	workDir := t.TempDir()
	tool := newManifestTool(workDir)
	args := `{
		"mode":"initialize",
		"title":"介绍延安",
		"theme":"government_red",
		"template":"current-affairs",
		"tasks":[
			{"task_id":"1","page_index":1,"title":"封面","content_type":"title_slide","description":"封面","output_file":"1_cover.pptx","status":"pending","content_plan":{"summary":"封面"}},
			{"task_id":"2","page_index":2,"title":"历史篇","content_type":"section_divider","description":"历史章节","output_file":"2_history.pptx","status":"pending","content_plan":{"summary":"历史章节"}},
			{"task_id":"3","page_index":3,"title":"文化篇","content_type":"section_divider","description":"文化章节","output_file":"3_culture.pptx","status":"pending","content_plan":{"summary":"文化章节"}},
			{"task_id":"4","page_index":4,"title":"文化说明","content_type":"image_text","description":"说明","output_file":"4_image.pptx","status":"pending","content_plan":{"summary":"说明","visual_intent":{"role":"supporting_photo"}}},
			{"task_id":"5","page_index":5,"title":"文化说明二","content_type":"image_text","description":"说明二","output_file":"5_image.pptx","status":"pending","content_plan":{"summary":"说明二","visual_intent":{"role":"supporting_photo"}}}
		]
	}`
	if _, err := tool.InvokableRun(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	got, err := ReadTasksManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Tasks[1].ContentPlan == nil || got.Tasks[1].ContentPlan.SectionNumber != "01" {
		t.Fatalf("first section number = %#v", got.Tasks[1].ContentPlan)
	}
	if got.Tasks[2].ContentPlan == nil || got.Tasks[2].ContentPlan.SectionNumber != "02" {
		t.Fatalf("second section number = %#v", got.Tasks[2].ContentPlan)
	}
	if got.Tasks[1].LayoutVariant == "" || got.Tasks[2].LayoutVariant == "" || got.Tasks[1].LayoutVariant != got.Tasks[2].LayoutVariant {
		t.Fatalf("section variants should be consistent: %q %q", got.Tasks[1].LayoutVariant, got.Tasks[2].LayoutVariant)
	}
	if got.Tasks[1].LayoutVariant != "number_sidebar" {
		t.Fatalf("default section variant = %q, want number_sidebar", got.Tasks[1].LayoutVariant)
	}
	if got.Tasks[3].LayoutVariant == "" || got.Tasks[4].LayoutVariant == "" || got.Tasks[3].LayoutVariant == got.Tasks[4].LayoutVariant {
		t.Fatalf("image_text variants were not rotated: %q %q", got.Tasks[3].LayoutVariant, got.Tasks[4].LayoutVariant)
	}
	if got.Tasks[3].ContentPlan.VisualIntent.PreferredVariant != got.Tasks[3].LayoutVariant {
		t.Fatalf("visual_intent preferred variant not synced: %#v", got.Tasks[3].ContentPlan.VisualIntent)
	}
}

func TestManifestToolHonorsExplicitSectionVariantAcrossDeck(t *testing.T) {
	workDir := t.TempDir()
	tool := newManifestTool(workDir)
	args := `{
		"mode":"initialize",
		"title":"介绍桂林",
		"theme":"sage_calm",
		"template":"travel",
		"tasks":[
			{"task_id":"1","page_index":1,"title":"封面","content_type":"title_slide","description":"封面","output_file":"1_cover.pptx","status":"pending"},
			{"task_id":"2","page_index":2,"title":"山水格局","content_type":"section_divider","layout_variant":"quiet_title","description":"第一章","output_file":"2_section.pptx","status":"pending"},
			{"task_id":"3","page_index":3,"title":"城市肌理","content_type":"content_slide","description":"内容","output_file":"3_content.pptx","status":"pending"},
			{"task_id":"4","page_index":4,"title":"文化体验","content_type":"section_divider","description":"第二章","output_file":"4_section.pptx","status":"pending","content_plan":{"summary":"第二章","visual_intent":{"preferred_variant":"number_sidebar"}}}
		]
	}`
	if _, err := tool.InvokableRun(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	got, err := ReadTasksManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Tasks[1].LayoutVariant != "quiet_title" || got.Tasks[3].LayoutVariant != "quiet_title" {
		t.Fatalf("section variant should follow first explicit section variant: %q %q", got.Tasks[1].LayoutVariant, got.Tasks[3].LayoutVariant)
	}
	if got.Tasks[3].ContentPlan == nil || got.Tasks[3].ContentPlan.VisualIntent == nil || got.Tasks[3].ContentPlan.VisualIntent.PreferredVariant != "quiet_title" {
		t.Fatalf("section visual intent should match deck-wide section variant: %#v", got.Tasks[3].ContentPlan)
	}
}

func TestManifestToolAcceptsComponentPlanContract(t *testing.T) {
	workDir := t.TempDir()
	tool := newManifestTool(workDir)
	args := `{
		"mode":"initialize",
		"title":"组件计划示例",
		"theme":"ocean_soft",
		"template":"generic",
		"tasks":[{
			"task_id":"1",
			"page_index":1,
			"title":"能力矩阵",
			"content_type":"card_grid",
			"layout_variant":"featured_card_plus_grid",
			"description":"说明产品能力矩阵",
			"output_file":"1_cards.pptx",
			"status":"pending",
			"content_plan":{
				"summary":"三层能力矩阵支撑端到端交付",
				"slide_intent":"说明产品能力矩阵",
				"components":[
					{"id":"headline_1","type":"headline","text":"三层能力矩阵支撑端到端交付","role":"main_point"},
					{"id":"feature_card_1","type":"feature_card","title":"数据接入","body":"统一连接业务库、文件和接口数据","emphasis":"primary"},
					{"id":"feature_card_2","type":"feature_card","title":"智能分析","body":"自动识别趋势、异常和关键指标","emphasis":"normal"}
				],
				"capacity_hint":{"estimated_density":"normal","overflow_risk":"low","component_count":99},
				"reviewer_status":{"planner_round":1,"locked":"true","issues":[]}
			}
		}]
	}`
	if _, err := tool.InvokableRun(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	got, err := ReadTasksManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	plan := got.Tasks[0].ContentPlan
	if plan == nil || plan.SlideIntent != "说明产品能力矩阵" || len(plan.Components) != 3 {
		t.Fatalf("component plan missing: %#v", plan)
	}
	if plan.CapacityHint == nil || plan.CapacityHint.ComponentCount != 3 {
		t.Fatalf("capacity hint not normalized: %#v", plan.CapacityHint)
	}
	if plan.ReviewerStatus == nil || !plan.ReviewerStatus.Locked {
		t.Fatalf("reviewer status not parsed: %#v", plan.ReviewerStatus)
	}
}

func TestManifestToolRejectsInvalidComponentPlan(t *testing.T) {
	workDir := t.TempDir()
	tool := newManifestTool(workDir)
	args := `{
		"mode":"initialize",
		"title":"坏组件",
		"tasks":[{
			"task_id":"1",
			"page_index":1,
			"title":"能力矩阵",
			"content_type":"card_grid",
			"description":"说明产品能力矩阵",
			"output_file":"1_cards.pptx",
			"status":"pending",
			"content_plan":{
				"summary":"摘要",
				"components":[{"id":"absolute_box","type":"x_y_positioned_box","title":"非法组件"}],
				"capacity_hint":{"estimated_density":"normal","overflow_risk":"low"}
			}
		}]
	}`
	if _, err := tool.InvokableRun(context.Background(), args); err == nil {
		t.Fatal("expected invalid component type error")
	}
	if _, err := os.Stat(filepath.Join(workDir, "tasks.json")); !os.IsNotExist(err) {
		t.Fatalf("invalid component plan should not be written, stat err=%v", err)
	}
}
