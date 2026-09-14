package model

import "time"

// TaskRecord — 持久化的任务元数据，用于历史记录和恢复。
type TaskRecord struct {
	ID string `gorm:"size:64;primaryKey" json:"id"`
	// ListTaskRecordsByUser filters by user_id then orders by created_at.
	// This composite index keeps the workbench task list from sorting a user's
	// full history in memory as the account accumulates conversations.
	UserID               uint       `gorm:"index;index:idx_task_records_user_created,priority:1;not null" json:"user_id"`
	Query                string     `gorm:"type:text" json:"query"`
	Status               string     `gorm:"size:20;not null;default:'running'" json:"status"`
	WorkDir              string     `gorm:"size:512" json:"work_dir"`
	DoneCount            int        `gorm:"default:0" json:"done_count"`
	TotalCount           int        `gorm:"default:0" json:"total_count"`
	Duration             string     `gorm:"size:50" json:"duration"`
	Error                string     `gorm:"type:text" json:"error"`
	PromptTokens         int64      `gorm:"default:0" json:"prompt_tokens"`
	CompletionTokens     int64      `gorm:"default:0" json:"completion_tokens"`
	TotalTokens          int64      `gorm:"default:0" json:"total_tokens"`
	Files                string     `gorm:"type:text" json:"files"`
	Intent               string     `gorm:"size:32;index" json:"intent"`
	ConversationID       string     `gorm:"size:64;index" json:"conversation_id"`
	SourceMessageID      string     `gorm:"size:64;index" json:"source_message_id"`
	ParentTaskID         string     `gorm:"size:64;index" json:"parent_task_id"`
	GenerationStartedAt  *time.Time `gorm:"index" json:"generation_started_at,omitempty"`
	GenerationFinishedAt *time.Time `json:"generation_finished_at,omitempty"`
	GenerationDurationMS int64      `gorm:"default:0" json:"generation_duration_ms"`
	FixerRunCount        int        `gorm:"default:0" json:"fixer_run_count"`
	CreatedAt            time.Time  `gorm:"index:idx_task_records_user_created,priority:2" json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// TaskFeedback stores one owner's reusable evaluation of a delivered ppt.
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
