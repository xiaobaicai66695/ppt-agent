package deck

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlannerManifestToolOnlyInitializesDraft(t *testing.T) {
	workDir := t.TempDir()
	planner := newPlannerManifestTool(workDir, nil, "组件化架构演进")

	result, err := planner.InvokableRun(context.Background(), `{"mode":"patch","tasks":[]}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "only supports initialize") {
		t.Fatalf("unexpected patch result: %s", result)
	}

	result, err = planner.InvokableRun(context.Background(), `{
		"mode":"initialize",
		"theme":"charcoal_light",
		"tasks":[{
			"task_id":"1","page_index":1,"title":"架构演进","content_type":"title_slide",
			"description":"说明系统架构演进的核心判断","output_file":"1_architecture.pptx","status":"pending"
		}]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"ok":true`) || !strings.Contains(result, `"target":"draft"`) {
		t.Fatalf("unexpected initialize result: %s", result)
	}
	manifest, err := ReadTasksDraftManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Title != "组件化架构演进" || len(manifest.Tasks) != 1 {
		t.Fatalf("unexpected draft: %#v", manifest)
	}
	if _, err := os.Stat(filepath.Join(workDir, "tasks.json")); !os.IsNotExist(err) {
		t.Fatalf("planner must not publish tasks.json: %v", err)
	}
}

func TestDraftPatchToolUpdatesExistingTaskWithoutPublishing(t *testing.T) {
	workDir := t.TempDir()
	manifest := &TasksManifest{
		Title: "初稿", Theme: "simple_gray",
		Tasks: []*TaskItem{{
			TaskID: "1", PageIndex: 1, Title: "旧标题", ContentType: "content_slide",
			Description: "旧内容", OutputFile: "1_content.pptx", Status: "pending",
		}},
	}
	if err := WriteTasksDraftManifest(workDir, manifest); err != nil {
		t.Fatal(err)
	}

	patcher := newDraftTasksPatchTool(workDir)
	result, err := patcher.InvokableRun(context.Background(), `{
		"tasks":[{"task_id":"1","title":"新标题","description":"新的结构化内容"}]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(result), &payload); err != nil || payload["ok"] != true {
		t.Fatalf("unexpected patch result: %s", result)
	}
	updated, err := ReadTasksDraftManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Tasks[0].Title != "新标题" || updated.Tasks[0].Description != "新的结构化内容" {
		t.Fatalf("patch was not applied: %#v", updated.Tasks[0])
	}
	if _, err := os.Stat(filepath.Join(workDir, "tasks.json")); !os.IsNotExist(err) {
		t.Fatalf("reviewer patch must not publish tasks.json: %v", err)
	}
}

func TestPlannerManifestToolRecoversTasksArrayFromStringSpill(t *testing.T) {
	workDir := t.TempDir()
	planner := newPlannerManifestTool(workDir, nil, "架构清理验证")

	result, err := planner.InvokableRun(context.Background(), `{
		"mode":"initialize",
		"title":"架构清理验证",
		"theme":"simple_gray",
		"tasks":"model draft: [{\"task_id\":\"1\",\"page_index\":1,\"title\":\"架构清理验证\",\"content_type\":\"content_slide\",\"description\":\"说明固定模板和旧自由工具已清理，当前链路以动态 DeckSpec 和确定性渲染为核心。\",\"output_file\":\"1_architecture_cleanup.pptx\",\"status\":\"pending\",\"content_plan\":{\"slide_intent\":\"确认架构清理范围和验证结果。\",\"components\":[{\"id\":\"p1\",\"type\":\"key_point\",\"body\":\"Planner 生成组件级计划，Go 质量门负责提交，Python generator 负责确定性渲染。\"}]}}], \"template\":\"legacy\""
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"ok":true`) {
		t.Fatalf("unexpected initialize result: %s", result)
	}
	manifest, err := ReadTasksDraftManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Tasks) != 1 || manifest.Tasks[0].Title != "架构清理验证" {
		t.Fatalf("embedded tasks array was not recovered: %#v", manifest)
	}
}
