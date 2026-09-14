package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cloudwego/ppt-agent/pkg/db"
	"github.com/cloudwego/ppt-agent/pkg/runtime/task"
	"github.com/cloudwego/ppt-agent/pkg/session"
)

func TestPublicTimelineEventKeepsCompleteToolPayloads(t *testing.T) {
	args := `{"notes":"` + strings.Repeat("x", persistedTimelineTextLimit+64) + `"}`
	result := strings.Repeat("tool output\n", persistedTimelineTextLimit+64)
	event := publicTimelineEvent(task.SSERichEvent{Type: task.SSEEventToolResult, ToolArgs: args, ToolResult: result})
	if event.ToolArgs != args || event.ToolResult != result {
		t.Fatalf("tool payload was truncated: %#v", event)
	}
}

func TestPersistedConversationMessagesPreservesStoredTurns(t *testing.T) {
	messages := []session.Message{
		{Role: "user", Content: "生成一套演示"},
		{Role: "user", Content: "生成一套演示"},
		{Role: "assistant", Content: "## 完成\n\n- 第一页"},
		{Role: "assistant", Content: "..."},
	}
	got := persistedConversationMessages(messages)
	if len(got) != 3 {
		t.Fatalf("len(messages) = %d, want 3", len(got))
	}
	if got[0].Content != got[1].Content || got[2].Role != "assistant" {
		t.Fatalf("stored turn order was changed: %#v", got)
	}
}

func TestConversationMessagesWithTaskFallbackKeepsHistoricalTaskVisible(t *testing.T) {
	createdAt := time.Date(2026, 9, 14, 8, 0, 0, 0, time.UTC)
	messages := conversationMessagesWithTaskFallback(nil, task.TaskInfo{Query: "为年度复盘制作演示", CreatedAt: createdAt})
	if len(messages) != 1 {
		t.Fatalf("len(messages) = %d, want 1", len(messages))
	}
	if messages[0].Role != "user" || messages[0].Content != "为年度复盘制作演示" || !messages[0].Timestamp.Equal(createdAt) {
		t.Fatalf("fallback message = %#v", messages[0])
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

func TestHandleGetConversationDoesNotLoadPersistedRuntimeTimeline(t *testing.T) {
	workDir := t.TempDir()
	oldListRuntimeEvents := listRuntimeEvents
	listRuntimeEvents = func(taskID string) ([]db.RuntimeEventRecord, error) {
		return nil, nil
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
	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if _, exists := payload["runtime_meta"]; exists {
		t.Fatal("conversation snapshot must omit runtime metadata")
	}
}

func TestConversationTimelineMergesMessagesWithDurableTrace(t *testing.T) {
	oldListTrace := listConversationTraceEvents
	listConversationTraceEvents = func(taskID string) ([]db.ConversationTraceEvent, error) {
		if taskID != "task-history" {
			t.Fatalf("task id = %q", taskID)
		}
		payload := func(event task.SSERichEvent) string {
			data, err := json.Marshal(event)
			if err != nil {
				t.Fatal(err)
			}
			return string(data)
		}
		base := time.Date(2026, 9, 13, 9, 0, 0, 0, time.UTC)
		return []db.ConversationTraceEvent{
			{ID: 1, TaskID: taskID, Type: task.SSEEventThought, Timestamp: base.Add(time.Second), Payload: payload(task.SSERichEvent{ID: 11, Type: task.SSEEventThought, Content: "正在规划", Phase: "planning"})},
			{ID: 2, TaskID: taskID, Type: task.SSEEventToolCall, Timestamp: base.Add(2 * time.Second), Payload: payload(task.SSERichEvent{ID: 12, Type: task.SSEEventToolCall, ToolCallID: "tool-12", ToolName: "read_file"})},
			{ID: 3, TaskID: taskID, Type: task.SSEEventToolResult, Timestamp: base.Add(3 * time.Second), Payload: payload(task.SSERichEvent{ID: 13, Type: task.SSEEventToolResult, ToolCallID: "tool-12", ToolName: "read_file", ToolResult: "已读取"})},
		}, nil
	}
	defer func() { listConversationTraceEvents = oldListTrace }()

	base := time.Date(2026, 9, 13, 9, 0, 0, 0, time.UTC)
	timeline := conversationTimeline("task-history", []session.Message{
		{Role: "user", Content: "做一份储能方案", Timestamp: base},
		{Role: "assistant", Content: "PPT 已完成交付", Timestamp: base.Add(4 * time.Second)},
	})
	if len(timeline) != 5 {
		t.Fatalf("timeline length = %d, want 5", len(timeline))
	}
	encoded, err := json.Marshal(timeline)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"type":"message"`, `"type":"thought"`, `"type":"tool_call"`, `"type":"tool_result"`, "PPT 已完成交付"} {
		if !bytes.Contains(encoded, []byte(want)) {
			t.Fatalf("timeline %s does not contain %q", encoded, want)
		}
	}
}

func TestConversationTimelineFallsBackToLegacyRuntimeTools(t *testing.T) {
	oldListTrace := listConversationTraceEvents
	oldListRuntimeEvents := listRuntimeEvents
	listConversationTraceEvents = func(string) ([]db.ConversationTraceEvent, error) { return nil, nil }
	listRuntimeEvents = func(taskID string) ([]db.RuntimeEventRecord, error) {
		base := time.Date(2026, 9, 13, 9, 0, 0, 0, time.UTC)
		return []db.RuntimeEventRecord{
			{ID: 20, TaskID: taskID, EventID: 101, Timestamp: base, Kind: "tool_start", Name: "search", Status: "running", Metadata: `{"args":"{\"query\":\"储能\"}"}`},
			{ID: 21, TaskID: taskID, EventID: 102, Timestamp: base.Add(time.Second), Kind: "tool_end", Name: "search", Status: "ok", Metadata: `{"args":"{\"query\":\"储能\"}","result":"已找到资料"}`},
		}, nil
	}
	defer func() {
		listConversationTraceEvents = oldListTrace
		listRuntimeEvents = oldListRuntimeEvents
	}()

	timeline := conversationTimeline("legacy-task", nil)
	if len(timeline) != 2 {
		t.Fatalf("legacy timeline length = %d, want 2", len(timeline))
	}
	encoded, err := json.Marshal(timeline)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(encoded, []byte(`"type":"tool_call"`)) || !bytes.Contains(encoded, []byte(`"type":"tool_result"`)) || !bytes.Contains(encoded, []byte(`"tool_call_id":"runtime-101"`)) {
		t.Fatalf("legacy runtime tools were not reconstructed: %s", encoded)
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
