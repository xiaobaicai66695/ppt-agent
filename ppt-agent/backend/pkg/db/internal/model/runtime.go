package model

import "time"

// TaskErrorAnalysis — 任务失败时的日志分析结果，用于迭代修复参考。
type TaskErrorAnalysis struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TaskID      string    `gorm:"size:64;index;not null" json:"task_id"`
	TriggerType string    `gorm:"size:20;not null" json:"trigger_type"` // "idle" 或 "failed"
	LogSnippet  string    `gorm:"type:longtext" json:"log_snippet"`     // 原始日志片段
	Analysis    string    `gorm:"type:longtext" json:"analysis"`        // LLM 分析结论
	RootCause   string    `gorm:"type:text" json:"root_cause"`          // 根本原因
	Suggestion  string    `gorm:"type:text" json:"suggestion"`          // 修复建议
	TokensUsed  int64     `gorm:"default:0" json:"tokens_used"`
	ModelUsed   string    `gorm:"size:100" json:"model_used"` // 分析使用的模型
	CreatedAt   time.Time `json:"created_at"`
}

// RuntimeEventRecord — 任务运行 Timeline 事件，保存完整 tool/LLM 入参、输出和元数据。
type RuntimeEventRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TaskID    string    `gorm:"size:64;uniqueIndex:idx_runtime_event_task_seq;index;not null" json:"task_id"`
	EventID   int64     `gorm:"uniqueIndex:idx_runtime_event_task_seq;not null" json:"event_id"`
	Timestamp time.Time `gorm:"index" json:"timestamp"`
	ElapsedMS int64     `gorm:"default:0" json:"elapsed_ms"`
	Kind      string    `gorm:"size:64;index" json:"kind"`
	Phase     string    `gorm:"size:64;index" json:"phase"`
	Name      string    `gorm:"size:128;index" json:"name"`
	Status    string    `gorm:"size:32;index" json:"status"`
	Detail    string    `gorm:"type:text" json:"detail"`
	Metadata  string    `gorm:"type:longtext" json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
}

// EvaluationSession is the private production-side root for evidence that may
// later become a reviewed benchmark case. It deliberately stores no user
// identity in exported payloads; UserID is retained here solely for access and
// withdrawal handling inside the production database.
type EvaluationSession struct {
	ID            string     `gorm:"size:64;primaryKey" json:"id"`
	SourceTaskID  string     `gorm:"size:64;uniqueIndex;index;not null" json:"source_task_id"`
	UserID        uint       `gorm:"index;not null" json:"user_id"`
	PrivacyStatus string     `gorm:"size:32;index;not null;default:'pending_redaction'" json:"privacy_status"`
	OutcomeJSON   string     `gorm:"type:longtext" json:"outcome_json"`
	WithdrawnAt   *time.Time `gorm:"index" json:"withdrawn_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// EvaluationArtifact indexes content-addressed JSON evidence. Artifact bytes
// live in evaluation storage rather than MySQL so retention and export do not
// inflate operational database backups.
type EvaluationArtifact struct {
	ID              string    `gorm:"size:64;primaryKey" json:"id"`
	SHA256          string    `gorm:"size:64;uniqueIndex;not null" json:"sha256"`
	StorageKey      string    `gorm:"size:512;uniqueIndex;not null" json:"storage_key"`
	Kind            string    `gorm:"size:64;index;not null" json:"kind"`
	ContentType     string    `gorm:"size:128;not null" json:"content_type"`
	SizeBytes       int64     `gorm:"not null" json:"size_bytes"`
	RedactionStatus string    `gorm:"size:32;index;not null;default:'pending_redaction'" json:"redaction_status"`
	CreatedAt       time.Time `json:"created_at"`
}

// EvaluationStageRun captures one authoritative Planner, Reviewer,
// PlannerRefiner, or Fixer boundary. Attempt is scoped by session and stage,
// allowing interrupted workflows to resume without overwriting evidence.
type EvaluationStageRun struct {
	ID               string    `gorm:"size:64;primaryKey" json:"id"`
	SessionID        string    `gorm:"size:64;index:idx_eval_stage_session_stage_attempt,priority:1;not null" json:"session_id"`
	Stage            string    `gorm:"size:32;index:idx_eval_stage_session_stage_attempt,priority:2;not null" json:"stage"`
	Attempt          int       `gorm:"index:idx_eval_stage_session_stage_attempt,priority:3;not null" json:"attempt"`
	ParentStageRunID string    `gorm:"size:64;index" json:"parent_stage_run_id"`
	InputArtifactID  string    `gorm:"size:64;index;not null" json:"input_artifact_id"`
	OutputArtifactID string    `gorm:"size:64;index;not null" json:"output_artifact_id"`
	Status           string    `gorm:"size:32;index;not null" json:"status"`
	ProvenanceJSON   string    `gorm:"type:longtext" json:"provenance_json"`
	CaptureError     string    `gorm:"type:text" json:"capture_error"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// EvaluationCandidate is the redaction and review queue entry derived from a
// stage run. It remains separate from benchmark cases because raw production
// evidence and a reviewed expected result have different lifecycles.
type EvaluationCandidate struct {
	ID                  string     `gorm:"size:64;primaryKey" json:"id"`
	SessionID           string     `gorm:"size:64;index;not null" json:"session_id"`
	StageRunID          string     `gorm:"size:64;uniqueIndex;not null" json:"stage_run_id"`
	Suite               string     `gorm:"size:32;index;not null" json:"suite"`
	Status              string     `gorm:"size:32;index;not null;default:'pending_review'" json:"status"`
	RedactionStatus     string     `gorm:"size:32;index;not null;default:'pending_redaction'" json:"redaction_status"`
	SelectionReason     string     `gorm:"type:text" json:"selection_reason"`
	QualitySignalsJSON  string     `gorm:"type:longtext" json:"quality_signals_json"`
	SemanticFingerprint string     `gorm:"size:128;index" json:"semantic_fingerprint"`
	SplitGroup          string     `gorm:"size:128;index" json:"split_group"`
	WithdrawnAt         *time.Time `gorm:"index" json:"withdrawn_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// EvaluationCaseRevision contains explicit reviewer-authored benchmark input,
// expected fields, and rubric. It is immutable once used by a frozen dataset.
type EvaluationCaseRevision struct {
	ID            string    `gorm:"size:64;primaryKey" json:"id"`
	CandidateID   string    `gorm:"size:64;index:idx_eval_case_candidate_revision,priority:1;not null" json:"candidate_id"`
	Revision      int       `gorm:"index:idx_eval_case_candidate_revision,priority:2;not null" json:"revision"`
	Suite         string    `gorm:"size:32;index;not null" json:"suite"`
	CaseJSON      string    `gorm:"type:longtext;not null" json:"case_json"`
	SplitGroup    string    `gorm:"size:128;index;not null" json:"split_group"`
	ApprovalState string    `gorm:"size:32;index;not null;default:'draft'" json:"approval_state"`
	CreatedAt     time.Time `json:"created_at"`
}

type EvaluationDatasetVersion struct {
	ID        string     `gorm:"size:64;primaryKey" json:"id"`
	Name      string     `gorm:"size:128;uniqueIndex;not null" json:"name"`
	Role      string     `gorm:"size:32;index;not null" json:"role"`
	State     string     `gorm:"size:32;index;not null;default:'draft'" json:"state"`
	ParentID  string     `gorm:"size:64;index" json:"parent_id"`
	FrozenAt  *time.Time `gorm:"index" json:"frozen_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type EvaluationDatasetMember struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	DatasetVersionID string    `gorm:"size:64;uniqueIndex:idx_eval_dataset_case,priority:1;index;not null" json:"dataset_version_id"`
	CaseRevisionID   string    `gorm:"size:64;uniqueIndex:idx_eval_dataset_case,priority:2;index;not null" json:"case_revision_id"`
	SplitGroup       string    `gorm:"size:128;index;not null" json:"split_group"`
	Ordinal          int       `gorm:"not null" json:"ordinal"`
	CreatedAt        time.Time `json:"created_at"`
}

type EvaluationExport struct {
	ID               string     `gorm:"size:64;primaryKey" json:"id"`
	DatasetVersionID string     `gorm:"size:64;index;not null" json:"dataset_version_id"`
	StorageKey       string     `gorm:"size:512;not null" json:"storage_key"`
	SHA256           string     `gorm:"size:64;not null" json:"sha256"`
	RevokedAt        *time.Time `gorm:"index" json:"revoked_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}
