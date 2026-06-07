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

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/logger"
)

// Classifier 意图分类器
type Classifier struct {
	modelFactory     func(ctx context.Context) (model.ToolCallingChatModel, error)
	textModelFactory func(ctx context.Context) (interface {
		Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
	}, error)
	useLLM bool
}

// NewClassifier 创建意图分类器
// modelFactory 用于创建 LLM 实例（可选，用于更精确的分类）
// textModelFactory 用于创建轻量级 LLM 实例（优先使用，节省成本）
func NewClassifier(
	modelFactory func(ctx context.Context) (model.ToolCallingChatModel, error),
	textModelFactory func(ctx context.Context) (interface {
		Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
	}, error),
) *Classifier {
	return &Classifier{
		modelFactory:     modelFactory,
		textModelFactory: textModelFactory,
		useLLM:          modelFactory != nil || textModelFactory != nil,
	}
}

// SetTextModelFactory 设置轻量级模型工厂
func (c *Classifier) SetTextModelFactory(
	factory func(ctx context.Context) (interface {
		Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
	}, error),
) {
	c.textModelFactory = factory
	if factory != nil {
		c.useLLM = true
	}
}

// Classify 对用户输入进行意图分类
// 优先使用规则匹配，fallback 到 LLM 分类
func (c *Classifier) Classify(ctx context.Context, query string, userID int) (*ClassificationResult, error) {
	// Step 1: 规则匹配（快速路径）
	result := c.ruleBasedClassification(query)

	// Step 2: 如果规则匹配置信度足够高，直接返回
	if result.Confidence >= 0.85 {
		logger.Debug("intent_classified_by_rules",
			"intent", result.Intent.String(),
			"confidence", result.Confidence,
			"query_len", len(query))
		return result, nil
	}

	// Step 3: LLM 增强分类（可选）
	if c.useLLM {
		// 优先使用轻量级 textModel，节省成本
		if c.textModelFactory != nil {
			llmResult, err := c.llmClassificationByTextModel(ctx, query)
			if err != nil {
				logger.Warn("intent_llm_text_classification_failed", "error", err.Error())
			} else {
				if llmResult.Confidence > result.Confidence {
					logger.Debug("intent_classified_by_text_model",
						"intent", llmResult.Intent.String(),
						"confidence", llmResult.Confidence)
					return llmResult, nil
				}
			}
		}
		// 回退到 ToolCallingChatModel
		if c.modelFactory != nil {
			llmResult, err := c.llmClassification(ctx, query)
			if err != nil {
				logger.Warn("intent_llm_classification_failed", "error", err.Error())
			} else {
				if llmResult.Confidence > result.Confidence {
					logger.Debug("intent_classified_by_llm",
						"intent", llmResult.Intent.String(),
						"confidence", llmResult.Confidence)
					return llmResult, nil
				}
			}
		}
	}

	return result, nil
}

// ruleBasedClassification 基于规则的意图分类。
// 规则只做它擅长的事——识别 4 类核心意图的关键词（create/edit/extend/regenerate）。
// query、customize、缺少关键词的 create 全部交给 LLM。
func (c *Classifier) ruleBasedClassification(query string) *ClassificationResult {
	query = strings.TrimSpace(query)
	queryLower := strings.ToLower(query)

	result := &ClassificationResult{
		Intent:       IntentUnknown,
		Confidence:   0.0,
		Domain:       DomainUnknown,
		Urgency:      UrgencyNormal,
		Complexity:   Complexity{Level: 5, TopicComplexity: 5, AudienceLevel: "intermediate"},
	}

	// === 仅 4 类核心意图 + continue，其余交给 LLM ===

	// 继续任务（简单可靠，保留）
	if matched, score := c.matchIntent(queryLower, []string{"继续", "继续做", "接着做", "resume", "continue"}); matched {
		result.Intent = IntentContinue
		result.Confidence = score
		result.IntentReasoning = "检测到继续任务关键字"
		c.fillMeta(result, queryLower)
		return result
	}

	// 1. 重新生成
	if matched, score := c.matchIntent(queryLower, []string{
		"重新生成", "重新做", "重做", "再来一次", "regenerate", "redo",
	}); matched {
		result.Intent = IntentRegenerate
		result.Confidence = score
		result.IntentReasoning = "检测到重新生成关键字"
		c.fillMeta(result, queryLower)
		return result
	}

	// 2. 编辑现有
	if matched, score := c.matchIntent(queryLower, []string{
		"修改", "编辑", "更改", "改成", "换成", "调整", "edit", "modify", "改一下",
	}); matched {
		result.Intent = IntentEdit
		result.Confidence = score
		result.IntentReasoning = "检测到编辑关键字"
		c.fillMeta(result, queryLower)
		return result
	}

	// 3. 扩展PPT（注意：不含 "add"，太宽泛容易误命中）
	if matched, score := c.matchIntent(queryLower, []string{
		"再加", "扩展", "增加", "加几页", "补充", "添加", "多几页", "extend",
	}); matched {
		result.Intent = IntentExtend
		result.Confidence = score
		result.IntentReasoning = "检测到扩展关键字"
		c.fillMeta(result, queryLower)
		return result
	}

	// 4. 新建PPT
	if matched, score := c.matchIntent(queryLower, []string{
		"做一个", "做个", "制作", "创建", "做一个关于", "帮我做", "写一个",
		"create", "make a", "generate a",
	}); matched {
		result.Intent = IntentCreate
		result.Confidence = score
		result.IntentReasoning = "检测到创建关键字"
		c.fillMeta(result, queryLower)
		return result
	}

	// 无任何关键词命中 → IntentUnknown + confidence 0.0 → 必走 LLM
	return result
}

// fillMeta 填充领域、复杂度、紧迫度等辅助信息（仅关键词命中后调用）。
func (c *Classifier) fillMeta(result *ClassificationResult, query string) {
	result.Domain = c.classifyDomain(query)
	result.Complexity = c.assessComplexity(query)
	result.Urgency = c.assessUrgency(query)
	c.enrichRecommendations(result, query)
}

// matchIntent 匹配意图关键字，返回(是否匹配, 置信度)。
// 置信度基于信号质量而非关键词排位：
//   - 特异性 (0~0.4)：关键词越长越不容易误匹配
//   - 位置 (0~0.3)：关键词在句首更可能是意图信号
//   - 独占性 (0~0.3)：整个 query 越接近关键词本身，信号越纯
func (c *Classifier) matchIntent(query string, keywords []string) (bool, float64) {
	for _, kw := range keywords {
		idx := strings.Index(query, kw)
		if idx < 0 {
			continue
		}

		// 特异性：按关键词长度 / 常见最大长度 (6) 归一化，越长的词越不容易误命中
		specificity := float64(len([]rune(kw))) / 6.0
		if specificity > 1.0 {
			specificity = 1.0
		}
		specificityScore := 0.4 * specificity

		// 位置：关键词出现在前 1/3 处 = 句首意图信号强；越靠后越弱
		posRatio := float64(idx) / float64(len([]rune(query))+1)
		positionScore := 0.3 * (1.0 - posRatio)
		if positionScore < 0 {
			positionScore = 0
		}

		// 独占性：关键词长度 / query 长度越接近 1，整句几乎就是关键词本身
		exclusivity := float64(len([]rune(kw))) / float64(len([]rune(query))+1)
		exclusivityScore := 0.3 * exclusivity
		if exclusivityScore > 0.3 {
			exclusivityScore = 0.3
		}

		confidence := 0.4 + specificityScore + positionScore + exclusivityScore
		if confidence > 0.98 {
			confidence = 0.98
		}
		return true, confidence
	}
	return false, 0
}

// classifyDomain 分类应用领域
func (c *Classifier) classifyDomain(query string) Domain {
	domainRules := map[Domain][]string{
		DomainBusiness: {
			"商业", "商务", "路演", "融资", "创业", "商业计划", "市场分析", "营销", "销售",
			"pitch", "business", "proposal", "客户", "投标",
		},
		DomainTechnical: {
			"技术", "架构", "代码", "开发", "系统", "技术分享", "培训", "工程师",
			"technical", "tech", "architecture", "coding",
		},
		DomainAcademic: {
			"学术", "研究", "论文", "答辩", "课程", "教学", "课件", "培训",
			"academic", "research", "course", "thesis", "presentation",
		},
		DomainGovernment: {
			"政务", "党建", "政府", "思政", "团课", "红色", "政策", "汇报",
			"government", "political", "party",
		},
		DomainPersonal: {
			"个人", "简历", "述职", "总结", "生活", "分享", "个人介绍",
			"resume", "summary", "personal", "bio",
		},
		DomainCreative: {
			"创意", "设计", "艺术", "活动", "策划", "创意", "品牌",
			"creative", "design", "art", "campaign",
		},
	}

	maxCount := 0
	var detectedDomain Domain

	for domain, keywords := range domainRules {
		count := 0
		for _, kw := range keywords {
			if strings.Contains(query, kw) {
				count++
			}
		}
		if count > maxCount {
			maxCount = count
			detectedDomain = domain
		}
	}

	if maxCount > 0 {
		return detectedDomain
	}
	return DomainUnknown
}

// assessComplexity 评估复杂度
func (c *Classifier) assessComplexity(query string) Complexity {
	cpl := Complexity{
		Level:            5,
		TopicComplexity:  5,
		PageCountEstimate: 10,
		AudienceLevel:    "intermediate",
	}

	// 页数估计
	pagePatterns := []struct {
		pattern *regexp.Regexp
		add     int
	}{
		{regexp.MustCompile(`(\d+)\s*页`), 0},
		{regexp.MustCompile(`(\d+)\s*张`), 0},
		{regexp.MustCompile(`(\d+)\s*slides?`), 0},
	}

	for _, p := range pagePatterns {
		if matches := p.pattern.FindStringSubmatch(query); len(matches) > 1 {
			if n, err := strconv.Atoi(matches[1]); err == nil && n > 0 && n <= 100 {
				cpl.PageCountEstimate = n
			}
		}
	}

	// 主题复杂度
	complexKeywords := []string{
		"详细", "深入", "全面", "系统", "完整", "复杂", "全面分析",
		"详细", "comprehensive", "detailed", "complex",
	}
	simpleKeywords := []string{
		"简单", "简洁", "简短", "概要", "快速", "简单介绍",
		"simple", "brief", "quick", "overview",
	}

	complexCount := 0
	simpleCount := 0
	for _, kw := range complexKeywords {
		if strings.Contains(query, kw) {
			complexCount++
		}
	}
	for _, kw := range simpleKeywords {
		if strings.Contains(query, kw) {
			simpleCount++
		}
	}

	if complexCount > simpleCount {
		cpl.Level += 2
		cpl.TopicComplexity += 2
	} else if simpleCount > complexCount {
		cpl.Level -= 2
		cpl.TopicComplexity -= 2
	}

	// 限制范围
	if cpl.Level < 1 {
		cpl.Level = 1
	}
	if cpl.Level > 10 {
		cpl.Level = 10
	}
	if cpl.TopicComplexity < 1 {
		cpl.TopicComplexity = 1
	}
	if cpl.TopicComplexity > 10 {
		cpl.TopicComplexity = 10
	}

	// 数据可视化需求
	if strings.Contains(query, "数据") || strings.Contains(query, "图表") ||
		strings.Contains(query, "统计") || strings.Contains(query, "分析") {
		cpl.NeedsDataViz = true
		cpl.Level++
	}

	// 研究需求
	if strings.Contains(query, "研究") || strings.Contains(query, "调研") ||
		strings.Contains(query, "搜索") || strings.Contains(query, "最新") {
		cpl.NeedsResearch = true
		cpl.Level++
	}

	// 多媒体需求
	if strings.Contains(query, "图片") || strings.Contains(query, "视频") ||
		strings.Contains(query, "图文") {
		cpl.NeedsMultiMedia = true
	}

	// 受众评估
	if strings.Contains(query, "专业") || strings.Contains(query, "专家") ||
		strings.Contains(query, "技术") {
		cpl.AudienceLevel = "expert"
	} else if strings.Contains(query, "小白") || strings.Contains(query, "入门") ||
		strings.Contains(query, "新手") {
		cpl.AudienceLevel = "beginner"
	}

	return cpl
}

// assessUrgency 评估紧迫度
func (c *Classifier) assessUrgency(query string) Urgency {
	urgentKeywords := []string{"紧急", "马上", "立刻", "尽快", "今天", "马上要", "急"}
	highKeywords := []string{"明天", "这周", "尽快", "赶", "deadline"}

	for _, kw := range urgentKeywords {
		if strings.Contains(query, kw) {
			return UrgencyUrgent
		}
	}
	for _, kw := range highKeywords {
		if strings.Contains(query, kw) {
			return UrgencyHigh
		}
	}
	return UrgencyNormal
}

// enrichRecommendations 丰富推荐信息
func (c *Classifier) enrichRecommendations(result *ClassificationResult, query string) {
	// 模板推荐
	result.SuggestedTemplates = c.suggestTemplates(result.Domain, result.Complexity)

	// 主题推荐
	result.SuggestedTheme = c.suggestTheme(result.Domain, query)

	// 页数推荐
	if result.Complexity.PageCountEstimate == 0 {
		result.SuggestedPageCount = c.suggestPageCount(result.Domain, result.Complexity)
	} else {
		result.SuggestedPageCount = result.Complexity.PageCountEstimate
	}

	// 动作推荐
	result.SuggestedActions = c.suggestActions(result.Intent, result.Complexity)
}

// suggestTemplates 推荐模板
func (c *Classifier) suggestTemplates(domain Domain, complexity Complexity) []string {
	templateMap := map[Domain][]string{
		DomainBusiness:   {"pitch-deck", "product-launch", "research-report"},
		DomainTechnical:  {"tech-sharing", "tech-intro", "course-module"},
		DomainAcademic:   {"course-module", "tech-intro", "design-defense"},
		DomainGovernment: {"politics-ideology", "current-affairs", "personal-summary"},
		DomainPersonal:   {"personal-summary", "weekly-report", "short-class-talk"},
		DomainCreative:   {"activity-plan", "product-launch", "course-module"},
		DomainUnknown:    {"tech-intro", "weekly-report", "short-class-talk"},
	}

	templates, ok := templateMap[domain]
	if !ok {
		return templateMap[DomainUnknown]
	}

	// 限制返回数量
	if len(templates) > 3 {
		templates = templates[:3]
	}
	return templates
}

// suggestTheme 推荐主题
func (c *Classifier) suggestTheme(domain Domain, query string) string {
	themeMap := map[Domain]string{
		DomainBusiness:   "charcoal_light",
		DomainTechnical:  "ocean_soft",
		DomainAcademic:   "sage_calm",
		DomainGovernment: "government_red",
		DomainPersonal:   "simple_gray",
		DomainCreative:   "berry_cream",
	}

	if theme, ok := themeMap[domain]; ok {
		return theme
	}
	return "simple_gray"
}

// suggestPageCount 推荐页数
func (c *Classifier) suggestPageCount(domain Domain, complexity Complexity) int {
	pageCountMap := map[Domain]int{
		DomainBusiness:   14,
		DomainTechnical:  18,
		DomainAcademic:   12,
		DomainGovernment: 16,
		DomainPersonal:   10,
		DomainCreative:   10,
		DomainUnknown:    12,
	}

	base := 12
	if count, ok := pageCountMap[domain]; ok {
		base = count
	}

	// 根据复杂度调整
	if complexity.Level > 7 {
		base += 4
	} else if complexity.Level < 4 {
		base -= 4
	}

	// 限制范围
	if base < 6 {
		base = 6
	}
	if base > 30 {
		base = 30
	}

	return base
}

// suggestActions 推荐动作
func (c *Classifier) suggestActions(intent Intent, complexity Complexity) []string {
	switch intent {
	case IntentCreate:
		if complexity.Level <= 3 {
			return []string{"quick_generate", "skip_qa"}
		}
		return []string{"plan", "generate", "qa"}
	case IntentEdit:
		return []string{"load_existing", "identify_changes", "apply_changes"}
	case IntentExtend:
		return []string{"load_existing", "plan_new_pages", "generate"}
	case IntentRegenerate:
		return []string{"load_page", "regenerate", "qa"}
	case IntentCustomize:
		return []string{"load_style", "apply_theme", "preview"}
	case IntentQuery:
		return []string{"answer_question"}
	default:
		return []string{"plan", "generate"}
	}
}

// llmClassification 使用 LLM 进行意图分类
func (c *Classifier) llmClassification(ctx context.Context, query string) (*ClassificationResult, error) {
	if c.modelFactory == nil {
		return nil, nil
	}

	m, err := c.modelFactory(ctx)
	if err != nil {
		return nil, err
	}

	systemPrompt := `你是一个PPT任务意图分类器。根据用户输入，分类其意图并评估任务复杂度。

意图类型：
- create: 新建PPT
- edit: 编辑现有PPT
- extend: 扩展PPT（增加页数）
- regenerate: 重新生成某页
- customize: 定制化调整
- query: 询问问题
- continue: 继续未完成任务

领域类型（根据输入中的关键词判断，无明确领域关键词时填 unknown）：
- business: 商业/商务/融资/路演/创业/市场/营销/客户/投标/产品发布
- technical: 技术/架构/代码/开发/系统/工程师/微服务/API/AI
- academic: 学术/论文/答辩/课程/教学/研究/培训/考试
- government: 政务/党建/政府/思政/团课/红色/政策/汇报/党风廉政
- personal: 个人/简历/述职/总结/自我介绍/求职/年终
- creative: 创意/设计/艺术/活动/策划/品牌/文化/旅游
- unknown: 无法确定领域（领域关键词不明确时使用）

复杂度评估规则（1-10分）：
- 1-3: 简单介绍/概览/基础说明/少于5页
- 4-6: 中等深度/包含数据和图表/5-15页
- 7-10: 深度分析/包含架构图和多章节/15页以上/要求详细和全面
- 输入包含"简单""简洁""快速""brief""simple"时降低1-2分
- 输入包含"详细""深入""全面""comprehensive""detailed"时提高1-2分
- 输入包含"数据""图表""architecture""代码"时提高1分

请返回JSON格式：
{
  "intent": "意图类型",
  "intent_reasoning": "判断理由",
  "domain": "领域类型",
  "complexity_level": 1-10,
  "page_count_estimate": 预估页数,
  "confidence": 置信度0-1,
  "suggested_theme": "推荐配色主题",
  "suggested_templates": ["推荐模板列表"]
}`

	resp, err := m.Generate(ctx, []*schema.Message{
		{Role: schema.System, Content: systemPrompt},
		{Role: schema.User, Content: "用户输入: " + query},
	})
	if err != nil {
		return nil, err
	}

	// 解析 JSON 响应
	var llmResult struct {
		Intent            string   `json:"intent"`
		IntentReasoning   string   `json:"intent_reasoning"`
		Domain            string   `json:"domain"`
		ComplexityLevel   int      `json:"complexity_level"`
		PageCountEstimate int      `json:"page_count_estimate"`
		Confidence        float64  `json:"confidence"`
		SuggestedTheme    string   `json:"suggested_theme"`
		SuggestedTemplates []string `json:"suggested_templates"`
	}

	if err := json.Unmarshal([]byte(resp.Content), &llmResult); err != nil {
		// 尝试从 markdown 代码块中提取
		content := strings.TrimSpace(resp.Content)
		if idx := strings.Index(content, "```"); idx >= 0 {
			start := idx + 3
			if strings.HasPrefix(content[start:], "json") {
				start += 4
			}
			end := strings.Index(content[start:], "```")
			if end >= 0 {
				content = content[start : start+end]
			}
		}
		if err := json.Unmarshal([]byte(content), &llmResult); err != nil {
			logger.Warn("intent_llm_parse_failed", "content", truncate(resp.Content, 200))
			return nil, err
		}
	}

	return &ClassificationResult{
		Intent:             ParseIntent(llmResult.Intent),
		IntentReasoning:    llmResult.IntentReasoning,
		Domain:             ParseDomain(llmResult.Domain),
		Complexity: Complexity{
			Level:            llmResult.ComplexityLevel,
			PageCountEstimate: llmResult.PageCountEstimate,
		},
		Confidence:          llmResult.Confidence,
		SuggestedTheme:     llmResult.SuggestedTheme,
		SuggestedTemplates: llmResult.SuggestedTemplates,
	}, nil
}

// llmClassificationByTextModel 使用轻量级模型（纯 Generate 接口）进行意图分类
func (c *Classifier) llmClassificationByTextModel(ctx context.Context, query string) (*ClassificationResult, error) {
	if c.textModelFactory == nil {
		return nil, nil
	}

	m, err := c.textModelFactory(ctx)
	if err != nil {
		return nil, err
	}

	systemPrompt := `你是一个PPT任务意图分类器。根据用户输入，分类其意图并评估任务复杂度。

意图类型：
- create: 新建PPT
- edit: 编辑现有PPT
- extend: 扩展PPT（增加页数）
- regenerate: 重新生成某页
- customize: 定制化调整
- query: 询问问题
- continue: 继续未完成任务

领域类型（根据输入中的关键词判断，无明确领域关键词时填 unknown）：
- business: 商业/商务/融资/路演/创业/市场/营销/客户/投标/产品发布
- technical: 技术/架构/代码/开发/系统/工程师/微服务/API/AI
- academic: 学术/论文/答辩/课程/教学/研究/培训/考试
- government: 政务/党建/政府/思政/团课/红色/政策/汇报/党风廉政
- personal: 个人/简历/述职/总结/自我介绍/求职/年终
- creative: 创意/设计/艺术/活动/策划/品牌/文化/旅游
- unknown: 无法确定领域（领域关键词不明确时使用）

复杂度评估规则（1-10分）：
- 1-3: 简单介绍/概览/基础说明/少于5页
- 4-6: 中等深度/包含数据和图表/5-15页
- 7-10: 深度分析/包含架构图和多章节/15页以上/要求详细和全面
- 输入包含"简单""简洁""快速""brief""simple"时降低1-2分
- 输入包含"详细""深入""全面""comprehensive""detailed"时提高1-2分
- 输入包含"数据""图表""architecture""代码"时提高1分

请返回JSON格式：
{
  "intent": "意图类型",
  "intent_reasoning": "判断理由",
  "domain": "领域类型",
  "complexity_level": 1-10,
  "page_count_estimate": 预估页数,
  "confidence": 置信度0-1,
  "suggested_theme": "推荐配色主题",
  "suggested_templates": ["推荐模板列表"]
}`

	resp, err := m.Generate(ctx, []*schema.Message{
		{Role: schema.System, Content: systemPrompt},
		{Role: schema.User, Content: "用户输入: " + query},
	})
	if err != nil {
		return nil, err
	}

	// 解析 JSON 响应
	var llmResult struct {
		Intent             string   `json:"intent"`
		IntentReasoning    string   `json:"intent_reasoning"`
		Domain             string   `json:"domain"`
		ComplexityLevel    int      `json:"complexity_level"`
		PageCountEstimate  int      `json:"page_count_estimate"`
		Confidence         float64  `json:"confidence"`
		SuggestedTheme     string   `json:"suggested_theme"`
		SuggestedTemplates []string `json:"suggested_templates"`
	}

	content := strings.TrimSpace(resp.Content)
	if idx := strings.Index(content, "```"); idx >= 0 {
		start := idx + 3
		if strings.HasPrefix(content[start:], "json") {
			start += 4
		}
		end := strings.Index(content[start:], "```")
		if end >= 0 {
			content = content[start : start+end]
		}
	}
	if err := json.Unmarshal([]byte(content), &llmResult); err != nil {
		logger.Warn("intent_llm_text_parse_failed", "content", truncate(resp.Content, 200))
		return nil, err
	}

	return &ClassificationResult{
		Intent:             ParseIntent(llmResult.Intent),
		IntentReasoning:     llmResult.IntentReasoning,
		Domain:             ParseDomain(llmResult.Domain),
		Complexity: Complexity{
			Level:            llmResult.ComplexityLevel,
			PageCountEstimate: llmResult.PageCountEstimate,
		},
		Confidence:          llmResult.Confidence,
		SuggestedTheme:     llmResult.SuggestedTheme,
		SuggestedTemplates: llmResult.SuggestedTemplates,
	}, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
