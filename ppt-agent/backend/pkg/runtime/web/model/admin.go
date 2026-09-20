package model

import "time"

type AdminTaskResponse struct {
	ID                   string     `json:"id"`
	UserID               uint       `json:"user_id"`
	UserEmail            string     `json:"user_email"`
	Query                string     `json:"query"`
	Status               string     `json:"status"`
	DoneCount            int        `json:"done_count"`
	TotalCount           int        `json:"total_count"`
	Duration             string     `json:"duration"`
	GenerationStartedAt  *time.Time `json:"generation_started_at,omitempty"`
	GenerationFinishedAt *time.Time `json:"generation_finished_at,omitempty"`
	GenerationDurationMS int64      `json:"generation_duration_ms"`
	FixerRunCount        int        `json:"fixer_run_count"`
	Error                string     `json:"error"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// AdminExecutionRecordResponse is the administrator-safe view of one durable
// task execution. Server-local paths and output file names are intentionally
// excluded.
type AdminExecutionRecordResponse struct {
	ID                   string     `json:"id"`
	UserID               uint       `json:"user_id"`
	UserEmail            string     `json:"user_email"`
	Query                string     `json:"query"`
	Status               string     `json:"status"`
	DoneCount            int        `json:"done_count"`
	TotalCount           int        `json:"total_count"`
	Duration             string     `json:"duration"`
	Error                string     `json:"error"`
	PromptTokens         int64      `json:"prompt_tokens"`
	CompletionTokens     int64      `json:"completion_tokens"`
	TotalTokens          int64      `json:"total_tokens"`
	Intent               string     `json:"intent"`
	ConversationID       string     `json:"conversation_id"`
	SourceMessageID      string     `json:"source_message_id"`
	ParentTaskID         string     `json:"parent_task_id"`
	GenerationStartedAt  *time.Time `json:"generation_started_at,omitempty"`
	GenerationFinishedAt *time.Time `json:"generation_finished_at,omitempty"`
	GenerationDurationMS int64      `json:"generation_duration_ms"`
	FixerRunCount        int        `json:"fixer_run_count"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type PlanDraftResponse struct {
	ID                string    `json:"id"`
	UserID            uint      `json:"user_id"`
	ConversationID    string    `json:"conversation_id"`
	SourceMessageID   string    `json:"source_message_id"`
	Query             string    `json:"query"`
	NormalizedRequest string    `json:"normalized_request"`
	DraftContent      string    `json:"draft_content"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
