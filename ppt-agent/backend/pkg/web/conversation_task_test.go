package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/ppt-agent/pkg/auth"
	"github.com/cloudwego/ppt-agent/pkg/db"
	"github.com/cloudwego/ppt-agent/pkg/session"
	"github.com/cloudwego/ppt-agent/pkg/task"
	"github.com/gin-gonic/gin"
)

func TestTaskGenerationQueryRetainsInitialTopicAndFollowup(t *testing.T) {
	sess := session.NewSession("task-qinggan", t.TempDir())
	if err := sess.AddUserMessage("我想做个青甘大环线的旅游项目介绍"); err != nil {
		t.Fatal(err)
	}
	if err := sess.AddAssistantMessage("建议先明确受众和风格。"); err != nil {
		t.Fatal(err)
	}
	if err := sess.AddUserMessage("你决定主题和风格吧"); err != nil {
		t.Fatal(err)
	}

	query := taskGenerationQuery(sess, "fallback")
	for _, want := range []string{"青甘大环线", "你决定主题和风格吧"} {
		if !strings.Contains(query, want) {
			t.Fatalf("generation query %q does not retain %q", query, want)
		}
	}
}

func TestHandleMessageManualPPTModeOverridesGenericChatRoute(t *testing.T) {
	previousDB := db.DB
	db.DB = nil
	t.Cleanup(func() { db.DB = previousDB })

	server := &Server{
		tasks:          task.NewTaskManager(t.TempDir(), nil, nil, nil),
		sessionManager: session.NewSessionManager(),
		taskIDGen:      func() string { return "manual-ppt" },
		textModelFactory: func(context.Context) (interface {
			Generate(context.Context, []*schema.Message, ...interface{}) (*schema.Message, error)
		}, error) {
			return fakeCreateRouteModel{response: `{"intent":"chat","mode":"chat","action":"reply","confidence":0.95,"reply":"普通回答"}`}, nil
		},
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/api/messages", bytes.NewBufferString(`{"message":"人工智能发展趋势","manual_mode":"pptagent"}`))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req.WithContext(auth.WithUser(req.Context(), &db.User{ID: 7}))

	server.handleMessage(ctx)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	var route MessageRouteResult
	if err := json.Unmarshal(recorder.Body.Bytes(), &route); err != nil {
		t.Fatal(err)
	}
	if route.Intent != messageIntentCreate || route.Mode != messageModePPTAgent || route.Action != messageActionPrepareCreate {
		t.Fatalf("manual PPT route = %#v, want create/pptagent/prepare_create", route)
	}
}

func TestHandleMessageStreamsChatTurnOverTaskTimeline(t *testing.T) {
	previousDB := db.DB
	db.DB = nil
	t.Cleanup(func() { db.DB = previousDB })

	manager := task.NewTaskManager(t.TempDir(), nil, nil, nil)
	sessions := session.NewSessionManager()
	manager.SetAssistantTurnCallback(func(taskID, workDir, content string) {
		if err := sessions.GetOrCreate(taskID, workDir).AddAssistantMessage(content); err != nil {
			t.Errorf("persist assistant turn: %v", err)
		}
	})
	server := &Server{
		tasks:          manager,
		sessionManager: sessions,
		taskIDGen:      func() string { return "chat-stream" },
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/api/messages", bytes.NewBufferString(`{"message":"你好"}`))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req.WithContext(auth.WithUser(req.Context(), &db.User{ID: 7}))

	server.handleMessage(ctx)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	var route MessageRouteResult
	if err := json.Unmarshal(recorder.Body.Bytes(), &route); err != nil {
		t.Fatal(err)
	}
	if !route.Streaming || route.TaskID != "chat-stream" {
		t.Fatalf("route = %#v, want streamed chat task", route)
	}

	ts := manager.GetTaskState(route.TaskID)
	if ts == nil {
		t.Fatal("chat task state was not created")
	}
	deadline := time.Now().Add(2 * time.Second)
	for {
		events, done := ts.SubscribeFrom("test", make(chan task.SSERichEvent, 8), 0)
		types := make(map[string]bool, len(events))
		for _, event := range events {
			types[event.Type] = true
		}
		if done && types["answer"] && types["answer_end"] && types["conversation_complete"] {
			messages := sessions.Get(route.TaskID).GetRecentMessages(0)
			assistantCount := 0
			for _, message := range messages {
				if message.Role == "assistant" {
					assistantCount++
				}
			}
			if assistantCount != 1 {
				t.Fatalf("assistant messages = %d, want one durable streamed reply: %#v", assistantCount, messages)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("chat timeline did not close through SSE, events=%#v done=%v", events, done)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestHandleMessageCreatesAndReusesConversationTask(t *testing.T) {
	previousDB := db.DB
	db.DB = nil
	t.Cleanup(func() { db.DB = previousDB })

	manager := task.NewTaskManager(t.TempDir(), nil, nil, nil)
	server := &Server{
		tasks:          manager,
		sessionManager: session.NewSessionManager(),
		taskIDGen:      func() string { return "task-qinggan" },
	}
	call := func(payload string, userID uint) (MessageRouteResult, int) {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req := httptest.NewRequest(http.MethodPost, "/api/messages", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		ctx.Request = req.WithContext(auth.WithUser(req.Context(), &db.User{ID: userID}))
		server.handleMessage(ctx)
		var route MessageRouteResult
		if err := json.Unmarshal(recorder.Body.Bytes(), &route); err != nil {
			return MessageRouteResult{}, recorder.Code
		}
		return route, recorder.Code
	}

	first, code := call(`{"message":"我想做个青甘大环线的旅游项目介绍"}`, 7)
	if code != http.StatusOK {
		t.Fatal("first message was not accepted")
	}
	if first.TaskID != "task-qinggan" {
		t.Fatalf("first task id = %q", first.TaskID)
	}
	second, code := call(`{"message":"你决定主题和风格吧","selected_task_id":"task-qinggan"}`, 7)
	if code != http.StatusOK {
		t.Fatal("follow-up message was not accepted")
	}
	if second.TaskID != first.TaskID || second.Intent != messageIntentCreate {
		t.Fatalf("second route = %#v, want create on the initial task", second)
	}
	info := manager.GetTask(first.TaskID)
	if info == nil || info.Status != task.TaskStatusConversation {
		t.Fatalf("conversation task = %#v", info)
	}
	messages := server.sessionManager.Get(first.TaskID).GetRecentMessages(0)
	hasInitial, hasFollowup := false, false
	for _, message := range messages {
		hasInitial = hasInitial || message.Content == "我想做个青甘大环线的旅游项目介绍"
		hasFollowup = hasFollowup || message.Content == "你决定主题和风格吧"
	}
	if !hasInitial || !hasFollowup {
		t.Fatalf("messages = %#v", messages)
	}
	if _, code = call(`{"message":"越权访问","selected_task_id":"task-qinggan"}`, 8); code != http.StatusNotFound {
		t.Fatalf("cross-user message status = %d, want %d", code, http.StatusNotFound)
	}
}

func TestNewServerRegistersPersistentConversationRoutes(t *testing.T) {
	server := NewServer(&ServerConfig{BaseDir: t.TempDir()})
	for _, path := range []string{"/api/messages", "/api/plan-drafts"} {
		recorder := httptest.NewRecorder()
		server.engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, path, nil))
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("POST %s status = %d, want authenticated route", path, recorder.Code)
		}
	}
}
