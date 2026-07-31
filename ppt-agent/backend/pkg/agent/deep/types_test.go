package deep

import (
	"os"
	"path/filepath"
	"testing"
)

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
