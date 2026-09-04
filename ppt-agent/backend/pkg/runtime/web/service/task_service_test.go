package service

import (
	"testing"

	"github.com/cloudwego/ppt-agent/pkg/agent/ppt"
	"github.com/cloudwego/ppt-agent/pkg/runtime/task"
)

func TestFindRecentDuplicateIgnoresFailedTasks(t *testing.T) {
	manager := task.NewTaskManager(t.TempDir(), nil, nil, nil)
	if _, err := manager.CreateConversationTask("task-1", "产品复盘 PPT", 7); err != nil {
		t.Fatal(err)
	}
	svc := NewTaskService(TaskServiceConfig{Tasks: manager})
	if got := svc.FindRecentDuplicate(7, " 产品复盘   PPT "); got == nil || got.ID != "task-1" {
		t.Fatalf("duplicate = %#v, want task-1", got)
	}

	state := manager.GetTaskState("task-1")
	state.Mu.Lock()
	state.Info.Status = task.TaskStatusFailed
	state.Mu.Unlock()
	if got := svc.FindRecentDuplicate(7, "产品复盘 PPT"); got != nil {
		t.Fatalf("failed task should not suppress a new request: %#v", got)
	}
}

func TestPrepareOutlineAppliesStableDefaults(t *testing.T) {
	svc := NewTaskService(TaskServiceConfig{})
	outline := &ppt.TaskOutline{Slides: []ppt.SlideOutline{{Title: "  标题  ", ContentType: "  content_slide  "}}}
	got, err := svc.PrepareOutline("  主题  ", outline)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "主题" || got.ContentMode != ppt.OutlineContentModeUserOutline || got.Slides[0].Title != "标题" || got.Slides[0].ContentType != "content_slide" {
		t.Fatalf("outline defaults = %#v", got)
	}
}
