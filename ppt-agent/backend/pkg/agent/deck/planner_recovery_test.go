package deck

import (
	"os"
	"path/filepath"
	"testing"

	agentintent "github.com/cloudwego/ppt-agent/pkg/agent/intent"
)

func TestRecoverMissingPlannerManifestFromThoughtOutput(t *testing.T) {
	workDir := t.TempDir()
	skillsDir := t.TempDir()
	bgDir := filepath.Join(skillsDir, "visual_designer", "background_templates", "eco_nature", "images")
	if err := os.MkdirAll(bgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"1.jpg", "2.jpg", "3.jpg", "4.jpg"} {
		if err := os.WriteFile(filepath.Join(bgDir, name), []byte("image"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	output := `{
"thought": "规划 8 页桂林介绍 PPT：\n1. 封面页 (title_slide)\n2. 目录页 (agenda)\n3. 章节1：山水桂林 (section_divider)\n4. 桂林概况 (image_text)\n5. 核心景点 (card_grid)\n6. 章节2：文化美食 (section_divider)\n7. 文化与美食 (two_column)\n8. 旅行贴士与总结 (summary_slide)\n\n背景主题：eco_nature（4张图轮换）\n配色：ocean_soft\n模板：product-intro"
}`
	cfg := &PPTTaskConfig{
		WorkDir:   workDir,
		SkillsDir: skillsDir,
		Query:     "介绍桂林",
		Outline: &TaskOutline{
			Template:              "product-intro",
			Theme:                 "sage_calm",
			ContentMode:           OutlineContentModeRecommendedStyle,
			UseBackground:         true,
			RecommendedBackground: "eco_nature",
			SuggestedPageCount:    8,
		},
		IntentResult: &agentintent.ClassificationResult{SuggestedPageCount: 8},
	}

	manifest, err := recoverMissingPlannerManifest(cfg, "介绍桂林", output)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Tasks) != 8 {
		t.Fatalf("slide count = %d", len(manifest.Tasks))
	}
	if manifest.Template != "product-intro" || manifest.Theme != "sage_calm" {
		t.Fatalf("unexpected manifest header: %#v", manifest)
	}
	if manifest.Tasks[2].ContentType != "section_divider" || manifest.Tasks[5].ContentType != "section_divider" {
		t.Fatalf("section pages not recovered: %#v %#v", manifest.Tasks[2], manifest.Tasks[5])
	}
	if manifest.Tasks[2].LayoutVariant == "" || manifest.Tasks[2].LayoutVariant != manifest.Tasks[5].LayoutVariant {
		t.Fatalf("section layout should be deck-wide: %q %q", manifest.Tasks[2].LayoutVariant, manifest.Tasks[5].LayoutVariant)
	}
	if manifest.Tasks[2].ContentPlan == nil || manifest.Tasks[2].ContentPlan.SectionNumber != "01" {
		t.Fatalf("first section number = %#v", manifest.Tasks[2].ContentPlan)
	}
	if manifest.Tasks[5].ContentPlan == nil || manifest.Tasks[5].ContentPlan.SectionNumber != "02" {
		t.Fatalf("second section number = %#v", manifest.Tasks[5].ContentPlan)
	}
	for i, task := range manifest.Tasks {
		if task.Background == "" || backgroundTheme(task.Background) != "eco_nature" {
			t.Fatalf("task %d background = %q", i+1, task.Background)
		}
	}
	if _, err := os.Stat(filepath.Join(workDir, "tasks.json")); err != nil {
		t.Fatalf("tasks.json not written: %v", err)
	}
}

func TestPlannerScratchThoughtIsNotEmittable(t *testing.T) {
	if !isPlannerScratchThought(`{"thought":"内部规划"}`) {
		t.Fatal("expected thought JSON to be classified as scratch")
	}
	if isPlannerScratchThought(`{"message":"用户可见内容"}`) {
		t.Fatal("ordinary JSON should remain visible")
	}
}
