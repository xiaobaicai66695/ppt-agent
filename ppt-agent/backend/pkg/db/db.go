package db

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/cloudwego/ppt-agent/pkg/logger"
)

var DB *gorm.DB

// User — 邮箱注册，可选密码。
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"size:120;uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"size:255" json:"-"`
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

// TaskRecord — persisted task metadata for history/recovery.
type TaskRecord struct {
	ID         string    `gorm:"size:64;primaryKey" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`
	Query      string    `gorm:"type:text" json:"query"`
	Status     string    `gorm:"size:20;not null;default:'running'" json:"status"`
	WorkDir    string    `gorm:"size:512" json:"work_dir"`
	DoneCount  int       `gorm:"default:0" json:"done_count"`
	TotalCount int       `gorm:"default:0" json:"total_count"`
	Duration          string    `gorm:"size:50" json:"duration"`
	Error             string    `gorm:"type:text" json:"error"`
	PromptTokens      int64     `gorm:"default:0" json:"prompt_tokens"`
	CompletionTokens  int64     `gorm:"default:0" json:"completion_tokens"`
	TotalTokens       int64     `gorm:"default:0" json:"total_tokens"`
	Files      string    `gorm:"type:text" json:"files"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ConversationMessage — a single message in a task's conversation history.
type ConversationMessage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TaskID    string    `gorm:"size:64;index;not null" json:"task_id"`
	Role      string    `gorm:"size:20;not null" json:"role"` // "user" or "assistant"
	Content   string    `gorm:"type:text;not null" json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// UserStyleProfile — persisted user style preferences learned from past tasks.
type UserStyleProfile struct {
	UserID            uint      `gorm:"primaryKey" json:"user_id"`
	PreferredThemes   string    `gorm:"type:text" json:"preferred_themes"`   // JSON array
	PreferredColors   string    `gorm:"type:text" json:"preferred_colors"`   // JSON array
	ContentPatterns   string    `gorm:"type:text" json:"content_patterns"` // JSON array
	LayoutPreferences string    `gorm:"type:text" json:"-"`                // deprecated
	LanguageTone      string    `gorm:"size:50" json:"language_tone"`
	TypicalPageCount  int       `gorm:"default:0" json:"typical_page_count"`
	ContentTypes      string    `gorm:"type:text" json:"content_types"` // JSON map
	SpecialNotes      string    `gorm:"type:text" json:"special_notes"` // JSON array
	TaskCount         int       `gorm:"default:0" json:"task_count"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// TaskEvent — a single SSE event persisted for later replay / continue-chat.
type TaskEvent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TaskID    string    `gorm:"size:64;index;not null" json:"task_id"`
	Type      string    `gorm:"size:32;not null" json:"type"`            // answer, tool_call, progress, file_ready, error, complete, etc.
	Content   string    `gorm:"type:mediumtext;not null" json:"content"` // JSON of task.SSERichEvent
	CreatedAt time.Time `json:"created_at"`
}

// ── TaskRecord CRUD ──────────────────────────────────────────────────────

func CreateTaskRecord(r *TaskRecord) error {
	return DB.Create(r).Error
}

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

// ── TaskEvent CRUD ─────────────────────────────────────────────────────────

func CreateTaskEvent(e *TaskEvent) error {
	return DB.Create(e).Error
}

// BatchCreateTaskEvents inserts multiple events in a single transaction.
func BatchCreateTaskEvents(events []TaskEvent) error {
	if len(events) == 0 {
		return nil
	}
	return DB.Create(&events).Error
}

func ListTaskEvents(taskID string, limit int) ([]TaskEvent, error) {
	var events []TaskEvent
	q := DB.Where("task_id = ?", taskID).Order("id ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&events).Error
	return events, err
}

func DeleteTaskEvents(taskID string) error {
	return DB.Where("task_id = ?", taskID).Delete(&TaskEvent{}).Error
}

// ── ConversationMessage CRUD ───────────────────────────────────────────────

func CreateConversationMessage(m *ConversationMessage) error {
	return DB.Create(m).Error
}

func ListConversationMessages(taskID string) ([]ConversationMessage, error) {
	var msgs []ConversationMessage
	err := DB.Where("task_id = ?", taskID).Order("timestamp ASC").Find(&msgs).Error
	return msgs, err
}

func DeleteConversationMessages(taskID string) error {
	return DB.Where("task_id = ?", taskID).Delete(&ConversationMessage{}).Error
}

// ── UserStyleProfile CRUD ─────────────────────────────────────────────────

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

// MarkZombieTasks sets all tasks with "running" status to "failed"
// (called on startup — the server process that owned them is gone).
func MarkZombieTasks() error {
	return DB.Model(&TaskRecord{}).
		Where("status = ?", "running").
		Updates(map[string]any{
			"status": "failed",
			"error":  "服务器重启，任务中断",
		}).Error
}

// Init initializes the database connection with production-grade settings.
// - maxOpenConns/maxIdleConns: limit concurrent DB load
// - connMaxLifetime: recycle connections before MySQL's wait_timeout expires
// - connMaxIdleTime: close connections left idle longer than MySQL's interactive_timeout
// - tcpKeepalive: detect dead connections at the TCP layer without waiting for MySQL timeout
func Init(dsn string) error {
	var err error
	DB, err = gorm.Open(mysql.Open(dsn+"&interpolateParams=true"), &gorm.Config{})
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

	if err := DB.AutoMigrate(&User{}, &VerificationCode{}, &TaskRecord{}, &TaskEvent{}, &ConversationMessage{}, &UserStyleProfile{}); err != nil {
		return fmt.Errorf("AutoMigrate: %w", err)
	}

	logger.Info("mysql_connected")
	return nil
}

