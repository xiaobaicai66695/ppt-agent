package task

import (
	"testing"
	"time"

	"github.com/cloudwego/ppt-agent/pkg/db"
)

func TestListAllTasksReturnsEveryInMemoryOwner(t *testing.T) {
	previousDB := db.DB
	db.DB = nil
	t.Cleanup(func() { db.DB = previousDB })

	tm := NewTaskManager(t.TempDir(), nil, nil, nil)
	tm.tasks["task-a"] = &TaskState{Info: TaskInfo{ID: "task-a", UserID: 1, CreatedAt: time.Now()}}
	tm.tasks["task-b"] = &TaskState{Info: TaskInfo{ID: "task-b", UserID: 2, CreatedAt: time.Now()}}

	got := tm.ListAllTasks()
	if len(got) != 2 {
		t.Fatalf("ListAllTasks() len = %d, want 2", len(got))
	}

	owners := map[int]bool{}
	for _, info := range got {
		owners[info.UserID] = true
	}
	if !owners[1] || !owners[2] {
		t.Fatalf("ListAllTasks() owners = %#v, want users 1 and 2", owners)
	}
}
