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

	if err := DB.AutoMigrate(&User{}, &VerificationCode{}, &TaskRecord{}); err != nil {
		return fmt.Errorf("AutoMigrate: %w", err)
	}

	logger.Info("mysql_connected")
	return nil
}

