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
		"同一 `content_type` 的背景关键词必须一致",
		"一个英文关键词",
		"image_left",
		"asset_purpose=\"scene\"",
		"section_marker",
		"440–840",
		"clean_text_only",
		"默认每页都必须先执行图片搜索",
		"不得把文本、图表或卡片可表达当成跳过图片搜索的理由",
		"搜索并下载外部背景",
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
