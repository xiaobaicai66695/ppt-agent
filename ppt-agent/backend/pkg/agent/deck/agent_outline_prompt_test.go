package deck

import (
	"strings"
	"testing"
)

func TestMainAgentPromptDistinguishesTemplateScaffold(t *testing.T) {
	outline := &TaskOutline{
		Template: "generic", Theme: "ocean_soft", Title: "主题",
		ContentMode:   OutlineContentModeTemplateScaffold,
		UseBackground: false,
		Slides:        []SlideOutline{{Title: "模板示例", ContentType: "title_slide"}},
	}
	prompt := buildPlannerInstruction("/tmp/work", "/tmp/skills", "", outline, "用户主题", false, 5)
	for _, want := range []string{"模板脚手架", "逐页围绕用户主题重写模板示例", "现有页数、顺序和 `content_type`", "theme=ocean_soft"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
	for _, want := range []string{"update_tasks_manifest", "一次 `update_tasks_manifest", "页面完成状态由后端依据代码元数据维护"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing fast-loop contract %q", want)
		}
	}
	if strings.Contains(prompt, "edit_file(path=") {
		t.Fatal("prompt still instructs the agent to edit tasks.json with edit_file")
	}
}

func TestMainAgentPromptKeepsLegacyBackgroundEmpty(t *testing.T) {
	outline := &TaskOutline{
		Template: "personal-summary", Theme: "charcoal_light", Title: "主题",
		ContentMode:   OutlineContentModeTemplateScaffold,
		UseBackground: true, RecommendedBackground: "party_government",
		Slides: []SlideOutline{{Title: "封面", ContentType: "title_slide"}},
	}
	prompt := buildPlannerInstructionWithImageSearch("/tmp/work", "/tmp/skills", "", outline, "两会总结", false, 5, false)
	for _, want := range []string{"没有可用图片搜索时只保留 `visual_intent.asset_query`", "`task.background` 默认留空", "`theme=charcoal_light` 是整套配色锚点"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
	for _, forbidden := range []string{"整套每页使用 `party_government` 同一主题目录", "为每一页主动填写该主题或其具体图片引用"} {
		if strings.Contains(prompt, forbidden) {
			t.Fatalf("prompt should not force local background, found %q", forbidden)
		}
	}

	withImageSearch := buildPlannerInstructionWithImageSearch("/tmp/work", "/tmp/skills", "", outline, "两会总结", false, 5, true)
	for _, want := range []string{"外部图片搜索可用", "`task.background` 默认留空", "背景氛围图"} {
		if !strings.Contains(withImageSearch, want) {
			t.Fatalf("image-search prompt missing %q", want)
		}
	}
	for _, forbidden := range []string{"整套每页使用 `party_government` 同一主题目录", "Planner 为每一页主动填写该主题或其具体图片引用"} {
		if strings.Contains(withImageSearch, forbidden) {
			t.Fatalf("image-search prompt should not force local background, found %q", forbidden)
		}
	}
}

func TestMainAgentPromptPreservesPopulatedUserOutline(t *testing.T) {
	outline := &TaskOutline{
		Template: "custom", Theme: "ocean_soft", Title: "主题",
		ContentMode: OutlineContentModeUserOutline,
		Slides:      []SlideOutline{{Title: "用户标题", ContentType: "content_slide", Description: "用户内容"}},
	}
	prompt := buildPlannerInstruction("/tmp/work", "/tmp/skills", "", outline, "用户主题", false, 5)
	for _, want := range []string{"用户大纲", "非空的 `title`、`description`、`content_plan`", "补齐空字段"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
}

func TestMainAgentPromptUsesMetadataCompletionWithoutQAOrPromptedReActLogs(t *testing.T) {
	prompt := buildPlannerInstruction("/tmp/work", "/tmp/skills", "", nil, "介绍大兴安岭", false, 5)
	for _, want := range []string{
		"页面完成状态由后端依据代码元数据维护",
		"Planner commit 后立即结束",
		"validate_deck_spec → render_worker_pool(generate_slide(task_id)) → reconcile_delivery",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
	for _, forbidden := range []string{"Fixer", "QA 质检", "bash(command=", "文件落地确认", "task | 并行调用 SlideExecutor", "ReAct 风格可见规划日志", "Thought:", "Action:", "Observation:"} {
		if strings.Contains(prompt, forbidden) {
			t.Fatalf("prompt still contains removed stage/tool %q", forbidden)
		}
	}
	if !strings.Contains(prompt, "Plan Reviewer") {
		t.Fatal("prompt should include plan-review quality gate")
	}
}

func TestMainAgentPromptUsesRecommendedStyleWithoutTemplateSlides(t *testing.T) {
	outline := &TaskOutline{
		Template: "research-report", Theme: "report_green", Title: "主题",
		ContentMode: OutlineContentModeRecommendedStyle, SuggestedPageCount: 11,
		UseBackground: true, RecommendedBackground: "snowy_mountain",
	}
	prompt := buildPlannerInstructionWithImageSearch("/tmp/work", "/tmp/skills", "模型推荐", outline, "生态报告", false, 5, false)
	for _, want := range []string{"智能推荐提供视觉方向，不预设页面结构", "建议页数：`11`", "`snowy_mountain` 仅作为后续图片搜索线索", "重新设计叙事结构"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
	withImageSearch := buildPlannerInstructionWithImageSearch("/tmp/work", "/tmp/skills", "模型推荐", outline, "生态报告", false, 5, true)
	if !strings.Contains(withImageSearch, "转换为外部背景图检索意图，不写本地 `background`") {
		t.Fatal("image-search prompt should turn recommended background into external image planning")
	}
	for _, forbidden := range []string{"整套每页使用", "同一主题目录"} {
		if strings.Contains(prompt, forbidden) || strings.Contains(withImageSearch, forbidden) {
			t.Fatalf("prompt should not carry legacy background contract, found %q", forbidden)
		}
	}
}

func TestMainAgentPromptAdvertisesImageSearchOnlyWhenConfigured(t *testing.T) {
	withoutTool := buildPlannerInstructionWithImageSearch("/tmp/work", "/tmp/skills", "", nil, "生态报告", false, 5, false)
	if strings.Contains(withoutTool, "search_images") {
		t.Fatal("prompt should not advertise image search when the provider is unavailable")
	}

	withTool := buildPlannerInstructionWithImageSearch("/tmp/work", "/tmp/skills", "", nil, "生态报告", false, 5, true)
	for _, want := range []string{"search_images", "download=true", "来源页和摄影师署名"} {
		if !strings.Contains(withTool, want) {
			t.Fatalf("configured prompt missing %q", want)
		}
	}
}
