package task

import "github.com/cloudwego/ppt-agent/pkg/agent/deck"

// ReadPlan is the persistence boundary used by recovery code. It accepts the
// historical tasks.json/tasks.draft.json names through the deck package.
func (tm *TaskManager) ReadPlan(id string) (*deck.TasksManifest, error) {
	return tm.ReadTasksManifestFile(id)
}
