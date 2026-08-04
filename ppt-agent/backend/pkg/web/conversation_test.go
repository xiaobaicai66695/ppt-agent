package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cloudwego/ppt-agent/pkg/session"
	"github.com/cloudwego/ppt-agent/pkg/task"
)

func TestConversationMessagesWithFallbackUsesFullAnswer(t *testing.T) {
	created := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	messages := []session.Message{{Role: "user", Content: "生成一套演示", Timestamp: created}}
	got := conversationMessagesWithFallback(messages, "## 完成\n\n- 第一页", "legacy", created.Add(time.Minute))
	if len(got) != 2 {
		t.Fatalf("len(messages) = %d, want 2", len(got))
	}
	if got[1].Role != "assistant" || got[1].Content != "## 完成\n\n- 第一页" {
		t.Fatalf("unexpected fallback message: %#v", got[1])
	}
}

func TestHandleGetConversationDeduplicatesCompletionFiles(t *testing.T) {
	manager := task.NewTaskManager(t.TempDir(), nil, nil, nil)
	manager.NewColdTaskState(task.TaskInfo{
		ID: "task-files", Status: task.TaskStatusCompleted,
		Files: []string{"/srv/task/1_cover.pptx", "1_cover.pptx", `C:\\task\\2_agenda.pptx`},
	})
	server := &Server{tasks: manager, sessionManager: session.NewSessionManager()}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/tasks/task-files/conversation", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "task-files"}}

	server.handleGetConversation(ctx)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Files []string `json:"files"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Files) != 2 || payload.Files[0] != "1_cover.pptx" || payload.Files[1] != "2_agenda.pptx" {
		t.Fatalf("files = %#v", payload.Files)
	}
}

func TestHandleGetConversationIncludesLatestEventID(t *testing.T) {
	manager := task.NewTaskManager(t.TempDir(), nil, nil, nil)
	ts := manager.NewColdTaskState(task.TaskInfo{ID: "task-cursor", Status: task.TaskStatusRunning})
	ts.Broadcast(task.SSERichEvent{Type: "answer", Content: "正在生成"})
	ts.Broadcast(task.SSERichEvent{Type: "progress", Phase: "generating"})
	ts.Broadcast(task.SSERichEvent{Type: "answer_end"})
	server := &Server{tasks: manager, sessionManager: session.NewSessionManager()}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/tasks/task-cursor/conversation", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "task-cursor"}}

	server.handleGetConversation(ctx)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		LatestEventID      uint64 `json:"latest_event_id"`
		ReplayAfterEventID uint64 `json:"replay_after_event_id"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.LatestEventID != 3 || payload.ReplayAfterEventID != 3 {
		t.Fatalf("event boundaries = (%d, %d), want (3, 3)", payload.LatestEventID, payload.ReplayAfterEventID)
	}
}

func TestConversationMessagesWithFallbackKeepsStructuredAssistant(t *testing.T) {
	messages := []session.Message{
		{Role: "user", Content: "修改第二页"},
		{Role: "assistant", Content: "- 已调整布局\n- 已更新文件"},
	}
	got := conversationMessagesWithFallback(messages, "legacy answer", "legacy content", time.Time{})
	if len(got) != 2 || got[1].Content != messages[1].Content {
		t.Fatalf("structured messages were replaced: %#v", got)
	}
}

func TestConversationMessagesWithFallbackDropsDuplicateFragments(t *testing.T) {
	messages := []session.Message{
		{Role: "assistant", Content: "完整回答"},
		{Role: "assistant", Content: "完整回答"},
		{Role: "assistant", Content: "..."},
	}
	got := conversationMessagesWithFallback(messages, "", "", time.Time{})
	if len(got) != 1 {
		t.Fatalf("len(messages) = %d, want 1: %#v", len(got), got)
	}
}
