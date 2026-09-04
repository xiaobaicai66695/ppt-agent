package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
	"github.com/cloudwego/ppt-agent/pkg/runtime/task"
	"github.com/cloudwego/ppt-agent/pkg/session"
)

func newContinueTestContext(taskID, message string) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	body, _ := json.Marshal(map[string]string{"message": message})
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/tasks/"+taskID+"/continue", bytes.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = gin.Params{{Key: "id", Value: taskID}}
	return ctx, recorder
}

func TestHandleContinueTaskReturnsAcceptedJSON(t *testing.T) {
	manager := task.NewTaskManager(t.TempDir(), nil, nil, nil)
	manager.NewColdTaskState(task.TaskInfo{ID: "task-1", Status: task.TaskStatusCompleted})
	started := false
	server := &Server{
		tasks: manager, sessionManager: session.NewSessionManager(),
		continueStarter: func(_ string, _ *task.TaskState, _ string, _ int, _ *session.ConversationSession) {
			started = true
		},
	}
	ctx, recorder := newContinueTestContext("task-1", "请调整第二页")

	server.handleContinueTask(ctx)
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusAccepted, recorder.Body.String())
	}
	if !started {
		t.Fatal("continuation starter was not called")
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "accepted" {
		t.Fatalf("response = %#v", body)
	}
}

func TestHandleContinueTaskQueuesRunningTask(t *testing.T) {
	manager := task.NewTaskManager(t.TempDir(), nil, nil, nil)
	ts := manager.NewColdTaskState(task.TaskInfo{ID: "task-2", Status: task.TaskStatusRunning})
	server := &Server{tasks: manager, sessionManager: session.NewSessionManager()}
	ctx, recorder := newContinueTestContext("task-2", "完成后换成深色标题")

	server.handleContinueTask(ctx)
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusAccepted, recorder.Body.String())
	}
	if !ts.HasPendingContinueMsg() {
		t.Fatal("running task did not retain the queued message")
	}
}

func TestResumeDraftCheckpointClaimsUncommittedDraft(t *testing.T) {
	workDir := t.TempDir()
	if err := deck.WriteTasksDraftManifest(workDir, &deck.TasksManifest{}); err != nil {
		t.Fatal(err)
	}
	server := &Server{}
	ts := &task.TaskState{Info: task.TaskInfo{ID: "task-draft", WorkDir: workDir}}
	claimed, err := server.resumeDraftCheckpoint("task-draft", ts, make(chan task.SSERichEvent, 1))
	if !claimed {
		t.Fatal("uncommitted draft checkpoint should be claimed by resume path")
	}
	if err == nil {
		t.Fatal("empty draft should report a recoverable checkpoint error")
	}
}

func TestResumeDraftCheckpointLeavesCommittedTaskToEditWorkflow(t *testing.T) {
	workDir := t.TempDir()
	if err := deck.WriteTasksManifest(workDir, &deck.TasksManifest{}); err != nil {
		t.Fatal(err)
	}
	server := &Server{}
	ts := &task.TaskState{Info: task.TaskInfo{ID: "task-committed", WorkDir: workDir}}
	claimed, err := server.resumeDraftCheckpoint("task-committed", ts, make(chan task.SSERichEvent, 1))
	if err != nil {
		t.Fatalf("resume checkpoint error = %v", err)
	}
	if claimed {
		t.Fatal("committed manifest should use the normal edit workflow")
	}
}

func TestResumeDraftCheckpointRecognizesRetryableRenderMarkerAfterStatusReopens(t *testing.T) {
	workDir := t.TempDir()
	if err := deck.WriteTasksManifest(workDir, &deck.TasksManifest{}); err != nil {
		t.Fatal(err)
	}
	server := &Server{}
	ts := &task.TaskState{Info: task.TaskInfo{
		ID: "task-render", WorkDir: workDir, Status: task.TaskStatusRunning,
		Error: "可恢复的 unexpected_eof：Post request: unexpected EOF。可在对话框输入“继续任务”从检查点恢复。",
	}}
	claimed, err := server.resumeDraftCheckpoint("task-render", ts, make(chan task.SSERichEvent, 1))
	if !claimed {
		t.Fatal("retryable render marker should be claimed after the status changes back to running")
	}
	if err == nil || !strings.Contains(err.Error(), "从渲染检查点恢复") {
		t.Fatalf("resume error = %v", err)
	}
}
