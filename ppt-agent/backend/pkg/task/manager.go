package task

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cloudwego/eino/adk"

	"github.com/cloudwego/ppt-agent/pkg/agent/deep"
	"github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/db"
	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/metrics"
)

// TaskStatus represents the overall status of a PPT generation task.
type TaskStatus string

const (
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
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
	Duration         string          `json:"duration,omitempty"`
	Status           TaskStatus      `json:"status,omitempty"`
	PromptTokens     int64           `json:"prompt_tokens,omitempty"`
	CompletionTokens int64           `json:"completion_tokens,omitempty"`
	TotalTokens      int64           `json:"total_tokens,omitempty"`
}

// TaskInfo is the public-facing summary of a task.
type TaskInfo struct {
	ID         string     `json:"id"`
	UserID     int        `json:"user_id"`
	Query      string     `json:"query"`
	Status     TaskStatus `json:"status"`
	WorkDir    string     `json:"work_dir"`
	CreatedAt  time.Time  `json:"created_at"`
	DoneCount  int        `json:"done_count"`
	TotalCount int        `json:"total_count"`
	Duration         string     `json:"duration,omitempty"`
	Error            string     `json:"error,omitempty"`
	Files            []string   `json:"files,omitempty"`
	PromptTokens     int64      `json:"prompt_tokens"`
	CompletionTokens int64      `json:"completion_tokens"`
	TotalTokens      int64      `json:"total_tokens"`
}

// TaskState holds the internal state of a single task.
type TaskState struct {
	Info          TaskInfo
	Events        []SSERichEvent
	listeners     map[string]chan SSERichEvent
	cancel        context.CancelFunc
	result        *deep.PPTTaskResult
	reportedFiles map[string]bool
	Mu            sync.Mutex
}

func (ts *TaskState) AddListener(id string, ch chan SSERichEvent) {
	ts.Mu.Lock()
	defer ts.Mu.Unlock()
	if ts.listeners == nil {
		ts.listeners = make(map[string]chan SSERichEvent)
	}
	ts.listeners[id] = ch
}

func (ts *TaskState) RemoveListener(id string) {
	ts.Mu.Lock()
	defer ts.Mu.Unlock()
	if ts.listeners != nil {
		delete(ts.listeners, id)
	}
}

func (ts *TaskState) Broadcast(event SSERichEvent) {
	ts.Mu.Lock()
	ts.Events = append(ts.Events, event)
	if len(ts.Events) > 500 {
		ts.Events = ts.Events[len(ts.Events)-300:]
	}
	for _, ch := range ts.listeners {
		select {
		case ch <- event:
		default:
		}
	}
	ts.Mu.Unlock()
}

func (ts *TaskState) Replay(listenerCh chan SSERichEvent) {
	ts.Mu.Lock()
	defer ts.Mu.Unlock()
	if ts.listeners == nil {
		return
	}
	for _, evt := range ts.Events {
		select {
		case listenerCh <- evt:
		default:
			return
		}
	}
}

// TaskManager manages the lifecycle of all PPT generation tasks.
type TaskManager struct {
	mu      sync.RWMutex
	tasks   map[string]*TaskState
	baseDir string
}

// NewTaskManager creates a new TaskManager. baseDir is the parent directory
// under which per-task output directories are created.
// If a MySQL DB is available, previously running tasks are marked as failed
// (the owning process no longer exists).
func NewTaskManager(baseDir string) *TaskManager {
	if db.DB != nil {
		if err := db.MarkZombieTasks(); err != nil {
			logger.Error("mark_zombie_tasks_failed", "error", err.Error())
		}
	}
	return &TaskManager{
		tasks:   make(map[string]*TaskState),
		baseDir: baseDir,
	}
}

// ── DB conversion helpers ────────────────────────────────────────────────

func taskInfoToRecord(info *TaskInfo) *db.TaskRecord {
	filesJSON, _ := json.Marshal(info.Files)
	return &db.TaskRecord{
		ID:               info.ID,
		UserID:           uint(info.UserID),
		Query:            info.Query,
		Status:           string(info.Status),
		WorkDir:          info.WorkDir,
		DoneCount:        info.DoneCount,
		TotalCount:       info.TotalCount,
		Duration:         info.Duration,
		Error:            info.Error,
		Files:            string(filesJSON),
		PromptTokens:     info.PromptTokens,
		CompletionTokens: info.CompletionTokens,
		TotalTokens:      info.TotalTokens,
	}
}

func recordToTaskInfo(r *db.TaskRecord) *TaskInfo {
	var files []string
	json.Unmarshal([]byte(r.Files), &files)
	if files == nil {
		files = []string{}
	}
	return &TaskInfo{
		ID:               r.ID,
		UserID:           int(r.UserID),
		Query:            r.Query,
		Status:           TaskStatus(r.Status),
		WorkDir:          r.WorkDir,
		DoneCount:        r.DoneCount,
		TotalCount:       r.TotalCount,
		Duration:         r.Duration,
		Error:            r.Error,
		Files:            files,
		CreatedAt:        r.CreatedAt,
		PromptTokens:     r.PromptTokens,
		CompletionTokens: r.CompletionTokens,
		TotalTokens:      r.TotalTokens,
	}
}

func (ts *TaskState) persist() {
	if db.DB == nil {
		return
	}
	ts.Mu.Lock()
	r := taskInfoToRecord(&ts.Info)
	ts.Mu.Unlock()
	if err := db.UpdateTaskRecord(r.ID, map[string]any{
		"status":            r.Status,
		"done_count":        r.DoneCount,
		"total_count":       r.TotalCount,
		"duration":          r.Duration,
		"error":             r.Error,
		"files":             r.Files,
		"prompt_tokens":     r.PromptTokens,
		"completion_tokens": r.CompletionTokens,
		"total_tokens":      r.TotalTokens,
	}); err != nil {
		logger.Error("db_persist_failed", "task_id", r.ID, "error", err.Error())
	}
}

// AgentFactory creates an agent for a specific task configuration.
type AgentFactory func(ctx context.Context, cfg *deep.PPTTaskConfig) (adk.Agent, error)

// ErrTaskAlreadyRunning is returned when trying to create a task while another is running.
var ErrTaskAlreadyRunning = fmt.Errorf("已有任务正在执行，请等待当前任务完成后再创建新任务")

// HasRunningTask returns true if the given user already has a running task.
func (tm *TaskManager) HasRunningTask(userID int) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	for _, ts := range tm.tasks {
		if ts.Info.UserID == userID && ts.Info.Status == TaskStatusRunning {
			return true
		}
	}
	return false
}

// CreateTask creates a new task, starts agent execution in a goroutine, and
// returns the task info.
func (tm *TaskManager) CreateTask(ctx context.Context, query string, userID int,
	factory AgentFactory, cfg *deep.PPTTaskConfig) (*TaskInfo, error) {

	if tm.HasRunningTask(userID) {
		return nil, ErrTaskAlreadyRunning
	}

	workDir := filepath.Join(tm.baseDir, fmt.Sprintf("%d-%s", userID, cfg.TaskID))
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
			UserID:    userID,
			Query:     query,
			Status:    TaskStatusRunning,
			WorkDir:   workDir,
			CreatedAt: time.Now(),
		},
		listeners:     make(map[string]chan SSERichEvent),
		reportedFiles: make(map[string]bool),
	}
	agentCtx, cancel := context.WithCancel(context.Background())
	// Attach token tracker to context for callbacks to accumulate usage.
	agentCtx, _ = utils.WithTokenTracker(agentCtx)
	type workDirSetter interface{ SetWorkDir(context.Context, string) context.Context }
	if setter, ok := cfg.Operator.(workDirSetter); ok {
		agentCtx = setter.SetWorkDir(agentCtx, workDir)
	}
	ts.cancel = cancel

	tm.mu.Lock()
	tm.tasks[cfg.TaskID] = ts
	tm.mu.Unlock()

	metrics.RecordTaskCreated()

	if db.DB != nil {
		if err := db.CreateTaskRecord(taskInfoToRecord(&ts.Info)); err != nil {
			logger.Error("task_create_persist_failed", "error", err.Error())
		}
	}

	go tm.runAgent(agentCtx, ts, agent, cfg, query)

	return &ts.Info, nil
}

func (tm *TaskManager) runAgent(ctx context.Context, ts *TaskState, agent adk.Agent,
	cfg *deep.PPTTaskConfig, query string) {

	defer tm.cleanupTask(ts)

	progressCtx, progressCancel := context.WithCancel(ctx)
	defer progressCancel()
	go tm.pollProgress(progressCtx, ts, cfg.WorkDir)

	result, err := deep.RunPPTTaskDeepAgentWithCallback(ctx, agent, cfg, query, func(event deep.AgentEvent) {
		ts.Broadcast(SSERichEvent{
			Type:     event.Type,
			Content:  event.Content,
			ToolName: event.ToolName,
			ToolArgs: event.ToolArgs,
			Error:    event.Error,
		})
	})

	ts.Mu.Lock()
	ts.result = result
	ts.Mu.Unlock()

	ts.Info.Status = TaskStatusCompleted
	if err != nil {
		if ctx.Err() == context.Canceled {
			ts.Info.Status = TaskStatusCancelled
			ts.Info.Error = "任务已被用户中断"
		} else {
			ts.Info.Status = TaskStatusFailed
			ts.Info.Error = err.Error()
		}
	}

	if result != nil {
		ts.Info.DoneCount = result.DoneSlides
		ts.Info.TotalCount = result.TotalSlides
		ts.Info.Duration = result.Duration.Round(time.Millisecond).String()
		ts.Info.Files = result.Files
	}

	// Record Prometheus metrics for task completion.
	durationSeconds := 0.0
	if result != nil {
		durationSeconds = result.Duration.Seconds()
	}
	metrics.RecordTaskCompleted(durationSeconds, ts.Info.DoneCount, ts.Info.TotalCount, string(ts.Info.Status))

	// Collect accumulated token usage from callbacks.
	if tt := utils.TokenTrackerFromContext(ctx); tt != nil {
		p, c, t := tt.TokenTotals()
		ts.Info.PromptTokens = p
		ts.Info.CompletionTokens = c
		ts.Info.TotalTokens = t
	}

	ts.persist()

	finalEvent := SSERichEvent{
		Type:            "complete",
		Status:          ts.Info.Status,
		Message:         ts.Info.Error,
		Done:            ts.Info.DoneCount,
		Total:           ts.Info.TotalCount,
		Files:           ts.Info.Files,
		Duration:        ts.Info.Duration,
		PromptTokens:    ts.Info.PromptTokens,
		CompletionTokens: ts.Info.CompletionTokens,
		TotalTokens:     ts.Info.TotalTokens,
	}
	if result != nil {
		finalEvent.Message = result.Message
	}
	ts.Broadcast(finalEvent)
}

func (tm *TaskManager) pollProgress(ctx context.Context, ts *TaskState, workDir string) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var lastDone int
	var lastTokens int64
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		// —— 实时 token 用量同步 ——
		if tt := utils.TokenTrackerFromContext(ctx); tt != nil {
			p, c, t := tt.TokenTotals()
			if t != lastTokens {
				lastTokens = t
				ts.Mu.Lock()
				ts.Info.PromptTokens = p
				ts.Info.CompletionTokens = c
				ts.Info.TotalTokens = t
				ts.Mu.Unlock()
				ts.Broadcast(SSERichEvent{
					Type:            "token_usage",
					PromptTokens:    p,
					CompletionTokens: c,
					TotalTokens:     t,
				})
			}
		}

		entries, err := os.ReadDir(workDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				if filepath.Ext(entry.Name()) == ".pptx" {
					ts.Mu.Lock()
					if !ts.reportedFiles[entry.Name()] {
						ts.reportedFiles[entry.Name()] = true
						ts.Mu.Unlock()
						ts.Broadcast(SSERichEvent{
							Type:     "file_ready",
							ToolName: entry.Name(),
							Files:    []string{entry.Name()},
						})
						continue
					}
					ts.Mu.Unlock()
				}
			}
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

		ts.Broadcast(SSERichEvent{
			Type:  "progress",
			Tasks: manifest.Tasks,
			Done:  done,
			Total: total,
		})
	}
}

func (tm *TaskManager) cleanupTask(ts *TaskState) {
	ts.Mu.Lock()
	defer ts.Mu.Unlock()
	for _, ch := range ts.listeners {
		close(ch)
	}
	ts.listeners = nil

	// Schedule removal from memory after 1 hour (MySQL + NewColdTaskState handles replay after that)
	id := ts.Info.ID
	time.AfterFunc(1*time.Hour, func() {
		tm.mu.Lock()
		if t, ok := tm.tasks[id]; ok && t != nil && t.Info.Status != TaskStatusRunning {
			delete(tm.tasks, id)
		}
		tm.mu.Unlock()
	})
}

// GetTask returns the stored task info, or nil if not found.
func (tm *TaskManager) GetTask(id string) *TaskInfo {
	tm.mu.RLock()
	ts := tm.tasks[id]
	tm.mu.RUnlock()
	if ts != nil {
		info := ts.Info
		return &info
	}
	if db.DB != nil {
		r, err := db.GetTaskRecord(id)
		if err == nil {
			return recordToTaskInfo(r)
		}
	}
	return nil
}

// GetTaskState returns the internal task state, or nil if not found.
func (tm *TaskManager) GetTaskState(id string) *TaskState {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.tasks[id]
}

// NewColdTaskState creates a minimal TaskState from a TaskInfo (no events, no cancel).
// Used to restore tasks from MySQL after a backend restart.
func (tm *TaskManager) NewColdTaskState(info TaskInfo) *TaskState {
	ts := &TaskState{
		Info:          info,
		listeners:     make(map[string]chan SSERichEvent),
		reportedFiles: make(map[string]bool),
	}
	tm.mu.Lock()
	tm.tasks[info.ID] = ts
	tm.mu.Unlock()
	return ts
}

// ListTasks returns summaries of tasks for the given user, newest first.
func (tm *TaskManager) ListTasks(userID int) []TaskInfo {
	seen := make(map[string]bool)
	var result []TaskInfo

	tm.mu.RLock()
	for _, ts := range tm.tasks {
		if ts.Info.UserID == userID {
			result = append(result, ts.Info)
			seen[ts.Info.ID] = true
		}
	}
	tm.mu.RUnlock()

	if db.DB != nil {
		records, err := db.ListTaskRecordsByUser(uint(userID))
		if err == nil {
			for i := len(records) - 1; i >= 0; i-- {
				if !seen[records[i].ID] {
					result = append(result, *recordToTaskInfo(&records[i]))
				}
			}
		}
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result
}

// CancelTask cancels a running task. Returns true if the task was found and running.
func (tm *TaskManager) CancelTask(id string) bool {
	tm.mu.RLock()
	ts := tm.tasks[id]
	tm.mu.RUnlock()
	if ts == nil || ts.cancel == nil {
		return false
	}
	ts.Mu.Lock()
	if ts.Info.Status == TaskStatusRunning {
		ts.Info.Status = TaskStatusCancelled
	}
	ts.Mu.Unlock()
	ts.persist()
	ts.cancel()
	return true
}

// DeleteTask removes a task from memory, the DB record, and its output directory.
// Running tasks are cancelled before deletion.
func (tm *TaskManager) DeleteTask(id string) error {
	ts := tm.GetTaskState(id)

	// If running, cancel first.
	if ts != nil && ts.Info.Status == TaskStatusRunning {
		tm.CancelTask(id)
		// Wait briefly for the goroutine to wind down.
		time.Sleep(100 * time.Millisecond)
	}

	// Remove from in-memory map.
	tm.mu.Lock()
	delete(tm.tasks, id)
	tm.mu.Unlock()

	// Delete DB record.
	if db.DB != nil {
		if err := db.DeleteTaskRecord(id); err != nil {
			logger.Error("db_delete_task_failed", "task_id", id, "error", err.Error())
		}
	}

	// Remove output directory.
	var workDir string
	if ts != nil {
		workDir = ts.Info.WorkDir
	} else {
		// Task no longer in memory (e.g., server restarted). Look up from DB.
		if db.DB != nil {
			if r, err := db.GetTaskRecord(id); err == nil {
				workDir = r.WorkDir
			}
		}
	}
	if workDir != "" {
		if err := os.RemoveAll(workDir); err != nil {
			logger.Error("delete_workdir_failed", "path", workDir, "error", err.Error())
			return fmt.Errorf("删除工作目录失败: %w", err)
		}
	}

	return nil
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
