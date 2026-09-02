package task

import "testing"

func TestApprovalPolicyAndLifecycle(t *testing.T) {
	if ShouldRequestApproval(ApprovalModeAuto, "重写全部页面") {
		t.Fatal("auto mode must not pause")
	}
	if !ShouldRequestApproval(ApprovalModeManual, "只修改第 2 页") {
		t.Fatal("manual mode must pause every continuation")
	}
	if !ShouldRequestApproval(ApprovalModeSensitive, "把整份 PPT 的主题全部重做") {
		t.Fatal("sensitive mode must pause global changes")
	}

	ts := &TaskState{
		Info:          TaskInfo{ID: "task-approval", Status: TaskStatusCompleted, ApprovalMode: ApprovalModeManual},
		listeners:     make(map[string]chan SSERichEvent),
		reportedFiles: make(map[string]bool),
	}
	approval := NewContinuationApproval(ApprovalModeManual, "调整第 2 页标题")
	if !ts.BeginApproval(approval) {
		t.Fatal("BeginApproval() = false")
	}
	if got := ts.SnapshotInfo(); got.Status != TaskStatusWaitingApproval || got.PendingApproval == nil {
		t.Fatalf("pending approval state = %#v", got)
	}
	if _, approved, err := ts.ResolveApproval("adjust_scope", "只调整第 2 页正文"); err != nil || approved {
		t.Fatalf("adjust scope = approved:%v err:%v", approved, err)
	}
	message, approved, err := ts.ResolveApproval("approve", "")
	if err != nil || !approved || message != "只调整第 2 页正文" {
		t.Fatalf("approve = message:%q approved:%v err:%v", message, approved, err)
	}
	if got := ts.SnapshotInfo(); got.Status != TaskStatusRunning || got.PendingApproval != nil {
		t.Fatalf("approved state = %#v", got)
	}
}

func TestRejectApprovalKeepsCompletedTask(t *testing.T) {
	ts := &TaskState{
		Info:          TaskInfo{ID: "task-reject", Status: TaskStatusCompleted},
		listeners:     make(map[string]chan SSERichEvent),
		reportedFiles: make(map[string]bool),
	}
	if !ts.BeginApproval(NewContinuationApproval(ApprovalModeSensitive, "重做整体风格")) {
		t.Fatal("BeginApproval() = false")
	}
	if _, approved, err := ts.ResolveApproval("reject", ""); err != nil || approved {
		t.Fatalf("reject = approved:%v err:%v", approved, err)
	}
	if got := ts.SnapshotInfo(); got.Status != TaskStatusCompleted || got.PendingApproval != nil {
		t.Fatalf("rejected state = %#v", got)
	}
}
