package task

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/adk"

	"github.com/cloudwego/ppt-agent/pkg/agent/deep"
	agentlearning "github.com/cloudwego/ppt-agent/pkg/agent/learning"
	"github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/db"
	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/metrics"
	"github.com/cloudwego/ppt-agent/pkg/session"
)

// TaskStatus 表示 PPT 生成任务的总体状态。
type TaskStatus string

const (
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// SSERichEvent 是 SSE 流式传输的增强事件。它封装了 agent 级别的
// AgentEvent，并附带额外的进度和生命周期信息。
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

// TaskInfo 是任务的公开可见摘要。
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
	ConversationContent string   `json:"conversation_content,omitempty"` // 拼接后的对话内容
}

// TaskState 保存单个任务的内部状态。
type TaskState struct {
	Info                 TaskInfo
	Events               []SSERichEvent
	listeners            map[string]chan SSERichEvent
	cancel               context.CancelFunc
	result               *deep.PPTTaskResult
	reportedFiles        map[string]bool
	Mu                   sync.Mutex

	// pendingContinueMsg 任务运行中时，用户提交的待处理消息（消费后清空）
	pendingContinueMsg string
	// pendingContinueQueued 已通知前端排队（避免重复通知）
	pendingContinueQueued bool
}

// Persist 将任务状态持久化到数据库。
func (ts *TaskState) Persist() {
	ts.persist()
}

// ReportedFiles 返回已上报文件的集合。
func (ts *TaskState) ReportedFiles() map[string]bool {
	return ts.reportedFiles
}

// SetReportedFile 将文件标记为已上报。
func (ts *TaskState) SetReportedFile(name string) {
	ts.reportedFiles[name] = true
}

// HasPendingContinueMsg 检查是否有等待处理的消息。
func (ts *TaskState) HasPendingContinueMsg() bool {
	ts.Mu.Lock()
	defer ts.Mu.Unlock()
	return ts.pendingContinueMsg != ""
}

// GetPendingContinueMsg 取出并清空等待消息。
func (ts *TaskState) GetPendingContinueMsg() string {
	ts.Mu.Lock()
	defer ts.Mu.Unlock()
	msg := ts.pendingContinueMsg
	ts.pendingContinueMsg = ""
	ts.pendingContinueQueued = false
	return msg
}

// SetPendingContinueMsg 设置等待消息（仅当当前为空时）。
// 返回是否设置成功（false 表示已有等待消息）。
func (ts *TaskState) SetPendingContinueMsg(msg string) bool {
	ts.Mu.Lock()
	defer ts.Mu.Unlock()
	if ts.pendingContinueMsg != "" {
		return false
	}
	ts.pendingContinueMsg = msg
	ts.pendingContinueQueued = false
	return true
}

// IsPendingContinueMsgFirst 检查当前等待消息是否为第一条（即之前没有排队）。
func (ts *TaskState) IsPendingContinueMsgFirst() bool {
	ts.Mu.Lock()
	defer ts.Mu.Unlock()
	return ts.pendingContinueQueued
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

// TaskManager 管理所有 PPT 生成任务的生命周期。
type TaskManager struct {
	mu               sync.RWMutex
	tasks            map[string]*TaskState
	baseDir          string
	onTaskComplete   func(userID int, workDir string, query string)
	onTaskFailed     func(taskID string)
	onTaskContinue   func(taskID string) // 任务完成且有待处理消息时触发
}

// NewTaskManager 创建一个新的 TaskManager。baseDir 是父目录，
// 每个任务的输出目录都创建在其下。
// 如果 MySQL 数据库可用，之前运行中的任务会被标记为失败
// （因为拥有它们的进程已不存在）。
func NewTaskManager(baseDir string, onTaskComplete func(userID int, workDir string, query string), onTaskFailed func(taskID string), onTaskContinue func(taskID string)) *TaskManager {
	if db.DB != nil {
		if err := db.MarkZombieTasks(); err != nil {
			logger.Error("mark_zombie_tasks_failed", "error", err.Error())
		}
	}
	return &TaskManager{
		tasks:          make(map[string]*TaskState),
		baseDir:        baseDir,
		onTaskComplete: onTaskComplete,
		onTaskFailed:   onTaskFailed,
	}
}

// ── 数据库转换辅助函数 ────────────────────────────────────────────────

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
		ConversationContent: info.ConversationContent,
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
		ConversationContent: r.ConversationContent,
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
		"status":                r.Status,
		"done_count":            r.DoneCount,
		"total_count":           r.TotalCount,
		"duration":              r.Duration,
		"error":                 r.Error,
		"files":                 r.Files,
		"prompt_tokens":          r.PromptTokens,
		"completion_tokens":      r.CompletionTokens,
		"total_tokens":           r.TotalTokens,
		"conversation_content":   r.ConversationContent,
	}); err != nil {
		logger.Error("db_persist_failed", "task_id", r.ID, "error", err.Error())
	}
}

// AgentFactory 为特定任务配置创建 agent。
type AgentFactory func(ctx context.Context, cfg *deep.PPTTaskConfig) (adk.Agent, error)

// ErrTaskAlreadyRunning 当尝试创建任务时如果另一个任务正在运行，则返回此错误。
var ErrTaskAlreadyRunning = fmt.Errorf("已有任务正在执行，请等待当前任务完成后再创建新任务")

// HasRunningTask 如果给定用户已有运行中的任务则返回 true。
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

// HasRunningTasks 如果有任何任务正在运行则返回 true。
func (tm *TaskManager) HasRunningTasks() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	for _, ts := range tm.tasks {
		if ts.Info.Status == TaskStatusRunning {
			return true
		}
	}
	return false
}

// CreateTask 创建一个新任务，启动 agent 执行（在一个 goroutine 中），
// 并返回任务信息。
func (tm *TaskManager) CreateTask(ctx context.Context, query string, userID int,
	factory AgentFactory, cfg *deep.PPTTaskConfig) (*TaskInfo, error) {

	if tm.HasRunningTask(userID) {
		return nil, ErrTaskAlreadyRunning
	}

	// ── 意图识别与路由 ──
	// 如果配置中没有意图结果，进行意图识别
	if cfg.IntentResult == nil {
		intentCfg, err := deep.ProcessUserIntent(ctx, query, userID)
		if err != nil {
			logger.Warn("intent_process_failed", "error", err.Error())
		} else if intentCfg != nil {
			// 合并意图识别结果
			if intentCfg.IntentResult != nil {
				cfg.IntentResult = intentCfg.IntentResult
			}
			if intentCfg.RoutingDecision != nil {
				cfg.RoutingDecision = intentCfg.RoutingDecision
			}
			if intentCfg.EnhancedProfile != nil {
				cfg.EnhancedProfile = intentCfg.EnhancedProfile
			}
			// 如果有推荐的样式上下文，追加到现有上下文
			if intentCfg.StyleContext != "" {
				if cfg.StyleContext != "" {
					cfg.StyleContext += "\n" + intentCfg.StyleContext
				} else {
					cfg.StyleContext = intentCfg.StyleContext
				}
			}
		}
	}

	workDir := filepath.Join(tm.baseDir, fmt.Sprintf("%d-%s", userID, cfg.TaskID))
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, err
	}
	cfg.WorkDir = workDir

	// 如果用户提供了 outline，先写入 tasks.json，跳过 AI 规划阶段
	if cfg.Outline != nil && len(cfg.Outline.Slides) > 0 {
		manifest := outlineToManifest(cfg.Outline, workDir)
		if err := deep.WriteTasksManifest(workDir, manifest); err != nil {
			return nil, fmt.Errorf("写入大纲失败: %w", err)
		}
	}

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
	defer func() {
		if r := recover(); r != nil {
			logger.Error("agent_panic", "task_id", ts.Info.ID, "panic", fmt.Sprintf("%v", r))
			ts.Mu.Lock()
			ts.Info.Status = TaskStatusFailed
			ts.Info.Error = fmt.Sprintf("agent internal panic: %v", r)
			ts.Mu.Unlock()
			ts.Broadcast(SSERichEvent{
				Type:   "complete",
				Status: ts.Info.Status,
				Error:  ts.Info.Error,
			})
		}
	}()

	progressCtx, progressCancel := context.WithCancel(ctx)
	defer progressCancel()
	go tm.pollProgress(progressCtx, ts, cfg.WorkDir)

	result, err := deep.RunPPTTaskDeepAgentWithCallback(ctx, agent, cfg, query, func(event deep.AgentEvent) {
		if event.Type != "tool_call" && event.Type != "token_usage" {
			ts.Broadcast(SSERichEvent{
				Type:     event.Type,
				Content:  event.Content,
				ToolName: event.ToolName,
				ToolArgs: event.ToolArgs,
				Error:    event.Error,
			})
		}
	})

	ts.Mu.Lock()
	ts.result = result
	ts.Mu.Unlock()

	ts.Info.Status = TaskStatusCompleted
	var taskFailed bool
	if err != nil {
		if ctx.Err() == context.Canceled {
			ts.Info.Status = TaskStatusCancelled
			ts.Info.Error = "任务已被用户中断"
		} else {
			ts.Info.Status = TaskStatusFailed
			ts.Info.Error = err.Error()
			taskFailed = true
		}
	}

	if result != nil {
		ts.Info.DoneCount = result.DoneSlides
		ts.Info.TotalCount = result.TotalSlides
		ts.Info.Duration = result.Duration.Round(time.Millisecond).String()
		ts.Info.Files = result.Files
	}

	// 记录任务完成的 Prometheus 指标。
	durationSeconds := 0.0
	if result != nil {
		durationSeconds = result.Duration.Seconds()
	}
	metrics.RecordTaskCompleted(durationSeconds, ts.Info.DoneCount, ts.Info.TotalCount, string(ts.Info.Status))

	// 从回调中收集累积的 token 使用量。
	if tt := utils.TokenTrackerFromContext(ctx); tt != nil {
		p, c, t := tt.TokenTotals()
		ts.Info.PromptTokens = p
		ts.Info.CompletionTokens = c
		ts.Info.TotalTokens = t
	}

	ts.persist()

	// 任务完成后，检查是否有等待中的继续消息，如有则触发自动继续处理
	if pendingMsg := ts.GetPendingContinueMsg(); pendingMsg != "" {
		ts.Broadcast(SSERichEvent{
			Type:    "continue_queued",
			Content: pendingMsg,
		})
		// 通过回调通知 Server 启动继续流程
		if tm.onTaskContinue != nil {
			go tm.onTaskContinue(ts.Info.ID)
		}
	}

	// 触发失败日志分析（异步，不阻塞任务完成通知）
	if taskFailed && tm.onTaskFailed != nil {
		go tm.onTaskFailed(ts.Info.ID)
	}

	// 提取对话内容用于存储（包含任务摘要、幻灯片概览、有意义的对话）
	ts.persistConversationContent()

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

	// 触发任务完成回调，更新用户风格偏好
	if tm.onTaskComplete != nil && ts.Info.UserID > 0 && ts.Info.Status == TaskStatusCompleted {
		go tm.onTaskComplete(ts.Info.UserID, ts.Info.WorkDir, ts.Info.Query)
	}

	// ── 记录学习信号 ──
	// 从 tasks.json 中收集 QA 结果，计算真实质量分数
	qualityScore := calculateQualityScoreFromQA(ts.Info.WorkDir)

	deep.UpdateUserProfileFromTask(ts.Info.UserID, &agentlearning.TaskContext{
		TaskID:       ts.Info.ID,
		UserID:      ts.Info.UserID,
		Duration:    result.Duration,
		Success:     ts.Info.Status == TaskStatusCompleted,
		QualityScore: qualityScore,
		PageCount:   ts.Info.TotalCount,
	})
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
	for _, ch := range ts.listeners {
		close(ch)
	}
	ts.listeners = nil

	// Flush conversation content to DB (once, on task end — not mid-stream)
	if db.DB != nil && ts.Info.ConversationContent != "" {
		go func() {
			if err := db.UpdateTaskRecord(ts.Info.ID, map[string]any{
				"conversation_content": ts.Info.ConversationContent,
			}); err != nil {
				logger.Error("persist_conversation_content_failed", "task_id", ts.Info.ID, "error", err.Error())
			}
		}()
	}

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

// GetTask 返回存储的任务信息，如果未找到则返回 nil。
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

// GetTaskState 返回内部任务状态，如果未找到则返回 nil。
func (tm *TaskManager) GetTaskState(id string) *TaskState {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.tasks[id]
}

// NewColdTaskState 从 TaskInfo 创建一个最小的 TaskState（无 events，无 cancel）。
// 用于在服务器重启后从 MySQL 恢复任务。
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

// ListTasks 返回给定用户任务的摘要列表，按最新时间排序。
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

	return result
}

// CancelTask 取消一个运行中的任务。如果任务被找到且正在运行则返回 true。
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

// DeleteTask 从内存、数据库记录和输出目录中删除任务。
// 删除前会先取消运行中的任务。
func (tm *TaskManager) DeleteTask(id string) error {
	ts := tm.GetTaskState(id)

	// 如果正在运行，先取消。
	if ts != nil && ts.Info.Status == TaskStatusRunning {
		tm.CancelTask(id)
		// 短暂等待，让 goroutine 关闭。
		time.Sleep(100 * time.Millisecond)
	}

	// 从内存 map 中移除。
	tm.mu.Lock()
	delete(tm.tasks, id)
	tm.mu.Unlock()

	// 删除数据库记录。
	if db.DB != nil {
		if err := db.DeleteTaskRecord(id); err != nil {
			logger.Error("db_delete_task_failed", "task_id", id, "error", err.Error())
		}
		// 级联删除会话消息。
		if err := session.DeleteSessionFromDB(id); err != nil {
			logger.Error("db_delete_conversation_failed", "task_id", id, "error", err.Error())
		}
	}

	// 删除输出目录。
	var workDir string
	if ts != nil {
		workDir = ts.Info.WorkDir
	} else {
		// 任务不在内存中（例如服务器重启了）。从数据库查找。
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

// ReadTasksManifestFile 读取并返回任务的 tasks.json。
func (tm *TaskManager) ReadTasksManifestFile(id string) (*deep.TasksManifest, error) {
	ts := tm.GetTaskState(id)
	if ts == nil {
		return nil, os.ErrNotExist
	}
	return deep.ReadTasksManifest(ts.Info.WorkDir)
}

// TaskFilesAsJSON 返回任务的 tasks.json 作为原始 JSON 字节。
func (tm *TaskManager) TaskFilesAsJSON(workDir string) ([]byte, error) {
	return json.Marshal(struct {
		Tasks []*deep.TaskItem `json:"tasks"`
	}{
		Tasks: nil,
	})
}

// outlineToManifest 将用户编排的 outline 转换为 TasksManifest
func outlineToManifest(outline *deep.TaskOutline, workDir string) *deep.TasksManifest {
	tasks := make([]*deep.TaskItem, 0, len(outline.Slides))
	for i, slide := range outline.Slides {
		safeTitle := sanitizeFilename(slide.Title)
		item := &deep.TaskItem{
			TaskID:      fmt.Sprintf("slide-%d", i+1),
			PageIndex:   i + 1,
			Title:       slide.Title,
			ContentType: slide.ContentType,
			Description: slide.Description,
			OutputFile:  fmt.Sprintf("%d_%s.pptx", i+1, safeTitle),
			Status:      deep.StatusPending,
			CreatedAt:   time.Now().Format(time.RFC3339),
			Background:  slide.Background,
		}
		// Carry through content_plan if present
		if slide.ContentPlan != nil {
			item.ContentPlan = &deep.ContentPlan{
				Summary:  slide.ContentPlan.Summary,
				Elements: make([]deep.ContentElement, len(slide.ContentPlan.Elements)),
			}
			copy(item.ContentPlan.Elements, slide.ContentPlan.Elements)
		}
		tasks = append(tasks, item)
	}
	return &deep.TasksManifest{
		Title:    outline.Title,
		Theme:    outline.Theme,
		Template: outline.Template,
		Tasks:    tasks,
	}
}

// persistConversationContent 从对话消息表（数据库）和内存中的 SSE 事件流中提取有意义的内容，
// 将它们合并为可读的对话文本，并存储到 ts.Info 中以便后续持久化到数据库。
// 在任务结束时调用（成功、取消或错误）。
func (ts *TaskState) persistConversationContent() {
	ts.Mu.Lock()
	defer ts.Mu.Unlock()

	// 从数据库加载所有对话消息（覆盖 continue 轮次）
	var dbMessages []db.ConversationMessage
	if db.DB != nil {
		var err error
		dbMessages, err = db.ListConversationMessages(ts.Info.ID)
		if err != nil {
			logger.Warn("load_conversation_messages_failed", "task_id", ts.Info.ID, "error", err.Error())
		}
	}

	var b strings.Builder
	b.WriteString("## 任务信息\n")
	b.WriteString(fmt.Sprintf("- 状态: %s | 进度: %d/%d | 耗时: %s\n",
		ts.Info.Status, ts.Info.DoneCount, ts.Info.TotalCount, ts.Info.Duration))
	if ts.Info.Error != "" {
		b.WriteString(fmt.Sprintf("- 错误: %s\n", ts.Info.Error))
	}

	manifest, err := deep.ReadTasksManifest(ts.Info.WorkDir)
	if err == nil && manifest != nil && len(manifest.Tasks) > 0 {
		b.WriteString("\n## 幻灯片概览\n")
		b.WriteString("| # | 标题 | 类型 | 状态 | QA |\n")
		b.WriteString("|---|------|------|------|----|\n")
		for _, t := range manifest.Tasks {
			qa := ""
			if t.QAReport != "" {
				qa = "有"
			}
			b.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s |\n",
				t.PageIndex, t.Title, t.ContentType, t.Status, qa))
		}
	}

	b.WriteString("\n## 对话内容\n")

	// 构建一个集合，追踪已从 SSE 添加的答案/错误内容（与数据库去重）
	sseAdded := make(map[string]bool)

	// 1. 先追加内存中的 SSE 事件（当前轮次，可能尚未写入数据库）
	for _, e := range ts.Events {
		switch e.Type {
		case "answer":
			trimmed := strings.TrimSpace(e.Content)
			if trimmed == "" || trimmed == "..." || trimmed == "……" {
				continue
			}
			if len(trimmed) > 2000 {
				trimmed = trimmed[:2000] + "\n...(内容已截断)"
			}
			sseAdded[trimmed] = true
			b.WriteString(fmt.Sprintf("\n**助手**: %s\n", trimmed))
		case "error":
			trimmed := strings.TrimSpace(e.Content)
			if trimmed != "" {
				sseAdded[trimmed] = true
				b.WriteString(fmt.Sprintf("\n**错误**: %s\n", trimmed))
			}
		case "complete":
			if e.Message != "" && e.Message != ts.Info.Error {
				b.WriteString(fmt.Sprintf("\n**完成摘要**: %s\n", e.Message))
			}
			if len(e.Files) > 0 {
				b.WriteString(fmt.Sprintf("\n**生成文件** (%d个): %s\n", len(e.Files), strings.Join(e.Files, ", ")))
			}
		}
	}

	// 2. 追加不在 SSE 事件中的数据库消息
	for _, m := range dbMessages {
		trimmed := strings.TrimSpace(m.Content)
		if trimmed == "" || trimmed == "..." || trimmed == "……" {
			continue
		}
		// 如果此确切内容已从 SSE 添加，则跳过
		if sseAdded[trimmed] {
			continue
		}
		if len(trimmed) > 2000 {
			trimmed = trimmed[:2000] + "\n...(内容已截断)"
		}
		roleTag := "**助手**"
		if m.Role == "user" {
			roleTag = "**用户**"
		}
		b.WriteString(fmt.Sprintf("\n%s: %s\n", roleTag, trimmed))
	}

	ts.Info.ConversationContent = b.String()
}

// BuildContinueContext 构建用于任务继续的紧凑 LLM 上下文。
// 它包括：任务摘要 → tasks.json 快照 → 最后几轮对话。
func (tm *TaskManager) BuildContinueContext(taskID string, lastMessages int) string {
	var b strings.Builder
	ts := tm.GetTaskState(taskID)
	if ts == nil {
		return ""
	}
	ts.Mu.Lock()
	defer ts.Mu.Unlock()

	b.WriteString("## 任务摘要\n")
	b.WriteString(fmt.Sprintf("- 状态: %s | 进度: %d/%d | 耗时: %s\n",
		ts.Info.Status, ts.Info.DoneCount, ts.Info.TotalCount, ts.Info.Duration))
	if ts.Info.Error != "" {
		b.WriteString(fmt.Sprintf("- 错误: %s\n", ts.Info.Error))
	}

	// tasks.json 快照
	manifest, err := deep.ReadTasksManifest(ts.Info.WorkDir)
	if err == nil && manifest != nil {
		b.WriteString("\n## 当前页面列表\n")
		b.WriteString("| # | 标题 | 类型 | 状态 | QA |\n")
		b.WriteString("|---|------|------|------|----|\n")
		for _, t := range manifest.Tasks {
			qa := ""
			if t.QAReport != "" {
				qa = "有报告"
			}
			b.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s |\n",
				t.PageIndex, t.Title, t.ContentType, t.Status, qa))
		}
	}

	// 最近 N 轮对话（仅答案/错误事件）
	b.WriteString(fmt.Sprintf("\n## 最近 %d 轮对话\n", lastMessages))
	count := 0
	for i := len(ts.Events) - 1; i >= 0 && count < lastMessages; i-- {
		e := ts.Events[i]
		if e.Type == "answer" || e.Type == "error" {
			prefix := "assistant"
			if e.Type == "error" {
				prefix = "error"
			}
			content := e.Content
			if len(content) > 500 {
				content = content[:500] + "..."
			}
			b.WriteString(fmt.Sprintf("[%s]: %s\n", prefix, content))
			count++
		}
	}

	return b.String()
}

// sanitizeFilename 移除/替换文件名中有问题的字符
func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_", "?", "_",
		"\"", "_", "<", "_", ">", "_", "|", "_", "\n", "_",
	)
	return replacer.Replace(name)
}

// calculateQualityScoreFromQA 从 tasks.json 中的 QA 结果计算质量分数
// 返回 0-5 的分数，0 表示没有 QA 结果
func calculateQualityScoreFromQA(workDir string) float64 {
	manifest, err := deep.ReadTasksManifest(workDir)
	if err != nil || manifest == nil || len(manifest.Tasks) == 0 {
		return 4.0 // 没有 QA 结果时返回默认值
	}

	var totalScore float64
	var qaCount int
	var highIssueCount int

	for _, task := range manifest.Tasks {
		// 只处理已 QA 的任务
		if task.QAReport == "" {
			continue
		}

		// 从 QA 报告中提取评分信息
		score := extractScoreFromQAReport(task.QAReport)
		if score > 0 {
			totalScore += score
			qaCount++
		}

		// 检查是否有 high 级别问题
		if strings.Contains(task.QAReport, "high") || strings.Contains(task.QAReport, "严重") {
			highIssueCount++
		}
	}

	if qaCount == 0 {
		return 4.0 // 没有有效的 QA 结果
	}

	// 计算平均分
	avgScore := totalScore / float64(qaCount)

	// 如果有 high 级别问题，降低分数
	if highIssueCount > 0 {
		penalty := float64(highIssueCount) * 0.5
		if penalty > 2.0 {
			penalty = 2.0
		}
		avgScore -= penalty
	}

	// 确保分数在有效范围内
	if avgScore < 1.0 {
		avgScore = 1.0
	}
	if avgScore > 5.0 {
		avgScore = 5.0
	}

	return avgScore
}

// extractScoreFromQAReport 从 QA 报告文本中提取评分
// 报告格式可能是自然语言，需要从中提取 1-5 的评分
func extractScoreFromQAReport(report string) float64 {
	// 尝试匹配常见的评分模式
	patterns := []string{
		`(?i)score[:\s]*(\d+(?:\.\d+)?)`,                          // score: 4.5 或 score 4
		`(?i)评分[:\s]*(\d+(?:\.\d+)?)`,                            // 评分: 4
		`(?i)rating[:\s]*(\d+(?:\.\d+)?)`,                          // rating: 4
		`(?i)quality[:\s]*(\d+(?:\.\d+)?)`,                         // quality: 4
		`(?i)质量[:\s]*(\d+(?:\.\d+)?)`,                            // 质量: 4
		`(?i)(\d+(?:\.\d+)?)\s*(?:分|分制)`,                        // 4分 或 4.5分
		`(?i)(?:优秀|good|excellent).*?(\d+(?:\.\d+)?)`,          // 优秀 4.5
		`(?i)(\d+(?:\.\d+)?).*?(?:优秀|good|excellent)`,          // 4.5 优秀
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(report); len(matches) > 1 {
			score, err := strconv.ParseFloat(matches[1], 64)
			if err == nil && score >= 1 && score <= 5 {
				return score
			}
		}
	}

	return 0 // 无法提取评分
}
