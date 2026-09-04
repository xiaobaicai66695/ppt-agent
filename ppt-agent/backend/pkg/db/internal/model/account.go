package model

import "time"

// User — 邮箱注册，可选密码。
type User struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Email       string    `gorm:"size:120;uniqueIndex;not null" json:"email"`
	Password    string    `gorm:"size:255" json:"-"`
	IsAdmin     bool      `gorm:"default:false" json:"is_admin"`
	GuestIPHash string    `gorm:"size:64;index" json:"-"`
	CreatedAt   time.Time `json:"created_at"`
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
