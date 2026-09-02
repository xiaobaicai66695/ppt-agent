package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/cloudwego/ppt-agent/pkg/task"
)

func TestHandleSaveTaskFeedbackRejectsInvalidRating(t *testing.T) {
	manager := task.NewTaskManager(t.TempDir(), nil, nil, nil)
	manager.NewColdTaskState(task.TaskInfo{ID: "task-feedback", Status: task.TaskStatusCompleted})
	server := &Server{tasks: manager}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "id", Value: "task-feedback"}}
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/tasks/task-feedback/feedback", bytes.NewBufferString(`{"rating":0}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	server.handleSaveTaskFeedback(ctx)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
}
