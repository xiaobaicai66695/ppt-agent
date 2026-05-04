package server

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cloudwego/eino/adk"

	"github.com/cloudwego/ppt-agent/pkg/agent/deep"
)

// TaskStatus represents the overall status of a PPT generation task.
type TaskStatus string

const (
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

// SSERichEvent is an enriched event for SSE streaming. It wraps agent-level
// AgentEvent with additional progress and lifecycle information.
type SSERichEvent struct {
	Type     string          `json:"type"`
	Content  string          `json:"content,omitempty"`
	ToolName string          `json:"tool_name,omitempty"`
	ToolArgs string          `json:"tool_args,omitempty"`
	Error    string          `json:"error,omitempty"`
	Tasks    []*deep.TaskItem `json:"tasks,omitempty"`
	Done     int             `json:"done,omitempty"`
	Total    int             `json:"total,omitempty"`
	Files    []string        `json:"files,omitempty"`
	Message  string          `json:"message,omitempty"`
	Duration string          `json:"duration,omitempty"`
	Status   TaskStatus      `json:"status,omitempty"`
}

// TaskInfo is the public-facing summary of a task.
type TaskInfo struct {
	ID         string     `json:"id"`
	Query      string     `json:"query"`
	Status     TaskStatus `json:"status"`
	WorkDir    string     `json:"work_dir"`
	CreatedAt  time.Time  `json:"created_at"`
	DoneCount  int        `json:"done_count"`
	TotalCount int        `json:"total_count"`
	Duration   string     `json:"duration,omitempty"`
	Error      string     `json:"error,omitempty"`
	Files      []string   `json:"files,omitempty"`
}

// TaskState holds the internal state of a single task.
type TaskState struct {
	Info      TaskInfo
	events    []SSERichEvent
	listeners map[string]chan SSERichEvent
	cancel    context.CancelFunc
	result    *deep.PPTTaskResult
	mu        sync.Mutex
}

func (ts *TaskState) addListener(id string, ch chan SSERichEvent) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.listeners[id] = ch
}

func (ts *TaskState) removeListener(id string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	delete(ts.listeners, id)
}

func (ts *TaskState) broadcast(event SSERichEvent) {
	ts.mu.Lock()
	ts.events = append(ts.events, event)
	if len(ts.events) > 500 {
		ts.events = ts.events[len(ts.events)-300:]
	}
	for _, ch := range ts.listeners {
		select {
		case ch <- event:
		default:
		}
	}
	ts.mu.Unlock()
}

func (ts *TaskState) replay(listenerCh chan SSERichEvent) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	for _, evt := range ts.events {
		select {
		case listenerCh <- evt:
		default:
			return
		}
	}
}

// TaskManager manages the lifecycle of all PPT generation tasks.
type TaskManager struct {
	mu       sync.RWMutex
	tasks    map[string]*TaskState
	baseDir  string
}

// NewTaskManager creates a new TaskManager. baseDir is the parent directory
// under which per-task output directories are created.
func NewTaskManager(baseDir string) *TaskManager {
	return &TaskManager{
		tasks:   make(map[string]*TaskState),
		baseDir: baseDir,
	}
}

// AgentFactory creates an agent for a specific task configuration.
type AgentFactory func(ctx context.Context, cfg *deep.PPTTaskConfig) (adk.Agent, error)

// CreateTask creates a new task, starts agent execution in a goroutine, and
// returns the task info.
func (tm *TaskManager) CreateTask(ctx context.Context, query string, factory AgentFactory,
	cfg *deep.PPTTaskConfig) (*TaskInfo, error) {

	workDir := filepath.Join(tm.baseDir, cfg.TaskID)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, err
	}
	cfg.WorkDir = workDir

	agent, err := factory(ctx, cfg)
	if err != nil {
		return nil, err
	}

	ts := &TaskState{
		Info: TaskInfo{
			ID:        cfg.TaskID,
			Query:     query,
			Status:    TaskStatusRunning,
			WorkDir:   workDir,
			CreatedAt: time.Now(),
		},
		listeners: make(map[string]chan SSERichEvent),
	}
	ctx, ts.cancel = context.WithCancel(ctx)

	tm.mu.Lock()
	tm.tasks[cfg.TaskID] = ts
	tm.mu.Unlock()

	go tm.runAgent(ctx, ts, agent, cfg, query)

	return &ts.Info, nil
}

func (tm *TaskManager) runAgent(ctx context.Context, ts *TaskState, agent adk.Agent,
	cfg *deep.PPTTaskConfig, query string) {

	defer tm.cleanupTask(ts)

	// Start progress poller.
	progressCtx, progressCancel := context.WithCancel(ctx)
	defer progressCancel()
	go tm.pollProgress(progressCtx, ts, cfg.WorkDir)

	result, err := deep.RunPPTTaskDeepAgentWithCallback(ctx, agent, cfg, query, func(event deep.AgentEvent) {
		ts.broadcast(SSERichEvent{
			Type:     event.Type,
			Content:  event.Content,
			ToolName: event.ToolName,
			ToolArgs: event.ToolArgs,
			Error:    event.Error,
		})
	})

	ts.mu.Lock()
	ts.result = result
	ts.mu.Unlock()

	ts.Info.Status = TaskStatusCompleted
	if err != nil {
		ts.Info.Status = TaskStatusFailed
		ts.Info.Error = err.Error()
	}

	if result != nil {
		ts.Info.DoneCount = result.DoneSlides
		ts.Info.TotalCount = result.TotalSlides
		ts.Info.Duration = result.Duration.Round(time.Millisecond).String()
		ts.Info.Files = result.Files
	}

	// Send a final event.
	finalEvent := SSERichEvent{
		Type:     "complete",
		Status:   ts.Info.Status,
		Message:  ts.Info.Error,
		Done:     ts.Info.DoneCount,
		Total:    ts.Info.TotalCount,
		Files:    ts.Info.Files,
		Duration: ts.Info.Duration,
	}
	if result != nil {
		finalEvent.Message = result.Message
	}
	ts.broadcast(finalEvent)
}

func (tm *TaskManager) pollProgress(ctx context.Context, ts *TaskState, workDir string) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var lastDone int
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		manifest, err := deep.ReadTasksManifest(workDir)
		if err != nil || manifest == nil {
			continue
		}

		done := manifest.CompletedCount()
		total := len(manifest.Tasks)
		if done == lastDone && total > 0 {
			continue
		}
		lastDone = done

		ts.broadcast(SSERichEvent{
			Type:  "progress",
			Tasks: manifest.Tasks,
			Done:  done,
			Total: total,
		})
	}
}

func (tm *TaskManager) cleanupTask(ts *TaskState) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	for _, ch := range ts.listeners {
		close(ch)
	}
	ts.listeners = nil
}

// GetTask returns the stored task info, or nil if not found.
func (tm *TaskManager) GetTask(id string) *TaskInfo {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	ts := tm.tasks[id]
	if ts == nil {
		return nil
	}
	info := ts.Info
	return &info
}

// GetTaskState returns the internal task state, or nil if not found.
func (tm *TaskManager) GetTaskState(id string) *TaskState {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.tasks[id]
}

// ListTasks returns summaries of all known tasks, newest first.
func (tm *TaskManager) ListTasks() []TaskInfo {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	var result []TaskInfo
	for _, ts := range tm.tasks {
		result = append(result, ts.Info)
	}
	// Reverse: newest first.
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result
}

// CancelTask cancels a running task.
func (tm *TaskManager) CancelTask(id string) bool {
	tm.mu.RLock()
	ts := tm.tasks[id]
	tm.mu.RUnlock()
	if ts == nil || ts.cancel == nil {
		return false
	}
	ts.cancel()
	return true
}

// ReadTasksManifestFile reads and returns the tasks.json for a task.
func (tm *TaskManager) ReadTasksManifestFile(id string) (*deep.TasksManifest, error) {
	ts := tm.GetTaskState(id)
	if ts == nil {
		return nil, os.ErrNotExist
	}
	return deep.ReadTasksManifest(ts.Info.WorkDir)
}

// TaskFilesAsJSON returns the task's tasks.json as raw JSON bytes.
func (tm *TaskManager) TaskFilesAsJSON(workDir string) ([]byte, error) {
	return json.Marshal(struct {
		Tasks []*deep.TaskItem `json:"tasks"`
	}{
		Tasks: nil,
	})
}
