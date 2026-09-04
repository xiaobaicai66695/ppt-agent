package task

import "github.com/cloudwego/ppt-agent/pkg/agent/ppt"

// ReadPlan is the persistence boundary used by recovery code. It accepts the
// historical tasks.json/tasks.draft.json names through the ppt package.
func (tm *TaskManager) ReadPlan(id string) (*ppt.TasksManifest, error) {
	return tm.ReadTasksManifestFile(id)
}
