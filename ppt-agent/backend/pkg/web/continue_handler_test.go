package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/cloudwego/ppt-agent/pkg/session"
	"github.com/cloudwego/ppt-agent/pkg/task"
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
