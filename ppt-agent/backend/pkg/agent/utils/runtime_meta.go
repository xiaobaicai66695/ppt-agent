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
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

const runtimeRecentEventLimit = 80
const runtimeMetadataPreviewStringLimit = 500
const runtimeMetadataRawStringLimit = 20000

var runtimeSecretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)((?:authorization|cookie)\s*[:=]\s*)([^\r\n]+)`),
	regexp.MustCompile(`(?i)("?(?:api[_-]?key|access[_-]?token|refresh[_-]?token|authorization|cookie|password|passwd|secret|private[_-]?key)"?\s*[:=]\s*)("[^"]*"|[^,\s}]+)`),
	regexp.MustCompile(`(?i)(bearer\s+)[a-z0-9._\-+/=]{12,}`),
	regexp.MustCompile(`(?i)(sk-)[a-z0-9]{12,}`),
}

type RuntimeEventSink func(event RuntimeEvent)

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

	CompressionBeforeTokens    int
	CompressionAfterTokens     int
	CompressionSavedPct        string
	CompressionBeforeMessages  int
	CompressionAfterMessages   int
	CompressionRemovedMessages int
	CompressionSavedTokens     int

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
	EventSink    RuntimeEventSink

	lastManifestValidation *manifestValidationState
}

type manifestValidationState struct {
	Done         int
	Total        int
	MissingFiles []string
	PendingTasks []string
	Status       string
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

	CompressionBeforeTokens    int    `json:"compression_before_tokens,omitempty"`
	CompressionAfterTokens     int    `json:"compression_after_tokens,omitempty"`
	CompressionSavedPct        string `json:"compression_saved_pct,omitempty"`
	CompressionBeforeMessages  int    `json:"compression_before_messages,omitempty"`
	CompressionAfterMessages   int    `json:"compression_after_messages,omitempty"`
	CompressionRemovedMessages int    `json:"compression_removed_messages,omitempty"`
	CompressionSavedTokens     int    `json:"compression_saved_tokens,omitempty"`

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
	m.recordEventLocked("intent_classified", "user_intent", "ok", anchor.Summary, map[string]any{
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
	m.recordEventLocked("deck_spec_frozen", "tasks.json", "ok", fmt.Sprintf("%d slides", len(slides)), map[string]any{
		"slide_count": len(m.PlanSlides),
		"slides":      m.PlanSlides,
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
		m.recordEventLocked("deck_spec_frozen", "tasks.json", "ok", fmt.Sprintf("%d slides", len(observed)), map[string]any{
			"slide_count": len(m.PlanSlides),
			"slides":      m.PlanSlides,
		})
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
	m.recordEventLocked("deck_spec_alignment", "plan_vs_render", status, fmt.Sprintf("%d deviations", len(warnings)), map[string]any{
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

func (m *RuntimeMeta) SetEventSink(sink RuntimeEventSink) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.EventSink = sink
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
	m.recordEventLocked(runtimeToolEventKind(name, "start"), name, "running", "", compactToolMetadata(name, args, ""))
}

func (m *RuntimeMeta) RecordToolEnd(name, args, result string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	name = strings.TrimSpace(name)
	if name == "" {
		name = "unknown"
	}
	m.recordEventLocked(runtimeToolEventKind(name, "end"), name, "ok", "", compactToolMetadata(name, args, result))
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
	m.recordEventLocked(runtimeToolEventKind(name, "error"), name, "error", m.LastError, nil)
}

func (m *RuntimeMeta) RecordLLMStart(name string) {
	m.RecordLLMStartDetails(name, nil)
}

func (m *RuntimeMeta) RecordLLMStartDetails(name string, metadata map[string]any) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	name = strings.TrimSpace(name)
	if name == "" {
		name = "chat_model"
	}
	m.recordEventLocked("llm_start", name, "running", "", metadata)
}

func (m *RuntimeMeta) RecordLLMTokens(prompt, completion, total int64) {
	m.RecordLLMEndDetails("chat_model", prompt, completion, total, nil)
}

func (m *RuntimeMeta) RecordLLMEndDetails(name string, prompt, completion, total int64, metadata map[string]any) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PromptTokens += prompt
	m.CompletionTokens += completion
	m.TotalTokens += total
	name = strings.TrimSpace(name)
	if name == "" {
		name = "chat_model"
	}
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadata["prompt_tokens"] = prompt
	metadata["completion_tokens"] = completion
	metadata["total_tokens"] = total
	m.recordEventLocked("llm_end", name, "ok", "", metadata)
}

func (m *RuntimeMeta) RecordLLMEnd(name string, metadata map[string]any) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	name = strings.TrimSpace(name)
	if name == "" {
		name = "chat_model"
	}
	m.recordEventLocked("llm_end", name, "ok", "", metadata)
}

func (m *RuntimeMeta) RecordLLMErrorDetails(name, errText string, metadata map[string]any) {
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
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadata["error"] = errText
	m.recordEventLocked("llm_error", name, "error", m.LastError, metadata)
}

func (m *RuntimeMeta) RecordToolErrorDetails(name, errText string, metadata map[string]any) {
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
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadata["error"] = errText
	m.recordEventLocked(runtimeToolEventKind(name, "error"), name, "error", m.LastError, metadata)
}

func (m *RuntimeMeta) RecordLLMTokensLegacy(prompt, completion, total int64) {
	m.RecordLLMEndDetails("chat_model", prompt, completion, total, map[string]any{
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
	m.RecordCompressionDetails(0, 0, beforeTokens, afterTokens, "", nil)
}

func (m *RuntimeMeta) RecordCompressionDetails(beforeMessages, afterMessages, beforeTokens, afterTokens int, userIntent string, requirements []string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CompressionBeforeTokens = beforeTokens
	m.CompressionAfterTokens = afterTokens
	m.CompressionBeforeMessages = beforeMessages
	m.CompressionAfterMessages = afterMessages
	m.CompressionRemovedMessages = maxInt(0, beforeMessages-afterMessages)
	m.CompressionSavedTokens = maxInt(0, beforeTokens-afterTokens)
	savedPct := "0.0%"
	if beforeTokens > 0 {
		savedPct = fmt.Sprintf("%.1f%%", 100*(1-float64(afterTokens)/float64(beforeTokens)))
	}
	m.CompressionSavedPct = savedPct
	m.recordEventLocked("planner_context_compressed", "context_compressor", "ok", "", map[string]any{
		"before_messages":        beforeMessages,
		"after_messages":         afterMessages,
		"removed_messages":       m.CompressionRemovedMessages,
		"before_tokens":          beforeTokens,
		"after_tokens":           afterTokens,
		"saved_tokens":           m.CompressionSavedTokens,
		"saved_pct":              savedPct,
		"user_intent_summary":    truncateString(strings.TrimSpace(userIntent), 240),
		"preserved_requirements": boundedStrings(requirements, 6, 180),
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
	m.recordEventLocked("delivery_progress", "tasks_manifest", "ok", "", map[string]any{
		"done":  done,
		"total": total,
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
	m.recordEventLocked("delivery_file_created", filepath.Base(fileName), "ok", fileName, nil)
}

func (m *RuntimeMeta) RecordManifestValidation(done, total int, missingFiles, pendingTasks []string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	status := "ok"
	if total <= 0 || len(pendingTasks) > 0 {
		status = "running"
	}
	if len(missingFiles) > 0 {
		status = "warning"
	}
	state := manifestValidationState{
		Done:         done,
		Total:        total,
		MissingFiles: append([]string(nil), missingFiles...),
		PendingTasks: append([]string(nil), pendingTasks...),
		Status:       status,
	}
	if m.lastManifestValidation != nil && reflect.DeepEqual(*m.lastManifestValidation, state) {
		return
	}
	m.lastManifestValidation = &state
	detail := fmt.Sprintf("已完成 %d/%d 页", done, total)
	if len(missingFiles) > 0 {
		detail += fmt.Sprintf("，缺少 %d 个文件", len(missingFiles))
	} else if len(pendingTasks) > 0 {
		detail += fmt.Sprintf("，还有 %d 页待生成", len(pendingTasks))
	}
	m.recordEventLocked("deck_spec_validated", "tasks.json", status, detail, map[string]any{
		"done":  done,
		"total": total,
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
		TaskID:                     m.TaskID,
		WorkDir:                    m.WorkDir,
		ElapsedMS:                  time.Since(started).Milliseconds(),
		Phase:                      m.Phase,
		PhaseDetail:                m.PhaseDetail,
		LastError:                  m.LastError,
		LastTool:                   m.LastTool,
		ToolCalls:                  cloneIntMap(m.ToolCalls),
		ToolErrors:                 cloneIntMap(m.ToolErrors),
		SameToolArgsStreak:         m.SameToolArgsStreak,
		PromptTokens:               m.PromptTokens,
		CompletionTokens:           m.CompletionTokens,
		TotalTokens:                m.TotalTokens,
		CompressionBeforeTokens:    m.CompressionBeforeTokens,
		CompressionAfterTokens:     m.CompressionAfterTokens,
		CompressionSavedPct:        m.CompressionSavedPct,
		CompressionBeforeMessages:  m.CompressionBeforeMessages,
		CompressionAfterMessages:   m.CompressionAfterMessages,
		CompressionRemovedMessages: m.CompressionRemovedMessages,
		CompressionSavedTokens:     m.CompressionSavedTokens,
		Budgets:                    m.Budgets,
		BudgetWarnings:             append([]string(nil), m.BudgetWarnings...),
		DoneSlides:                 m.DoneSlides,
		TotalSlides:                m.TotalSlides,
		MissingFiles:               m.MissingFiles,
		QAHighIssues:               m.QAHighIssues,
		QAMediumIssues:             m.QAMediumIssues,
		QALowIssues:                m.QALowIssues,
		IntentAnchor:               m.IntentAnchor,
		PlanSlides:                 clonePlanSlides(m.PlanSlides),
		CurrentSlide:               clonePlanSlide(m.CurrentSlide),
		AlignmentStatus:            m.AlignmentStatus,
		AlignmentWarnings:          append([]AlignmentWarning(nil), m.AlignmentWarnings...),
		EventCounts:                cloneIntMap(m.EventCounts),
		RecentEvents:               runtimeEventSummaries(m.RecentEvents),
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
	m.mu.RLock()
	done, total := m.DoneSlides, m.TotalSlides
	m.mu.RUnlock()
	return fmt.Sprintf("<agent_progress>\ngenerated_slides: %d\ntotal_slides: %d\n</agent_progress>", done, total)
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
	snap.RecentEvents = nil
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

func LoadRuntimeMetaSnapshot(workDir string) (*RuntimeMetaSnapshot, error) {
	workDir = strings.TrimSpace(workDir)
	if workDir == "" {
		return nil, fmt.Errorf("workDir is required")
	}
	reportPath := filepath.Join(workDir, "runtime_report.json")
	data, reportErr := os.ReadFile(reportPath)
	if reportErr != nil {
		return nil, reportErr
	}
	var report RuntimeReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}
	snap := report.Snapshot
	if snap.WorkDir == "" {
		snap.WorkDir = workDir
	}
	return &snap, nil
}

func countRuntimeEvents(events []RuntimeEvent) map[string]int {
	if len(events) == 0 {
		return nil
	}
	counts := make(map[string]int)
	for _, event := range events {
		kind := strings.TrimSpace(event.Kind)
		if kind == "" {
			kind = "event"
		}
		counts[kind]++
	}
	return counts
}

func runtimeEventSummaries(events []RuntimeEvent) []RuntimeEvent {
	if len(events) == 0 {
		return nil
	}
	summaries := make([]RuntimeEvent, len(events))
	for i, event := range events {
		summaries[i] = RuntimeEventSummary(event)
	}
	return summaries
}

// RuntimeEventSummary removes heavy/auditable payloads from timeline summaries while
// preserving assistant-visible LLM output for the chat transcript.
func RuntimeEventSummary(event RuntimeEvent) RuntimeEvent {
	summary := event
	summary.Metadata = nil
	if safe := publicRuntimeEventMetadata(event); len(safe) > 0 {
		summary.Metadata = safe
	}
	return summary
}

func publicRuntimeEventMetadata(event RuntimeEvent) map[string]any {
	kind := strings.ToLower(strings.TrimSpace(event.Kind))
	if !strings.Contains(kind, "llm") || strings.Contains(kind, "start") {
		return nil
	}
	output, ok := event.Metadata["assistant_output"].(string)
	if !ok {
		return nil
	}
	output = strings.TrimSpace(output)
	if output == "" {
		return nil
	}
	return map[string]any{"assistant_output": truncateString(redactRuntimeMetadataString("assistant_output", output), runtimeMetadataRawStringLimit)}
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
		Metadata:  compactRuntimeMetadata(metadata, 0),
	}
	if m.EventCounts == nil {
		m.EventCounts = map[string]int{}
	}
	m.EventCounts[kind]++
	m.RecentEvents = append(m.RecentEvents, event)
	if len(m.RecentEvents) > runtimeRecentEventLimit {
		m.RecentEvents = m.RecentEvents[len(m.RecentEvents)-runtimeRecentEventLimit:]
	}
	if m.EventSink != nil {
		m.EventSink(event)
	}
}

func runtimeToolEventKind(name, suffix string) string {
	if strings.EqualFold(strings.TrimSpace(name), "generate_slide") {
		return "slide_render_" + suffix
	}
	return "tool_" + suffix
}

func compactToolMetadata(name, args, result string) map[string]any {
	metadata := map[string]any{}
	if strings.TrimSpace(args) != "" {
		metadata["args"] = strings.TrimSpace(args)
		metadata["args_preview"] = truncateString(strings.TrimSpace(args), 220)
	}
	if strings.TrimSpace(result) != "" {
		metadata["result"] = strings.TrimSpace(result)
		metadata["result_preview"] = truncateString(strings.TrimSpace(result), 220)
	}
	addToolObservationFields(metadata, name, args, result)
	if strings.EqualFold(strings.TrimSpace(name), "generate_slide") {
		addJSONField(metadata, args, "task_id")
		addJSONField(metadata, result, "task_id")
		addJSONField(metadata, result, "content_type")
		addJSONField(metadata, result, "output_file")
	}
	if len(metadata) == 0 {
		return nil
	}
	return metadata
}

func addToolObservationFields(metadata map[string]any, name, args, result string) {
	if metadata == nil {
		return
	}
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "search":
		addJSONFieldAs(metadata, args, "query", "search_query")
		addJSONFieldAs(metadata, args, "reason", "search_reason")
		if count, titles, errText := extractSearchSummary(result, 5); count > 0 || errText != "" {
			metadata["source_count"] = count
			if len(titles) > 0 {
				metadata["source_titles"] = titles
			}
			if errText != "" {
				metadata["error"] = errText
			}
		}
		if urls := extractSearchURLs(result, 5); len(urls) > 0 {
			metadata["source_urls"] = urls
		}
	case "search_images":
		addJSONFieldAs(metadata, args, "query", "image_query")
		addJSONFieldAs(metadata, args, "asset_purpose", "asset_purpose")
		addJSONFieldAs(metadata, args, "asset_subject", "asset_subject")
		addJSONFieldAs(metadata, args, "composition", "composition")
		addJSONFieldAs(metadata, args, "reason", "search_reason")
		addJSONFieldAs(metadata, result, "provider", "provider")
		addJSONFieldAs(metadata, result, "asset_query", "asset_query")
		addJSONFieldAs(metadata, result, "total", "total")
		addJSONFieldAs(metadata, result, "total_pages", "total_pages")
		addJSONFieldAs(metadata, result, "error", "error")
		if previews, urls := extractImageSearchResults(result, 6); len(previews) > 0 {
			metadata["image_results"] = previews
			if len(urls) > 0 {
				metadata["source_urls"] = urls
			}
		}
	case "read_file":
		addJSONFieldAs(metadata, args, "path", "file_path")
	case "update_tasks_manifest":
		if count := countManifestTasks(args); count > 0 {
			metadata["slide_count"] = count
		}
		addJSONFieldAs(metadata, args, "template", "template")
		addJSONFieldAs(metadata, args, "theme", "theme")
		addJSONFieldAs(metadata, args, "background", "background")
	}
}

func addJSONField(metadata map[string]any, raw, key string) {
	addJSONFieldAs(metadata, raw, key, key)
}

func addJSONFieldAs(metadata map[string]any, raw, key, target string) {
	if strings.TrimSpace(raw) == "" || metadata == nil {
		return
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return
	}
	if value, ok := parsed[key]; ok {
		metadata[target] = value
	}
}

func countManifestTasks(raw string) int {
	if strings.TrimSpace(raw) == "" {
		return 0
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return 0
	}
	if tasks, ok := parsed["tasks"].([]any); ok {
		return len(tasks)
	}
	if tasks, ok := parsed["slides"].([]any); ok {
		return len(tasks)
	}
	return 0
}

func extractSearchSummary(raw string, limit int) (int, []string, string) {
	if strings.TrimSpace(raw) == "" {
		return 0, nil, ""
	}
	var parsed struct {
		Results []struct {
			Title string `json:"title"`
		} `json:"results"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return 0, nil, ""
	}
	titles := make([]string, 0, minInt(len(parsed.Results), limit))
	for _, result := range parsed.Results {
		title := truncateString(strings.TrimSpace(result.Title), 120)
		if title == "" {
			continue
		}
		titles = append(titles, title)
		if len(titles) >= limit {
			break
		}
	}
	return len(parsed.Results), titles, truncateString(strings.TrimSpace(parsed.Error), 180)
}

func extractSearchURLs(raw string, limit int) []string {
	if strings.TrimSpace(raw) == "" || limit <= 0 {
		return nil
	}
	var parsed struct {
		Results []struct {
			URL string `json:"url"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil
	}
	urls := make([]string, 0, minInt(len(parsed.Results), limit))
	seen := map[string]struct{}{}
	for _, result := range parsed.Results {
		url := truncateString(strings.TrimSpace(result.URL), 240)
		if url == "" {
			continue
		}
		if _, ok := seen[url]; ok {
			continue
		}
		seen[url] = struct{}{}
		urls = append(urls, url)
		if len(urls) >= limit {
			break
		}
	}
	return urls
}

func extractImageSearchResults(raw string, limit int) ([]map[string]any, []string) {
	if strings.TrimSpace(raw) == "" || limit <= 0 {
		return nil, nil
	}
	var parsed struct {
		Provider     string `json:"provider"`
		AssetPurpose string `json:"asset_purpose"`
		AssetQuery   string `json:"asset_query"`
		Photos       []struct {
			ID             string `json:"id"`
			Description    string `json:"description"`
			AltDescription string `json:"alt_description"`
			ImageURL       string `json:"image_url"`
			PreviewURL     string `json:"preview_url"`
			SourceURL      string `json:"source_url"`
			Photographer   string `json:"photographer"`
			Attribution    string `json:"attribution"`
			LocalPath      string `json:"local_path"`
			DownloadError  string `json:"download_error"`
		} `json:"photos"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, nil
	}
	previews := make([]map[string]any, 0, minInt(len(parsed.Photos), limit))
	urls := make([]string, 0, minInt(len(parsed.Photos), limit))
	seenURLs := map[string]struct{}{}
	for _, photo := range parsed.Photos {
		item := map[string]any{
			"id":            truncateString(strings.TrimSpace(photo.ID), 80),
			"provider":      truncateString(strings.TrimSpace(parsed.Provider), 40),
			"asset_purpose": truncateString(strings.TrimSpace(parsed.AssetPurpose), 40),
			"asset_query":   truncateString(strings.TrimSpace(parsed.AssetQuery), 180),
			"preview_url":   truncateString(strings.TrimSpace(firstRuntimeString(photo.PreviewURL, photo.ImageURL)), 240),
			"image_url":     truncateString(strings.TrimSpace(photo.ImageURL), 240),
			"source_url":    truncateString(strings.TrimSpace(photo.SourceURL), 240),
			"photographer":  truncateString(strings.TrimSpace(photo.Photographer), 120),
			"attribution":   truncateString(strings.TrimSpace(photo.Attribution), 180),
			"local_path":    truncateString(strings.TrimSpace(photo.LocalPath), 240),
		}
		if strings.TrimSpace(photo.Description) != "" || strings.TrimSpace(photo.AltDescription) != "" {
			item["description"] = truncateString(strings.TrimSpace(firstRuntimeString(photo.Description, photo.AltDescription)), 180)
		}
		if strings.TrimSpace(photo.DownloadError) != "" {
			item["download_error"] = truncateString(strings.TrimSpace(photo.DownloadError), 180)
		}
		previews = append(previews, item)
		if url := strings.TrimSpace(photo.SourceURL); url != "" {
			if _, ok := seenURLs[url]; !ok {
				seenURLs[url] = struct{}{}
				urls = append(urls, truncateString(url, 240))
			}
		}
		if len(previews) >= limit {
			break
		}
	}
	return previews, urls
}

func firstRuntimeString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func compactRuntimeMetadata(metadata map[string]any, depth int) map[string]any {
	if len(metadata) == 0 || depth > 2 {
		return nil
	}
	out := make(map[string]any, len(metadata))
	for key, value := range metadata {
		key = strings.TrimSpace(key)
		if key == "" || shouldDropRuntimeMetadataKey(key) {
			continue
		}
		if compacted, ok := compactRuntimeValue(key, value, depth); ok {
			out[key] = compacted
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func shouldDropRuntimeMetadataKey(key string) bool {
	switch strings.ToLower(key) {
	case "output", "config", "extra", "tools", "output_config", "output_extra", "reasoning_content":
		return true
	default:
		return false
	}
}

func compactRuntimeValue(key string, value any, depth int) (any, bool) {
	switch v := value.(type) {
	case nil:
		return nil, false
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			return "", false
		}
		limit := runtimeMetadataPreviewStringLimit
		if isRawRuntimeMetadataKey(key) {
			limit = runtimeMetadataRawStringLimit
		}
		return truncateString(redactRuntimeMetadataString(key, v), limit), true
	case int, int64, float64, bool:
		return v, true
	case []string:
		return boundedStrings(v, 12, 160), len(v) > 0
	case []map[string]any:
		limit := len(v)
		if limit > 12 {
			limit = 12
		}
		items := make([]any, 0, limit)
		for i := 0; i < limit; i++ {
			if item, ok := compactRuntimeValue(key, v[i], depth+1); ok {
				items = append(items, item)
			}
		}
		return items, len(items) > 0
	case []any:
		limit := len(v)
		if limit > 12 {
			limit = 12
		}
		items := make([]any, 0, limit)
		for i := 0; i < limit; i++ {
			if item, ok := compactRuntimeValue(key, v[i], depth+1); ok {
				items = append(items, item)
			}
		}
		return items, len(items) > 0
	case map[string]any:
		nested := compactRuntimeMetadata(v, depth+1)
		return nested, len(nested) > 0
	default:
		text := fmt.Sprint(v)
		if strings.TrimSpace(text) == "" {
			return nil, false
		}
		return truncateString(text, 240), true
	}
}

func isRawRuntimeMetadataKey(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "args", "result", "content", "arguments", "assistant_output":
		return true
	default:
		return false
	}
}

func isSensitiveRuntimeMetadataKey(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, marker := range []string{"api_key", "apikey", "token", "authorization", "cookie", "password", "passwd", "secret", "private_key"} {
		if strings.Contains(key, marker) {
			return true
		}
	}
	return false
}

func redactRuntimeMetadataString(key, value string) string {
	if isSensitiveRuntimeMetadataKey(key) {
		return "[REDACTED]"
	}
	redacted := value
	for _, pattern := range runtimeSecretPatterns {
		redacted = pattern.ReplaceAllString(redacted, "${1}[REDACTED]")
	}
	return redacted
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func boundedStrings(values []string, limit, maxLen int) []string {
	result := make([]string, 0, minInt(len(values), limit))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = truncateString(strings.TrimSpace(value), maxLen)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
		if len(result) >= limit {
			break
		}
	}
	return result
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
