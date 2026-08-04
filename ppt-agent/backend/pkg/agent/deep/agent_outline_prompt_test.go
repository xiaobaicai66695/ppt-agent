package deep

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
	prompt := buildDeepAgentInstruction("/tmp/work", "/tmp/skills", "", outline, "用户主题", false, 5)
	for _, want := range []string{"模板脚手架模式", "重写 title、description", "禁止再启动独立的“填充模板”阶段", "不自动添加背景", "原本为空的 background 必须继续为空", "theme=ocean_soft"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
	for _, want := range []string{"update_tasks_manifest", "一次批量提交", "页面完成状态由后端根据输出文件自动更新"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing fast-loop contract %q", want)
		}
	}
	if strings.Contains(prompt, "edit_file(path=") {
		t.Fatal("prompt still instructs the agent to edit tasks.json with edit_file")
	}
}

func TestMainAgentPromptUsesOnlyRecommendedBackground(t *testing.T) {
	outline := &TaskOutline{
		Template: "personal-summary", Theme: "charcoal_light", Title: "主题",
		ContentMode:   OutlineContentModeTemplateScaffold,
		UseBackground: true, RecommendedBackground: "party_government",
		Slides: []SlideOutline{{Title: "封面", ContentType: "title_slide"}},
	}
	prompt := buildDeepAgentInstruction("/tmp/work", "/tmp/skills", "", outline, "两会总结", false, 5)
	for _, want := range []string{"仅使用 `party_government`", "只允许将该值补到", "不得选择其他背景", "不得根据 `party_government` 改写 `theme=charcoal_light`"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
}

func TestMainAgentPromptPreservesPopulatedUserOutline(t *testing.T) {
	outline := &TaskOutline{
		Template: "custom", Theme: "ocean_soft", Title: "主题",
		ContentMode: OutlineContentModeUserOutline,
		Slides:      []SlideOutline{{Title: "用户标题", ContentType: "content_slide", Description: "用户内容"}},
	}
	prompt := buildDeepAgentInstruction("/tmp/work", "/tmp/skills", "", outline, "用户主题", false, 5)
	for _, want := range []string{"用户大纲模式", "非空的 title、description、content_plan", "只对空字段"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
}

func TestMainAgentPromptUsesMetadataCompletionWithoutQAOrDiskVerification(t *testing.T) {
	prompt := buildDeepAgentInstruction("/tmp/work", "/tmp/skills", "", nil, "介绍大兴安岭", false, 5)
	for _, want := range []string{
		"页面交付状态由后端代码维护",
		"不要使用任何工具或额外模型轮次验证文件",
		"任务管理器会在元数据达到 N/N 时自动完成交付",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
	for _, forbidden := range []string{"Reviewer", "Fixer", "QA 质检", "bash(command=", "文件落地确认"} {
		if strings.Contains(prompt, forbidden) {
			t.Fatalf("prompt still contains removed stage/tool %q", forbidden)
		}
	}
}
