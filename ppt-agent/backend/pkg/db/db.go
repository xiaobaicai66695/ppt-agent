package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	drivermysql "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/cloudwego/ppt-agent/pkg/logger"
)

var DB *gorm.DB

const dbOperationTimeout = 5 * time.Second

func withOperationTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), dbOperationTimeout)
}

// User — 邮箱注册，可选密码。
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"size:120;uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"size:255" json:"-"`
	IsAdmin   bool      `gorm:"default:false" json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
}

// UserAPIKey stores an account-level model API key override.
// Empty rows are not kept; absence means the service falls back to env config.
type UserAPIKey struct {
	UserID    uint      `gorm:"primaryKey" json:"user_id"`
	Provider  string    `gorm:"size:40;not null;default:'ark'" json:"provider"`
	APIKey    string    `gorm:"type:text;not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VerificationCode — 邮箱验证码。
type VerificationCode struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	Email     string    `gorm:"size:120;index;not null" json:"email"`
	Code      string    `gorm:"size:10;not null" json:"-"`
	Used      bool      `gorm:"default:false" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// TaskRecord — 持久化的任务元数据，用于历史记录和恢复。
type TaskRecord struct {
	ID                  string    `gorm:"size:64;primaryKey" json:"id"`
	UserID              uint      `gorm:"index;not null" json:"user_id"`
	Query               string    `gorm:"type:text" json:"query"`
	Status              string    `gorm:"size:20;not null;default:'running'" json:"status"`
	WorkDir             string    `gorm:"size:512" json:"work_dir"`
	DoneCount           int       `gorm:"default:0" json:"done_count"`
	TotalCount          int       `gorm:"default:0" json:"total_count"`
	Duration            string    `gorm:"size:50" json:"duration"`
	Error               string    `gorm:"type:text" json:"error"`
	PromptTokens        int64     `gorm:"default:0" json:"prompt_tokens"`
	CompletionTokens    int64     `gorm:"default:0" json:"completion_tokens"`
	TotalTokens         int64     `gorm:"default:0" json:"total_tokens"`
	Files               string    `gorm:"type:text" json:"files"`
	ConversationContent string    `gorm:"type:longtext" json:"conversation_content"` // 拼接后的对话内容
	FullAnswer          string    `gorm:"type:longtext" json:"full_answer"`          // 完整拼接的 LLM 回答（用于冷加载恢复）
	Intent              string    `gorm:"size:32;index" json:"intent"`
	ConversationID      string    `gorm:"size:64;index" json:"conversation_id"`
	SourceMessageID     string    `gorm:"size:64;index" json:"source_message_id"`
	ParentTaskID        string    `gorm:"size:64;index" json:"parent_task_id"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// TaskFeedback stores one owner's reusable evaluation of a delivered deck.
type TaskFeedback struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TaskID     string    `gorm:"size:64;uniqueIndex:idx_task_feedback_task_user;index;not null" json:"task_id"`
	UserID     uint      `gorm:"uniqueIndex:idx_task_feedback_task_user;index;not null" json:"user_id"`
	Rating     int       `gorm:"not null" json:"rating"`
	Suggestion string    `gorm:"type:text" json:"suggestion"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// PlanDraftRecord stores PPT Agent planning results that must not trigger PPT rendering.
type PlanDraftRecord struct {
	ID                string    `gorm:"size:64;primaryKey" json:"id"`
	UserID            uint      `gorm:"index;not null" json:"user_id"`
	ConversationID    string    `gorm:"size:64;index" json:"conversation_id"`
	SourceMessageID   string    `gorm:"size:64;index" json:"source_message_id"`
	Query             string    `gorm:"type:text" json:"query"`
	NormalizedRequest string    `gorm:"type:text" json:"normalized_request"`
	DraftContent      string    `gorm:"type:longtext" json:"draft_content"`
	Status            string    `gorm:"size:32;index;not null;default:'draft'" json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ConversationMessage — 任务对话历史中的单条消息。
type ConversationMessage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TaskID    string    `gorm:"size:64;index;not null" json:"task_id"`
	Role      string    `gorm:"size:20;not null" json:"role"` // "user" or "assistant"
	Content   string    `gorm:"type:text;not null" json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// TaskErrorAnalysis — 任务失败时的日志分析结果，用于迭代修复参考。
type TaskErrorAnalysis struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TaskID      string    `gorm:"size:64;index;not null" json:"task_id"`
	TriggerType string    `gorm:"size:20;not null" json:"trigger_type"` // "idle" 或 "failed"
	LogSnippet  string    `gorm:"type:longtext" json:"log_snippet"`     // 原始日志片段
	Analysis    string    `gorm:"type:longtext" json:"analysis"`        // LLM 分析结论
	RootCause   string    `gorm:"type:text" json:"root_cause"`          // 根本原因
	Suggestion  string    `gorm:"type:text" json:"suggestion"`          // 修复建议
	TokensUsed  int64     `gorm:"default:0" json:"tokens_used"`
	ModelUsed   string    `gorm:"size:100" json:"model_used"` // 分析使用的模型
	CreatedAt   time.Time `json:"created_at"`
}

// RuntimeEventRecord — 任务运行 Timeline 事件，保存完整 tool/LLM 入参、输出和元数据。
type RuntimeEventRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TaskID    string    `gorm:"size:64;uniqueIndex:idx_runtime_event_task_seq;index;not null" json:"task_id"`
	EventID   int64     `gorm:"uniqueIndex:idx_runtime_event_task_seq;not null" json:"event_id"`
	Timestamp time.Time `gorm:"index" json:"timestamp"`
	ElapsedMS int64     `gorm:"default:0" json:"elapsed_ms"`
	Kind      string    `gorm:"size:64;index" json:"kind"`
	Phase     string    `gorm:"size:64;index" json:"phase"`
	Name      string    `gorm:"size:128;index" json:"name"`
	Status    string    `gorm:"size:32;index" json:"status"`
	Detail    string    `gorm:"type:text" json:"detail"`
	Metadata  string    `gorm:"type:longtext" json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
}

// ── TaskRecord 增删改查 ──────────────────────────────────────────────────

func CreateTaskRecord(r *TaskRecord) error {
	return DB.Create(r).Error
}

// UpsertTaskRecord persists a task that may already have been created as a
// conversation.  A conversation task is promoted in-place when the user later
// decides to generate a deck, so inserting a second row with the same task ID
// would otherwise fail and leave the durable task state stale.
func UpsertTaskRecord(r *TaskRecord) error {
	if DB == nil {
		return nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Save(r).Error
}

// UpdateTaskRecord 向可更新字段添加 conversation_content。
func UpdateTaskRecord(id string, updates map[string]any) error {
	return DB.Model(&TaskRecord{}).Where("id = ?", id).Updates(updates).Error
}

func GetTaskRecord(id string) (*TaskRecord, error) {
	var r TaskRecord
	err := DB.Where("id = ?", id).First(&r).Error
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func ListTaskRecordsByUser(userID uint) ([]TaskRecord, error) {
	var records []TaskRecord
	err := DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&records).Error
	return records, err
}

func DeleteTaskRecord(id string) error {
	return DB.Where("id = ?", id).Delete(&TaskRecord{}).Error
}

func UpsertTaskFeedback(taskID string, userID uint, rating int, suggestion string) (*TaskFeedback, error) {
	if DB == nil { return nil, fmt.Errorf("database unavailable") }
	ctx, cancel := withOperationTimeout(); defer cancel()
	now := time.Now()
	record := TaskFeedback{TaskID: taskID, UserID: userID, Rating: rating, Suggestion: suggestion, CreatedAt: now, UpdatedAt: now}
	if err := DB.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "task_id"}, {Name: "user_id"}}, DoUpdates: clause.Assignments(map[string]any{"rating": rating, "suggestion": suggestion, "updated_at": now})}).Create(&record).Error; err != nil { return nil, err }
	return GetTaskFeedback(taskID, userID)
}

func GetTaskFeedback(taskID string, userID uint) (*TaskFeedback, error) {
	if DB == nil { return nil, nil }
	ctx, cancel := withOperationTimeout(); defer cancel()
	var record TaskFeedback
	err := DB.WithContext(ctx).Where("task_id = ? AND user_id = ?", taskID, userID).First(&record).Error
	if err == gorm.ErrRecordNotFound { return nil, nil }
	if err != nil { return nil, err }
	return &record, nil
}

// ── PlanDraftRecord CRUD ──────────────────────────────────────────────────

func CreatePlanDraft(r *PlanDraftRecord) error {
	if DB == nil {
		return nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Create(r).Error
}

func GetPlanDraft(id string) (*PlanDraftRecord, error) {
	if DB == nil {
		return nil, gorm.ErrRecordNotFound
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var r PlanDraftRecord
	err := DB.WithContext(ctx).Where("id = ?", id).First(&r).Error
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func ListPlanDraftsByUser(userID uint) ([]PlanDraftRecord, error) {
	if DB == nil {
		return nil, nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var records []PlanDraftRecord
	err := DB.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&records).Error
	return records, err
}

func UpdatePlanDraft(id string, updates map[string]any) error {
	if DB == nil {
		return nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Model(&PlanDraftRecord{}).Where("id = ?", id).Updates(updates).Error
}

// ── ConversationMessage 增删改查 ─────────────────────────────────────────────

func CreateConversationMessage(m *ConversationMessage) error {
	if DB == nil {
		return nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Create(m).Error
}

func ListConversationMessages(taskID string) ([]ConversationMessage, error) {
	if DB == nil {
		return nil, nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var msgs []ConversationMessage
	err := DB.WithContext(ctx).Where("task_id = ?", taskID).Order("timestamp ASC").Find(&msgs).Error
	return msgs, err
}

func DeleteConversationMessages(taskID string) error {
	if DB == nil {
		return nil
	}
	return DB.Where("task_id = ?", taskID).Delete(&ConversationMessage{}).Error
}

// ── RuntimeEventRecord 增删改查 ─────────────────────────────────────────────

func CreateRuntimeEvent(r *RuntimeEventRecord) error {
	if DB == nil || r == nil {
		return nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Create(r).Error
}

func ListRuntimeEvents(taskID string) ([]RuntimeEventRecord, error) {
	if DB == nil {
		return nil, nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var events []RuntimeEventRecord
	err := DB.WithContext(ctx).Where("task_id = ?", taskID).Order("event_id ASC").Find(&events).Error
	return events, err
}

func ListRuntimeEventSummaries(taskID string) ([]RuntimeEventRecord, error) {
	if DB == nil {
		return nil, nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var events []RuntimeEventRecord
	// Metadata is needed here so the web layer can derive a bounded, redacted
	// public summary for inline tool previews. Raw payloads are never returned by
	// the conversation endpoint.
	err := DB.WithContext(ctx).Select("id", "task_id", "event_id", "timestamp", "elapsed_ms", "kind", "phase", "name", "status", "detail", "metadata", "created_at").
		Where("task_id = ?", taskID).
		Order("event_id ASC").
		Find(&events).Error
	return events, err
}

func GetRuntimeEvent(taskID string, eventID int64) (*RuntimeEventRecord, error) {
	if DB == nil {
		return nil, nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var event RuntimeEventRecord
	err := DB.WithContext(ctx).Where("task_id = ? AND event_id = ?", taskID, eventID).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func DeleteRuntimeEvents(taskID string) error {
	if DB == nil {
		return nil
	}
	return DB.Where("task_id = ?", taskID).Delete(&RuntimeEventRecord{}).Error
}

// ── TaskErrorAnalysis 增删改查 ────────────────────────────────────────────────

func CreateTaskErrorAnalysis(a *TaskErrorAnalysis) error {
	return DB.Create(a).Error
}

func GetTaskErrorAnalysis(taskID string) ([]TaskErrorAnalysis, error) {
	var records []TaskErrorAnalysis
	err := DB.Where("task_id = ?", taskID).Order("created_at DESC").Find(&records).Error
	return records, err
}

func ListRecentErrorAnalyses(limit int) ([]TaskErrorAnalysis, error) {
	var records []TaskErrorAnalysis
	err := DB.Order("created_at DESC").Limit(limit).Find(&records).Error
	return records, err
}

// MarkZombieTasks 将所有状态为 "running" 的任务设置为 "failed"
// （在启动时调用 — 拥有这些任务的服务器进程已不存在）。
func MarkZombieTasks() error {
	return DB.Model(&TaskRecord{}).
		Where("status = ?", "running").
		Updates(map[string]any{
			"status": "failed",
			"error":  "服务器重启，任务中断",
		}).Error
}

// Init 使用生产级设置初始化数据库连接。
// - maxOpenConns/maxIdleConns：限制并发数据库负载
// - connMaxLifetime：在 MySQL 的 wait_timeout 过期前回收连接
// - connMaxIdleTime：关闭空闲时间超过 MySQL 的 interactive_timeout 的连接
// - tcpKeepalive：在 TCP 层检测死连接，而无需等待 MySQL 超时
func Init(dsn string) error {
	dsn, err := dsnWithTimeouts(dsn)
	if err != nil {
		return fmt.Errorf("parse MySQL DSN: %w", err)
	}

	DB, err = gorm.Open(gormmysql.New(gormmysql.Config{
		DSN:                       dsn,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("gorm.Open: %w", err)
	}

	sqlDB, _ := DB.DB()
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(3 * time.Minute)
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("DB ping failed: %w", err)
	}
	var currentDatabase string
	if err := sqlDB.QueryRow("SELECT DATABASE()").Scan(&currentDatabase); err != nil {
		return fmt.Errorf("resolve current database: %w", err)
	}
	if !isBusinessDatabase(currentDatabase) {
		return fmt.Errorf("refusing to migrate non-business database %q", currentDatabase)
	}

	if err := DB.AutoMigrate(&User{}, &UserAPIKey{}, &VerificationCode{}, &TaskRecord{}, &TaskFeedback{}, &PlanDraftRecord{}, &ConversationMessage{}, &RuntimeEventRecord{}, &TaskErrorAnalysis{}); err != nil {
		return fmt.Errorf("AutoMigrate: %w", err)
	}

	logger.Info("mysql_connected")
	return nil
}

func isBusinessDatabase(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "information_schema", "mysql", "performance_schema", "sys":
		return false
	default:
		return true
	}
}

func dsnWithTimeouts(dsn string) (string, error) {
	cfg, err := drivermysql.ParseDSN(dsn)
	if err != nil {
		return "", err
	}
	cfg.InterpolateParams = true
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 10 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 10 * time.Second
	}
	return cfg.FormatDSN(), nil
}

// ListAllUsers 返回所有用户（供管理员查看）。
func ListAllUsers() ([]User, error) {
	var users []User
	err := DB.Order("created_at DESC").Find(&users).Error
	return users, err
}

// ListAllTaskRecords 返回所有任务记录（供管理员查看）。
func ListAllTaskRecords(limit int) ([]TaskRecord, error) {
	var records []TaskRecord
	err := DB.Order("created_at DESC").Limit(limit).Find(&records).Error
	return records, err
}

// DeleteErrorAnalysis 删除指定 ID 的日志分析记录。
func DeleteErrorAnalysis(id uint) error {
	return DB.Where("id = ?", id).Delete(&TaskErrorAnalysis{}).Error
}

// UpsertUserAPIKey stores or replaces the current user's model key override.
func UpsertUserAPIKey(userID uint, provider, apiKey string) error {
	if DB == nil {
		return fmt.Errorf("database unavailable")
	}
	record, updates := buildUserAPIKeyUpsert(userID, provider, apiKey, time.Now())
	return DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(updates),
	}).Create(&record).Error
}

func buildUserAPIKeyUpsert(userID uint, provider, apiKey string, now time.Time) (UserAPIKey, map[string]any) {
	provider = strings.TrimSpace(provider)
	if provider == "" {
		provider = "ark"
	}
	apiKey = strings.TrimSpace(apiKey)
	record := UserAPIKey{
		UserID:    userID,
		Provider:  provider,
		APIKey:    apiKey,
		CreatedAt: now,
		UpdatedAt: now,
	}
	updates := map[string]any{
		"provider":   provider,
		"api_key":    apiKey,
		"updated_at": now,
	}
	return record, updates
}

// GetUserAPIKey returns the configured key override, or nil if the account uses defaults.
func GetUserAPIKey(userID uint) (*UserAPIKey, error) {
	if DB == nil {
		return nil, nil
	}
	var key UserAPIKey
	err := DB.Where("user_id = ?", userID).First(&key).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &key, err
}

// DeleteUserAPIKey removes the user's model key override.
func DeleteUserAPIKey(userID uint) error {
	if DB == nil {
		return nil
	}
	return DB.Where("user_id = ?", userID).Delete(&UserAPIKey{}).Error
}
