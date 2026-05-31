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
	"context"
	"strings"

	"github.com/cloudwego/ppt-agent/pkg/agent/intent"
	"github.com/cloudwego/ppt-agent/pkg/logger"
)

// Engine 路由引擎
type Engine struct {
	classifier *intent.Classifier
	profileMatcher *ProfileMatcher
}

// NewEngine 创建路由引擎
func NewEngine(classifier *intent.Classifier) *Engine {
	return &Engine{
		classifier:     classifier,
		profileMatcher: NewProfileMatcher(),
	}
}

// Route 根据意图分类结果和用户上下文进行路由决策
func (e *Engine) Route(ctx context.Context, query string, userID int, profile interface{}) (*intent.RoutingDecision, error) {
	// Step 1: 获取意图分类结果
	classification, err := e.classifier.Classify(ctx, query, userID)
	if err != nil {
		logger.Warn("intent_classification_failed", "error", err.Error())
		// 使用默认决策
		classification = &intent.ClassificationResult{
			Intent:    intent.IntentCreate,
			Confidence: 0.5,
		}
	}

	// Step 2: 匹配用户画像（如果有）
	if profile != nil {
		e.profileMatcher.EnhanceWithProfile(classification, profile)
	}

	// Step 3: 根据意图和复杂度生成路由决策
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

	// === Agent 类型选择 ===
	switch {
	case classification.Intent == intent.IntentQuery:
		// 简单查询，使用简单模式
		decision.AgentType = "simple"
		decision.Pipeline = []string{"answer"}
		decision.SkipQA = true
		decision.SkipFix = true
		decision.Concurrency = 1

	case classification.Intent == intent.IntentCustomize:
		// 定制化调整
		decision.AgentType = "customize"
		decision.Pipeline = []string{"load_style", "apply_theme", "preview"}

	case classification.Intent == intent.IntentRegenerate:
		// 重新生成
		decision.AgentType = "regenerate"
		decision.Pipeline = []string{"load_page", "regenerate", "qa", "fix"}
		decision.Concurrency = 1

	case classification.Intent == intent.IntentContinue:
		// 继续任务
		decision.AgentType = "continue"
		decision.Pipeline = []string{"load_state", "resume"}

	case classification.Intent == intent.IntentEdit:
		// 编辑现有
		decision.AgentType = "edit"
		decision.Pipeline = []string{"load_existing", "identify_changes", "apply_changes", "qa"}

	case classification.Intent == intent.IntentExtend:
		// 扩展PPT
		decision.AgentType = "extend"
		decision.Pipeline = []string{"load_existing", "plan_new_pages", "generate", "qa"}

	case classification.Intent == intent.IntentCreate:
		// 新建PPT - 根据复杂度选择
		if classification.IsSimpleTask() {
			decision.AgentType = "quick"
			decision.Pipeline = []string{"quick_generate"}
			decision.SkipQA = classification.Complexity.Level <= 2
			decision.SkipFix = true
			decision.Concurrency = 5
		} else if classification.NeedsFullPipeline() {
			decision.AgentType = "deep"
			decision.Pipeline = []string{"plan", "generate", "qa", "fix"}
			decision.Concurrency = min(5, max(3, classification.Complexity.PageCountEstimate/3))
		} else {
			decision.AgentType = "standard"
			decision.Pipeline = []string{"plan", "generate", "qa"}
			decision.Concurrency = 4
		}

	default:
		// 未知意图，默认流程
		decision.AgentType = "deep"
		decision.Pipeline = []string{"plan", "generate", "qa", "fix"}
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
		// 使用用户历史偏好增强推荐
		if templates := ep.GetPreferredTemplates(); len(templates) > 0 {
			// 将用户偏好模板添加到推荐列表前面
			classification.SuggestedTemplates = append(templates, classification.SuggestedTemplates...)
			// 去重
			classification.SuggestedTemplates = deduplicateStringSlice(classification.SuggestedTemplates, 5)
		}

		// 使用用户的典型页数作为参考
		if typicalPages := ep.GetTypicalPageCount(); typicalPages > 0 {
			// 如果用户没有明确指定页数，使用用户偏好
			if classification.Complexity.PageCountEstimate == 0 {
				classification.SuggestedPageCount = typicalPages
			} else {
				// 取平均值
				classification.SuggestedPageCount = (classification.SuggestedPageCount + typicalPages) / 2
			}
		}

		// 使用用户的主题偏好
		if theme := ep.GetPreferredTheme(); theme != "" {
			// 如果没有明确的主题建议，使用用户偏好
			if classification.SuggestedTheme == "" {
				classification.SuggestedTheme = theme
			}
		}
	}
}

// EnhancedProfileProvider 用户画像接口
type EnhancedProfileProvider interface {
	GetPreferredTemplates() []string
	GetTypicalPageCount() int
	GetPreferredTheme() string
}

func deduplicateStringSlice(ss []string, maxLen int) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range ss {
		lower := strings.ToLower(s)
		if !seen[lower] {
			seen[lower] = true
			result = append(result, s)
			if len(result) >= maxLen {
				break
			}
		}
	}
	return result
}
