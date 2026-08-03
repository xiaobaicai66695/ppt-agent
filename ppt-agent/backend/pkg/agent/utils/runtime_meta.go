/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
)

const runtimeRecentEventLimit = 80

// RuntimeMeta is a code-maintained status bar for an agent run. It keeps
// frequently needed operational facts out of long trajectory scans.
type RuntimeMeta struct {
	mu sync.RWMutex

	TaskID    string
	WorkDir   string
	StartedAt time.Time

	Phase        string
	PhaseDetail  string
	LastError    string
	LastTool     string
	LastToolArgs string

	ToolCalls          map[string]int
	ToolErrors         map[string]int
	SameToolArgsStreak int

	PromptTokens     int64
	CompletionTokens int64
	TotalTokens      int64

	CompressionBeforeTokens int
	CompressionAfterTokens  int
	CompressionSavedPct     string

	Budgets        RuntimeBudgets
	BudgetWarnings []string

	DoneSlides     int
	TotalSlides    int
	MissingFiles   int
	QAHighIssues   int
	QAMediumIssues int
	QALowIssues    int

	IntentAnchor      IntentAnchor
	PlanSlides        []PlanSlide
	CurrentSlide      *PlanSlide
	AlignmentStatus   string
	AlignmentWarnings []AlignmentWarning

	EventSeq     int64
	EventCounts  map[string]int
	RecentEvents []RuntimeEvent
}

type RuntimeBudgets struct {
	SameToolArgsWarn     int `json:"same_tool_args_warn,omitempty"`
	MaxToolCallsPerTool  int `json:"max_tool_calls_per_tool,omitempty"`
	MaxTotalToolCalls    int `json:"max_total_tool_calls,omitempty"`
	TokenWarn            int `json:"token_warn,omitempty"`
	PhaseDurationWarnSec int `json:"phase_duration_warn_sec,omitempty"`
}

type IntentAnchor struct {
	Summary        string `json:"summary,omitempty"`
	OriginalLength int    `json:"original_length,omitempty"`
	Intent         string `json:"intent,omitempty"`
	Domain         string `json:"domain,omitempty"`
	SuggestedPages int    `json:"suggested_pages,omitempty"`
	Template       string `json:"template,omitempty"`
	Theme          string `json:"theme,omitempty"`
	UseBackground  bool   `json:"use_background,omitempty"`
	Background     string `json:"background,omitempty"`
	Recommendation string `json:"recommendation,omitempty"`
}

type PlanSlide struct {
	PageIndex   int    `json:"page_index,omitempty"`
	TaskID      string `json:"task_id,omitempty"`
	Title       string `json:"title,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	OutputFile  string `json:"output_file,omitempty"`
	Status      string `json:"status,omitempty"`
}

type AlignmentWarning struct {
	Code      string `json:"code"`
	Step      string `json:"step"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	PageIndex int    `json:"page_index,omitempty"`
	Expected  string `json:"expected,omitempty"`
	Observed  string `json:"observed,omitempty"`
}

// RuntimeMetaSnapshot is the serializable form exposed to prompts, SSE and UI.
type RuntimeMetaSnapshot struct {
	TaskID    string `json:"task_id,omitempty"`
	WorkDir   string `json:"work_dir,omitempty"`
	ElapsedMS int64  `json:"elapsed_ms"`

	Phase       string `json:"phase,omitempty"`
	PhaseDetail string `json:"phase_detail,omitempty"`
	LastError   string `json:"last_error,omitempty"`

	LastTool           string         `json:"last_tool,omitempty"`
	ToolCalls          map[string]int `json:"tool_calls,omitempty"`
	ToolErrors         map[string]int `json:"tool_errors,omitempty"`
	SameToolArgsStreak int            `json:"same_tool_args_streak,omitempty"`

	PromptTokens     int64 `json:"prompt_tokens,omitempty"`
	CompletionTokens int64 `json:"completion_tokens,omitempty"`
	TotalTokens      int64 `json:"total_tokens,omitempty"`

	CompressionBeforeTokens int    `json:"compression_before_tokens,omitempty"`
	CompressionAfterTokens  int    `json:"compression_after_tokens,omitempty"`
	CompressionSavedPct     string `json:"compression_saved_pct,omitempty"`

	Budgets        RuntimeBudgets `json:"budgets,omitempty"`
	BudgetWarnings []string       `json:"budget_warnings,omitempty"`

	DoneSlides     int `json:"done_slides,omitempty"`
	TotalSlides    int `json:"total_slides,omitempty"`
	MissingFiles   int `json:"missing_files,omitempty"`
	QAHighIssues   int `json:"qa_high_issues,omitempty"`
	QAMediumIssues int `json:"qa_medium_issues,omitempty"`
	QALowIssues    int `json:"qa_low_issues,omitempty"`

	IntentAnchor      IntentAnchor       `json:"intent_anchor,omitempty"`
	PlanSlides        []PlanSlide        `json:"plan_slides,omitempty"`
	CurrentSlide      *PlanSlide         `json:"current_slide,omitempty"`
	AlignmentStatus   string             `json:"alignment_status,omitempty"`
	AlignmentWarnings []AlignmentWarning `json:"alignment_warnings,omitempty"`

	EventCounts  map[string]int `json:"event_counts,omitempty"`
	RecentEvents []RuntimeEvent `json:"recent_events,omitempty"`
}

type RuntimeEvent struct {
	ID        int64          `json:"id"`
	TaskID    string         `json:"task_id,omitempty"`
	Timestamp string         `json:"timestamp"`
	ElapsedMS int64          `json:"elapsed_ms"`
	Kind      string         `json:"kind"`
	Phase     string         `json:"phase,omitempty"`
	Name      string         `json:"name,omitempty"`
	Status    string         `json:"status,omitempty"`
	Detail    string         `json:"detail,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type RuntimeReport struct {
	TaskID      string              `json:"task_id,omitempty"`
	WorkDir     string              `json:"work_dir,omitempty"`
	Status      string              `json:"status"`
	WrittenAt   string              `json:"written_at"`
	Snapshot    RuntimeMetaSnapshot `json:"snapshot"`
	EventCounts map[string]int      `json:"event_counts,omitempty"`
}

type runtimeMetaKey struct{}

func NewRuntimeMeta(taskID, workDir string) *RuntimeMeta {
	return &RuntimeMeta{
		TaskID:          taskID,
		WorkDir:         workDir,
		StartedAt:       time.Now(),
		Phase:           "preparing",
		ToolCalls:       map[string]int{},
		ToolErrors:      map[string]int{},
		EventCounts:     map[string]int{},
		AlignmentStatus: "pending",
		Budgets: RuntimeBudgets{
			SameToolArgsWarn:     EnvInt("PPT_SAME_TOOL_ARGS_WARN", 3),
			MaxToolCallsPerTool:  EnvInt("PPT_MAX_TOOL_CALLS_PER_TOOL", 20),
			MaxTotalToolCalls:    EnvInt("PPT_MAX_TOTAL_TOOL_CALLS", 120),
			TokenWarn:            EnvInt("PPT_TOKEN_WARN", 180000),
			PhaseDurationWarnSec: EnvInt("PPT_PHASE_DURATION_WARN_SEC", 600),
		},
	}
}

func (m *RuntimeMeta) RecordIntent(anchor IntentAnchor) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	anchor.Summary = truncateString(strings.TrimSpace(anchor.Summary), 160)
	m.IntentAnchor = anchor
	m.recordEventLocked("intent_captured", "user_intent", "ok", anchor.Summary, map[string]any{
		"intent": anchor.Intent, "domain": anchor.Domain, "suggested_pages": anchor.SuggestedPages,
		"template": anchor.Template, "theme": anchor.Theme, "background": anchor.Background,
		"use_background": anchor.UseBackground, "recommendation": anchor.Recommendation,
	})
}

func (m *RuntimeMeta) FreezePlan(slides []PlanSlide) {
	if m == nil || len(slides) == 0 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.PlanSlides) > 0 {
		return
	}
	m.PlanSlides = clonePlanSlides(slides)
	m.AlignmentStatus = "aligned"
	m.recordEventLocked("plan_frozen", "tasks.json", "ok", fmt.Sprintf("%d slides", len(slides)), map[string]any{
		"slides": m.PlanSlides,
	})
}

func (m *RuntimeMeta) RecordCurrentSlide(slide *PlanSlide) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if slide == nil {
		m.CurrentSlide = nil
		return
	}
	copy := *slide
	m.CurrentSlide = &copy
}

func (m *RuntimeMeta) ComparePlan(observed []PlanSlide, missingFiles []string) {
	if m == nil || len(observed) == 0 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.PlanSlides) == 0 {
		m.PlanSlides = clonePlanSlides(observed)
		m.AlignmentStatus = "aligned"
		m.recordEventLocked("plan_frozen", "tasks.json", "ok", fmt.Sprintf("%d slides", len(observed)), map[string]any{"slides": m.PlanSlides})
		return
	}

	warnings := comparePlanSlides(m.PlanSlides, observed, missingFiles)
	nextStatus := "aligned"
	if len(warnings) > 0 {
		nextStatus = "warning"
	}
	if m.AlignmentStatus == nextStatus && reflect.DeepEqual(m.AlignmentWarnings, warnings) {
		return
	}
	m.AlignmentWarnings = warnings
	m.AlignmentStatus = nextStatus
	status := "ok"
	if len(warnings) > 0 {
		status = "warning"
	}
	m.recordEventLocked("intent_alignment", "plan_vs_execution", status, fmt.Sprintf("%d deviations", len(warnings)), map[string]any{
		"warnings": warnings,
	})
}

func WithRuntimeMeta(ctx context.Context, meta *RuntimeMeta) context.Context {
	if meta == nil {
		return ctx
	}
	return context.WithValue(ctx, runtimeMetaKey{}, meta)
}

func RuntimeMetaFromContext(ctx context.Context) *RuntimeMeta {
	if meta, ok := ctx.Value(runtimeMetaKey{}).(*RuntimeMeta); ok {
		return meta
	}
	return nil
}

func (m *RuntimeMeta) RecordPhase(phase, detail string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if strings.TrimSpace(phase) != "" {
		m.Phase = strings.TrimSpace(phase)
	}
	if strings.TrimSpace(detail) != "" {
		m.PhaseDetail = strings.TrimSpace(detail)
	}
	m.recordEventLocked("phase_changed", phase, "ok", detail, nil)
}

func (m *RuntimeMeta) RecordToolStart(name, args string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	name = strings.TrimSpace(name)
	if name == "" {
		name = "unknown"
	}
	m.ToolCalls[name]++
	if m.LastTool == name && m.LastToolArgs == args {
		m.SameToolArgsStreak++
	} else {
		m.SameToolArgsStreak = 1
	}
	m.LastTool = name
	m.LastToolArgs = args
	m.refreshBudgetWarningsLocked()
	m.recordEventLocked("tool_start", name, "running", "", map[string]any{
		"args_preview": truncateString(args, 220),
	})
}

func (m *RuntimeMeta) RecordToolEnd(name, resultPreview string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	name = strings.TrimSpace(name)
	if name == "" {
		name = "unknown"
	}
	m.recordEventLocked("tool_end", name, "ok", "", map[string]any{
		"result_preview": truncateString(strings.TrimSpace(resultPreview), 220),
	})
}

func (m *RuntimeMeta) RecordToolError(name, errText string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	name = strings.TrimSpace(name)
	if name == "" {
		name = "unknown"
	}
	m.ToolErrors[name]++
	m.LastError = truncateString(strings.TrimSpace(errText), 240)
	m.refreshBudgetWarningsLocked()
	m.recordEventLocked("tool_error", name, "error", m.LastError, nil)
}

func (m *RuntimeMeta) RecordLLMStart(name string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	name = strings.TrimSpace(name)
	if name == "" {
		name = "chat_model"
	}
	m.recordEventLocked("llm_start", name, "running", "", nil)
}

func (m *RuntimeMeta) RecordLLMTokens(prompt, completion, total int64) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PromptTokens += prompt
	m.CompletionTokens += completion
	m.TotalTokens += total
	m.recordEventLocked("llm_end", "chat_model", "ok", "", map[string]any{
		"prompt_tokens":     prompt,
		"completion_tokens": completion,
		"total_tokens":      total,
	})
}

func (m *RuntimeMeta) RecordLLMError(name, errText string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	name = strings.TrimSpace(name)
	if name == "" {
		name = "chat_model"
	}
	m.LastError = truncateString(strings.TrimSpace(errText), 240)
	m.recordEventLocked("llm_error", name, "error", m.LastError, nil)
}

func (m *RuntimeMeta) SetLLMTokens(prompt, completion, total int64) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PromptTokens = prompt
	m.CompletionTokens = completion
	m.TotalTokens = total
	m.refreshBudgetWarningsLocked()
}

func (m *RuntimeMeta) RecordCompression(beforeTokens, afterTokens int, savedPct string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CompressionBeforeTokens = beforeTokens
	m.CompressionAfterTokens = afterTokens
	m.CompressionSavedPct = savedPct
	m.recordEventLocked("compression", "context_compressor", "ok", "", map[string]any{
		"before_tokens": beforeTokens,
		"after_tokens":  afterTokens,
		"saved_pct":     savedPct,
	})
}

func (m *RuntimeMeta) RecordSlideProgress(done, total, missing int) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.DoneSlides == done && m.TotalSlides == total && m.MissingFiles == missing {
		return
	}
	m.DoneSlides = done
	m.TotalSlides = total
	m.MissingFiles = missing
	m.recordEventLocked("slide_progress", "tasks_manifest", "ok", "", map[string]any{
		"done":    done,
		"total":   total,
		"missing": missing,
	})
}

func (m *RuntimeMeta) RecordQAIssues(high, medium, low int) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.QAHighIssues += high
	m.QAMediumIssues += medium
	m.QALowIssues += low
	if high+medium+low > 0 {
		m.recordEventLocked("qa_issues", "qa", "warning", "", map[string]any{
			"high":   high,
			"medium": medium,
			"low":    low,
		})
	}
}

func (m *RuntimeMeta) RecordFileCreated(fileName string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordEventLocked("file_created", filepath.Base(fileName), "ok", fileName, nil)
}

func (m *RuntimeMeta) RecordManifestValidation(done, total int, missingFiles, pendingTasks []string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	status := "ok"
	if len(missingFiles) > 0 || len(pendingTasks) > 0 {
		status = "warning"
	}
	m.recordEventLocked("manifest_validated", "tasks.json", status, "", map[string]any{
		"done":          done,
		"total":         total,
		"missing_files": missingFiles,
		"pending_tasks": pendingTasks,
	})
}

func (m *RuntimeMeta) RecordTaskTerminal(status, detail string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	eventStatus := "ok"
	if status == "failed" || status == "cancelled" {
		eventStatus = status
	}
	m.recordEventLocked("task_terminal", "task", eventStatus, detail, map[string]any{
		"status": status,
	})
}

func (m *RuntimeMeta) RecordEvent(kind, name, status, detail string, metadata map[string]any) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordEventLocked(kind, name, status, detail, metadata)
}

func (m *RuntimeMeta) Snapshot() RuntimeMetaSnapshot {
	if m == nil {
		return RuntimeMetaSnapshot{}
	}
	m.RefreshBudgetWarnings()
	m.mu.RLock()
	defer m.mu.RUnlock()
	started := m.StartedAt
	if started.IsZero() {
		started = time.Now()
	}
	return RuntimeMetaSnapshot{
		TaskID:                  m.TaskID,
		WorkDir:                 m.WorkDir,
		ElapsedMS:               time.Since(started).Milliseconds(),
		Phase:                   m.Phase,
		PhaseDetail:             m.PhaseDetail,
		LastError:               m.LastError,
		LastTool:                m.LastTool,
		ToolCalls:               cloneIntMap(m.ToolCalls),
		ToolErrors:              cloneIntMap(m.ToolErrors),
		SameToolArgsStreak:      m.SameToolArgsStreak,
		PromptTokens:            m.PromptTokens,
		CompletionTokens:        m.CompletionTokens,
		TotalTokens:             m.TotalTokens,
		CompressionBeforeTokens: m.CompressionBeforeTokens,
		CompressionAfterTokens:  m.CompressionAfterTokens,
		CompressionSavedPct:     m.CompressionSavedPct,
		Budgets:                 m.Budgets,
		BudgetWarnings:          append([]string(nil), m.BudgetWarnings...),
		DoneSlides:              m.DoneSlides,
		TotalSlides:             m.TotalSlides,
		MissingFiles:            m.MissingFiles,
		QAHighIssues:            m.QAHighIssues,
		QAMediumIssues:          m.QAMediumIssues,
		QALowIssues:             m.QALowIssues,
		IntentAnchor:            m.IntentAnchor,
		PlanSlides:              clonePlanSlides(m.PlanSlides),
		CurrentSlide:            clonePlanSlide(m.CurrentSlide),
		AlignmentStatus:         m.AlignmentStatus,
		AlignmentWarnings:       append([]AlignmentWarning(nil), m.AlignmentWarnings...),
		EventCounts:             cloneIntMap(m.EventCounts),
		RecentEvents:            append([]RuntimeEvent(nil), m.RecentEvents...),
	}
}

func (m *RuntimeMeta) RefreshBudgetWarnings() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.refreshBudgetWarningsLocked()
}

func (m *RuntimeMeta) StatusBar() string {
	if m == nil {
		return ""
	}
	snap := m.Snapshot()
	lines := []string{
		"<agent_status>",
		fmt.Sprintf("task_id: %s", snap.TaskID),
		fmt.Sprintf("elapsed_ms: %d", snap.ElapsedMS),
		fmt.Sprintf("phase: %s", snap.Phase),
		fmt.Sprintf("slides: %d/%d done, %d missing", snap.DoneSlides, snap.TotalSlides, snap.MissingFiles),
		fmt.Sprintf("tokens: prompt=%d completion=%d total=%d", snap.PromptTokens, snap.CompletionTokens, snap.TotalTokens),
		fmt.Sprintf("last_tool: %s same_args_streak=%d", snap.LastTool, snap.SameToolArgsStreak),
		fmt.Sprintf("tool_calls: %s", formatIntMap(snap.ToolCalls)),
		fmt.Sprintf("tool_errors: %s", formatIntMap(snap.ToolErrors)),
		fmt.Sprintf("qa_issues: high=%d medium=%d low=%d", snap.QAHighIssues, snap.QAMediumIssues, snap.QALowIssues),
		fmt.Sprintf("intent: %s", snap.IntentAnchor.Summary),
		fmt.Sprintf("alignment: %s warnings=%d", snap.AlignmentStatus, len(snap.AlignmentWarnings)),
		fmt.Sprintf("compression: before=%d after=%d saved=%s", snap.CompressionBeforeTokens, snap.CompressionAfterTokens, snap.CompressionSavedPct),
	}
	if len(snap.BudgetWarnings) > 0 {
		lines = append(lines, "budget_warnings: "+strings.Join(snap.BudgetWarnings, " | "))
	}
	if snap.PhaseDetail != "" {
		lines = append(lines, "phase_detail: "+snap.PhaseDetail)
	}
	if snap.LastError != "" {
		lines = append(lines, "last_error: "+snap.LastError)
	}
	lines = append(lines,
		"policy: if same_args_streak>=3 or a tool has repeated errors, stop repeating the same action; inspect state, change strategy, or report a blocker.",
		"</agent_status>",
	)
	return strings.Join(lines, "\n")
}

func (m *RuntimeMeta) refreshBudgetWarningsLocked() {
	warnings := make([]string, 0, 4)
	b := m.Budgets
	if b.SameToolArgsWarn > 0 && m.SameToolArgsStreak >= b.SameToolArgsWarn {
		warnings = append(warnings, fmt.Sprintf("same tool+args repeated %d times; change strategy before retrying", m.SameToolArgsStreak))
	}
	totalCalls := 0
	for toolName, count := range m.ToolCalls {
		totalCalls += count
		if b.MaxToolCallsPerTool > 0 && count >= b.MaxToolCallsPerTool {
			warnings = append(warnings, fmt.Sprintf("tool %s reached %d calls", toolName, count))
		}
	}
	if b.MaxTotalToolCalls > 0 && totalCalls >= b.MaxTotalToolCalls {
		warnings = append(warnings, fmt.Sprintf("total tool calls reached %d", totalCalls))
	}
	if b.TokenWarn > 0 && m.TotalTokens >= int64(b.TokenWarn) {
		warnings = append(warnings, fmt.Sprintf("token usage reached %d", m.TotalTokens))
	}
	if b.PhaseDurationWarnSec > 0 && !m.StartedAt.IsZero() && time.Since(m.StartedAt) >= time.Duration(b.PhaseDurationWarnSec)*time.Second {
		warnings = append(warnings, fmt.Sprintf("run duration exceeded %ds; prefer finishing or asking for help", b.PhaseDurationWarnSec))
	}
	m.BudgetWarnings = warnings
}

func (m *RuntimeMeta) WriteReport(status string) error {
	if m == nil {
		return nil
	}
	snap := m.Snapshot()
	report := RuntimeReport{
		TaskID:      snap.TaskID,
		WorkDir:     snap.WorkDir,
		Status:      status,
		WrittenAt:   time.Now().Format(time.RFC3339Nano),
		Snapshot:    snap,
		EventCounts: snap.EventCounts,
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(m.WorkDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(m.WorkDir, "runtime_report.json"), data, 0644)
}

func (m *RuntimeMeta) recordEventLocked(kind, name, status, detail string, metadata map[string]any) {
	kind = strings.TrimSpace(kind)
	if kind == "" {
		kind = "event"
	}
	status = strings.TrimSpace(status)
	if status == "" {
		status = "ok"
	}
	started := m.StartedAt
	if started.IsZero() {
		started = time.Now()
	}
	m.EventSeq++
	event := RuntimeEvent{
		ID:        m.EventSeq,
		TaskID:    m.TaskID,
		Timestamp: time.Now().Format(time.RFC3339Nano),
		ElapsedMS: time.Since(started).Milliseconds(),
		Kind:      kind,
		Phase:     m.Phase,
		Name:      truncateString(strings.TrimSpace(name), 80),
		Status:    status,
		Detail:    truncateString(strings.TrimSpace(detail), 240),
		Metadata:  metadata,
	}
	if m.EventCounts == nil {
		m.EventCounts = map[string]int{}
	}
	m.EventCounts[kind]++
	m.RecentEvents = append(m.RecentEvents, event)
	if len(m.RecentEvents) > runtimeRecentEventLimit {
		m.RecentEvents = m.RecentEvents[len(m.RecentEvents)-runtimeRecentEventLimit:]
	}
	if m.WorkDir == "" {
		return
	}
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	if err := os.MkdirAll(m.WorkDir, 0755); err != nil {
		return
	}
	f, err := os.OpenFile(filepath.Join(m.WorkDir, "runtime_events.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(data, '\n'))
}

func (s RuntimeMetaSnapshot) MarshalJSONBytes() []byte {
	data, _ := json.Marshal(s)
	return data
}

func cloneIntMap(in map[string]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func clonePlanSlides(in []PlanSlide) []PlanSlide {
	return append([]PlanSlide(nil), in...)
}

func clonePlanSlide(in *PlanSlide) *PlanSlide {
	if in == nil {
		return nil
	}
	copy := *in
	return &copy
}

func comparePlanSlides(expected, observed []PlanSlide, missingFiles []string) []AlignmentWarning {
	warnings := make([]AlignmentWarning, 0)
	if len(expected) != len(observed) {
		warnings = append(warnings, AlignmentWarning{
			Code: "page_count_changed", Step: "planning", Severity: "warning",
			Message: "执行页数与冻结计划不一致", Expected: fmt.Sprintf("%d", len(expected)), Observed: fmt.Sprintf("%d", len(observed)),
		})
	}
	expectedByPage := make(map[int]PlanSlide, len(expected))
	for _, slide := range expected {
		expectedByPage[slide.PageIndex] = slide
	}
	observedPages := make(map[int]bool, len(observed))
	for _, slide := range observed {
		if observedPages[slide.PageIndex] {
			warnings = append(warnings, AlignmentWarning{
				Code: "duplicate_page_index", Step: "planning", Severity: "error", PageIndex: slide.PageIndex,
				Message: "多个页面使用了同一页码", Observed: slide.TaskID,
			})
			continue
		}
		observedPages[slide.PageIndex] = true
		plan, ok := expectedByPage[slide.PageIndex]
		if !ok {
			warnings = append(warnings, AlignmentWarning{
				Code: "unexpected_page", Step: "planning", Severity: "warning", PageIndex: slide.PageIndex,
				Message: "执行中出现计划外页面", Observed: slide.Title,
			})
			continue
		}
		if strings.TrimSpace(plan.Title) != strings.TrimSpace(slide.Title) {
			warnings = append(warnings, AlignmentWarning{
				Code: "title_changed", Step: "execution", Severity: "warning", PageIndex: slide.PageIndex,
				Message: "页面标题偏离冻结计划", Expected: plan.Title, Observed: slide.Title,
			})
		}
		if strings.TrimSpace(plan.ContentType) != strings.TrimSpace(slide.ContentType) {
			warnings = append(warnings, AlignmentWarning{
				Code: "content_type_changed", Step: "execution", Severity: "warning", PageIndex: slide.PageIndex,
				Message: "页面布局类型偏离冻结计划", Expected: plan.ContentType, Observed: slide.ContentType,
			})
		}
	}
	for _, slide := range expected {
		if !observedPages[slide.PageIndex] {
			warnings = append(warnings, AlignmentWarning{
				Code: "planned_page_missing", Step: "execution", Severity: "error", PageIndex: slide.PageIndex,
				Message: "冻结计划中的页面未进入执行清单", Expected: slide.Title,
			})
		}
	}
	for _, file := range missingFiles {
		warnings = append(warnings, AlignmentWarning{
			Code: "output_file_missing", Step: "delivery", Severity: "error",
			Message: "页面输出文件缺失", Expected: filepath.Base(file), Observed: "missing",
		})
	}
	if len(warnings) > 24 {
		warnings = warnings[:24]
	}
	return warnings
}

func formatIntMap(values map[string]int) string {
	if len(values) == 0 {
		return "{}"
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", k, values[k]))
	}
	return strings.Join(parts, ", ")
}
