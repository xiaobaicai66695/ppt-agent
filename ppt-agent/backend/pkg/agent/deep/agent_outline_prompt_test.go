package deep

import (
	"strings"
	"testing"
)

func TestMainAgentPromptDistinguishesTemplateScaffold(t *testing.T) {
	outline := &TaskOutline{
		Template: "generic", Theme: "ocean_soft", Title: "主题",
		ContentMode: OutlineContentModeTemplateScaffold,
		Slides:      []SlideOutline{{Title: "模板示例", ContentType: "title_slide"}},
	}
	prompt := buildDeepAgentInstruction("/tmp/work", "/tmp/skills", "", outline, "用户主题", false, 5)
	for _, want := range []string{"模板脚手架模式", "重写 title、description", "禁止再启动独立的“填充模板”阶段"} {
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
