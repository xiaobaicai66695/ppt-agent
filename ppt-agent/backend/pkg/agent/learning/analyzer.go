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

package learning

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/style"
)

// PatternType 模式类型
type PatternType int

const (
	PatternTimeDistribution PatternType = iota // 时间分布模式
	PatternDomainPreference                    // 领域偏好模式
	PatternTemplateSuccess                     // 模板成功率模式
	PatternQualityTrend                        // 质量趋势模式
	PatternEditPattern                         // 编辑行为模式
)

// Pattern 识别到的模式
type Pattern struct {
	Type       PatternType            // 模式类型
	Confidence float64                // 置信度 0-1
	Data       map[string]interface{} // 模式数据
	Insight    string                 // 洞察描述
	Suggestion string                 // 建议
}

// Analyzer 模式分析器
type Analyzer struct {
	mu sync.RWMutex

	// 任务历史（内存缓存）
	taskHistory map[string]*TaskPattern // taskID -> pattern
	signals     []*LearningSignal       // 最近的学习信号

	// LLM 驱动的模式分析（轻量级模型，节省成本）
	modelFactory func(ctx context.Context) (interface {
		Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
	}, error)

	// 统计
	stats *AnalyzerStats
}

// AnalyzerStats 分析器统计
type AnalyzerStats struct {
	TotalSignals  int
	PatternsFound int
	LastAnalysis  time.Time
}

// TaskPattern 任务模式
type TaskPattern struct {
	TaskID       string
	UserID       int
	Domain       string
	Template     string
	Theme        string
	QualityScore float64
	Duration     time.Duration
	EditCount    int
	CompletedAt  time.Time
}

// NewAnalyzer 创建模式分析器。
// modelFactory 为 nil 时使用纯规则分析。
func NewAnalyzer(modelFactory func(ctx context.Context) (interface {
	Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
}, error)) *Analyzer {
	return &Analyzer{
		taskHistory:  make(map[string]*TaskPattern),
		signals:      make([]*LearningSignal, 0),
		modelFactory: modelFactory,
		stats: &AnalyzerStats{
			LastAnalysis: time.Now(),
		},
	}
}

// RecordSignal 记录学习信号
func (a *Analyzer) RecordSignal(signal *LearningSignal) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.signals = append(a.signals, signal)
	// 限制信号数量
	if len(a.signals) > 1000 {
		a.signals = a.signals[len(a.signals)-1000:]
	}

	a.stats.TotalSignals++
}

// RecordSuccess 记录成功任务
func (a *Analyzer) RecordSuccess(taskID string, profile *style.EnhancedProfile) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.taskHistory[taskID] = &TaskPattern{
		TaskID:       taskID,
		UserID:       profile.UserID,
		QualityScore: 4.5, // 默认高质量
		CompletedAt:  time.Now(),
	}
}

// AnalyzeUserPatterns 分析用户模式。
// 优先使用 LLM 分析（如果有 modelFactory），纯规则分析作为无模型时的降级。
func (a *Analyzer) AnalyzeUserPatterns(userID int) []*Pattern {
	a.mu.RLock()
	userSignals := a.collectUserSignals(userID)
	a.mu.RUnlock()

	if len(userSignals) == 0 {
		return nil
	}

	// LLM 深度分析
	if a.modelFactory != nil {
		patterns := a.llmAnalyzePatterns(userID, userSignals)
		if len(patterns) > 0 {
			a.mu.Lock()
			a.stats.PatternsFound += len(patterns)
			a.stats.LastAnalysis = time.Now()
			a.mu.Unlock()
			return patterns
		}
	}

	// 降级：纯规则分析
	return a.ruleBasedAnalysis(userSignals)
}

// collectUserSignals 收集某用户的所有信号（读锁外调用）
func (a *Analyzer) collectUserSignals(userID int) []*LearningSignal {
	var userSignals []*LearningSignal
	for _, s := range a.signals {
		if s.UserID == userID {
			userSignals = append(userSignals, s)
		}
	}
	return userSignals
}

// ruleBasedAnalysis 纯规则分析（无 LLM 时的降级路径）
func (a *Analyzer) ruleBasedAnalysis(userSignals []*LearningSignal) []*Pattern {
	var patterns []*Pattern

	// 分析时间模式
	if timePatterns := a.analyzeTimePatterns(userSignals); len(timePatterns) > 0 {
		patterns = append(patterns, timePatterns...)
	}

	// 分析质量趋势
	if qualityPatterns := a.analyzeQualityTrends(userSignals); len(qualityPatterns) > 0 {
		patterns = append(patterns, qualityPatterns...)
	}

	// 分析编辑行为
	if editPatterns := a.analyzeEditPatterns(userSignals); len(editPatterns) > 0 {
		patterns = append(patterns, editPatterns...)
	}

	a.mu.Lock()
	a.stats.PatternsFound += len(patterns)
	a.stats.LastAnalysis = time.Now()
	a.mu.Unlock()

	return patterns
}

// analyzeTimePatterns 分析时间分布模式
func (a *Analyzer) analyzeTimePatterns(signals []*LearningSignal) []*Pattern {
	var patterns []*Pattern

	// 统计各时段的信号数量
	hourlyCount := make(map[int]int)
	for _, s := range signals {
		if s.Timestamp.IsZero() {
			continue
		}
		hour := s.Timestamp.Hour()
		hourlyCount[hour]++
	}

	// 找出高峰时段
	var peakHours []int
	var maxCount int
	for hour, count := range hourlyCount {
		if count > maxCount {
			maxCount = count
			peakHours = []int{hour}
		} else if count == maxCount && count > 2 {
			peakHours = append(peakHours, hour)
		}
	}

	if len(peakHours) > 0 {
		pattern := &Pattern{
			Type:       PatternTimeDistribution,
			Confidence: float64(maxCount) / float64(len(signals)+1),
			Data: map[string]interface{}{
				"peak_hours":    peakHours,
				"total_signals": len(signals),
			},
			Insight:    "用户活跃时段分析",
			Suggestion: "可以在用户活跃时段提前准备资源，提高响应速度",
		}
		patterns = append(patterns, pattern)
	}

	return patterns
}

// analyzeQualityTrends 分析质量趋势
func (a *Analyzer) analyzeQualityTrends(signals []*LearningSignal) []*Pattern {
	var patterns []*Pattern

	var totalScore float64
	var scoreCount int
	var completionSignals []*LearningSignal

	for _, s := range signals {
		if s.Type == SignalCompletion && s.Context != nil {
			completionSignals = append(completionSignals, s)
			if s.Context.QualityScore > 0 {
				totalScore += s.Context.QualityScore
				scoreCount++
			}
		}
	}

	if scoreCount >= 3 {
		avgScore := totalScore / float64(scoreCount)

		// 检测质量趋势
		var trend string
		var confidence float64

		if len(completionSignals) >= 5 {
			// 至少5个数据点才能判断趋势
			recentSignals := completionSignals[len(completionSignals)-5:]
			var recentAvg, olderAvg float64

			for i, s := range recentSignals {
				if s.Context != nil && s.Context.QualityScore > 0 {
					if i < 3 {
						olderAvg += s.Context.QualityScore
					} else {
						recentAvg += s.Context.QualityScore
					}
				}
			}
			olderAvg /= 2
			recentAvg /= 2

			if recentAvg > olderAvg+0.3 {
				trend = "improving"
				confidence = 0.8
			} else if recentAvg < olderAvg-0.3 {
				trend = "declining"
				confidence = 0.8
			} else {
				trend = "stable"
				confidence = 0.9
			}
		}

		pattern := &Pattern{
			Type:       PatternQualityTrend,
			Confidence: confidence,
			Data: map[string]interface{}{
				"avg_score":   avgScore,
				"trend":       trend,
				"sample_size": scoreCount,
			},
			Insight:    "用户任务质量分析",
			Suggestion: a.getQualitySuggestion(trend, avgScore),
		}
		patterns = append(patterns, pattern)
	}

	return patterns
}

// analyzeEditPatterns 分析编辑行为模式
func (a *Analyzer) analyzeEditPatterns(signals []*LearningSignal) []*Pattern {
	var patterns []*Pattern

	editCount := 0
	for _, s := range signals {
		if s.Type == SignalEditAction {
			editCount++
		}
	}

	if len(signals) > 0 {
		editRatio := float64(editCount) / float64(len(signals))

		if editRatio > 0.3 {
			pattern := &Pattern{
				Type:       PatternEditPattern,
				Confidence: min(editRatio*2, 1.0),
				Data: map[string]interface{}{
					"edit_ratio":    editRatio,
					"total_signals": len(signals),
				},
				Insight:    "用户编辑行为较多",
				Suggestion: "建议生成更符合用户期望的初始内容，或提供更多定制选项",
			}
			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

// getQualitySuggestion 根据质量趋势给出建议
func (a *Analyzer) getQualitySuggestion(trend string, avgScore float64) string {
	if trend == "improving" {
		return "用户满意度持续提升，当前策略有效"
	}
	if trend == "declining" {
		return "用户满意度下降，建议优化生成策略或更新模板"
	}
	if avgScore < 3.5 {
		return "整体质量偏低，建议检查生成流程和模板质量"
	}
	if avgScore < 4.0 {
		return "质量中等，可以尝试更精细的模板匹配"
	}
	return "质量良好，继续保持当前策略"
}

// DetectPreferenceDrift 检测偏好漂移
func (a *Analyzer) DetectPreferenceDrift(userID int, currentProfile, historicalProfile *style.EnhancedProfile) bool {
	if historicalProfile == nil || currentProfile == nil {
		return false
	}

	// 检查主题偏好变化
	themeDrift := a.calculateSliceDrift(
		currentProfile.PreferredThemes,
		historicalProfile.PreferredThemes,
	)

	// 检查内容模式变化
	patternDrift := a.calculateSliceDrift(
		currentProfile.ContentPatterns,
		historicalProfile.ContentPatterns,
	)

	// 如果漂移超过阈值，认为发生了偏好漂移
	driftThreshold := 0.5
	return themeDrift > driftThreshold || patternDrift > driftThreshold
}

// calculateSliceDrift 计算列表漂移程度
func (a *Analyzer) calculateSliceDrift(current, historical []string) float64 {
	if len(historical) == 0 {
		return 0
	}

	currentSet := make(map[string]bool)
	for _, s := range current {
		currentSet[s] = true
	}

	matchCount := 0
	for _, s := range historical {
		if currentSet[s] {
			matchCount++
		}
	}

	return 1.0 - float64(matchCount)/float64(len(historical))
}

// GenerateInsights 生成洞察报告
func (a *Analyzer) GenerateInsights(userID int) *InsightsReport {
	patterns := a.AnalyzeUserPatterns(userID)

	report := &InsightsReport{
		UserID:          userID,
		GeneratedAt:     time.Now(),
		Patterns:        patterns,
		Summary:         "",
		Recommendations: []string{},
	}

	// 生成总结
	if len(patterns) == 0 {
		report.Summary = "数据不足，无法生成洞察"
	} else {
		report.Summary = "基于历史数据分析，发现以下模式"
	}

	// 生成建议
	for _, p := range patterns {
		if p.Suggestion != "" {
			report.Recommendations = append(report.Recommendations, p.Suggestion)
		}
	}

	return report
}

// InsightsReport 洞察报告
type InsightsReport struct {
	UserID          int
	GeneratedAt     time.Time
	Patterns        []*Pattern
	Summary         string
	Recommendations []string
}

// ── LLM 驱动的模式分析 ───────────────────────────────────────────────────

const patternAnalysisSystemPrompt = `你是一个用户行为模式分析引擎，负责从用户的 PPT 生成行为信号中提取有价值的洞察。

你会收到一个用户最近的的行为信号列表（JSON 数组），每条信号包含 type（信号类型）、timestamp、context 等字段。

信号类型说明：
- explicit_feedback: 用户对任务/页面的评分（context.quality_score 字段，1-5分）
- implicit_feedback: 隐式行为（context.action_type 如 "task_start"）
- edit_action: 用户编辑了PPT某页（context.page_index, context.action_type="edit"）
- qa_result: QA审阅结果（data.has_issue=true 表示发现质量问题）
- completion: 任务完成（context.quality_score, context.duration）
- abandon_task: 任务被放弃（data.reason, data.progress）

请分析这些信号，输出 JSON：

{
  "patterns": [
    {
      "type": "quality_trend" | "edit_pattern" | "time_pattern" | "content_preference" | "domain_preference" | "complexity_pattern",
      "confidence": 0.0-1.0,
      "insight": "一句话描述发现的模式",
      "suggestion": "针对该模式的优化建议，1-2句话",
      "data": {
        // 具体数据，如编辑次数、评分变化等
      }
    }
  ],
  "summary": "总分析总结，2-3句话"
}

## 规则
1. patterns 数组最多 5 条，按 confidence 从高到低排序
2. 只有 confidence >= 0.5 的模式才输出
3. suggestion 要具体可操作，不要空泛的建议
4. 如果信号太少（<3条），返回空的 patterns 数组
5. type 为 quality_trend 时，data 应包含 avg_score 和 trend（improving/stable/declining）
6. type 为 edit_pattern 时，data 应包含 edit_count 和 ratio`

// llmAnalyzePatterns 使用 LLM 分析用户信号，提取模式。
func (a *Analyzer) llmAnalyzePatterns(userID int, signals []*LearningSignal) []*Pattern {
	if len(signals) < 3 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	m, err := a.modelFactory(ctx)
	if err != nil {
		logger.Warn("analyzer_llm_model_create_failed", "error", err.Error())
		return nil
	}

	prompt := a.buildSignalAnalysisPrompt(signals)

	resp, err := m.Generate(ctx, []*schema.Message{
		{Role: schema.System, Content: patternAnalysisSystemPrompt},
		{Role: schema.User, Content: prompt},
	})
	if err != nil {
		logger.Warn("analyzer_llm_generate_failed", "error", err.Error())
		return nil
	}

	result, err := a.parsePatterns(resp.Content)
	if err != nil {
		logger.Warn("analyzer_llm_parse_failed", "content", truncateAnalyzedContent(resp.Content))
		return nil
	}

	return result
}

// buildSignalAnalysisPrompt 构建发送给 LLM 的分析 prompt。
func (a *Analyzer) buildSignalAnalysisPrompt(signals []*LearningSignal) string {
	var lines []string
	for _, s := range signals {
		entry := map[string]interface{}{
			"type":      s.Type.String(),
			"timestamp": s.Timestamp.Format("2006-01-02 15:04:05"),
		}
		if s.Context != nil {
			entry["context"] = map[string]interface{}{
				"phase":         s.Context.TaskPhase,
				"quality_score": s.Context.QualityScore,
				"duration_sec":  s.Context.Duration,
				"action_type":   s.Context.ActionType,
				"page_index":    s.Context.PageIndex,
			}
		}
		if s.Data != nil {
			entry["data"] = s.Data
		}
		b, _ := json.Marshal(entry)
		lines = append(lines, string(b))
	}

	return fmt.Sprintf("请分析以下用户行为信号（%d 条）：\n```json\n[%s\n]\n```", len(signals), strings.Join(lines, ",\n"))
}

// llmPatternResult LLM 返回的模式分析结果。
type llmPatternResult struct {
	Patterns []struct {
		Type       string                 `json:"type"`
		Confidence float64                `json:"confidence"`
		Insight    string                 `json:"insight"`
		Suggestion string                 `json:"suggestion"`
		Data       map[string]interface{} `json:"data"`
	} `json:"patterns"`
	Summary string `json:"summary"`
}

var patternTypeMap = map[string]PatternType{
	"quality_trend":      PatternQualityTrend,
	"edit_pattern":       PatternEditPattern,
	"time_pattern":       PatternTimeDistribution,
	"content_preference": PatternTemplateSuccess,
	"domain_preference":  PatternDomainPreference,
	"complexity_pattern": PatternQualityTrend,
}

func (a *Analyzer) parsePatterns(content string) ([]*Pattern, error) {
	content = strings.TrimSpace(content)
	if idx := strings.Index(content, "```"); idx >= 0 {
		start := idx + 3
		if rest := content[start:]; len(rest) >= 4 && rest[:4] == "json" {
			start += 4
		}
		end := strings.Index(content[start:], "```")
		if end >= 0 {
			content = content[start : start+end]
		}
	}

	var result llmPatternResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}

	var patterns []*Pattern
	for _, p := range result.Patterns {
		pt, ok := patternTypeMap[p.Type]
		if !ok {
			pt = PatternQualityTrend
		}
		patterns = append(patterns, &Pattern{
			Type:       pt,
			Confidence: p.Confidence,
			Data:       p.Data,
			Insight:    p.Insight,
			Suggestion: p.Suggestion,
		})
	}
	return patterns, nil
}

func truncateAnalyzedContent(s string) string {
	if len(s) <= 300 {
		return s
	}
	return s[:300] + "..."
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
