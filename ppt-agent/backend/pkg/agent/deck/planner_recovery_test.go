package deck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecoverMissingPlannerManifestFromThoughtOutput(t *testing.T) {
	workDir := t.TempDir()
	skillsDir := t.TempDir()

	output := `{
"thought": "规划 8 页桂林介绍 PPT：\n1. 封面页 (title_slide)\n2. 目录页 (agenda)\n3. 章节1：山水桂林 (section_divider)\n4. 桂林概况 (image_text)\n5. 核心景点 (card_grid)\n6. 章节2：文化美食 (section_divider)\n7. 文化与美食 (content_slide)\n8. 旅行贴士与总结 (content_slide)\n\n背景素材：landscape（同类型页面复用）"
}`
	cfg := &PPTTaskConfig{
		WorkDir:   workDir,
		SkillsDir: skillsDir,
		Query:     "介绍桂林",
		Outline: &TaskOutline{
			ContentMode: OutlineContentModeUserOutline,
		},
	}

	manifest, err := recoverMissingPlannerManifest(cfg, "介绍桂林", output)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Tasks) != 8 {
		t.Fatalf("slide count = %d", len(manifest.Tasks))
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
	if _, err := os.Stat(filepath.Join(workDir, tasksDraftFileName)); err != nil {
		t.Fatalf("tasks.draft.json not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, "tasks.json")); !os.IsNotExist(err) {
		t.Fatalf("recovery must not bypass review by publishing tasks.json: %v", err)
	}
}

func TestPlannerScratchThoughtFormatsForUser(t *testing.T) {
	if !isPlannerScratchThought(`{"thought":"内部规划"}`) {
		t.Fatal("expected thought JSON to be classified as scratch")
	}
	if isPlannerScratchThought(`{"message":"用户可见内容"}`) {
		t.Fatal("ordinary JSON should remain visible")
	}
	visible := plannerVisibleThought(`{"thought":"规划 3 页延安介绍 PPT：\n1. 封面页 (title_slide)\n2. 革命历史 (image_text)\n3. 总结 (content_slide)\n\n背景素材：historical site"}`)
	if visible == "" || visible == `{"thought":"内部规划"}` {
		t.Fatalf("thought should be formatted for the user, got %q", visible)
	}
	for _, want := range []string{"规划草案", "1. 封面页 (title_slide)", "- 背景素材：historical site", "正在写入 DeckSpec"} {
		if !strings.Contains(visible, want) {
			t.Fatalf("visible thought missing %q: %s", want, visible)
		}
	}
}
