package task

import (
	"reflect"
	"testing"

	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
)

func TestCanonicalOutputFileHandlesPathForms(t *testing.T) {
	tests := map[string]string{
		"1_cover.pptx":                         "1_cover.pptx",
		"/srv/ppt/weboutput/task/1_cover.pptx": "1_cover.pptx",
		`C:\ppt\task\1_cover.pptx`:             "1_cover.pptx",
		"  ./slides/2_agenda.pptx  ":           "2_agenda.pptx",
	}
	for input, want := range tests {
		if got := CanonicalOutputFile(input); got != want {
			t.Errorf("CanonicalOutputFile(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestDeduplicateOutputFilesPreservesFirstSeenOrder(t *testing.T) {
	files := []string{
		"/srv/task/1_cover.pptx",
		"1_cover.pptx",
		`C:\task\2_agenda.pptx`,
		"./2_agenda.pptx",
	}
	want := []string{"1_cover.pptx", "2_agenda.pptx"}
	if got := DeduplicateOutputFiles(files); !reflect.DeepEqual(got, want) {
		t.Fatalf("DeduplicateOutputFiles() = %#v, want %#v", got, want)
	}
}

func TestManifestOutputFilesDeduplicatesLogicalPage(t *testing.T) {
	manifest := &deck.TasksManifest{Tasks: []*deck.TaskItem{
		{TaskID: "slide-1", PageIndex: 1, OutputFile: "1_cover.pptx", Status: deck.StatusDone},
		{TaskID: "legacy-copy", PageIndex: 1, OutputFile: "/srv/task/1_cover.pptx", Status: deck.StatusDone},
		{TaskID: "slide-2", PageIndex: 2, OutputFile: "2_agenda.pptx", Status: deck.StatusQADone},
		{TaskID: "slide-3", PageIndex: 3, OutputFile: "3_pending.pptx", Status: deck.StatusPending},
	}}
	want := []string{"1_cover.pptx", "2_agenda.pptx"}
	if got := ManifestOutputFiles(manifest); !reflect.DeepEqual(got, want) {
		t.Fatalf("ManifestOutputFiles() = %#v, want %#v", got, want)
	}
}
