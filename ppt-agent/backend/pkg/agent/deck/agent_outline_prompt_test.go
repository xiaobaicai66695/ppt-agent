package deck

import (
	"strings"
	"testing"
)

func TestPlannerPromptAlwaysInitializesCompleteDraft(t *testing.T) {
	outlines := []*TaskOutline{
		nil,
		{
			Template: "generic", Theme: "ocean_soft", Title: "主题",
			ContentMode: OutlineContentModeUserOutline,
			Slides:      []SlideOutline{{Title: "用户标题", ContentType: "content_slide", Description: "用户内容"}},
		},
	}
	for _, outline := range outlines {
		prompt := buildPlannerInstruction("/tmp/work", "/tmp/skills", "", outline, "用户主题", false)
		for _, want := range []string{
			"无论是否有用户大纲",
			"update_tasks_manifest(mode=\"initialize\")",
			"一次性提交完整页面数组",
			"不要逐页 patch",
			"tasks.draft.json",
		} {
			if !strings.Contains(prompt, want) {
				t.Fatalf("planner prompt missing %q", want)
			}
		}
		for _, forbidden := range []string{"review_tasks_manifest", "mode=\"commit\"", "Fixer", "failure_recovery.tmpl", "color_policy.tmpl"} {
			if strings.Contains(prompt, forbidden) {
				t.Fatalf("planner prompt crosses boundary with %q", forbidden)
			}
		}
	}
}

func TestPlannerPromptKeepsBackgroundPolicyInSkill(t *testing.T) {
	withoutTool := buildPlannerInstruction("/tmp/work", "/tmp/skills", "", nil, "生态报告", false)
	for _, want := range []string{"背景图片策略", "以 skill 为准", "visual_intent.asset_query"} {
		if !strings.Contains(withoutTool, want) {
			t.Fatalf("planner prompt missing skill boundary %q", want)
		}
	}
	if strings.Contains(withoutTool, "search_images(download=true)") {
		t.Fatal("planner should not advertise unavailable image search")
	}

	withTool := buildPlannerInstruction("/tmp/work", "/tmp/skills", "", nil, "生态报告", true)
	for _, want := range []string{"search_images(download=true)", "来源页和摄影师署名"} {
		if !strings.Contains(withTool, want) {
			t.Fatalf("configured planner prompt missing %q", want)
		}
	}
}

func TestReviewerAndFixerPromptsHaveSeparateScopes(t *testing.T) {
	cfg := &PPTTaskConfig{WorkDir: "/tmp/work", SkillsDir: "/tmp/skills", Query: "用户主题"}
	reviewer := buildReviewerInstruction(cfg)
	for _, want := range []string{"TaskPlanReviewer", "tasks.draft.json", "patch_tasks_draft", "所有必要修正合并为一次 patch"} {
		if !strings.Contains(reviewer, want) {
			t.Fatalf("reviewer prompt missing %q", want)
		}
	}
	if strings.Contains(reviewer, "PPTFixer") {
		t.Fatal("reviewer prompt should not include generated-deck fixer")
	}

	fixer := buildFixerInstruction(cfg)
	for _, want := range []string{"PPTFixer", "PPT 已生成后", "patch_selected_tasks", "不修改未授权页面"} {
		if !strings.Contains(fixer, want) {
			t.Fatalf("fixer prompt missing %q", want)
		}
	}
	if strings.Contains(fixer, "tasks.draft.json") {
		t.Fatal("fixer must only consume the committed tasks.json")
	}
}
