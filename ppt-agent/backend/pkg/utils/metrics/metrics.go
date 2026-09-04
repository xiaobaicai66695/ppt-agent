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
	// TasksTotal 统计创建的任务总数
	TasksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ppt_agent",
			Subsystem: "task",
			Name:      "total",
			Help:      "Total number of PPT generation tasks created.",
		},
		[]string{"status"}, // running, completed, failed, cancelled
	)

	// TaskDuration 追踪每个任务的完成时间
	TaskDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ppt_agent",
			Subsystem: "task",
			Name:      "duration_seconds",
			Help:      "Duration of each PPT generation task in seconds.",
			Buckets:   []float64{10, 30, 60, 120, 300, 600, 1200, 1800, 3600},
		},
		[]string{"status"},
	)

	// TaskSlidesGenerated 追踪每个任务生成的幻灯片数量
	TaskSlidesGenerated = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ppt_agent",
			Subsystem: "task",
			Name:      "slides_generated",
			Help:      "Number of slides generated per task.",
			Buckets:   []float64{1, 5, 10, 15, 20, 30, 50},
		},
		[]string{"status"},
	)

	// QAIssuesTotal 按严重程度统计发现的 QA 问题数量
	QAIssuesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ppt_agent",
			Subsystem: "qa",
			Name:      "issues_total",
			Help:      "Total number of QA issues found.",
		},
		[]string{"severity"}, // high, medium, low
	)

	// QASlideScore 追踪每张幻灯片的质量评分（1-5分），来自视觉 QA 审查
	QASlideScore = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ppt_agent",
			Subsystem: "qa",
			Name:      "slide_score",
			Help:      "Per-slide visual quality score (1=unusable, 5=excellent).",
			Buckets:   []float64{1, 2, 3, 4, 5, 6}, // 6 catches any score>5
		},
		[]string{"content_type"}, // title_slide, content_slide, two_column, ...
	)

	// QAFixesTotal 统计成功修复的 QA 问题数量
	QAFixesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ppt_agent",
			Subsystem: "qa",
			Name:      "fixes_total",
			Help:      "Total number of QA issues fixed.",
		},
		[]string{"result"}, // success, failed
	)

	// LLMTokensTotal 追踪所有代理的累计 token 使用量
	LLMTokensTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ppt_agent",
			Subsystem: "llm",
			Name:      "tokens_total",
			Help:      "Total number of LLM tokens consumed.",
		},
		[]string{"type"}, // prompt, completion, total
	)

	// LLMCallsTotal 追踪 LLM API 调用总数
	LLMCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ppt_agent",
			Subsystem: "llm",
			Name:      "calls_total",
			Help:      "Total number of LLM API calls.",
		},
		[]string{"status"}, // success, error, rate_limit
	)

	// AgentCallsTotal 追踪每种代理类型的调用次数
	AgentCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ppt_agent",
			Subsystem: "agent",
			Name:      "calls_total",
			Help:      "Total number of agent invocations.",
		},
		[]string{"agent"}, // planner, renderer, workflow
	)

	// ToolCallsTotal 追踪每种工具类型的调用次数
	ToolCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ppt_agent",
			Subsystem: "tool",
			Name:      "calls_total",
			Help:      "Total number of tool invocations.",
		},
		[]string{"tool", "status"}, // read_file, search, manifest tools
	)

	// ActiveTasks 追踪当前正在运行的任务数量
	ActiveTasks = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "ppt_agent",
			Subsystem: "task",
			Name:      "active",
			Help:      "Number of currently running PPT generation tasks.",
		},
	)

	// HTTPRequestsTotal 追踪 HTTP 请求总数
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ppt_agent",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status_code"},
	)

	// HTTPRequestDuration 追踪 HTTP 请求延迟
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ppt_agent",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request latency in seconds.",
			Buckets:   []float64{0.005, 0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"method", "path"},
	)
)

// RecordTaskCreated 增加创建任务计数器（指定状态）
func RecordTaskCreated() {
	TasksTotal.WithLabelValues("running").Inc()
	ActiveTasks.Inc()
}

// RecordTaskCompleted 任务完成时记录指标
func RecordTaskCompleted(durationSeconds float64, slidesGenerated, slidesTotal int, status string) {
	TasksTotal.WithLabelValues(status).Inc()
	ActiveTasks.Dec()
	TaskDuration.WithLabelValues(status).Observe(durationSeconds)
	if slidesGenerated > 0 {
		TaskSlidesGenerated.WithLabelValues(status).Observe(float64(slidesGenerated))
	}
}

// RecordTokens 记录 token 使用量
func RecordTokens(prompt, completion, total int64) {
	LLMTokensTotal.WithLabelValues("prompt").Add(float64(prompt))
	LLMTokensTotal.WithLabelValues("completion").Add(float64(completion))
	LLMTokensTotal.WithLabelValues("total").Add(float64(total))
}

// RecordLLMCall 记录一次 LLM 调用结果
func RecordLLMCall(status string) {
	LLMCallsTotal.WithLabelValues(status).Inc()
}

// RecordAgentCall 记录一次代理调用
func RecordAgentCall(agent string) {
	AgentCallsTotal.WithLabelValues(agent).Inc()
}

// RecordToolCall 记录一次工具调用
func RecordToolCall(tool, status string) {
	ToolCallsTotal.WithLabelValues(tool, status).Inc()
}

// RecordQAIssue 记录一个发现的 QA 问题
func RecordQAIssue(severity string) {
	QAIssuesTotal.WithLabelValues(severity).Inc()
}

// RecordQAFix 记录一次 QA 修复尝试结果
func RecordQAFix(success bool) {
	if success {
		QAFixesTotal.WithLabelValues("success").Inc()
	} else {
		QAFixesTotal.WithLabelValues("failed").Inc()
	}
}

// RecordSlideScore 记录 QA 审查后的每张幻灯片质量评分
func RecordSlideScore(score float64, contentType string) {
	QASlideScore.WithLabelValues(contentType).Observe(score)
}
