package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cloudwego/ppt-agent/pkg/db"
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

func TestHandleGetConversationBypassesTerminalInMemoryLock(t *testing.T) {
	manager := task.NewTaskManager(t.TempDir(), nil, nil, nil)
	ts := manager.NewColdTaskState(task.TaskInfo{
		ID: "task-terminal", Status: task.TaskStatusCompleted,
		Files: []string{"1_cover.pptx"},
	})
	ts.Mu.Lock()
	defer ts.Mu.Unlock()

	server := &Server{tasks: manager, sessionManager: session.NewSessionManager()}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/tasks/task-terminal/conversation", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "task-terminal"}}

	done := make(chan struct{})
	go func() {
		server.handleGetConversation(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("conversation blocked on terminal in-memory TaskState lock")
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestHandleGetConversationIncludesPersistedRuntimeTimeline(t *testing.T) {
	workDir := t.TempDir()
	oldListRuntimeEvents := listRuntimeEvents
	listRuntimeEvents = func(taskID string) ([]db.RuntimeEventRecord, error) {
		if taskID != "task-runtime" {
			return nil, nil
		}
		return []db.RuntimeEventRecord{
			{TaskID: taskID, EventID: 1, Timestamp: time.Date(2026, 8, 4, 10, 0, 0, 0, time.UTC), ElapsedMS: 12, Kind: "tool_start", Name: "python3", Status: "running", Metadata: `{"args":"{\"code\":\"print(1)\"}"}`},
			{TaskID: taskID, EventID: 2, Timestamp: time.Date(2026, 8, 4, 10, 0, 1, 0, time.UTC), ElapsedMS: 30, Kind: "tool_end", Name: "python3", Status: "ok", Metadata: `{"result":"stdout:\n1"}`},
			{TaskID: taskID, EventID: 3, Timestamp: time.Date(2026, 8, 4, 10, 0, 2, 0, time.UTC), ElapsedMS: 42, Kind: "llm_end", Name: "planner", Status: "ok", Metadata: `{"assistant_output":"## 规划\n\n开始拆分页面。","output_preview":"规划摘要"}`},
		}, nil
	}
	defer func() { listRuntimeEvents = oldListRuntimeEvents }()
	manager := task.NewTaskManager(t.TempDir(), nil, nil, nil)
	manager.NewColdTaskState(task.TaskInfo{ID: "task-runtime", Status: task.TaskStatusCompleted, WorkDir: workDir})
	server := &Server{tasks: manager, sessionManager: session.NewSessionManager()}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/tasks/task-runtime/conversation", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "task-runtime"}}

	server.handleGetConversation(ctx)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		RuntimeMeta struct {
			RecentEvents []struct {
				Kind     string         `json:"kind"`
				Metadata map[string]any `json:"metadata"`
			} `json:"recent_events"`
		} `json:"runtime_meta"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.RuntimeMeta.RecentEvents) != 3 {
		t.Fatalf("recent_events = %#v", payload.RuntimeMeta.RecentEvents)
	}
	if payload.RuntimeMeta.RecentEvents[0].Metadata["args_preview"] == nil || payload.RuntimeMeta.RecentEvents[1].Metadata["result_preview"] == nil {
		t.Fatalf("conversation timeline should expose bounded tool previews: %#v", payload.RuntimeMeta.RecentEvents)
	}
	if payload.RuntimeMeta.RecentEvents[0].Metadata["args"] != nil || payload.RuntimeMeta.RecentEvents[1].Metadata["result"] != nil {
		t.Fatalf("conversation timeline should omit raw tool payloads: %#v", payload.RuntimeMeta.RecentEvents)
	}
	if got := payload.RuntimeMeta.RecentEvents[2].Metadata["assistant_output"]; got != "## 规划\n\n开始拆分页面。" {
		t.Fatalf("conversation timeline should keep assistant_output only, got %#v", payload.RuntimeMeta.RecentEvents[2].Metadata)
	}
}

func TestHandleGetRuntimeEventReturnsPersistedMetadata(t *testing.T) {
	oldGetRuntimeEvent := getRuntimeEvent
	getRuntimeEvent = func(taskID string, eventID int64) (*db.RuntimeEventRecord, error) {
		if taskID != "task-runtime" || eventID != 2 {
			return nil, nil
		}
		return &db.RuntimeEventRecord{
			TaskID: taskID, EventID: 2,
			Timestamp: time.Date(2026, 8, 4, 10, 0, 1, 0, time.UTC),
			ElapsedMS: 30, Kind: "tool_end", Name: "python3", Status: "ok",
			Metadata: `{"args":"{\"code\":\"print(1)\"}","result":"stdout:\n1"}`,
		}, nil
	}
	defer func() { getRuntimeEvent = oldGetRuntimeEvent }()

	server := &Server{}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/tasks/task-runtime/runtime-events/2", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "task-runtime"}, {Key: "event_id", Value: "2"}}

	server.handleGetRuntimeEvent(ctx)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Kind     string         `json:"kind"`
		Metadata map[string]any `json:"metadata"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Kind != "tool_end" || payload.Metadata["args"] == "" || payload.Metadata["result"] == "" {
		t.Fatalf("runtime event metadata missing: %#v", payload)
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

func TestConversationMessagesWithFallbackTrimsCumulativeAssistantPrefix(t *testing.T) {
	prefix := "我将为您创建一个关于微服务项目治理的20页PPT。首先，让我读取组件契约文件以了解可用的组件类型和版式。"
	suffix := "我已了解组件契约。现在让我规划20页的微服务项目治理PPT。"
	messages := []session.Message{
		{Role: "assistant", Content: prefix, Timestamp: time.Date(2026, 8, 24, 10, 0, 0, 0, time.UTC)},
		{Role: "assistant", Content: prefix + "\n\n" + suffix, Timestamp: time.Date(2026, 8, 24, 10, 0, 1, 0, time.UTC)},
	}

	got := conversationMessagesWithFallback(messages, "", "", time.Time{})

	if len(got) != 2 {
		t.Fatalf("len(messages) = %d, want 2: %#v", len(got), got)
	}
	if got[0].Content != prefix || got[1].Content != suffix {
		t.Fatalf("messages = %#v, want prefix then suffix", got)
	}
}

func TestConversationMessagesWithFallbackDropsContainedAssistantFragments(t *testing.T) {
	full := "我将为您生成一份12页的PPT，主题为中小后端项目Git版本控制规范方案。\nPPT规划已完成，12页内容已提交。以下是各页概要：\n| 页码 | 标题 |\n| --- | --- |\n| 12 | 致谢 |"
	fragment := "PPT规划已完成，12页内容已提交。以下是各页概要：\n| 页码 | 标题 |\n| --- | --- |\n| 12 | 致谢 |"
	messages := []session.Message{
		{Role: "assistant", Content: fragment, Timestamp: time.Date(2026, 8, 24, 10, 0, 0, 0, time.UTC)},
		{Role: "assistant", Content: full, Timestamp: time.Date(2026, 8, 24, 10, 0, 1, 0, time.UTC)},
	}

	got := conversationMessagesWithFallback(messages, "", "", time.Time{})

	if len(got) != 1 {
		t.Fatalf("len(messages) = %d, want 1: %#v", len(got), got)
	}
	if got[0].Content != full {
		t.Fatalf("kept content = %q, want full answer", got[0].Content)
	}
}
