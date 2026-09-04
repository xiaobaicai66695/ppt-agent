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
