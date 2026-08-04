package deep

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestManifestToolPatchesMultipleTasksAtomically(t *testing.T) {
	workDir := t.TempDir()
	manifest := &TasksManifest{
		Title: "示例", Theme: "charcoal_light", Template: "personal-summary",
		Tasks: []*TaskItem{
			{TaskID: "1", PageIndex: 1, Title: "旧标题", ContentType: "title_slide", Description: "旧描述", OutputFile: "1_cover.pptx", Status: StatusPending},
			{TaskID: "2", PageIndex: 2, Title: "旧目录", ContentType: "agenda", Description: "旧描述", OutputFile: "2_agenda.pptx", Status: StatusPending},
		},
	}
	if err := WriteTasksManifest(workDir, manifest); err != nil {
		t.Fatal(err)
	}

	tool := newManifestTool(workDir)
	result, err := tool.InvokableRun(context.Background(), `{
		"mode":"patch",
		"tasks":[
			{"task_id":"1","title":"两会工作报告","description":"概括年度目标","content_plan":{"summary":"年度目标","elements":[{"type":"bullet_list","items":["目标一","目标二"]}]}},
			{"task_id":"2","title":"核心议程","description":"列出四项议程"}
		]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if result == "" {
		t.Fatal("expected structured result")
	}
	got, err := ReadTasksManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Tasks[0].Title != "两会工作报告" || got.Tasks[1].Title != "核心议程" {
		t.Fatalf("patch not applied: %#v", got.Tasks)
	}
	if got.Tasks[0].ContentPlan == nil || len(got.Tasks[0].ContentPlan.Elements) != 1 {
		t.Fatalf("content plan missing: %#v", got.Tasks[0].ContentPlan)
	}
	if got.Tasks[0].Status != StatusPending || got.Tasks[1].OutputFile != "2_agenda.pptx" {
		t.Fatalf("runtime fields were not preserved: %#v", got.Tasks)
	}
}

func TestManifestToolLeavesFileUnchangedOnInvalidPatch(t *testing.T) {
	workDir := t.TempDir()
	manifest := &TasksManifest{Title: "示例", Tasks: []*TaskItem{{
		TaskID: "1", PageIndex: 1, Title: "封面", ContentType: "title_slide",
		Description: "描述", OutputFile: "1_cover.pptx", Status: StatusPending,
	}}}
	if err := WriteTasksManifest(workDir, manifest); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(workDir, "tasks.json")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	tool := newManifestTool(workDir)
	if _, err := tool.InvokableRun(context.Background(), `{"mode":"patch","tasks":[{"task_id":"missing","title":"不应写入"}]}`); err == nil {
		t.Fatal("expected missing task error")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("manifest changed after invalid patch\nbefore=%s\nafter=%s", before, after)
	}
}
