package task

import "github.com/cloudwego/ppt-agent/pkg/agent/ppt"

// ReadPlanFile returns the task's PPT plan. The old method is retained for
// callers compiled against the previous terminology.
func (tm *TaskManager) ReadPlanFile(id string) (*ppt.TasksManifest, error) {
	return tm.ReadTasksManifestFile(id)
}
