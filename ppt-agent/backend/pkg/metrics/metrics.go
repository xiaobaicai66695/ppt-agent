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

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// TasksTotal counts the total number of tasks created.
	TasksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ppt_agent",
			Subsystem: "task",
			Name:      "total",
			Help:     "Total number of PPT generation tasks created.",
		},
		[]string{"status"}, // running, completed, failed, cancelled
	)

	// TaskDuration tracks how long each task takes to complete.
	TaskDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ppt_agent",
			Subsystem: "task",
			Name:      "duration_seconds",
			Help:     "Duration of each PPT generation task in seconds.",
			Buckets:   []float64{10, 30, 60, 120, 300, 600, 1200, 1800, 3600},
		},
		[]string{"status"},
	)

	// TaskSlidesGenerated tracks slides generated per task.
	TaskSlidesGenerated = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ppt_agent",
			Subsystem: "task",
			Name:      "slides_generated",
			Help:     "Number of slides generated per task.",
			Buckets:   []float64{1, 5, 10, 15, 20, 30, 50},
		},
		[]string{"status"},
	)

	// QAIssuesTotal counts QA issues found by severity.
	QAIssuesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ppt_agent",
			Subsystem: "qa",
			Name:      "issues_total",
			Help:     "Total number of QA issues found.",
		},
		[]string{"severity"}, // high, medium, low
	)

	// QASlideScore tracks per-slide quality scores (1-5) from visual QA review.
	QASlideScore = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ppt_agent",
			Subsystem: "qa",
			Name:      "slide_score",
			Help:     "Per-slide visual quality score (1=unusable, 5=excellent).",
			Buckets:   []float64{1, 2, 3, 4, 5, 6}, // 6 catches any score>5
		},
		[]string{"content_type"}, // title_slide, content_slide, two_column, ...
	)

	// QAFixesTotal counts how many QA issues were successfully fixed.
	QAFixesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ppt_agent",
			Subsystem: "qa",
			Name:      "fixes_total",
			Help:     "Total number of QA issues fixed.",
		},
		[]string{"result"}, // success, failed
	)

	// LLMTokensTotal tracks cumulative token usage across all agents.
	LLMTokensTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ppt_agent",
			Subsystem: "llm",
			Name:      "tokens_total",
			Help:     "Total number of LLM tokens consumed.",
		},
		[]string{"type"}, // prompt, completion, total
	)

	// LLMCallsTotal tracks total LLM API calls.
	LLMCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ppt_agent",
			Subsystem: "llm",
			Name:      "calls_total",
			Help:     "Total number of LLM API calls.",
		},
		[]string{"status"}, // success, error, rate_limit
	)

	// AgentCallsTotal tracks calls per agent type.
	AgentCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ppt_agent",
			Subsystem: "agent",
			Name:      "calls_total",
			Help:     "Total number of agent invocations.",
		},
		[]string{"agent"}, // master, slide_executor, reviewer, fixer, planner, executor
	)

	// ToolCallsTotal tracks calls per tool type.
	ToolCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ppt_agent",
			Subsystem: "tool",
			Name:      "calls_total",
			Help:     "Total number of tool invocations.",
		},
		[]string{"tool", "status"}, // python3, search, read_file, edit_file, etc. + success/error
	)

	// ActiveTasks tracks currently running tasks.
	ActiveTasks = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "ppt_agent",
			Subsystem: "task",
			Name:      "active",
			Help:     "Number of currently running PPT generation tasks.",
		},
	)

	// HTTPRequestsTotal tracks HTTP request counts.
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ppt_agent",
			Subsystem: "http",
			Name:      "requests_total",
			Help:     "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status_code"},
	)

	// HTTPRequestDuration tracks HTTP request latency.
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ppt_agent",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:     "HTTP request latency in seconds.",
			Buckets:   []float64{0.005, 0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"method", "path"},
	)
)

// RecordTaskCreated increments the tasks total counter for the given status.
func RecordTaskCreated() {
	TasksTotal.WithLabelValues("running").Inc()
	ActiveTasks.Inc()
}

// RecordTaskCompleted records metrics when a task finishes.
func RecordTaskCompleted(durationSeconds float64, slidesGenerated, slidesTotal int, status string) {
	TasksTotal.WithLabelValues(status).Inc()
	ActiveTasks.Dec()
	TaskDuration.WithLabelValues(status).Observe(durationSeconds)
	if slidesGenerated > 0 {
		TaskSlidesGenerated.WithLabelValues(status).Observe(float64(slidesGenerated))
	}
}

// RecordTokens records token usage.
func RecordTokens(prompt, completion, total int64) {
	LLMTokensTotal.WithLabelValues("prompt").Add(float64(prompt))
	LLMTokensTotal.WithLabelValues("completion").Add(float64(completion))
	LLMTokensTotal.WithLabelValues("total").Add(float64(total))
}

// RecordLLMCall records an LLM call result.
func RecordLLMCall(status string) {
	LLMCallsTotal.WithLabelValues(status).Inc()
}

// RecordAgentCall records an agent invocation.
func RecordAgentCall(agent string) {
	AgentCallsTotal.WithLabelValues(agent).Inc()
}

// RecordToolCall records a tool invocation.
func RecordToolCall(tool, status string) {
	ToolCallsTotal.WithLabelValues(tool, status).Inc()
}

// RecordQAIssue records a QA issue found.
func RecordQAIssue(severity string) {
	QAIssuesTotal.WithLabelValues(severity).Inc()
}

// RecordQAFix records a QA fix attempt result.
func RecordQAFix(success bool) {
	if success {
		QAFixesTotal.WithLabelValues("success").Inc()
	} else {
		QAFixesTotal.WithLabelValues("failed").Inc()
	}
}

// RecordSlideScore records a per-slide quality score from QA review.
func RecordSlideScore(score float64, contentType string) {
	QASlideScore.WithLabelValues(contentType).Observe(score)
}
