package task

// TaskStatusSnapshot returns a race-free status for callers that only need the
// state boundary and must not hold TaskState.Mu while doing I/O.
func (ts *TaskState) TaskStatusSnapshot() TaskStatus {
	if ts == nil {
		return ""
	}
	ts.Mu.Lock()
	defer ts.Mu.Unlock()
	return ts.Info.Status
}
