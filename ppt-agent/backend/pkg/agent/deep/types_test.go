package deep

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestTaskOutlineUnmarshalNormalizesObjectItems(t *testing.T) {
	var outline TaskOutline
	err := json.Unmarshal([]byte(`{
		"template":"custom","theme":"ocean_soft","title":"测试",
		"slides":[{"title":"数据","content_type":"content_slide","description":"",
		"content_plan":{"elements":[{"type":"bullet_list","items":[{"title":"增长","description":"20%"},"稳定"]}]}}]
	}`), &outline)
	if err != nil {
		t.Fatalf("unmarshal outline: %v", err)
	}
	items := outline.Slides[0].ContentPlan.Elements[0].Items
	if len(items) != 2 || items[0] != "增长: 20%" || items[1] != "稳定" {
		t.Fatalf("items = %#v", items)
	}
}

func TestWriteTasksManifestKeepsExplicitRuntimeStatus(t *testing.T) {
	workDir := t.TempDir()

	initial := &TasksManifest{Tasks: []*TaskItem{{
		TaskID:      "slide-1",
		PageIndex:   1,
		Title:       "Initial",
		ContentType: "content_slide",
		Description: "initial",
		OutputFile:  "slide-1.pptx",
		Status:      StatusPending,
		CreatedAt:   "2026-07-31T10:00:00Z",
	}}}
	if err := WriteTasksManifest(workDir, initial); err != nil {
		t.Fatalf("write initial manifest: %v", err)
	}

	updated := &TasksManifest{Tasks: []*TaskItem{{
		TaskID:      "slide-1",
		PageIndex:   1,
		Title:       "Updated",
		ContentType: "content_slide",
		Description: "updated",
		OutputFile:  "slide-1.pptx",
		Status:      StatusDone,
	}}}
	if err := WriteTasksManifest(workDir, updated); err != nil {
		t.Fatalf("write updated manifest: %v", err)
	}

	got, err := ReadTasksManifest(workDir)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	item := got.GetTask("slide-1")
	if item == nil {
		t.Fatal("task missing")
	}
	if item.Status != StatusDone {
		t.Fatalf("status = %q, want %q", item.Status, StatusDone)
	}
	if item.CreatedAt != "2026-07-31T10:00:00Z" {
		t.Fatalf("created_at = %q, want original timestamp", item.CreatedAt)
	}
}

func TestReconcileTasksManifestOutputFiles(t *testing.T) {
	workDir := t.TempDir()

	manifest := &TasksManifest{Tasks: []*TaskItem{{
		TaskID:      "slide-1",
		PageIndex:   1,
		Title:       "Slide 1",
		ContentType: "content_slide",
		OutputFile:  "slide-1.pptx",
		Status:      StatusPending,
	}}}
	if err := WriteTasksManifest(workDir, manifest); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workDir, "slide-1.pptx"), []byte("pptx"), 0644); err != nil {
		t.Fatalf("write output file: %v", err)
	}

	got, err := ReconcileTasksManifestOutputFiles(workDir)
	if err != nil {
		t.Fatalf("reconcile manifest: %v", err)
	}
	item := got.GetTask("slide-1")
	if item == nil {
		t.Fatal("task missing")
	}
	if item.Status != StatusDone {
		t.Fatalf("status = %q, want %q", item.Status, StatusDone)
	}
}

func TestReconcileTasksManifestOutputFilesNormalizesUniquePageArtifact(t *testing.T) {
	workDir := t.TempDir()
	manifest := &TasksManifest{Tasks: []*TaskItem{{
		TaskID: "slide-6", PageIndex: 6, Title: "2025年经济社会发展成就",
		ContentType: "content_slide", OutputFile: "6_2025年经济社会发展成就.pptx", Status: StatusDone,
	}}}
	if err := WriteTasksManifest(workDir, manifest); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	sourceFile := "6_2025 年经济社会发展成就.pptx"
	if err := os.WriteFile(filepath.Join(workDir, sourceFile), []byte("pptx"), 0644); err != nil {
		t.Fatalf("write drifted output: %v", err)
	}
	qaDir := filepath.Join(workDir, "qa_images")
	if err := os.MkdirAll(qaDir, 0755); err != nil {
		t.Fatalf("mkdir qa_images: %v", err)
	}
	if err := os.WriteFile(filepath.Join(qaDir, "6_2025 年经济社会发展成就.jpg"), []byte("jpg"), 0644); err != nil {
		t.Fatalf("write drifted thumbnail: %v", err)
	}

	got, err := ReconcileTasksManifestOutputFiles(workDir)
	if err != nil {
		t.Fatalf("reconcile manifest: %v", err)
	}
	if got.GetTask("slide-6").Status != StatusDone {
		t.Fatalf("status = %q, want %q", got.GetTask("slide-6").Status, StatusDone)
	}
	for _, expected := range []string{
		"6_2025年经济社会发展成就.pptx",
		filepath.Join("qa_images", "6_2025年经济社会发展成就.jpg"),
	} {
		if _, err := os.Stat(filepath.Join(workDir, expected)); err != nil {
			t.Fatalf("normalized artifact %q missing: %v", expected, err)
		}
	}
	if _, err := os.Stat(filepath.Join(workDir, sourceFile)); !os.IsNotExist(err) {
		t.Fatalf("drifted source still exists or stat failed unexpectedly: %v", err)
	}
}

func TestReconcileTasksManifestOutputFilesRejectsAmbiguousPageArtifacts(t *testing.T) {
	workDir := t.TempDir()
	manifest := &TasksManifest{Tasks: []*TaskItem{{
		TaskID: "slide-9", PageIndex: 9, Title: "目标",
		ContentType: "content_slide", OutputFile: "9_2026年主要预期目标.pptx", Status: StatusPending,
	}}}
	if err := WriteTasksManifest(workDir, manifest); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	for _, name := range []string{"9_候选一.pptx", "9_候选二.pptx"} {
		if err := os.WriteFile(filepath.Join(workDir, name), []byte("pptx"), 0644); err != nil {
			t.Fatalf("write candidate %q: %v", name, err)
		}
	}

	got, err := ReconcileTasksManifestOutputFiles(workDir)
	if err != nil {
		t.Fatalf("reconcile manifest: %v", err)
	}
	if got.GetTask("slide-9").Status != StatusPending {
		t.Fatalf("status = %q, want %q", got.GetTask("slide-9").Status, StatusPending)
	}
	if _, err := os.Stat(filepath.Join(workDir, "9_2026年主要预期目标.pptx")); !os.IsNotExist(err) {
		t.Fatalf("ambiguous target should not be created: %v", err)
	}
}

func TestValidateTasksManifestOutcomeRejectsEmptyManifest(t *testing.T) {
	workDir := t.TempDir()
	if err := WriteTasksManifest(workDir, &TasksManifest{Tasks: []*TaskItem{}}); err != nil {
		t.Fatalf("write empty manifest: %v", err)
	}

	report, err := ValidateTasksManifestOutcome(workDir)
	if err != nil {
		t.Fatalf("validate empty manifest: %v", err)
	}
	if !report.Invalid || report.Total != 0 {
		t.Fatalf("report = %#v, want invalid empty delivery", report)
	}
}
