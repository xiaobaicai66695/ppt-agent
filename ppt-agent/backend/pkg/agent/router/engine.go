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

package router

import (
	"strings"

	"github.com/cloudwego/ppt-agent/pkg/agent/intent"
	"github.com/cloudwego/ppt-agent/pkg/logger"
)

// Engine 路由引擎
type Engine struct {
	profileMatcher *ProfileMatcher
}

// NewEngine 创建路由引擎
func NewEngine(classifier *intent.Classifier) *Engine {
	_ = classifier
	return &Engine{
		profileMatcher: NewProfileMatcher(),
	}
}

// Route validates the single model classification and turns it into an
// executable initial-generation decision. It never classifies the query again.
func (e *Engine) Route(classification *intent.ClassificationResult, profile interface{}) (*intent.RoutingDecision, error) {
	if classification == nil {
		classification = &intent.ClassificationResult{
			Intent: intent.IntentCreate, AgentType: "planner", Pipeline: []string{"plan", "generate"},
			Concurrency: 5, RoutingSource: "fallback",
		}
	}

	if profile != nil {
		e.profileMatcher.EnhanceWithProfile(classification, profile)
	}

	decision := e.makeRoutingDecision(classification)

	logger.Info("routing_decision_made",
		"intent", classification.Intent.String(),
		"domain", classification.Domain.String(),
		"complexity", classification.Complexity.Level,
		"agent_type", decision.AgentType,
		"pipeline", strings.Join(decision.Pipeline, ","),
		"concurrency", decision.Concurrency)

	return decision, nil
}

// makeRoutingDecision 根据分类结果生成路由决策
func (e *Engine) makeRoutingDecision(classification *intent.ClassificationResult) *intent.RoutingDecision {
	decision := intent.NewDefaultRoutingDecision()
	decision.Source = classification.RoutingSource
	if decision.Source == "" {
		decision.Source = "fallback"
	}
	decision.Reason = classification.IntentReasoning

	// The executable path is Planner + renderer workflow.
	if strings.EqualFold(strings.TrimSpace(classification.AgentType), "planner") {
		decision.AgentType = "planner"
	} else {
		logger.Warn("routing_agent_normalized", "requested", classification.AgentType, "selected", "planner")
		decision.AgentType = "planner"
	}
	decision.Pipeline = []string{"plan", "generate"}
	decision.SkipQA = true
	decision.SkipFix = true
	if classification.Concurrency > 0 {
		decision.Concurrency = classification.Concurrency
		if decision.Concurrency > 10 {
			decision.Concurrency = 10
		}
	}

	// === 优先级设置 ===
	decision.Priority = e.calculatePriority(classification)

	// === 预估消耗 ===
	decision.EstimatedTime = e.estimateTime(classification)
	decision.EstimatedCost = e.estimateCost(classification)

	return decision
}

// calculatePriority 计算任务优先级
func (e *Engine) calculatePriority(classification *intent.ClassificationResult) int {
	priority := 5 // 默认优先级

	// 紧迫度影响
	switch classification.Urgency {
	case intent.UrgencyUrgent:
		priority += 3
	case intent.UrgencyHigh:
		priority += 1
	}

	// 复杂度影响（高复杂度需要更高优先级以获得更多资源）
	if classification.Complexity.Level >= 7 {
		priority += 2
	} else if classification.Complexity.Level >= 5 {
		priority += 1
	}

	// 意图影响
	switch classification.Intent {
	case intent.IntentRegenerate:
		priority += 1 // 重新生成通常更急迫
	case intent.IntentQuery:
		priority += 2 // 查询通常需要快速响应
	}

	// 限制范围
	if priority > 10 {
		priority = 10
	}
	if priority < 1 {
		priority = 1
	}

	return priority
}

// estimateTime 估算执行时间（秒）
func (e *Engine) estimateTime(classification *intent.ClassificationResult) int {
	baseTime := 60 // 基础时间60秒

	// 根据页数估算
	baseTime += classification.Complexity.PageCountEstimate * 10

	// 根据复杂度调整
	if classification.Complexity.Level >= 7 {
		baseTime *= 2
	} else if classification.Complexity.Level >= 5 {
		baseTime += 60
	}

	// 根据 pipeline 长度调整
	pipelineLen := len(classification.SuggestedActions)
	if pipelineLen > 3 {
		baseTime += (pipelineLen - 3) * 30
	}

	return baseTime
}

// estimateCost 估算 token 消耗
func (e *Engine) estimateCost(classification *intent.ClassificationResult) int {
	baseCost := 5000 // 基础消耗 5000 tokens

	// 根据页数估算
	baseCost += classification.Complexity.PageCountEstimate * 800

	// 根据复杂度调整
	if classification.Complexity.Level >= 7 {
		baseCost += 5000
	} else if classification.Complexity.Level >= 5 {
		baseCost += 2000
	}

	// 数据可视化需求增加消耗
	if classification.Complexity.NeedsDataViz {
		baseCost += 3000
	}

	// 研究需求增加消耗
	if classification.Complexity.NeedsResearch {
		baseCost += 4000
	}

	return baseCost
}

// ProfileMatcher 用户画像匹配器
type ProfileMatcher struct{}

func NewProfileMatcher() *ProfileMatcher {
	return &ProfileMatcher{}
}

// EnhanceWithProfile 根据用户画像增强分类结果
func (m *ProfileMatcher) EnhanceWithProfile(classification *intent.ClassificationResult, profile interface{}) {
	if classification == nil {
		return
	}

	// 尝试转换为 EnhancedProfile
	if ep, ok := profile.(EnhancedProfileProvider); ok {
		domain := classification.Domain.String()

		// 典型页数只在当前任务没有页数估算时补位；显式/模型识别出的页数优先。
		if typicalPages := ep.GetTypicalPageCount(); typicalPages > 0 {
			if classification.Complexity.PageCountEstimate == 0 && classification.SuggestedPageCount == 0 {
				classification.SuggestedPageCount = typicalPages
			}
		}

		// 配色/主题属于高场景敏感字段，只在同领域历史中补位。
		if theme := preferredThemeForDomain(ep, domain); theme != "" {
			if classification.SuggestedTheme == "" {
				classification.SuggestedTheme = theme
			}
		}
	}
}

// EnhancedProfileProvider 用户画像接口
type EnhancedProfileProvider interface {
	GetTypicalPageCount() int
	GetPreferredTheme() string
}

type domainAwareProfileProvider interface {
	GetPreferredThemeForDomain(domain string) string
}

func preferredThemeForDomain(profile EnhancedProfileProvider, domain string) string {
	if p, ok := profile.(domainAwareProfileProvider); ok {
		return p.GetPreferredThemeForDomain(domain)
	}
	return ""
}
