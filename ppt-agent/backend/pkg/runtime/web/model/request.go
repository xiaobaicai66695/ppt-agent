package model

import "github.com/cloudwego/ppt-agent/pkg/agent/ppt"

// Authentication request payloads.
type SendCodeRequest struct {
	Email string `json:"email"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Code     string `json:"code"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Code     string `json:"code"`
	Password string `json:"password"`
}

type SetPasswordRequest struct {
	Password string `json:"password"`
}

// CreateTaskRequest is the public payload accepted by POST /api/tasks.
type CreateTaskRequest struct {
	Query   string           `json:"query"`
	Outline *ppt.TaskOutline `json:"outline,omitempty"`
}

// TaskCreateInput is the service-layer form of CreateTaskRequest.  It keeps
// authentication-derived values out of the HTTP decoder.
type TaskCreateInput struct {
	TaskID         string
	Query          string
	UserID         int
	Credential     ModelCredential
	Intent         string
	ConversationID string
	Outline        *ppt.TaskOutline
}

type StartConversationInput struct {
	TaskID     string
	Query      string
	UserID     int
	Credential ModelCredential
}

type ContinueTaskRequest struct {
	Message string `json:"message"`
}

type MessageRequest struct {
	Message        string `json:"message"`
	SelectedTaskID string `json:"selected_task_id,omitempty"`
	ManualMode     string `json:"manual_mode,omitempty"`
	WebSearch      bool   `json:"web_search,omitempty"`
	ImageSearch    bool   `json:"image_search,omitempty"`
}

type TaskFeedbackRequest struct {
	Rating     int    `json:"rating"`
	Suggestion string `json:"suggestion"`
}

type UpdateAPIKeyRequest struct {
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
}

type PlanDraftRequest struct {
	Query             string `json:"query"`
	NormalizedRequest string `json:"normalized_request,omitempty"`
	DraftContent      string `json:"draft_content,omitempty"`
	ConversationID    string `json:"conversation_id,omitempty"`
	SourceMessageID   string `json:"source_message_id,omitempty"`
}
