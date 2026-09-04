// Package model contains the transport models shared by the Web router and
// application services.  Keeping these types outside the HTTP adapter makes
// the JSON contract explicit without coupling services to Gin.
package model

import "time"

// Intent, mode and action identifiers are part of the message API contract.
const (
	IntentChat   = "chat"
	IntentCreate = "create"
	IntentPlan   = "plan"
	IntentFix    = "fix"

	ModeChat     = "chat"
	ModePPTAgent = "pptagent"

	ActionReply            = "reply"
	ActionPrepareCreate    = "prepare_create"
	ActionSavePlan         = "save_plan"
	ActionUpdateTask       = "update_task"
	ActionAskClarification = "ask_clarification"

	IntentClarifyTopic = "clarify_topic"
)

// CreateRequestRoute is the normalized result used by the create endpoint.
type CreateRequestRoute struct {
	Intent                string  `json:"intent"`
	Reason                string  `json:"reason"`
	ClarificationQuestion string  `json:"clarification_question,omitempty"`
	Confidence            float64 `json:"confidence,omitempty"`
}

// MessageRouteResult is the stable response contract for the unified message
// endpoint.  It is also consumed by the benchmark command.
type MessageRouteResult struct {
	Intent            string          `json:"intent"`
	Mode              string          `json:"mode"`
	Confidence        float64         `json:"confidence"`
	NeedsConfirmation bool            `json:"needs_confirmation"`
	NormalizedRequest string          `json:"normalized_request"`
	TaskID            string          `json:"task_id"`
	DraftID           string          `json:"draft_id,omitempty"`
	MissingFields     []string        `json:"missing_fields"`
	Action            string          `json:"action"`
	Reason            string          `json:"reason,omitempty"`
	Reply             string          `json:"reply,omitempty"`
	TaskCandidates    []TaskCandidate `json:"task_candidates,omitempty"`
	Streaming         bool            `json:"streaming,omitempty"`
	AfterEventID      uint64          `json:"after_event_id,omitempty"`
}

// TaskCandidate is a compact task projection shown when a message needs a
// user to select an existing ppt.
type TaskCandidate struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// RouteResult is the continuation intent contract.
type RouteResult struct {
	Intent string `json:"intent"`
	Reason string `json:"reason"`

	TargetPages           []int       `json:"target_pages,omitempty"`
	TargetTaskIDs         []string    `json:"target_task_ids,omitempty"`
	NeedsClarification    bool        `json:"needs_clarification,omitempty"`
	ClarificationQuestion string      `json:"clarification_question,omitempty"`
	FixDetails            *FixDetails `json:"fix_details,omitempty"`
	RegenerateScope       []int       `json:"regenerate_scope,omitempty"`
	SuggestFix            bool        `json:"suggest_fix,omitempty"`
}

// FixDetails describes a visual adjustment requested for a page.
type FixDetails struct {
	Aspect         string `json:"aspect"`
	Detail         string `json:"detail"`
	TargetElements string `json:"target_elements,omitempty"`
}

// ModelCredential is the provider/key pair selected for one request.
// APIKey is never serialized; the type is intentionally kept separate from
// HTTP response DTOs to make accidental leakage harder.
type ModelCredential struct {
	Provider string `json:"-"`
	APIKey   string `json:"-"`
}
