package deck

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

func TestManifestToolNormalizesNarrativeAndListComponentAliases(t *testing.T) {
	workDir := t.TempDir()
	tool := newConfiguredManifestTool(workDir, filepath.Join(t.TempDir(), "skills"), nil, "说明主流程优化", false)
	args := `{
		"mode":"initialize",
		"theme":"ocean_soft",
		"template":"tech-intro",
		"tasks":[{
			"task_id":"1",
			"page_index":1,
			"title":"为什么要优化主流程",
			"content_type":"content_slide",
			"description":"用一页说明优化必要性",
			"output_file":"1_flow.pptx",
			"status":"pending",
			"content_plan":{
				"summary":"主流程需要先规划再润色",
				"components":[
					{"id":"arg_1","type":"long_paragraph","title":"核心判断","body":"当前问题不是单页渲染能力不足，而是规划阶段没有给出足够完整的论述、证据和结论，导致后续只能渲染出泛化短句。"},
					{"id":"steps_1","type":"ordered_list","title":"改造顺序","items":["先规划叙事结构","再审查并润色内容","最后按组件生成页面"]},
					{"id":"list_1","type":"list_group","title":"验收点","items":["论点完整","列表清晰"]}
				],
				"capacity_hint":{"estimated_density":"normal","overflow_risk":"low","component_count":3},
				"reviewer_status":{"planner_round":1,"locked":true,"issues":[]}
			}
		}]
	}`
	if _, err := tool.InvokableRun(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	if _, err := tool.InvokableRun(context.Background(), `{"mode":"commit"}`); err != nil {
		t.Fatal(err)
	}
	got, err := ReadTasksManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	components := got.Tasks[0].ContentPlan.Components
	if components[0].Type != "argument_block" || components[1].Type != "numbered_list" || components[2].Type != "list" {
		t.Fatalf("component aliases not normalized: %#v", components)
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

func TestManifestToolInfersInitializeTitleFromQuery(t *testing.T) {
	workDir := t.TempDir()
	tool := newConfiguredManifestTool(workDir, filepath.Join(t.TempDir(), "skills"), nil, "介绍延安", false)
	args := `{
		"mode":"initialize",
		"theme":"government_red",
		"template":"current-affairs",
		"tasks":[
			{"task_id":"1","page_index":1,"title":"封面","content_type":"title_slide","description":"封面描述","output_file":"1_cover.pptx","status":"pending"},
			{"task_id":"2","page_index":2,"title":"红色历史","content_type":"content_slide","description":"介绍延安红色历史","output_file":"2_history.pptx","status":"pending"}
		]
	}`

	if _, err := tool.InvokableRun(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	got, err := ReadTasksDraftManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "介绍延安" {
		t.Fatalf("title = %q, want query fallback", got.Title)
	}
	if _, err := os.Stat(filepath.Join(workDir, "tasks.json")); !os.IsNotExist(err) {
		t.Fatalf("configured manifest tool should write draft before commit, stat err=%v", err)
	}
	if _, err := tool.InvokableRun(context.Background(), `{"mode":"commit"}`); err != nil {
		t.Fatal(err)
	}
	finalManifest, err := ReadTasksManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if finalManifest.Title != "介绍延安" || len(finalManifest.Tasks) != 2 {
		t.Fatalf("commit did not publish draft: %#v", finalManifest)
	}
	if _, err := os.Stat(filepath.Join(workDir, "tasks.draft.json")); !os.IsNotExist(err) {
		t.Fatalf("draft should be removed after commit, stat err=%v", err)
	}
}

func TestManifestToolInfersInitializeTitleFromSlideTitle(t *testing.T) {
	workDir := t.TempDir()
	tool := newManifestTool(workDir)
	args := `{
		"mode":"initialize",
		"theme":"sage_calm",
		"template":"travel",
		"tasks":[
			{"task_id":"1","page_index":1,"title":"封面","content_type":"title_slide","description":"封面描述","output_file":"1_cover.pptx","status":"pending"},
			{"task_id":"2","page_index":2,"title":"桂林山水格局","content_type":"content_slide","description":"介绍桂林山水格局","output_file":"2_landscape.pptx","status":"pending"}
		]
	}`

	if _, err := tool.InvokableRun(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	got, err := ReadTasksManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "桂林山水格局" {
		t.Fatalf("title = %q, want slide title fallback", got.Title)
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

func TestConfiguredManifestLeavesLocalBackgroundEmptyWhenImageSearchAvailable(t *testing.T) {
	workDir := t.TempDir()
	skillsDir := filepath.Join(t.TempDir(), "skills")
	imageDir := filepath.Join(skillsDir, "visual_designer", "background_templates", "minimalist_blue", "images")
	if err := os.MkdirAll(imageDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(imageDir, "1.jpg"), []byte("image"), 0o644); err != nil {
		t.Fatal(err)
	}

	outline := &TaskOutline{
		ContentMode:           OutlineContentModeRecommendedStyle,
		UseBackground:         true,
		RecommendedBackground: "minimalist_blue",
	}
	tool := newConfiguredManifestTool(workDir, skillsDir, outline, "低空经济", true)
	args := `{
		"mode":"initialize",
		"title":"低空经济",
		"theme":"ocean_soft",
		"template":"research-report",
		"tasks":[
			{"task_id":"1","page_index":1,"title":"封面","content_type":"title_slide","description":"封面","output_file":"1_cover.pptx","status":"pending","content_plan":{"summary":"封面"}},
			{"task_id":"2","page_index":2,"title":"场景","content_type":"image_text","description":"场景","background":"missing/images/1.jpg","output_file":"2_scene.pptx","status":"pending","content_plan":{"summary":"场景","visual_intent":{"asset_purpose":"background","asset_query":"aerial city skyline at blue hour, wide landscape, clean negative space"}}}
		]
	}`
	if _, err := tool.InvokableRun(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	got, err := ReadTasksDraftManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Tasks[0].Background != "" {
		t.Fatalf("empty local background should stay empty when image search is available, got %q", got.Tasks[0].Background)
	}
	if got.Tasks[1].Background != "missing/images/1.jpg" {
		t.Fatalf("draft should not be rewritten by local background fallback, got %q", got.Tasks[1].Background)
	}
}

func TestManifestToolBackfillsOnlyImplementedLayoutVariants(t *testing.T) {
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
	if got.Tasks[3].LayoutVariant != "" || got.Tasks[4].LayoutVariant != "" {
		t.Fatalf("unsupported image_text variants should not be backfilled: %q %q", got.Tasks[3].LayoutVariant, got.Tasks[4].LayoutVariant)
	}
	if got.Tasks[3].ContentPlan.VisualIntent.PreferredVariant != "" {
		t.Fatalf("unsupported image_text preferred variant should not be synced: %#v", got.Tasks[3].ContentPlan.VisualIntent)
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
	if got.Tasks[1].LayoutVariant != "number_sidebar" || got.Tasks[3].LayoutVariant != "number_sidebar" {
		t.Fatalf("unsupported section variants should normalize to implemented default: %q %q", got.Tasks[1].LayoutVariant, got.Tasks[3].LayoutVariant)
	}
	if got.Tasks[3].ContentPlan == nil || got.Tasks[3].ContentPlan.VisualIntent == nil || got.Tasks[3].ContentPlan.VisualIntent.PreferredVariant != "number_sidebar" {
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

func TestManifestToolAcceptsAtomicRenderingComponents(t *testing.T) {
	workDir := t.TempDir()
	tool := newManifestTool(workDir)
	args := `{
		"mode":"initialize",
		"title":"组件化架构页",
		"theme":"charcoal_light",
		"template":"generic",
		"tasks":[{
			"task_id":"1",
			"page_index":1,
			"title":"规划到渲染的执行链路",
			"content_type":"deep_dive",
			"description":"用架构模块和连接关系说明 PPT 生成链路",
			"output_file":"1_architecture.pptx",
			"status":"pending",
			"content_plan":{
				"summary":"Planner 只规划语义组件，Python generator 负责稳定排版。",
				"slide_intent":"说明组件式生成链路",
				"components":[
					{"id":"tag_1","type":"tag","text":"准备阶段"},
					{"id":"icon_1","type":"icon","icon":"LLM","title":"规划器"},
					{"id":"divider_1","type":"divider","role":"区分规划与渲染"},
					{"id":"shape_1","type":"shape","role":"强调组件计划已锁定"},
					{"id":"box_1","type":"architecture_box","title":"Deck Planner","body":"负责整套叙事、章节和页面角色，不接触坐标、字号或颜色。","role":"规划层"},
					{"id":"box_2","type":"architecture_box","title":"Component Planner","body":"把每页拆成 headline、卡片、指标、架构模块等可执行语义组件。","role":"规划层"},
					{"id":"arrow_1","type":"arrow","relation":"输出 DeckSpec","target":"box_2"}
				],
				"capacity_hint":{"estimated_density":"normal","overflow_risk":"low","component_count":7},
				"reviewer_status":{"planner_round":1,"locked":true,"issues":[]}
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
	components := got.Tasks[0].ContentPlan.Components
	if len(components) != 7 || components[6].Target != "box_2" || components[1].Icon != "LLM" {
		t.Fatalf("atomic components not preserved: %#v", components)
	}
}

func TestManifestToolAcceptsUnifiedComponentContractTypes(t *testing.T) {
	workDir := t.TempDir()
	tool := newManifestTool(workDir)
	args := `{
		"mode":"initialize",
		"title":"统一组件契约",
		"theme":"sage_calm",
		"template":"generic",
		"tasks":[{
			"task_id":"1",
			"page_index":1,
			"title":"桂林山水甲天下",
			"content_type":"title_slide",
			"description":"封面页",
			"output_file":"1_cover.pptx",
			"status":"pending",
			"content_plan":{
				"summary":"介绍桂林",
				"components":[
					{"id":"deck_title_1","type":"deck_title","text":"桂林山水甲天下"},
					{"id":"eyebrow_1","type":"eyebrow","text":"城市介绍"},
					{"id":"fact_card_1","type":"fact_card","title":"山水名片","body":"喀斯特峰林与漓江构成城市识别"}
				],
				"capacity_hint":{"estimated_density":"normal","overflow_risk":"low","component_count":3}
			}
		},{
			"task_id":"2",
			"page_index":2,
			"title":"区域路线",
			"content_type":"region_map",
			"description":"说明桂林周边路线",
			"output_file":"2_region.pptx",
			"status":"pending",
			"content_plan":{
				"summary":"区域关系",
				"components":[
					{"id":"map_1","type":"map","title":"市区-漓江-阳朔","body":"从市区向南串联漓江和阳朔"},
					{"id":"image_1","type":"image","asset_query":"桂林 漓江 山水 背景图","caption":"漓江山水"},
					{"id":"stat_1","type":"stat","title":"短途路线","text":"3天"},
					{"id":"recommendation_1","type":"recommendation","title":"建议","body":"按市区、漓江、阳朔组织路线"}
				],
				"capacity_hint":{"estimated_density":"normal","overflow_risk":"low","component_count":3}
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
	if got.Tasks[0].LayoutVariant != "" || got.Tasks[1].LayoutVariant != "" {
		t.Fatalf("unsupported layout variants should not be backfilled: %q %q", got.Tasks[0].LayoutVariant, got.Tasks[1].LayoutVariant)
	}
}

func TestConfiguredManifestToolDefersDraftValidationUntilCommit(t *testing.T) {
	workDir := t.TempDir()
	tool := newConfiguredManifestTool(workDir, filepath.Join(t.TempDir(), "skills"), nil, "中间草稿", false)
	args := `{
		"mode":"initialize",
		"title":"中间草稿",
		"theme":"ocean_soft",
		"template":"generic",
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

	if _, err := tool.InvokableRun(context.Background(), args); err != nil {
		t.Fatalf("draft write should defer hard validation: %v", err)
	}
	if _, err := ReadTasksDraftManifest(workDir); err != nil {
		t.Fatal(err)
	}
	result, err := tool.InvokableRun(context.Background(), `{"mode":"commit"}`)
	if err != nil {
		t.Fatalf("commit validation failure should be returned to planner, got tool error %v", err)
	}
	if !strings.Contains(result, `"ok":false`) || !strings.Contains(result, `unsupported type \"x_y_positioned_box\"`) {
		t.Fatalf("commit should return structured validation failure, got %s", result)
	}
	if _, statErr := os.Stat(filepath.Join(workDir, "tasks.json")); !os.IsNotExist(statErr) {
		t.Fatalf("invalid draft should not publish final tasks.json, stat err=%v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(workDir, "tasks.draft.json")); statErr != nil {
		t.Fatalf("invalid draft should remain for planner retry, stat err=%v", statErr)
	}
}

func TestConfiguredManifestToolNormalizesSectionComponentAliasesOnCommit(t *testing.T) {
	workDir := t.TempDir()
	tool := newConfiguredManifestTool(workDir, filepath.Join(t.TempDir(), "skills"), nil, "章节页", false)
	args := `{
		"mode":"initialize",
		"title":"章节页",
		"theme":"ocean_soft",
		"template":"generic",
		"tasks":[{
			"task_id":"1",
			"page_index":1,
			"title":"背景与目标",
			"content_type":"section_divider",
			"description":"章节分隔页",
			"output_file":"1_section.pptx",
			"status":"pending",
			"content_plan":{
				"summary":"进入背景章节",
				"components":[
					{"id":"section_title","type":"section_title","text":"背景与目标"},
					{"id":"subtitle_1","type":"subtitle","text":"以数字化服务串联赛事体验和运营效率"}
				],
				"capacity_hint":{"estimated_density":"sparse","overflow_risk":"low","component_count":2}
			}
		}]
	}`

	if _, err := tool.InvokableRun(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	if _, err := tool.InvokableRun(context.Background(), `{"mode":"commit"}`); err != nil {
		t.Fatal(err)
	}
	got, err := ReadTasksManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	componentType := got.Tasks[0].ContentPlan.Components[0].Type
	if componentType != "headline" {
		t.Fatalf("section_title alias = %q, want headline", componentType)
	}
	subtitleType := got.Tasks[0].ContentPlan.Components[1].Type
	if subtitleType != "subheadline" {
		t.Fatalf("subtitle alias = %q, want subheadline", subtitleType)
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
