package deck

import "fmt"

// ReadPlan reads the current PPT plan. The tasks.json filename remains the
// storage contract; this name makes the runtime intent clear to new callers.
func ReadPlan(workDir string) (*TasksManifest, error) { return ReadTasksManifest(workDir) }

// WritePlan persists the current PPT plan using the existing tasks.json file.
func WritePlan(workDir string, plan *TasksManifest) error { return WriteTasksManifest(workDir, plan) }

// ReadPlanDraft reads the resumable draft while accepting historical names.
func ReadPlanDraft(workDir string) (*TasksManifest, error) { return ReadTasksDraftManifest(workDir) }

// WritePlanDraft writes a resumable plan draft.
func WritePlanDraft(workDir string, plan *TasksManifest) error {
	return WriteTasksDraftManifest(workDir, plan)
}

// SyncPlanOutputFiles refreshes page completion metadata from generated files.
func SyncPlanOutputFiles(workDir string) (*TasksManifest, error) {
	plan, err := ReconcileTasksManifestOutputFiles(workDir)
	if err != nil {
		return nil, fmt.Errorf("sync plan output files: %w", err)
	}
	return plan, nil
}
