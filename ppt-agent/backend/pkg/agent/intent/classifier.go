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
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/logger"
)

// Classifier 意图分类器
type Classifier struct {
	modelFactory func(ctx context.Context) (model.ChatModel, error)
	useLLM       bool
}

// NewClassifier 创建意图分类器
// modelFactory 用于创建 LLM 实例（可选，用于更精确的分类）
func NewClassifier(modelFactory func(ctx context.Context) (model.ChatModel, error)) *Classifier {
	return &Classifier{
		modelFactory: modelFactory,
		useLLM:      modelFactory != nil,
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
		llmResult, err := c.llmClassification(ctx, query)
		if err != nil {
			logger.Warn("intent_llm_classification_failed", "error", err.Error())
		} else {
			// 合并结果：优先使用高置信度的结果
			if llmResult.Confidence > result.Confidence {
				logger.Debug("intent_classified_by_llm",
					"intent", llmResult.Intent.String(),
					"confidence", llmResult.Confidence)
				return llmResult, nil
			}
		}
	}

	return result, nil
}

// ruleBasedClassification 基于规则的意图分类
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

	// === 意图识别规则 ===
	// 继续任务
	if matched, score := c.matchIntent(queryLower, []string{"继续", "继续做", "接着做", "resume", "continue"}); matched {
		result.Intent = IntentContinue
		result.Confidence = score
		result.IntentReasoning = "检测到继续任务的关键字"
		return result
	}

	// 重新生成
	if matched, score := c.matchIntent(queryLower, []string{"重新生成", "重新做", "再来一次", "regenerate", "redo"}); matched {
		result.Intent = IntentRegenerate
		result.Confidence = score
		result.IntentReasoning = "检测到重新生成的关键字"
		return result
	}

	// 编辑现有
	if matched, score := c.matchIntent(queryLower, []string{"编辑", "修改", "调整", "edit", "modify", "改一下", "调整一下"}); matched {
		result.Intent = IntentEdit
		result.Confidence = score
		result.IntentReasoning = "检测到编辑现有PPT的关键字"
		return result
	}

	// 扩展PPT
	if matched, score := c.matchIntent(queryLower, []string{"加几页", "扩展", "增加", "添加", "多几页", "再加", "补充", "extend", "add"}); matched {
		result.Intent = IntentExtend
		result.Confidence = score
		result.IntentReasoning = "检测到扩展PPT的关键字"
		return result
	}

	// 定制化调整
	if matched, score := c.matchIntent(queryLower, []string{"换个配色", "换个风格", "改颜色", "定制", "customize", "风格"}); matched {
		result.Intent = IntentCustomize
		result.Confidence = score
		result.IntentReasoning = "检测到定制化调整的关键字"
		return result
	}

	// 询问问题
	if matched, score := c.matchIntent(queryLower, []string{"怎么", "如何", "请问", "是什么", "多少", "how", "what", "why", "question"}); matched {
		// 需要确保不是创建PPT的上下文
		if !c.containsCreateKeywords(queryLower) {
			result.Intent = IntentQuery
			result.Confidence = score
			result.IntentReasoning = "检测到询问类关键字且无创建意图"
			return result
		}
	}

	// 默认：新建PPT
	result.Intent = IntentCreate
	result.IntentReasoning = "未检测到特定意图关键字，默认为新建PPT"
	result.Confidence = 0.7

	// === 领域识别 ===
	result.Domain = c.classifyDomain(queryLower)

	// === 复杂度评估 ===
	result.Complexity = c.assessComplexity(queryLower)

	// === 紧迫度评估 ===
	result.Urgency = c.assessUrgency(queryLower)

	// === 建议 ===
	c.enrichRecommendations(result, queryLower)

	return result
}

// matchIntent 匹配意图关键字，返回(是否匹配, 置信度)
func (c *Classifier) matchIntent(query string, keywords []string) (bool, float64) {
	for i, kw := range keywords {
		if strings.Contains(query, kw) {
			// 越靠前的关键字置信度越高
			confidence := 0.9 - float64(i)*0.05
			if confidence < 0.7 {
				confidence = 0.7
			}
			return true, confidence
		}
	}
	return false, 0
}

// containsCreateKeywords 检查是否包含创建PPT的关键字
func (c *Classifier) containsCreateKeywords(query string) bool {
	createKeywords := []string{"做一个", "做一个ppt", "做ppt", "帮我做", "帮我做ppt", "创建", "生成", "制作", "写一个",
		"做", "写", "ppt", "presentation", "演示", "幻灯片", "帮我", "做一个关于"}
	for _, kw := range createKeywords {
		if strings.Contains(query, kw) {
			return true
		}
	}
	return false
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
			var pageCount int
			if _, err := parseInt(matches[1]); err == nil {
				pageCount = int(matches[1][0]-'0') * 10
				if pageCount > 0 && pageCount <= 100 {
					cpl.PageCountEstimate = pageCount
				}
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

领域类型：
- business: 商业/商务
- technical: 技术/工程
- academic: 学术/教育
- government: 政务/党建
- personal: 个人/生活
- creative: 创意/艺术

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

// parseInt 解析字符串中的数字
func parseInt(s string) (int, error) {
	re := regexp.MustCompile(`\d+`)
	matches := re.FindString(s)
	if matches == "" {
		return 0, nil
	}
	var n int
	for _, c := range matches {
		n = n*10 + int(c-'0')
	}
	return n, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
