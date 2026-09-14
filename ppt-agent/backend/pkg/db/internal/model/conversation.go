package model

import "time"

// ConversationMessage — 任务对话历史中的单条消息。
type ConversationMessage struct {
	// The composite index is ordered exactly like the history query:
	// WHERE task_id = ? ORDER BY timestamp ASC, id ASC.  The older single
	// task_id index is left for AutoMigrate compatibility and can be removed
	// separately after production EXPLAIN evidence has been collected.
	ID     uint   `gorm:"primaryKey;index:idx_conversation_messages_task_timestamp_id,priority:3" json:"id"`
	TaskID string `gorm:"size:64;index;index:idx_conversation_messages_task_timestamp_id,priority:1;not null" json:"task_id"`
	Role   string `gorm:"size:20;not null" json:"role"` // "user" or "assistant"
	// A row is one complete conversation message. Stream chunks stay in memory
	// until an answer boundary, then are persisted as a single assistant row.
	Content   string    `gorm:"type:longtext;not null" json:"content"`
	Timestamp time.Time `gorm:"index:idx_conversation_messages_task_timestamp_id,priority:2" json:"timestamp"`
}

// ConversationTraceEvent is one user-visible, safe timeline event. It is kept
// separately from conversation messages because a tool/action trace is not an
// LLM turn, but must survive task eviction and service restart.
type ConversationTraceEvent struct {
	// History reads use the same task/time/ID ordering as messages.  Keeping
	// this index avoids a filesort as a task's visible execution trace grows.
	ID            uint      `gorm:"primaryKey;index:idx_conversation_trace_task_timestamp_id,priority:3" json:"id"`
	TaskID        string    `gorm:"size:64;index;index:idx_conversation_trace_task_timestamp_id,priority:1;not null" json:"task_id"`
	SourceEventID uint64    `gorm:"index" json:"source_event_id"`
	Type          string    `gorm:"size:48;index;not null" json:"type"`
	Payload       string    `gorm:"type:longtext;not null" json:"payload"`
	Timestamp     time.Time `gorm:"index;index:idx_conversation_trace_task_timestamp_id,priority:2" json:"timestamp"`
	CreatedAt     time.Time `json:"created_at"`
}
