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

package intent

import "strings"

// Intent 用户意图类型
type Intent int

const (
	IntentUnknown    Intent = iota
	IntentCreate            // 新建PPT
	IntentEdit              // 编辑现有PPT
	IntentExtend            // 扩展PPT（增加页数）
	IntentRegenerate        // 重新生成某页
	IntentQuery             // 询问问题
	IntentCustomize         // 定制化调整
	IntentContinue          // 继续未完成的任务
)

func (i Intent) String() string {
	switch i {
	case IntentCreate:
		return "create"
	case IntentEdit:
		return "edit"
	case IntentExtend:
		return "extend"
	case IntentRegenerate:
		return "regenerate"
	case IntentQuery:
		return "query"
	case IntentCustomize:
		return "customize"
	case IntentContinue:
		return "continue"
	default:
		return "unknown"
	}
}

func ParseIntent(s string) Intent {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "create":
		return IntentCreate
	case "edit":
		return IntentEdit
	case "extend":
		return IntentExtend
	case "regenerate":
		return IntentRegenerate
	case "query":
		return IntentQuery
	case "customize":
		return IntentCustomize
	case "continue":
		return IntentContinue
	default:
		return IntentUnknown
	}
}

// Domain 应用领域
type Domain string

const (
	DomainBusiness   Domain = "business"   // 商业/商务
	DomainAcademic   Domain = "academic"   // 学术/教育
	DomainPersonal   Domain = "personal"   // 个人/生活
	DomainCreative   Domain = "creative"   // 创意/艺术
	DomainGovernment Domain = "government" // 政务/党建
	DomainTechnical  Domain = "technical"  // 技术/工程
	DomainUnknown    Domain = "unknown"
)

func (d Domain) String() string {
	return string(d)
}

func ParseDomain(s string) Domain {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "business":
		return DomainBusiness
	case "academic", "education":
		return DomainAcademic
	case "personal":
		return DomainPersonal
	case "creative", "art":
		return DomainCreative
	case "government", "politics":
		return DomainGovernment
	case "technical", "tech":
		return DomainTechnical
	default:
		return DomainUnknown
	}
}

// Complexity 任务复杂度评估
type Complexity struct {
	Level             int    // 1-10 综合复杂度
	TopicComplexity   int    // 1-10 主题复杂度
	PageCountEstimate int    // 预估页数
	NeedsResearch     bool   // 是否需要搜索研究
	NeedsDataViz      bool   // 是否需要数据可视化
	NeedsMultiMedia   bool   // 是否需要多媒体元素
	AudienceLevel     string // 受众专业程度: beginner/intermediate/expert
}

// Urgency 紧迫度
type Urgency int

const (
	UrgencyNormal Urgency = iota
	UrgencyHigh
	UrgencyUrgent
)

func (u Urgency) String() string {
	switch u {
	case UrgencyNormal:
		return "normal"
	case UrgencyHigh:
		return "high"
	case UrgencyUrgent:
		return "urgent"
	default:
		return "normal"
	}
}

// ClassificationResult 意图分类结果
type ClassificationResult struct {
	Intent              Intent     // 识别到的意图
	IntentReasoning     string     // 意图判断的理由
	Complexity          Complexity // 复杂度评估
	Domain              Domain     // 应用领域
	Urgency             Urgency    // 紧迫度
	Confidence          float64    // 分类置信度 0-1
	SuggestedActions    []string   // 建议的下一步动作
	SuggestedTemplates  []string   // 建议的模板
	SuggestedTheme      string     // 建议的配色主题
	SuggestedBackground string     // 建议的图片检索线索（历史字段名）
	UseBackground       *bool      // 是否建议在适合页面规划外部图片
	SuggestedPageCount  int        // 建议的页数
	AgentType           string     // LLM 选择的可执行 Agent 类型
	Pipeline            []string   // LLM 建议的执行阶段
	Concurrency         int        // LLM 建议的页面生成并发数
	RoutingSource       string     // llm 或 fallback
}

// HasHighConfidence 判断是否有高置信度
func (r *ClassificationResult) HasHighConfidence() bool {
	return r.Confidence >= 0.7
}

// IsSimpleTask 判断是否为简单任务
func (r *ClassificationResult) IsSimpleTask() bool {
	return r.Complexity.Level <= 3 && r.Complexity.PageCountEstimate <= 5
}

// NeedsFullPipeline 判断是否需要完整流程
func (r *ClassificationResult) NeedsFullPipeline() bool {
	return r.Complexity.Level >= 5 || r.Complexity.PageCountEstimate > 10
}

// RoutingDecision 路由决策
type RoutingDecision struct {
	AgentType     string   // Agent 类型：planner
	Pipeline      []string // 执行的pipeline阶段
	Concurrency   int      // 并发数
	Source        string   // llm 或 fallback
	Reason        string   // 路由理由
	SkipQA        bool     // 是否跳过QA
	SkipFix       bool     // 是否跳过修复
	UseCustomFlow bool     // 是否使用自定义流程
	CacheProfile  bool     // 是否使用缓存的偏好
	Priority      int      // 优先级 1-10
	EstimatedCost int      // 预估消耗 tokens
	EstimatedTime int      // 预估时间（秒）
}

// NewDefaultRoutingDecision 返回默认路由决策
func NewDefaultRoutingDecision() *RoutingDecision {
	return &RoutingDecision{
		AgentType:    "planner",
		Pipeline:     []string{"plan", "generate"},
		Concurrency:  5,
		Source:       "fallback",
		SkipQA:       true,
		SkipFix:      true,
		CacheProfile: true,
		Priority:     5,
	}
}

// ActionRecommendation 动作推荐
type ActionRecommendation struct {
	Action     string                 // 动作类型
	Priority   int                    // 优先级
	Reason     string                 // 推荐理由
	Confidence float64                // 置信度
	Parameters map[string]interface{} // 动作参数
}
