package task

import (
	"errors"
	"fmt"

	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
)

var errDeliveryMetadataComplete = errors.New("delivery metadata complete")

// DeliverySnapshot is the code-owned terminal contract for a generated deck.
// The LLM may create pages, but it never decides whether this snapshot is done.
type DeliverySnapshot struct {
	Total        int
	Done         int
	Files        []string
	PendingTasks []string
}

func (s DeliverySnapshot) Complete() bool {
	return s.Total > 0 && s.Done == s.Total && len(s.PendingTasks) == 0
}

func deliverySnapshotFromManifest(manifest *deck.TasksManifest) DeliverySnapshot {
	if manifest == nil {
		return DeliverySnapshot{}
	}
	snapshot := DeliverySnapshot{
		Total: len(manifest.Tasks),
		Done:  manifest.CompletedCount(),
		Files: ManifestOutputFiles(manifest),
	}
	for _, item := range manifest.Tasks {
		if item == nil || isCompletedSlideStatus(item.Status) {
			continue
		}
		snapshot.PendingTasks = append(snapshot.PendingTasks, item.TaskID)
	}
	return snapshot
}

func (ts *TaskState) updateDelivery(snapshot DeliverySnapshot) {
	ts.Mu.Lock()
	defer ts.Mu.Unlock()
	ts.delivery = snapshot
	ts.Info.DoneCount = snapshot.Done
	ts.Info.TotalCount = snapshot.Total
	ts.Info.Files = append([]string(nil), snapshot.Files...)
}

func (ts *TaskState) deliverySnapshot() DeliverySnapshot {
	ts.Mu.Lock()
	defer ts.Mu.Unlock()
	snapshot := ts.delivery
	snapshot.Files = append([]string(nil), ts.delivery.Files...)
	snapshot.PendingTasks = append([]string(nil), ts.delivery.PendingTasks...)
	return snapshot
}

func (tm *TaskManager) syncDeliveryMetadata(ts *TaskState, workDir string) (DeliverySnapshot, error) {
	manifest, err := deck.ReconcileTasksManifestOutputFiles(workDir)
	if err != nil {
		return DeliverySnapshot{}, err
	}
	if manifest == nil {
		return DeliverySnapshot{}, fmt.Errorf("delivery manifest is missing")
	}
	snapshot := deliverySnapshotFromManifest(manifest)
	ts.updateDelivery(snapshot)
	if ts.runtimeMeta != nil {
		ts.runtimeMeta.RecordSlideProgress(snapshot.Done, snapshot.Total, len(snapshot.PendingTasks))
		ts.runtimeMeta.RecordManifestValidation(snapshot.Done, snapshot.Total, nil, snapshot.PendingTasks)
	}
	return snapshot, nil
}
