package prompts

import (
	"strings"
	"testing"
)

func TestPlannerPromptCarriesFirstDraftQualityGate(t *testing.T) {
	prompt, err := RenderPlanner("master_instruction", &TemplateData{
		SkillsDir:            "/tmp/skills",
		TasksJSON:            "/tmp/tasks.draft.json",
		OutlineQuery:         "AI 产业趋势",
		ImageSearchAvailable: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"首次草稿质量门",
		"quality_gate_passed=false",
		"asset_purpose=\"background\"",
		"背景 `asset_query` 最多 2 个",
		"image_left",
		"asset_purpose=\"scene\"",
		"section_marker",
		"440–840",
		"clean_text_only",
	} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("planner prompt missing %q", expected)
		}
	}
}

func TestReviewerPromptSuppressesVerbosePreToolAnalysis(t *testing.T) {
	prompt, err := RenderReviewer("master_instruction", &TemplateData{
		SkillsDir:    "/tmp/skills",
		TasksJSON:    "/tmp/tasks.draft.json",
		OutlineQuery: "AI 产业趋势",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(prompt, "不要在工具调用前输出") || !strings.Contains(prompt, "不重复审查报告") {
		t.Fatalf("reviewer prompt does not constrain visible analysis: %s", prompt)
	}
}
