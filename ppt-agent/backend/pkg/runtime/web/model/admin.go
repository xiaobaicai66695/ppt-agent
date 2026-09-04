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
