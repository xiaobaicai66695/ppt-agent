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

// ConversationTraceEvent is one user-visible, safe timeline event. It is kept
// separately from conversation messages because a tool/action trace is not an
// LLM turn, but must survive task eviction and service restart.
type ConversationTraceEvent struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TaskID        string    `gorm:"size:64;index;not null" json:"task_id"`
	SourceEventID uint64    `gorm:"index" json:"source_event_id"`
	Type          string    `gorm:"size:48;index;not null" json:"type"`
	Payload       string    `gorm:"type:longtext;not null" json:"payload"`
	Timestamp     time.Time `gorm:"index" json:"timestamp"`
	CreatedAt     time.Time `json:"created_at"`
}
