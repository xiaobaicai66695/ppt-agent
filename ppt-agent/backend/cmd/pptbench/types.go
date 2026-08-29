package main

import (
	"encoding/json"
	"time"

	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
)

type options struct {
	dataset   string
	suite     string
	step      string
	casesPath string
	outPath   string
	limit     int
	caseID    string
	timeout   time.Duration
}

type benchCase struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Input      json.RawMessage `json:"input"`
	Expected   json.RawMessage `json:"expected,omitempty"`
	JudgeFocus []string        `json:"judge_focus,omitempty"`
	Raw        json.RawMessage `json:"-"`
}

type caseInput struct {
	UserRequest         string                 `json:"user_request"`
	UserMessage         string                 `json:"user_message"`
	HasOutline          bool                   `json:"has_outline"`
	HasExistingTask     bool                   `json:"has_existing_task"`
	TasksSummary        string                 `json:"tasks_summary"`
	ConversationContext []string               `json:"conversation_context"`
	DraftTasks          *deck.TasksManifest    `json:"draft_tasks"`
	BaseTasks           *deck.TasksManifest    `json:"base_tasks"`
	ReviewIssues        []deck.PlanReviewIssue `json:"review_issues"`
	AllowedPageIndexes  []int                  `json:"allowed_page_indexes"`
	SourceMaterials     []any                  `json:"source_materials"`
	Requirements        []string               `json:"requirements"`
}

type agentOutput struct {
	CaseID              string                 `json:"case_id"`
	Suite               string                 `json:"suite"`
	StartedAt           string                 `json:"started_at"`
	DurationMS          int64                  `json:"duration_ms"`
	Output              any                    `json:"output,omitempty"`
	Before              any                    `json:"before,omitempty"`
	After               any                    `json:"after,omitempty"`
	Events              []deck.AgentEvent      `json:"events,omitempty"`
	DeterministicReview *deck.PlanReviewReport `json:"deterministic_review,omitempty"`
	Error               string                 `json:"error,omitempty"`
}

type modelOutput struct {
	CaseID              string                 `json:"case_id"`
	Suite               string                 `json:"suite"`
	Output              any                    `json:"output,omitempty"`
	Before              any                    `json:"before,omitempty"`
	After               any                    `json:"after,omitempty"`
	DeterministicReview *deck.PlanReviewReport `json:"deterministic_review,omitempty"`
	Error               string                 `json:"error,omitempty"`
}

type judgeInput struct {
	Case                 any      `json:"case"`
	Suite                string   `json:"suite"`
	Rubric               string   `json:"rubric"`
	ModelOutput          any      `json:"model_output"`
	RequiredOutputSchema any      `json:"required_output_schema"`
	ScoringScale         []string `json:"scoring_scale"`
}

type judgeResult struct {
	CaseID           string         `json:"case_id"`
	Suite            string         `json:"suite"`
	Score            int            `json:"score"`
	Pass             bool           `json:"pass"`
	DimensionScores  map[string]int `json:"dimension_scores,omitempty"`
	Strengths        []string       `json:"strengths,omitempty"`
	Weaknesses       []string       `json:"weaknesses,omitempty"`
	CriticalFailures []string       `json:"critical_failures,omitempty"`
	RecommendedFix   string         `json:"recommended_fix,omitempty"`
	RawContent       string         `json:"raw_content,omitempty"`
	Error            string         `json:"error,omitempty"`
}

type caseSummary struct {
	CaseID string `json:"case_id"`
	Suite  string `json:"suite"`
	Score  int    `json:"score,omitempty"`
	Pass   bool   `json:"pass,omitempty"`
	Judged bool   `json:"judged,omitempty"`
	Error  string `json:"error,omitempty"`
}

type runSummary struct {
	Suite      string        `json:"suite"`
	StartedAt  string        `json:"started_at"`
	FinishedAt string        `json:"finished_at"`
	Total      int           `json:"total"`
	Judged     int           `json:"judged"`
	Passed     int           `json:"passed"`
	Average    float64       `json:"average_score,omitempty"`
	Cases      []caseSummary `json:"cases"`
}
