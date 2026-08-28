package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/cloudwego/eino/adk"

	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
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
	ID               uint64                     `json:"id,omitempty"`
	Type             string                     `json:"type"`
	Content          string                     `json:"content,omitempty"`
	ToolName         string                     `json:"tool_name,omitempty"`
	ToolArgs         string                     `json:"tool_args,omitempty"`
	Error            string                     `json:"error,omitempty"`
	Tasks            []*deck.TaskItem           `json:"tasks,omitempty"`
	Done             int                        `json:"done,omitempty"`
	Total            int                        `json:"total,omitempty"`
	Files            []string                   `json:"files,omitempty"`
	Message          string                     `json:"message,omitempty"`
	Duration         string                     `json:"duration,omitempty"`
	Status           TaskStatus                 `json:"status,omitempty"`
	PromptTokens     int64                      `json:"prompt_tokens,omitempty"`
	CompletionTokens int64                      `json:"completion_tokens,omitempty"`
	TotalTokens      int64                      `json:"total_tokens,omitempty"`
	Phase            string                     `json:"phase,omitempty"`
	PhaseDetail      string                     `json:"phase_detail,omitempty"`
	RuntimeMeta      *utils.RuntimeMetaSnapshot `json:"runtime_meta,omitempty"`
}

// TaskInfo 是任务的公开可见摘要。
type TaskInfo struct {
	ID                  string     `json:"id"`
	UserID              int        `json:"user_id"`
	Query               string     `json:"query"`
	Status              TaskStatus `json:"status"`
	WorkDir             string     `json:"work_dir"`
	CreatedAt           time.Time  `json:"created_at"`
	DoneCount           int        `json:"done_count"`
	TotalCount          int        `json:"total_count"`
	Duration            string     `json:"duration,omitempty"`
	Error               string     `json:"error,omitempty"`
	Files               []string   `json:"files,omitempty"`
	PromptTokens        int64      `json:"prompt_tokens"`
	CompletionTokens    int64      `json:"completion_tokens"`
	TotalTokens         int64      `json:"total_tokens"`
	ConversationContent string     `json:"conversation_content,omitempty"` // 拼接后的对话内容
	FullAnswer          string     `json:"full_answer,omitempty"`          // 完整累积的 LLM 回答
}

// TaskState 保存单个任务的内部状态。
type TaskState struct {
	Info          TaskInfo
	Events        []SSERichEvent
	nextEventID   uint64
	turnEventID   uint64
	listeners     map[string]chan SSERichEvent
	cancel        context.CancelFunc
	result        *deck.PPTTaskResult
	reportedFiles map[string]bool
	runtimeMeta   *utils.RuntimeMeta
	delivery      DeliverySnapshot
	Mu            sync.Mutex

	// pendingContinueMsg 任务运行中时，用户提交的待处理消息（消费后清空）
	pendingContinueMsg string
	// pendingContinueQueued 已通知前端排队（避免重复通知）
	pendingContinueQueued bool

	// fullAnswer 累积全部 LLM answer SSE 输出，任务结束时一次性存入 DB
	fullAnswer      strings.Builder
	answerTurn      strings.Builder
	assistantTurnFn func(taskID, workDir, content string)
}

// Persist 将任务状态持久化到数据库。
func (ts *TaskState) Persist() {
	ts.persist()
}

// FullAnswer 返回已累积的完整 LLM 回答内容。
func (ts *TaskState) FullAnswer() string {
	ts.Mu.Lock()
	defer ts.Mu.Unlock()
	return ts.fullAnswer.String()
}

// LatestEventID lets clients begin a continuation stream after the previous
// terminal event instead of replaying the completed generation turn.
func (ts *TaskState) LatestEventID() uint64 {
	ts.Mu.Lock()
	defer ts.Mu.Unlock()
	return ts.nextEventID
}

// ReplayAfterEventID returns the last event that closed a persisted assistant
// turn. Clients restore structured messages, then replay only the unfinished
// turn following this boundary.
func (ts *TaskState) ReplayAfterEventID() uint64 {
	ts.Mu.Lock()
	defer ts.Mu.Unlock()
	return ts.turnEventID
}

// EventBoundaries returns a consistent pair for conversation snapshot clients.
func (ts *TaskState) EventBoundaries() (latest, replayAfter uint64) {
	ts.Mu.Lock()
	defer ts.Mu.Unlock()
	return ts.nextEventID, ts.turnEventID
}

// SnapshotInfo returns a race-free public task snapshot.
func (ts *TaskState) SnapshotInfo() TaskInfo {
	ts.Mu.Lock()
	defer ts.Mu.Unlock()
	info := ts.Info
	info.Files = append([]string(nil), ts.Info.Files...)
	return info
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
	if event.ID == 0 {
		ts.nextEventID++
		event.ID = ts.nextEventID
	} else if event.ID > ts.nextEventID {
		ts.nextEventID = event.ID
	}
	// 累积 LLM 回答内容到 fullAnswer，任务结束时一次性写入 DB。
	// 有些模型/框架会把 answer chunk 作为累计文本发出，这里先转成
	// 可显示的增量，避免 SSE 实时流、事件回放和会话持久化重复展示。
	if event.Type == "answer" && event.Content != "" {
		event.Content = normalizeAnswerChunk(ts.answerTurn.String(), event.Content)
		ts.fullAnswer.WriteString(event.Content)
		ts.answerTurn.WriteString(event.Content)
		if ts.runtimeMeta != nil && strings.TrimSpace(event.Content) != "" {
			ts.runtimeMeta.RecordAssistantOutput(event.Content)
		}
	}
	ts.Events = append(ts.Events, event)
	if len(ts.Events) > 500 {
		ts.Events = ts.Events[len(ts.Events)-300:]
	}
	var completedTurn string
	isTurnBoundary := event.Type == "answer_end" || event.Type == "complete" || event.Type == "continue_complete"
	if isTurnBoundary {
		completedTurn = strings.TrimSpace(ts.answerTurn.String())
		ts.answerTurn.Reset()
	}
	for _, ch := range ts.listeners {
		select {
		case ch <- event:
		default:
		}
	}
	turnCallback := ts.assistantTurnFn
	taskID := ts.Info.ID
	workDir := ts.Info.WorkDir
	ts.Mu.Unlock()
	if completedTurn != "" && turnCallback != nil {
		turnCallback(taskID, workDir, completedTurn)
	}
	if isTurnBoundary {
		ts.Mu.Lock()
		if event.ID > ts.turnEventID {
			ts.turnEventID = event.ID
		}
		ts.Mu.Unlock()
	}
}

func normalizeAnswerChunk(currentTurn, chunk string) string {
	current := strings.TrimSpace(currentTurn)
	incoming := strings.TrimSpace(chunk)
	if incoming == "" {
		return ""
	}
	if current == "" {
		return chunk
	}
	if chunk == currentTurn || incoming == current {
		return ""
	}
	if strings.HasPrefix(chunk, currentTurn) {
		suffix := strings.TrimPrefix(chunk, currentTurn)
		if strings.TrimSpace(suffix) == "" {
			return ""
		}
		return preserveAnswerChunkBoundary(currentTurn, suffix)
	}
	if suffix, ok := cumulativeAnswerSuffix(currentTurn, chunk); ok {
		if strings.TrimSpace(suffix) == "" {
			return ""
		}
		return preserveAnswerChunkBoundary(currentTurn, suffix)
	}
	currentNormalized := normalizeAnswerForCompare(current)
	incomingNormalized := normalizeAnswerForCompare(incoming)
	if len(incomingNormalized) >= 20 && strings.Contains(currentNormalized, incomingNormalized) {
		return ""
	}
	return preserveAnswerChunkBoundary(currentTurn, chunk)
}

func normalizeAnswerForCompare(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), ""))
}

func cumulativeAnswerSuffix(existing, incoming string) (string, bool) {
	existingNormalized := []rune(normalizeAnswerForCompare(existing))
	incomingNormalized := []rune(normalizeAnswerForCompare(incoming))
	if len(existingNormalized) == 0 || len(incomingNormalized) < len(existingNormalized) {
		return "", false
	}
	for index, char := range existingNormalized {
		if incomingNormalized[index] != char {
			return "", false
		}
	}

	matched := 0
	suffixStart := 0
	for index, char := range incoming {
		if unicode.IsSpace(char) {
			continue
		}
		if matched >= len(existingNormalized) {
			break
		}
		if unicode.ToLower(char) != existingNormalized[matched] {
			return "", false
		}
		matched++
		suffixStart = index + len(string(char))
	}
	if matched != len(existingNormalized) {
		return "", false
	}
	return incoming[suffixStart:], true
}

func preserveAnswerChunkBoundary(currentTurn, chunk string) string {
	if shouldInsertASCIIWordSpace(currentTurn, chunk) {
		return " " + chunk
	}
	return chunk
}

func shouldInsertASCIIWordSpace(left, right string) bool {
	first, ok := firstRune(right)
	if !ok || unicode.IsSpace(first) || !isASCIIWordRune(first) {
		return false
	}
	last, ok := lastRune(left)
	if !ok || !isASCIIWordRune(last) {
		return false
	}
	return true
}

func firstRune(value string) (rune, bool) {
	for _, char := range value {
		return char, true
	}
	return 0, false
}

func lastRune(value string) (rune, bool) {
	var last rune
	ok := false
	for _, char := range value {
		last = char
		ok = true
	}
	return last, ok
}

func isASCIIWordRune(char rune) bool {
	return (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')
}

// SubscribeFrom atomically snapshots buffered events newer than afterEventID
// and registers a live listener when the task is still running. Keeping these
// operations under one lock prevents events from falling between replay and
// subscription.
func (ts *TaskState) SubscribeFrom(listenerID string, listenerCh chan SSERichEvent, afterEventID uint64) ([]SSERichEvent, bool) {
	ts.Mu.Lock()
	defer ts.Mu.Unlock()

	events := make([]SSERichEvent, 0, len(ts.Events))
	for _, event := range ts.Events {
		if event.ID > afterEventID {
			events = append(events, event)
		}
	}

	done := ts.Info.Status != TaskStatusRunning
	if !done {
		if ts.listeners == nil {
			ts.listeners = make(map[string]chan SSERichEvent)
		}
		ts.listeners[listenerID] = listenerCh
	}
	return events, done
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
	mu              sync.RWMutex
	tasks           map[string]*TaskState
	baseDir         string
	onTaskComplete  func(userID int, workDir string, query string)
	onTaskFailed    func(taskID string)
	onTaskContinue  func(taskID string) // 任务完成且有待处理消息时触发
	onFileReady     func(taskID string, workDir string, filename string)
	onAssistantTurn func(taskID string, workDir string, content string)
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
		onTaskContinue: onTaskContinue,
	}
}

// SetFileReadyCallback 注册单页 PPTX 落盘后的异步处理回调。
// 回调不得阻塞任务进度轮询；TaskManager 会在独立 goroutine 中调用它。
func (tm *TaskManager) SetFileReadyCallback(callback func(taskID string, workDir string, filename string)) {
	tm.mu.Lock()
	tm.onFileReady = callback
	tm.mu.Unlock()
}

// SetAssistantTurnCallback registers the single persistence sink used when an
// answer turn reaches an explicit lifecycle boundary.
func (tm *TaskManager) SetAssistantTurnCallback(callback func(taskID string, workDir string, content string)) {
	tm.mu.Lock()
	tm.onAssistantTurn = callback
	for _, ts := range tm.tasks {
		ts.Mu.Lock()
		ts.assistantTurnFn = callback
		ts.Mu.Unlock()
	}
	tm.mu.Unlock()
}

func (tm *TaskManager) reportFileReady(ts *TaskState, workDir, filename string) {
	ts.Broadcast(SSERichEvent{
		Type:     "file_ready",
		ToolName: filename,
		Files:    []string{filename},
	})

	tm.mu.RLock()
	callback := tm.onFileReady
	tm.mu.RUnlock()
	if callback != nil {
		go callback(ts.Info.ID, workDir, filename)
	}
}

// ── 数据库转换辅助函数 ────────────────────────────────────────────────

func taskInfoToRecord(info *TaskInfo) *db.TaskRecord {
	filesJSON, _ := json.Marshal(DeduplicateOutputFiles(info.Files))
	return &db.TaskRecord{
		ID:                  info.ID,
		UserID:              uint(info.UserID),
		Query:               mysqlSafeText(info.Query),
		Status:              string(info.Status),
		WorkDir:             info.WorkDir,
		DoneCount:           info.DoneCount,
		TotalCount:          info.TotalCount,
		Duration:            info.Duration,
		Error:               mysqlSafeText(info.Error),
		Files:               mysqlSafeText(string(filesJSON)),
		PromptTokens:        info.PromptTokens,
		CompletionTokens:    info.CompletionTokens,
		TotalTokens:         info.TotalTokens,
		ConversationContent: mysqlSafeText(info.ConversationContent),
		FullAnswer:          mysqlSafeText(info.FullAnswer),
	}
}

func mysqlSafeText(value string) string {
	value = strings.ToValidUTF8(value, "")
	return strings.Map(func(r rune) rune {
		// Some existing deployments still use MySQL utf8/utf8mb3 columns.
		// Keep ordinary BMP text, including Chinese, and drop emoji/symbol
		// glyphs before writing task records.
		if r > 0xFFFF || isEmojiSymbol(r) {
			return -1
		}
		return r
	}, value)
}

func isEmojiSymbol(r rune) bool {
	switch {
	case r >= 0x2600 && r <= 0x27BF:
		return true
	case r >= 0x2B00 && r <= 0x2BFF:
		return true
	default:
		return false
	}
}

func recordToTaskInfo(r *db.TaskRecord) *TaskInfo {
	var files []string
	json.Unmarshal([]byte(r.Files), &files)
	files = DeduplicateOutputFiles(files)
	if files == nil {
		files = []string{}
	}
	return &TaskInfo{
		ID:                  r.ID,
		UserID:              int(r.UserID),
		Query:               r.Query,
		Status:              TaskStatus(r.Status),
		WorkDir:             r.WorkDir,
		DoneCount:           r.DoneCount,
		TotalCount:          r.TotalCount,
		Duration:            r.Duration,
		Error:               r.Error,
		Files:               files,
		CreatedAt:           r.CreatedAt,
		PromptTokens:        r.PromptTokens,
		CompletionTokens:    r.CompletionTokens,
		TotalTokens:         r.TotalTokens,
		ConversationContent: r.ConversationContent,
		FullAnswer:          r.FullAnswer,
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
		"status":               r.Status,
		"done_count":           r.DoneCount,
		"total_count":          r.TotalCount,
		"duration":             r.Duration,
		"error":                r.Error,
		"files":                r.Files,
		"prompt_tokens":        r.PromptTokens,
		"completion_tokens":    r.CompletionTokens,
		"total_tokens":         r.TotalTokens,
		"conversation_content": r.ConversationContent,
		"full_answer":          r.FullAnswer,
	}); err != nil {
		logger.Error("db_persist_failed", "task_id", r.ID, "error", err.Error())
	}
}

// AgentFactory 为特定任务配置创建 agent。
type AgentFactory func(ctx context.Context, cfg *deck.PPTTaskConfig) (adk.Agent, error)

// ErrTaskAlreadyRunning is retained for API compatibility with older handlers.
// New tasks are no longer rejected globally; model calls are limited per upstream resource.
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
	factory AgentFactory, cfg *deck.PPTTaskConfig) (*TaskInfo, error) {

	workDir := filepath.Join(tm.baseDir, fmt.Sprintf("%d-%s", userID, cfg.TaskID))
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, err
	}
	cfg.WorkDir = workDir
	runtimeMeta := utils.NewRuntimeMeta(cfg.TaskID, workDir)
	runtimeMeta.SetEventSink(persistRuntimeEvent)
	taskInput := utils.TaskInputAnchor{
		Summary: compactRequestSummary(query, 120), OriginalLength: len([]rune(query)),
	}
	if cfg.Outline != nil {
		taskInput.Recommendation = cfg.Outline.RecommendationReason
	}
	runtimeMeta.RecordTaskInput(taskInput)
	runtimeMeta.RecordEvent("task_created", "task", "ok", query, map[string]any{
		"user_id": userID,
	})
	agentCtx, tokenTracker := utils.WithTokenTracker(context.Background())
	agentCtx = utils.WithRuntimeMeta(agentCtx, runtimeMeta)
	agentCtx, cancel := context.WithCancel(agentCtx)
	cfg.CompressorTracker = tokenTracker
	cfg.RuntimeMeta = runtimeMeta

	// outline 只作为 Planner 输入草稿。无论是否有大纲，都必须经过
	// Planner 补全、Task Reviewer 审查和 Go commit 后才能发布 tasks.json。
	if cfg.Outline != nil && len(cfg.Outline.Slides) > 0 {
		manifest := outlineToManifest(cfg.Outline, workDir)
		if err := deck.WriteTasksDraftManifest(workDir, manifest); err != nil {
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
		listeners:       make(map[string]chan SSERichEvent),
		reportedFiles:   make(map[string]bool),
		runtimeMeta:     runtimeMeta,
		assistantTurnFn: tm.onAssistantTurn,
	}
	runtimeMeta.SetEventSink(func(event utils.RuntimeEvent) {
		persistRuntimeEvent(event)
		go func() {
			if event.Kind == "phase_changed" && event.Phase != "" {
				ts.Broadcast(SSERichEvent{
					Type:        "progress",
					Phase:       event.Phase,
					PhaseDetail: event.Detail,
				})
			}
			if event.Kind == "planner_context_compressing" {
				ts.Broadcast(SSERichEvent{
					Type:        "progress",
					Phase:       "compressing_context",
					PhaseDetail: firstRuntimeDetail(event.Detail, "正在压缩较早对话，保留你的最新要求"),
				})
			}
			snap := runtimeMeta.Snapshot()
			ts.Broadcast(SSERichEvent{
				Type:        "runtime_meta",
				RuntimeMeta: &snap,
			})
		}()
	})
	type workDirSetter interface {
		SetWorkDir(context.Context, string) context.Context
	}
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

func firstRuntimeDetail(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func (tm *TaskManager) runAgent(ctx context.Context, ts *TaskState, agent adk.Agent,
	cfg *deck.PPTTaskConfig, query string) {
	startedAt := time.Now()
	if ts.runtimeMeta != nil {
		ts.runtimeMeta.RecordPhase("preparing", "初始化任务运行环境")
	}

	// ── 步骤1：任务输入 ──────────────────────────────
	var step1 strings.Builder
	step1.WriteString("【步骤1/2】任务输入\n")
	if cfg.Outline != nil && len(cfg.Outline.Slides) > 0 {
		step1.WriteString(fmt.Sprintf("  • 用户大纲: %d 页\n", len(cfg.Outline.Slides)))
	} else {
		step1.WriteString("  • 用户大纲: 未提供，由 Planner 规划\n")
	}

	// ── 步骤2：执行链路 ────────────────────────────
	var step2 strings.Builder
	step2.WriteString("【步骤2/2】执行链路\n")
	step2.WriteString("  • Planner 规划 → Task Reviewer 审查 → Go 校验提交 → Worker Pool 渲染\n")
	step2.WriteString("  • 用户风格不再自动读取；如需全局风格，请在任务提示词中手动说明\n")

	ts.Broadcast(SSERichEvent{Type: "system_step", Content: step1.String()})
	ts.Broadcast(SSERichEvent{Type: "system_step_end", Content: ""})
	ts.Broadcast(SSERichEvent{Type: "system_step", Content: step2.String()})
	ts.Broadcast(SSERichEvent{Type: "system_step_end", Content: ""})

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
			if ts.runtimeMeta != nil {
				ts.runtimeMeta.RecordTaskTerminal(string(TaskStatusFailed), ts.Info.Error)
				if err := ts.runtimeMeta.WriteReport(string(TaskStatusFailed)); err != nil {
					logger.Warn("runtime_report_write_failed", "task_id", ts.Info.ID, "error", err.Error())
				}
			}
		}
	}()

	runCtx, cancelRun := context.WithCancelCause(ctx)
	defer cancelRun(nil)
	go tm.pollProgress(runCtx, ts, cfg.WorkDir, func(snapshot DeliverySnapshot) {
		logger.Info("delivery_metadata_complete", "task_id", ts.Info.ID, "done", snapshot.Done, "total", snapshot.Total)
		cancelRun(errDeliveryMetadataComplete)
	})

	result, err := deck.RunPPTPlannerWithCallback(runCtx, agent, cfg, query, func(event deck.AgentEvent) {
		if event.Type == deck.AgentEventProgress {
			ts.Broadcast(SSERichEvent{
				Type:        "progress",
				Phase:       event.Phase,
				PhaseDetail: event.PhaseDetail,
			})
			return
		}
		if event.Type == "tool_call" || event.Type == "token_usage" {
			// 从 tool_call 推断阶段
			detectAndBroadcastPhase(ts, event)
			return
		}
		ts.Broadcast(SSERichEvent{
			Type:     event.Type,
			Content:  event.Content,
			ToolName: event.ToolName,
			ToolArgs: event.ToolArgs,
			Error:    event.Error,
		})
	})
	ts.Broadcast(SSERichEvent{Type: "answer_end"})

	if err == nil && ctx.Err() == nil {
		renderResult, renderErr := deck.RenderDeckByTaskIDWorkflow(ctx, cfg, func(event deck.DeckRenderEvent) {
			switch event.Type {
			case "workflow_start":
				ts.Broadcast(SSERichEvent{Type: "progress", Phase: "rendering", PhaseDetail: event.Detail})
			case "slide_start":
				ts.Broadcast(SSERichEvent{
					Type:        "progress",
					Phase:       "rendering",
					PhaseDetail: fmt.Sprintf("开始生成第 %d 页：%s", event.PageIndex, event.Detail),
				})
			case "slide_done":
				ts.Broadcast(SSERichEvent{
					Type:        "progress",
					Phase:       "rendering",
					PhaseDetail: fmt.Sprintf("第 %d 页生成完成：%s", event.PageIndex, event.OutputFile),
				})
			case "slide_error":
				ts.Broadcast(SSERichEvent{
					Type:        "error",
					Error:       fmt.Sprintf("第 %d 页生成失败：%s", event.PageIndex, event.Error),
					Phase:       "rendering",
					PhaseDetail: event.OutputFile,
				})
			}
		})
		if renderResult != nil {
			result = renderResult
		}
		if renderErr != nil {
			err = renderErr
		}
	}

	ts.Mu.Lock()
	ts.result = result
	ts.Mu.Unlock()

	metadataStoppedAgent := errors.Is(context.Cause(runCtx), errDeliveryMetadataComplete)
	if metadataStoppedAgent {
		err = nil
	}
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
		ts.Info.Duration = result.Duration.Round(time.Millisecond).String()
	} else {
		ts.Info.Duration = time.Since(startedAt).Round(time.Millisecond).String()
	}

	delivery := ts.deliverySnapshot()
	if !delivery.Complete() && ts.Info.Status != TaskStatusCancelled && shouldSyncDeliveryAfterRun(cfg.WorkDir, taskFailed) {
		if synced, syncErr := tm.syncDeliveryMetadata(ts, cfg.WorkDir); syncErr == nil {
			delivery = synced
		} else {
			logger.Warn("delivery_metadata_sync_failed", "task_id", ts.Info.ID, "error", syncErr.Error())
		}
	}
	if delivery.Complete() && ts.Info.Status != TaskStatusCancelled {
		ts.Info.Status = TaskStatusCompleted
		ts.Info.Error = ""
		taskFailed = false
	}
	if applyDeliverySnapshotOutcome(ts, delivery) {
		taskFailed = true
	}
	if delivery.Total > 0 {
		if result == nil {
			result = &deck.PPTTaskResult{Duration: time.Since(startedAt)}
		}
		result.TotalSlides = delivery.Total
		result.DoneSlides = delivery.Done
		result.Files = append([]string(nil), delivery.Files...)
		ts.Mu.Lock()
		ts.result = result
		ts.Mu.Unlock()
	}

	if ts.runtimeMeta != nil {
		switch ts.Info.Status {
		case TaskStatusCompleted:
			ts.runtimeMeta.RecordPhase("complete", "任务执行完成")
		case TaskStatusCancelled:
			ts.runtimeMeta.RecordPhase("cancelled", ts.Info.Error)
		case TaskStatusFailed:
			ts.runtimeMeta.RecordPhase("failed", ts.Info.Error)
		}
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

	// 将累积的完整 LLM 回答写入持久化字段
	ts.Info.FullAnswer = ts.fullAnswer.String()

	ts.persist()

	// 任务完成后，检查是否有等待中的继续消息，如有则触发自动继续处理
	if ts.HasPendingContinueMsg() {
		ts.Broadcast(SSERichEvent{
			Type:    "continue_queued",
			Content: "queued",
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

	finalEvent := SSERichEvent{
		Type:             "complete",
		Status:           ts.Info.Status,
		Message:          ts.Info.Error,
		Done:             ts.Info.DoneCount,
		Total:            ts.Info.TotalCount,
		Files:            DeduplicateOutputFiles(ts.Info.Files),
		Duration:         ts.Info.Duration,
		PromptTokens:     ts.Info.PromptTokens,
		CompletionTokens: ts.Info.CompletionTokens,
		TotalTokens:      ts.Info.TotalTokens,
	}
	if ts.Info.Status == TaskStatusCompleted {
		finalEvent.Message = fmt.Sprintf("PPT 已完成交付，共 %d 页，交付元数据 %d/%d。", ts.Info.TotalCount, ts.Info.DoneCount, ts.Info.TotalCount)
	}
	if ts.runtimeMeta != nil {
		ts.runtimeMeta.RecordTaskTerminal(string(ts.Info.Status), finalEvent.Message)
		if err := ts.runtimeMeta.WriteReport(string(ts.Info.Status)); err != nil {
			logger.Warn("runtime_report_write_failed", "task_id", ts.Info.ID, "error", err.Error())
		}
		snap := ts.runtimeMeta.Snapshot()
		finalEvent.RuntimeMeta = &snap
	}
	ts.Broadcast(finalEvent)

	// 在 complete 事件被所有监听者处理完毕后（触发 flushAnswerToDB + flushCompleteToDB），
	// 再构建并写入 conversation_content。
	ts.persistConversationContent()

	// 触发任务完成回调。
	if tm.onTaskComplete != nil && ts.Info.UserID > 0 && ts.Info.Status == TaskStatusCompleted {
		go tm.onTaskComplete(ts.Info.UserID, ts.Info.WorkDir, ts.Info.Query)
	}
}

func shouldSyncDeliveryAfterRun(workDir string, taskFailed bool) bool {
	if !taskFailed {
		return true
	}
	if _, err := os.Stat(filepath.Join(workDir, "tasks.json")); err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func applyDeliverySnapshotOutcome(ts *TaskState, delivery DeliverySnapshot) bool {
	if ts == nil || ts.Info.Status != TaskStatusCompleted {
		return false
	}

	message := ""
	switch {
	case delivery.Total == 0:
		message = "交付元数据中没有有效页面，任务未完成，请重试"
	case !delivery.Complete():
		message = fmt.Sprintf("生成未完成：交付元数据显示已交付 %d/%d 页", delivery.Done, delivery.Total)
		if len(delivery.PendingTasks) > 0 {
			message += fmt.Sprintf("，仍有 %d 页待完成", len(delivery.PendingTasks))
		}
	}
	if message == "" {
		return false
	}

	ts.Info.Status = TaskStatusFailed
	ts.Info.Error = message
	return true
}

func persistRuntimeEvent(event utils.RuntimeEvent) {
	if event.TaskID == "" {
		return
	}
	metadata := ""
	if len(event.Metadata) > 0 {
		if data, err := json.Marshal(event.Metadata); err == nil {
			metadata = string(data)
		}
	}
	timestamp := time.Now()
	if parsed, err := time.Parse(time.RFC3339Nano, event.Timestamp); err == nil {
		timestamp = parsed
	}
	if err := db.CreateRuntimeEvent(&db.RuntimeEventRecord{
		TaskID:    event.TaskID,
		EventID:   event.ID,
		Timestamp: timestamp,
		ElapsedMS: event.ElapsedMS,
		Kind:      event.Kind,
		Phase:     event.Phase,
		Name:      event.Name,
		Status:    event.Status,
		Detail:    event.Detail,
		Metadata:  metadata,
	}); err != nil {
		logger.Warn("persist_runtime_event_failed", "task_id", event.TaskID, "event_id", event.ID, "error", err.Error())
	}
}

// detectAndBroadcastPhase 从 Planner 工具事件推断当前阶段并广播进度事件。
func detectAndBroadcastPhase(ts *TaskState, event deck.AgentEvent) {
	detail := event.PhaseDetail
	if detail == "" {
		detail = event.ToolArgs
	}

	phase := "planning"
	phaseDetail := ""

	switch {
	case event.ToolName == "update_tasks_manifest":
		phase = "planning"
		phaseDetail = "Planner 正在一次性写入 DeckSpec 草稿"
	case event.ToolName == "patch_tasks_draft":
		phase = "reviewing"
		phaseDetail = "Task Reviewer 正在批量修正规划问题"
	case event.ToolName == "search":
		phase = "planning"
		if query := extractToolStringArg(event.ToolArgs, "query"); query != "" {
			phaseDetail = "正在搜索：" + query
		} else {
			phaseDetail = "正在检索并核实资料"
		}
	case strings.Contains(detail, "tasks.json") || strings.Contains(detail, "TasksJSON"):
		phase = "planning"
		phaseDetail = "读取任务清单"
	case event.ToolName == "read_file":
		phase = "preparing"
		phaseDetail = "正在读取模板与设计规范"
	case strings.Contains(detail, ".py") && strings.Contains(detail, "read_file"):
		phase = "preparing"
		phaseDetail = "读取模板"
	}

	if ts.runtimeMeta != nil {
		ts.runtimeMeta.RecordPhase(phase, phaseDetail)
	}
	ts.Broadcast(SSERichEvent{
		Type:        "progress",
		Phase:       phase,
		PhaseDetail: phaseDetail,
	})
}

func extractToolStringArg(raw, key string) string {
	var parsed map[string]any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return ""
	}
	value, _ := parsed[key].(string)
	return strings.TrimSpace(value)
}

func (tm *TaskManager) pollProgress(ctx context.Context, ts *TaskState, workDir string, onDeliveryComplete func(DeliverySnapshot)) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	lastDone, lastTotal := -1, -1
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
			if ts.runtimeMeta != nil {
				ts.runtimeMeta.SetLLMTokens(p, c, t)
			}
			if t != lastTokens {
				lastTokens = t
				ts.Mu.Lock()
				ts.Info.PromptTokens = p
				ts.Info.CompletionTokens = c
				ts.Info.TotalTokens = t
				ts.Mu.Unlock()
				ts.Broadcast(SSERichEvent{
					Type:             "token_usage",
					PromptTokens:     p,
					CompletionTokens: c,
					TotalTokens:      t,
				})
			}
		}
		if ts.runtimeMeta != nil {
			snap := ts.runtimeMeta.Snapshot()
			ts.Broadcast(SSERichEvent{
				Type:        "runtime_meta",
				RuntimeMeta: &snap,
			})
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
						if ts.runtimeMeta != nil {
							ts.runtimeMeta.RecordFileCreated(entry.Name())
						}
						tm.reportFileReady(ts, workDir, entry.Name())
						continue
					}
					ts.Mu.Unlock()
				}
			}
		}

		manifest, err := deck.ReconcileTasksManifestOutputFiles(workDir)
		if err != nil || manifest == nil {
			continue
		}

		snapshot := deliverySnapshotFromManifest(manifest)
		ts.updateDelivery(snapshot)
		pendingFiles := make([]string, 0, len(snapshot.PendingTasks))
		var currentSlide *utils.PlanSlide
		var nextPendingSlide *utils.PlanSlide
		for _, item := range manifest.Tasks {
			if item == nil {
				continue
			}
			if !isCompletedSlideStatus(item.Status) && item.OutputFile != "" {
				pendingFiles = append(pendingFiles, item.OutputFile)
			}
			if currentSlide == nil && item.Status == deck.StatusGenerating {
				slide := runtimePlanSlide(item)
				currentSlide = &slide
			}
			if nextPendingSlide == nil && item.Status == deck.StatusPending {
				slide := runtimePlanSlide(item)
				nextPendingSlide = &slide
			}
		}
		if currentSlide == nil {
			currentSlide = nextPendingSlide
		}
		if ts.runtimeMeta != nil {
			ts.runtimeMeta.RecordSlideProgress(snapshot.Done, snapshot.Total, len(snapshot.PendingTasks))
			ts.runtimeMeta.RecordManifestValidation(snapshot.Done, snapshot.Total, nil, snapshot.PendingTasks)
			ts.runtimeMeta.RecordCurrentSlide(currentSlide)
			ts.runtimeMeta.ComparePlan(runtimePlanSlides(manifest.Tasks), pendingFiles)
		}

		if snapshot.Done != lastDone || snapshot.Total != lastTotal {
			lastDone, lastTotal = snapshot.Done, snapshot.Total
			ts.Broadcast(SSERichEvent{
				Type:  "progress",
				Tasks: manifest.Tasks,
				Done:  snapshot.Done,
				Total: snapshot.Total,
			})
		}
		if snapshot.Complete() {
			if onDeliveryComplete != nil {
				onDeliveryComplete(snapshot)
			}
			return
		}
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
		conversationContent := mysqlSafeText(ts.Info.ConversationContent)
		go func() {
			if err := db.UpdateTaskRecord(ts.Info.ID, map[string]any{
				"conversation_content": conversationContent,
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

// GetWorkDir 返回任务的工作目录，先从内存查再查 DB。
// 用于冷加载场景下下载文件/缩略图不因内存过期而 404。
func (tm *TaskManager) GetWorkDir(id string) string {
	ts := tm.GetTaskState(id)
	if ts != nil {
		return ts.Info.WorkDir
	}
	if db.DB != nil {
		if r, err := db.GetTaskRecord(id); err == nil {
			return r.WorkDir
		}
	}
	return ""
}

// NewColdTaskState 从 TaskInfo 创建一个最小的 TaskState（无 events，无 cancel）。
// 用于在服务器重启后从 MySQL 恢复任务。
func (tm *TaskManager) NewColdTaskState(info TaskInfo) *TaskState {
	ts := &TaskState{
		Info:            info,
		listeners:       make(map[string]chan SSERichEvent),
		reportedFiles:   make(map[string]bool),
		assistantTurnFn: tm.onAssistantTurn,
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

// ListAllTasks returns task summaries for administrators, merging hot in-memory
// tasks with persisted records.
func (tm *TaskManager) ListAllTasks() []TaskInfo {
	seen := make(map[string]bool)
	var result []TaskInfo

	tm.mu.RLock()
	for _, ts := range tm.tasks {
		result = append(result, ts.Info)
		seen[ts.Info.ID] = true
	}
	tm.mu.RUnlock()

	if db.DB != nil {
		records, err := db.ListAllTaskRecords(500)
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
		if err := db.DeleteRuntimeEvents(id); err != nil {
			logger.Error("db_delete_runtime_events_failed", "task_id", id, "error", err.Error())
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
func (tm *TaskManager) ReadTasksManifestFile(id string) (*deck.TasksManifest, error) {
	ts := tm.GetTaskState(id)
	if ts == nil {
		return nil, os.ErrNotExist
	}
	return deck.ReadTasksManifest(ts.Info.WorkDir)
}

// TaskFilesAsJSON 返回任务的 tasks.json 作为原始 JSON 字节。
func (tm *TaskManager) TaskFilesAsJSON(workDir string) ([]byte, error) {
	return json.Marshal(struct {
		Tasks []*deck.TaskItem `json:"tasks"`
	}{
		Tasks: nil,
	})
}

// outlineToManifest 将用户编排的 outline 转换为 TasksManifest
func outlineToManifest(outline *deck.TaskOutline, workDir string) *deck.TasksManifest {
	tasks := make([]*deck.TaskItem, 0, len(outline.Slides))
	for i, slide := range outline.Slides {
		safeTitle := sanitizeFilename(slide.Title)
		if strings.TrimSpace(safeTitle) == "" {
			safeTitle = fmt.Sprintf("slide-%d", i+1)
		}
		item := &deck.TaskItem{
			TaskID:      fmt.Sprintf("slide-%d", i+1),
			PageIndex:   i + 1,
			Title:       slide.Title,
			ContentType: slide.ContentType,
			OutputFile:  fmt.Sprintf("%d_%s.pptx", i+1, safeTitle),
			Status:      deck.StatusPending,
		}
		// Carry through content_plan if present
		if slide.ContentPlan != nil {
			copiedPlan := *slide.ContentPlan
			if slide.ContentPlan.Components != nil {
				copiedPlan.Components = append([]deck.PlanComponent(nil), slide.ContentPlan.Components...)
			}
			item.ContentPlan = &copiedPlan
		}
		tasks = append(tasks, item)
	}
	return &deck.TasksManifest{
		Title: outline.Title,
		Tasks: tasks,
	}
}

func runtimePlanSlides(items []*deck.TaskItem) []utils.PlanSlide {
	slides := make([]utils.PlanSlide, 0, len(items))
	for _, item := range items {
		if item != nil {
			slides = append(slides, runtimePlanSlide(item))
		}
	}
	return slides
}

func runtimePlanSlide(item *deck.TaskItem) utils.PlanSlide {
	return utils.PlanSlide{
		PageIndex: item.PageIndex, TaskID: item.TaskID, Title: item.Title,
		ContentType: item.ContentType, OutputFile: CanonicalOutputFile(item.OutputFile), Status: item.Status,
	}
}

func isCompletedSlideStatus(status string) bool {
	return status == deck.StatusDone || status == deck.StatusQADone || status == deck.StatusFixed
}

func compactRequestSummary(query string, limit int) string {
	query = strings.ReplaceAll(query, "\r", "\n")
	parts := strings.FieldsFunc(query, func(r rune) bool { return r == '\n' || r == '。' || r == '！' || r == '？' })
	summary := strings.TrimSpace(query)
	if len(parts) > 0 {
		summary = strings.TrimSpace(parts[0])
	}
	summary = strings.TrimLeft(summary, "#>-* `\t")
	summary = strings.Join(strings.Fields(summary), " ")
	runes := []rune(summary)
	if limit > 0 && len(runes) > limit {
		summary = string(runes[:limit]) + "..."
	}
	return summary
}

// persistConversationContent 从对话消息表（数据库）和内存中的 SSE 事件流中提取有意义的内容，
// 将它们合并为可读的对话文本，并存储到 ts.Info 中以便后续持久化到数据库。
// 在任务结束时调用（成功、取消或错误）。
var listConversationMessagesForSummary = db.ListConversationMessages

func (ts *TaskState) persistConversationContent() {
	ts.Mu.Lock()
	info := ts.Info
	ts.Mu.Unlock()

	// Database and filesystem I/O stay outside TaskState.Mu so readers can keep
	// serving the in-memory task and answer while persistence is slow.
	dbMessages, err := listConversationMessagesForSummary(info.ID)
	if err != nil {
		logger.Warn("load_conversation_messages_failed", "task_id", info.ID, "error", err.Error())
	}

	manifest, manifestErr := deck.ReadTasksManifest(info.WorkDir)

	var b strings.Builder
	b.WriteString("## 任务信息\n")
	b.WriteString(fmt.Sprintf("- 状态: %s | 进度: %d/%d | 耗时: %s\n",
		info.Status, info.DoneCount, info.TotalCount, info.Duration))
	if info.Error != "" {
		b.WriteString(fmt.Sprintf("- 错误: %s\n", info.Error))
	}

	if manifestErr == nil && manifest != nil && len(manifest.Tasks) > 0 {
		b.WriteString("\n## 幻灯片概览\n")
		b.WriteString("| # | 标题 | 类型 | 状态 |\n")
		b.WriteString("|---|------|------|------|\n")
		for _, t := range manifest.Tasks {
			b.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n",
				t.PageIndex, t.Title, t.ContentType, t.Status))
		}
	}

	b.WriteString("\n## 对话内容\n")

	// 从数据库直接构建对话（每条消息都是 flush 后的完整内容，无碎片）。
	for _, m := range dbMessages {
		trimmed := strings.TrimSpace(m.Content)
		if trimmed == "" || trimmed == "..." || trimmed == "……" {
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

	ts.Mu.Lock()
	defer ts.Mu.Unlock()
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
	manifest, err := deck.ReadTasksManifest(ts.Info.WorkDir)
	if err == nil && manifest != nil {
		b.WriteString("\n## 当前页面列表\n")
		b.WriteString("| # | 标题 | 类型 | 状态 |\n")
		b.WriteString("|---|------|------|------|\n")
		for _, t := range manifest.Tasks {
			b.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n",
				t.PageIndex, t.Title, t.ContentType, t.Status))
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
