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
		"必须且只允许再执行一次完整 `initialize`",
		"用户要求/素材事实 → 页码 → title/content_type → 组件",
		"visual_policy",
		"确定性搜索、下载、去重",
		"Planner 不填写或虚构它们",
		"section_marker",
		"component_text_density",
		"page_content_density",
		"当前 `content_type.capacity`",
		"作为观点锚点",
		"只填写一个英文单词",
		"clean_text_only",
	} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("planner prompt missing %q", expected)
		}
	}
	for _, forbidden := range []string{"440–840", "70–130 字", "至少 240 字", "至少写到 260 个中文字符", "整页正文至少 300 字"} {
		if strings.Contains(prompt, forbidden) {
			t.Fatalf("planner prompt must delegate numeric content limits to the skill, found %q", forbidden)
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
	for _, expected := range []string{"不要在工具调用前输出", "不重复审查报告", "信息密度是阻断性质量门", "至少 240 个中文字符", "至少 220 个中文字符", "440–840", "必须恰好是一个英文单词"} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("reviewer prompt missing %q: %s", expected, prompt)
		}
	}
}
