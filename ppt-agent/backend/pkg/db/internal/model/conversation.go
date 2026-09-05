package model

import "time"

// ConversationMessage — 任务对话历史中的单条消息。
type ConversationMessage struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	TaskID string `gorm:"size:64;index;not null" json:"task_id"`
	Role   string `gorm:"size:20;not null" json:"role"` // "user" or "assistant"
	// A row is one complete conversation message. Stream chunks stay in memory
	// until an answer boundary, then are persisted as a single assistant row.
	Content   string    `gorm:"type:longtext;not null" json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
