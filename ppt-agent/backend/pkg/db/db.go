package db

import (
	"fmt"
	"strings"
	"time"

	drivermysql "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/cloudwego/ppt-agent/pkg/logger"
)

var DB *gorm.DB

// User — 邮箱注册，可选密码。
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"size:120;uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"size:255" json:"-"`
	IsAdmin   bool      `gorm:"default:false" json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
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
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// ConversationMessage — 任务对话历史中的单条消息。
type ConversationMessage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TaskID    string    `gorm:"size:64;index;not null" json:"task_id"`
	Role      string    `gorm:"size:20;not null" json:"role"` // "user" or "assistant"
	Content   string    `gorm:"type:text;not null" json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// UserStyleProfile — 从过去任务中学习的持久化用户风格偏好。
type UserStyleProfile struct {
	UserID              uint      `gorm:"primaryKey" json:"user_id"`
	PreferredThemes     string    `gorm:"type:text" json:"preferred_themes"` // JSON array
	PreferredColors     string    `gorm:"type:text" json:"preferred_colors"` // JSON array
	ContentPatterns     string    `gorm:"type:text" json:"content_patterns"` // JSON array
	LayoutPreferences   string    `gorm:"type:text" json:"-"`                // deprecated
	LanguageTone        string    `gorm:"size:50" json:"language_tone"`
	TypicalPageCount    int       `gorm:"default:0" json:"typical_page_count"`
	ContentTypes        string    `gorm:"type:text" json:"content_types"`        // JSON map
	SpecialNotes        string    `gorm:"type:text" json:"special_notes"`        // JSON array
	ExtendedPreferences string    `gorm:"type:text" json:"extended_preferences"` // JSON for enhanced profile
	TaskCount           int       `gorm:"default:0" json:"task_count"`
	UpdatedAt           time.Time `json:"updated_at"`
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

// ── TaskRecord 增删改查 ──────────────────────────────────────────────────

func CreateTaskRecord(r *TaskRecord) error {
	return DB.Create(r).Error
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

// ── ConversationMessage 增删改查 ─────────────────────────────────────────────

func CreateConversationMessage(m *ConversationMessage) error {
	if DB == nil {
		return nil
	}
	return DB.Create(m).Error
}

func ListConversationMessages(taskID string) ([]ConversationMessage, error) {
	if DB == nil {
		return nil, nil
	}
	var msgs []ConversationMessage
	err := DB.Where("task_id = ?", taskID).Order("timestamp ASC").Find(&msgs).Error
	return msgs, err
}

func DeleteConversationMessages(taskID string) error {
	if DB == nil {
		return nil
	}
	return DB.Where("task_id = ?", taskID).Delete(&ConversationMessage{}).Error
}

// ── UserStyleProfile 增删改查 ─────────────────────────────────────────────────

func UpsertUserStyleProfile(p *UserStyleProfile) error {
	return DB.Save(p).Error
}

func GetUserStyleProfile(userID uint) (*UserStyleProfile, error) {
	var p UserStyleProfile
	err := DB.Where("user_id = ?", userID).First(&p).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &p, err
}

func DeleteUserStyleProfile(userID uint) error {
	return DB.Where("user_id = ?", userID).Delete(&UserStyleProfile{}).Error
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

	if err := DB.AutoMigrate(&User{}, &VerificationCode{}, &TaskRecord{}, &ConversationMessage{}, &UserStyleProfile{}, &TaskErrorAnalysis{}); err != nil {
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

// ListAllStyleProfiles 返回所有用户风格偏好。
func ListAllStyleProfiles() ([]UserStyleProfile, error) {
	var profiles []UserStyleProfile
	err := DB.Find(&profiles).Error
	return profiles, err
}

// DeleteErrorAnalysis 删除指定 ID 的日志分析记录。
func DeleteErrorAnalysis(id uint) error {
	return DB.Where("id = ?", id).Delete(&TaskErrorAnalysis{}).Error
}
