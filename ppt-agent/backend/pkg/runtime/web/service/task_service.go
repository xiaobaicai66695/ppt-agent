// Package service contains Web-facing application services.  Services own
// workflow orchestration and runtime dependencies; routers should only decode
// requests and map service results to HTTP responses.
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/ppt-agent/pkg/agent/ppt"
	"github.com/cloudwego/ppt-agent/pkg/runtime/task"
	webmodel "github.com/cloudwego/ppt-agent/pkg/runtime/web/model"
	"github.com/cloudwego/ppt-agent/pkg/session"
	"github.com/cloudwego/ppt-agent/pkg/templates"
)

// ErrInvalidOutline marks a client-supplied outline that cannot be accepted
// by the configured component contract.
var ErrInvalidOutline = errors.New("invalid outline")

// TaskService is the application boundary for task creation and lifecycle
// operations used by HTTP handlers.  It deliberately exposes the existing
// TaskManager rather than duplicating its state machine.
type TaskService struct {
	tasks          *task.TaskManager
	sessions       *session.SessionManager
	agentFactory   task.AgentFactory
	makeTaskConfig func(taskID string) *ppt.PPTTaskConfig
	templateLoader *templates.Loader
}

// TaskServiceConfig wires runtime dependencies into TaskService.
type TaskServiceConfig struct {
	Tasks          *task.TaskManager
	Sessions       *session.SessionManager
	AgentFactory   task.AgentFactory
	MakeTaskConfig func(taskID string) *ppt.PPTTaskConfig
	TemplateLoader *templates.Loader
}

// NewTaskService creates a task application service.  Dependencies are kept
// explicit so unit tests can provide an in-memory TaskManager.
func NewTaskService(cfg TaskServiceConfig) *TaskService {
	return &TaskService{
		tasks:          cfg.Tasks,
		sessions:       cfg.Sessions,
		agentFactory:   cfg.AgentFactory,
		makeTaskConfig: cfg.MakeTaskConfig,
		templateLoader: cfg.TemplateLoader,
	}
}

// FindRecentDuplicate returns a still-active task created for the same query
// in the recent duplicate-suppression window.
func (s *TaskService) FindRecentDuplicate(userID int, query string) *task.TaskInfo {
	queryKey := normalizeMessageKey(query)
	if queryKey == "" || s == nil || s.tasks == nil {
		return nil
	}
	deadline := time.Now().Add(-2 * time.Minute)
	for _, info := range s.tasks.ListTasks(userID) {
		if info.CreatedAt.Before(deadline) {
			continue
		}
		if info.Status == task.TaskStatusCancelled || info.Status == task.TaskStatusFailed {
			continue
		}
		if normalizeMessageKey(info.Query) == queryKey {
			copy := info
			return &copy
		}
	}
	return nil
}

// Create creates and asynchronously starts a PPT task.  Route decisions are
// intentionally made by the router layer; this method only applies the
// selected request to the runtime config and task manager.
func (s *TaskService) Create(ctx context.Context, input webmodel.TaskCreateInput) (*task.TaskInfo, error) {
	if s == nil || s.tasks == nil {
		return nil, fmt.Errorf("task service is not configured")
	}
	if s.makeTaskConfig == nil {
		return nil, fmt.Errorf("task config factory is not configured")
	}
	cfg := s.makeTaskConfig(input.TaskID)
	if cfg == nil {
		return nil, fmt.Errorf("task config factory returned nil")
	}
	cfg.Query = input.Query
	cfg.UserID = input.UserID
	cfg.ModelAPIKey = input.Credential.APIKey
	cfg.ModelProvider = input.Credential.Provider
	cfg.Intent = input.Intent
	cfg.ConversationID = input.ConversationID
	if input.Outline != nil && len(input.Outline.Slides) > 0 {
		outline, err := s.PrepareOutline(input.Query, input.Outline)
		if err != nil {
			return nil, err
		}
		cfg.Outline = outline
	}

	info, err := s.tasks.CreateTask(ctx, input.Query, input.UserID, s.agentFactory, cfg)
	if err != nil {
		return nil, err
	}
	if s.sessions != nil {
		if err := s.sessions.GetOrCreate(input.TaskID, info.WorkDir).AddUserMessage(input.Query); err != nil {
			s.tasks.CancelTask(input.TaskID)
			return nil, fmt.Errorf("保存会话消息失败: %w", err)
		}
	}
	return info, nil
}

// StartConversation promotes an existing conversation task to PPT generation
// without changing its identity.
func (s *TaskService) StartConversation(ctx context.Context, input webmodel.StartConversationInput) (*task.TaskInfo, error) {
	if s == nil || s.tasks == nil {
		return nil, fmt.Errorf("task service is not configured")
	}
	if s.makeTaskConfig == nil {
		return nil, fmt.Errorf("task config factory is not configured")
	}
	cfg := s.makeTaskConfig(input.TaskID)
	if cfg == nil {
		return nil, fmt.Errorf("task config factory returned nil")
	}
	cfg.Query = input.Query
	cfg.UserID = input.UserID
	cfg.ModelAPIKey = input.Credential.APIKey
	cfg.ModelProvider = input.Credential.Provider
	cfg.Intent = webmodel.IntentCreate
	cfg.ConversationID = input.TaskID
	return s.tasks.StartConversationTask(ctx, input.TaskID, input.Query, input.UserID, s.agentFactory, cfg)
}

// PrepareOutline applies the server-side defaults and validates layout IDs.
func (s *TaskService) PrepareOutline(query string, outline *ppt.TaskOutline) (*ppt.TaskOutline, error) {
	if outline == nil || len(outline.Slides) == 0 {
		return outline, nil
	}
	if strings.TrimSpace(outline.Title) == "" {
		outline.Title = strings.TrimSpace(query)
	}
	outline.ContentMode = ppt.OutlineContentModeUserOutline
	for i := range outline.Slides {
		slide := &outline.Slides[i]
		slide.Title = strings.TrimSpace(slide.Title)
		slide.ContentType = strings.TrimSpace(slide.ContentType)
		if s.templateLoader != nil && s.templateLoader.GetLayout(slide.ContentType) == nil {
			return nil, fmt.Errorf("%w: 第%d页 content_type=%q 不存在", ErrInvalidOutline, i+1, slide.ContentType)
		}
	}
	return outline, nil
}

// GetWorkDir resolves an in-memory or persisted task working directory.
func (s *TaskService) GetWorkDir(taskID string) string {
	if s == nil || s.tasks == nil {
		return ""
	}
	return s.tasks.GetWorkDir(taskID)
}

// TaskManager exposes the underlying state manager for handlers that need
// streaming subscriptions or read-only projections. Mutations should prefer
// the methods above.
func (s *TaskService) TaskManager() *task.TaskManager {
	if s == nil {
		return nil
	}
	return s.tasks
}

func normalizeMessageKey(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), ""))
}
