package deck

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlannerManifestToolOnlyInitializesDraft(t *testing.T) {
	workDir := t.TempDir()
	planner := newPlannerManifestTool(workDir, nil, "组件化架构演进")

	result, err := planner.InvokableRun(context.Background(), `{"mode":"patch","tasks":[]}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "only supports initialize") {
		t.Fatalf("unexpected patch result: %s", result)
	}

	result, err = planner.InvokableRun(context.Background(), `{
		"mode":"initialize",
		"tasks":[{
			"page_index":1,"title":"架构演进","content_type":"title_slide",
			"content_plan":{
				"summary":"说明系统架构演进的核心判断",
				"slide_intent":"用核心判断建立整套演示的架构演进主线。",
				"visual_intent":{"asset_purpose":"background","asset_subject":"software architecture blueprint","asset_query":"architecture","composition":"wide landscape, clean negative space on left"},
				"components":[{"id":"point-1","type":"key_point","body":"架构演进的核心不是堆叠更多模块，而是持续收敛职责边界和交付路径。"}]
			}
		}]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"ok":true`) || !strings.Contains(result, `"target":"draft"`) {
		t.Fatalf("unexpected initialize result: %s", result)
	}
	manifest, err := ReadTasksDraftManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Title != "组件化架构演进" || len(manifest.Tasks) != 1 {
		t.Fatalf("unexpected draft: %#v", manifest)
	}
	if got := manifest.Tasks[0]; got.TaskID != "slide-1" || got.OutputFile != "slide_01.pptx" || got.Status != StatusPending {
		t.Fatalf("runtime fields were not derived: %#v", got)
	}
	if _, err := os.Stat(filepath.Join(workDir, "tasks.json")); !os.IsNotExist(err) {
		t.Fatalf("planner must not publish tasks.json: %v", err)
	}
}

func TestDraftPatchToolUpdatesExistingTaskWithoutPublishing(t *testing.T) {
	workDir := t.TempDir()
	manifest := &TasksManifest{
		Title: "初稿",
		Tasks: []*TaskItem{{
			TaskID: "1", PageIndex: 1, Title: "旧标题", ContentType: "content_slide",
			OutputFile: "1_content.pptx", Status: "pending",
		}},
	}
	if err := WriteTasksDraftManifest(workDir, manifest); err != nil {
		t.Fatal(err)
	}

	patcher := newDraftTasksPatchTool(workDir)
	result, err := patcher.InvokableRun(context.Background(), `{
		"tasks":[{"page_index":1,"title":"新标题"}]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(result), &payload); err != nil || payload["ok"] != true {
		t.Fatalf("unexpected patch result: %s", result)
	}
	updated, err := ReadTasksDraftManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Tasks[0].Title != "新标题" {
		t.Fatalf("patch was not applied: %#v", updated.Tasks[0])
	}
	if _, err := os.Stat(filepath.Join(workDir, "tasks.json")); !os.IsNotExist(err) {
		t.Fatalf("reviewer patch must not publish tasks.json: %v", err)
	}
}

func TestDraftPatchPreservesDownloadedVisualMetadata(t *testing.T) {
	workDir := t.TempDir()
	manifest := &TasksManifest{
		Title: "初稿",
		Tasks: []*TaskItem{{
			TaskID: "1", PageIndex: 1, Title: "封面", ContentType: "title_slide",
			OutputFile: "1_title.pptx", Status: "pending",
			ContentPlan: &ContentPlan{
				SlideIntent: "建立主题",
				VisualIntent: &VisualIntent{
					AssetPurpose: "background",
					AssetQuery:   "old query",
					LocalPath:    "assets/images/cover.jpg",
					SourceURL:    "https://unsplash.com/photos/cover",
					Attribution:  "Photo by Tester on Unsplash",
				},
				Components: []PlanComponent{{ID: "old", Type: "key_point", Body: "旧观点"}},
			},
		}},
	}
	if err := WriteTasksDraftManifest(workDir, manifest); err != nil {
		t.Fatal(err)
	}

	patcher := newDraftTasksPatchTool(workDir)
	result, err := patcher.InvokableRun(context.Background(), `{
		"tasks":[{"page_index":1,"content_plan":{
			"slide_intent":"强化主题判断",
			"visual_intent":{"asset_query":"new query","composition":"wide landscape"},
			"components":[{"id":"new","type":"insight","body":"新观点"}]
		}}]
	}`)
	if err != nil || !strings.Contains(result, `"ok":true`) {
		t.Fatalf("unexpected patch result: %s err=%v", result, err)
	}
	updated, err := ReadTasksDraftManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	visual := updated.Tasks[0].ContentPlan.VisualIntent
	if visual.AssetQuery != "new query" || visual.Composition != "wide landscape" {
		t.Fatalf("reviewer visual patch was not applied: %#v", visual)
	}
	if visual.LocalPath != "assets/images/cover.jpg" || visual.SourceURL == "" || visual.Attribution == "" {
		t.Fatalf("downloaded visual metadata was lost: %#v", visual)
	}
	if components := updated.Tasks[0].ContentPlan.Components; len(components) != 1 || components[0].ID != "new" {
		t.Fatalf("component replacement was not applied: %#v", components)
	}
}

func TestPlannerManifestToolRejectsStringSpillAndDeprecatedFields(t *testing.T) {
	workDir := t.TempDir()
	planner := newPlannerManifestTool(workDir, nil, "架构清理验证")

	result, err := planner.InvokableRun(context.Background(), `{
		"mode":"initialize",
		"title":"架构清理验证",
		"theme":"simple_gray",
		"tasks":"model draft: [{\"task_id\":\"1\",\"page_index\":1,\"title\":\"架构清理验证\",\"content_type\":\"content_slide\",\"description\":\"说明固定模板和旧自由工具已清理，当前链路以动态 DeckSpec 和确定性渲染为核心。\",\"output_file\":\"1_architecture_cleanup.pptx\",\"status\":\"pending\",\"content_plan\":{\"slide_intent\":\"确认架构清理范围和验证结果。\",\"visual_intent\":{\"asset_purpose\":\"background\",\"asset_query\":\"architecture\",\"asset_subject\":\"software architecture blueprint\",\"composition\":\"wide landscape clean negative space\"},\"components\":[{\"id\":\"p1\",\"type\":\"key_point\",\"body\":\"Planner 生成组件级计划，Go 质量门负责提交，Python generator 负责确定性渲染。\"}]}}], \"template\":\"legacy\""
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"ok":false`) || !strings.Contains(result, "theme is not part of the DeckSpec contract") {
		t.Fatalf("deprecated fields and string spill should be rejected: %s", result)
	}
}

func TestPlannerManifestToolRejectsDeprecatedThemeTemplateAndRuntimeFields(t *testing.T) {
	workDir := t.TempDir()
	planner := newPlannerManifestTool(workDir, nil, "AI 产业趋势")

	result, err := planner.InvokableRun(context.Background(), `{
		"mode":"initialize",
		"theme":"government_red",
		"template":"generic",
		"tasks":[{
			"task_id":"1","page_index":1,"title":"AI 产业趋势","content_type":"title_slide",
			"description":"建立整套演示的核心判断和视觉方向。","output_file":"1_title.pptx","status":"pending",
			"content_plan":{
				"slide_intent":"用封面建立 AI 产业趋势分析的主线。",
				"visual_intent":{"asset_purpose":"background","asset_query":"technology","asset_subject":"AI data center","composition":"wide landscape clean negative space"},
				"components":[{"id":"point-1","type":"key_point","body":"AI 产业趋势分析需要同时覆盖技术成熟度、商业化路径和组织采纳条件。"}]
			}
		}]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"ok":false`) || !strings.Contains(result, "theme is not part of the DeckSpec contract") {
		t.Fatalf("deprecated fields should be rejected, got: %s", result)
	}
}

func TestPlannerManifestToolNormalizesAgendaBeforePreflight(t *testing.T) {
	workDir := t.TempDir()
	planner := newPlannerManifestTool(workDir, nil, "国际局势分析")

	result, err := planner.InvokableRun(context.Background(), `{
		"mode":"initialize","title":"国际局势分析","tasks":[{
			"page_index":2,"title":"目录","content_type":"agenda",
			"content_plan":{
				"summary":"围绕冲突、安全、经济和治理四条主线组织内容。",
				"slide_intent":"帮助观众先理解整套报告的章节顺序和判断框架。",
				"visual_intent":{"asset_purpose":"background","asset_query":"diplomacy","asset_subject":"global diplomacy","composition":"wide landscape clean negative space"},
				"components":[
					{"id":"toc_1","type":"toc_item","title":"冲突热点"},
					{"id":"toc_2","type":"toc_item","title":"大国关系"},
					{"id":"toc_3","type":"toc_item","title":"区域安全"},
					{"id":"toc_4","type":"toc_item","title":"能源粮食"},
					{"id":"toc_5","type":"toc_item","title":"全球治理"},
					{"id":"toc_6","type":"toc_item","title":"产业链重组"},
					{"id":"toc_7","type":"toc_item","title":"技术竞争"},
					{"id":"toc_8","type":"toc_item","title":"风险展望"},
					{"id":"toc_9","type":"toc_item","title":"行动建议"}
				]
			}
		}]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"ok":true`) {
		t.Fatalf("agenda should be normalized before preflight, got: %s", result)
	}
	manifest, err := ReadTasksDraftManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	components := manifest.Tasks[0].ContentPlan.Components
	if len(components) > maxComponentsForContentType("agenda") {
		t.Fatalf("agenda components were not compacted: %d", len(components))
	}
	if components[0].Type != "insight" {
		t.Fatalf("agenda should keep/add a narrative anchor first, got %#v", components[0])
	}
}

func TestPlannerManifestToolNormalizesStatSlideAnchorBeforePreflight(t *testing.T) {
	workDir := t.TempDir()
	planner := newPlannerManifestTool(workDir, nil, "国际局势分析")

	result, err := planner.InvokableRun(context.Background(), `{
		"mode":"initialize","title":"国际局势分析","tasks":[{
			"page_index":4,"title":"冲突数量上升","content_type":"stat_slide",
			"content_plan":{
				"summary":"关键数字说明安全风险没有随外交斡旋自然降温。",
				"slide_intent":"用指标建立后续冲突分析的事实基线。",
				"visual_intent":{"asset_purpose":"background","asset_query":"diplomacy","asset_subject":"global diplomacy","composition":"wide landscape clean negative space"},
				"components":[
					{"id":"stat_1","type":"stat","title":"冲突热点","body":"多地延续"},
					{"id":"stat_2","type":"stat","title":"安全风险","body":"保持高位"}
				]
			}
		}]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"ok":true`) {
		t.Fatalf("stat slide should receive an insight before preflight, got: %s", result)
	}
	manifest, err := ReadTasksDraftManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if !hasNarrativeAnchor(manifest.Tasks[0].ContentPlan.Components) {
		t.Fatalf("stat slide still lacks narrative anchor: %#v", manifest.Tasks[0].ContentPlan.Components)
	}
}

func TestPlannerManifestToolNormalizesShortArgumentBlockWhenLongArgumentIsNotRequired(t *testing.T) {
	workDir := t.TempDir()
	planner := newPlannerManifestTool(workDir, nil, "国际局势分析")

	result, err := planner.InvokableRun(context.Background(), `{
		"mode":"initialize","title":"国际局势分析","tasks":[{
			"page_index":5,"title":"风险判断","content_type":"content_slide",
			"content_plan":{
				"summary":"风险不是单点爆发，而是多条压力线同时抬升。",
				"slide_intent":"形成一页可带走的核心判断。",
				"visual_intent":{"asset_purpose":"background","asset_query":"diplomacy","asset_subject":"global diplomacy","composition":"wide landscape clean negative space"},
				"components":[{"id":"arg_1","type":"argument_block","body":"多条压力线同时抬升，局势判断需要同时看冲突、能源、供应链和治理机制。"}]
			}
		}]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"ok":true`) {
		t.Fatalf("short argument should be normalized for content_slide, got: %s", result)
	}
	manifest, err := ReadTasksDraftManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if got := manifest.Tasks[0].ContentPlan.Components[0].Type; got != "insight" {
		t.Fatalf("short argument_block type = %q, want insight", got)
	}
}

func TestPlannerManifestToolWritesDraftWhenPreflightFindsQualityIssues(t *testing.T) {
	workDir := t.TempDir()
	planner := newPlannerManifestTool(workDir, nil, "AI 产业趋势")

	result, err := planner.InvokableRun(context.Background(), `{
		"mode":"initialize","title":"AI 产业趋势","tasks":[{
			"page_index":1,"title":"产业阶段","content_type":"section_divider",
			"content_plan":{"slide_intent":"切换到产业阶段章节。","components":[{"id":"marker-1","type":"section_marker"}]}
		}]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"quality_gate_passed":false`) ||
		!strings.Contains(result, `"invalid_component_schema"`) ||
		!strings.Contains(result, `"missing_background_image"`) {
		t.Fatalf("unexpected preflight result: %s", result)
	}
	manifest, err := ReadTasksDraftManifest(workDir)
	if err != nil {
		t.Fatalf("planner preflight issues should still leave a reviewable draft: %v", err)
	}
	if len(manifest.Tasks) != 1 || manifest.Tasks[0].TaskID != "slide-1" {
		t.Fatalf("unexpected draft after preflight issues: %#v", manifest)
	}
}
