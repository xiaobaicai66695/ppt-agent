package prompts

import (
	"strings"
	"testing"
)

func TestSlideExecutorPromptLocksManifestThemeAndOutputFile(t *testing.T) {
	prompt, err := RenderDeepAgent("slide_executor_instruction", &TemplateData{
		WorkDir: "/tmp/work", SkillsDir: "/tmp/skills",
	})
	if err != nil {
		t.Fatalf("render slide executor prompt: %v", err)
	}
	for _, want := range []string{
		`deck_theme = manifest["theme"]`,
		`task["output_file"]`,
		"background 为空时必须保持为空",
		"禁止按 title 重新拼接文件名",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
	if strings.Contains(prompt, "from generators import get_palette_for_background") || strings.Contains(prompt, "palette = get_palette_for_background") {
		t.Fatal("prompt still demonstrates background palette replacement")
	}
}

func TestSlideExecutorContinuePromptKeepsStyleContract(t *testing.T) {
	prompt, err := RenderSlideExecutorContinueInstruction(&TemplateData{
		WorkDir: "/tmp/work", SkillsDir: "/tmp/skills",
	})
	if err != nil {
		t.Fatalf("render continuation prompt: %v", err)
	}
	for _, want := range []string{
		`deck_theme = manifest["theme"]`,
		`script_dir / task["output_file"]`,
		"为空时必须保持为空",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
	if strings.Contains(prompt, "from generators import get_palette_for_background") || strings.Contains(prompt, "palette = get_palette_for_background") {
		t.Fatal("continuation prompt still demonstrates background palette replacement")
	}
}
