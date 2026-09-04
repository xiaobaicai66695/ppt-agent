package model

import "time"

// ConversationMessage — 任务对话历史中的单条消息。
type ConversationMessage struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	TaskID string `gorm:"size:64;index;not null" json:"task_id"`
	Role   string `gorm:"size:20;not null" json:"role"` // "user" or "assistant"
	// Content remains a legacy fallback for short messages. Larger messages use
	// ConversationMessageChunk so MySQL row limits do not truncate assistant turns.
	Content   string                     `gorm:"type:longtext;not null" json:"content"`
	Timestamp time.Time                  `json:"timestamp"`
	Chunks    []ConversationMessageChunk `gorm:"foreignKey:MessageID" json:"-"`
}

type ConversationMessageChunk struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	MessageID uint   `gorm:"uniqueIndex:idx_message_chunk_sequence;index;not null" json:"message_id"`
	Sequence  int    `gorm:"uniqueIndex:idx_message_chunk_sequence;not null" json:"sequence"`
	Content   string `gorm:"type:longtext;not null" json:"content"`
}
