package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	if len(messages) < 2 || messages[0].Content != "我想做个青甘大环线的旅游项目介绍" || messages[len(messages)-1].Content != "你决定主题和风格吧" {
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
