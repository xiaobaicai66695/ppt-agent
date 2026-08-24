package web

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
	agentintent "github.com/cloudwego/ppt-agent/pkg/agent/intent"
	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/templates"
)

type TemplateSelection struct {
	Mode     string `json:"mode"`
	Template string `json:"template,omitempty"`
}

type TemplateStrategy struct {
	Mode          string `json:"mode"`
	Template      string `json:"template"`
	Theme         string `json:"theme"`
	UseBackground bool   `json:"use_background"`
	Background    string `json:"background,omitempty"`
	Reason        string `json:"reason"`
	PageCount     int    `json:"page_count,omitempty"`
}

type TemplateRecommendation struct {
	Strategy        TemplateStrategy               `json:"strategy"`
	PrimaryTemplate TemplateCandidate              `json:"primary_template"`
	RankedTemplates []TemplateCandidate            `json:"ranked_templates"`
	Theme           *templates.ThemeInfo           `json:"theme,omitempty"`
	VisualPolicy    string                         `json:"visual_policy"`
	ComponentFocus  []string                       `json:"component_focus"`
	Risks           []string                       `json:"risks,omitempty"`
}

type TemplateCandidate struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Thumbnail   string   `json:"thumbnail"`
	SlideCount  int      `json:"slide_count"`
	Tags        []string `json:"tags"`
	Reason      string   `json:"reason"`
}

var explicitPageCountPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:生成|制作|创建|做|需要|共|总共|一共|页数\s*(?:为|是|[:：=])?|generate|create|make|need|total)\s*(?:一套|a\s+)?\s*(\d{1,2})[-\s]*(?:页|頁|pages?|slides?)`),
	regexp.MustCompile(`(?i)(\d{1,2})\s*(?:页|頁|pages?|slides?)\s*(?:的)?\s*(?:PPT|演示文稿|幻灯片|deck|presentation)`),
}

func (s *Server) resolveTemplateSelection(query string, selection *TemplateSelection) (*deck.TaskOutline, TemplateStrategy, error) {
	return s.resolveTemplateSelectionWithIntent(query, selection, nil)
}

func (s *Server) resolveTemplateSelectionWithIntent(query string, selection *TemplateSelection, intent *agentintent.ClassificationResult) (*deck.TaskOutline, TemplateStrategy, error) {
	if selection == nil {
		return nil, TemplateStrategy{}, nil
	}
	mode := strings.ToLower(strings.TrimSpace(selection.Mode))
	switch mode {
	case "preset":
		name := strings.TrimSpace(selection.Template)
		preset := s.templateLoader.GetPreset(name)
		if preset == nil {
			return nil, TemplateStrategy{}, fmt.Errorf("template %q not found", name)
		}
		strategy := TemplateStrategy{
			Mode: "preset", Template: preset.Name, Theme: s.validThemeOrFallback(preset.DefaultPalette),
			Reason: "使用用户在创作首页明确选择的模板",
		}
		return s.outlineFromTemplate(query, preset, strategy), strategy, nil
	case "recommended":
		strategy, preset, err := s.recommendTemplateStrategyWithIntent(intent)
		if err != nil {
			return nil, TemplateStrategy{}, err
		}
		if pageCount, ok := explicitRequestedPageCount(query); ok {
			strategy.PageCount = pageCount
		}
		return s.recommendedOutlineFromTemplate(query, preset, strategy), strategy, nil
	default:
		return nil, TemplateStrategy{}, fmt.Errorf("template_selection.mode must be preset or recommended")
	}
}

func (s *Server) handleRecommendTemplate(c *gin.Context) {
	var req struct {
		Query string `json:"query"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Query) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}

	var intent *agentintent.ClassificationResult
	ctx, cancel := context.WithTimeout(c.Request.Context(),
		time.Duration(agentutils.EnvInt("INTENT_ROUTE_TIMEOUT_SECONDS", 30))*time.Second)
	if cfg, err := deck.ProcessUserIntent(ctx, req.Query, userIDGin(c)); err == nil && cfg != nil {
		intent = cfg.IntentResult
	}
	cancel()

	recommendation, err := s.buildTemplateRecommendation(req.Query, intent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "模板推荐失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, recommendation)
}

func (s *Server) buildTemplateRecommendation(query string, intent *agentintent.ClassificationResult) (*TemplateRecommendation, error) {
	strategy, preset, err := s.recommendTemplateStrategyWithIntent(intent)
	if err != nil {
		return nil, err
	}
	if pageCount, ok := explicitRequestedPageCount(query); ok {
		strategy.PageCount = pageCount
	}
	candidates := s.rankTemplateCandidates(intent, preset)
	if len(candidates) == 0 && preset != nil {
		candidates = append(candidates, templateCandidateFromInfo(preset, "通用兜底模板，交由 Planner 动态规划章节和组件"))
	}
	rec := &TemplateRecommendation{
		Strategy:        strategy,
		RankedTemplates: candidates,
		VisualPolicy:    visualPolicyText(strategy),
		ComponentFocus:  componentFocusForStrategy(strategy, intent),
		Risks:           recommendationRisks(strategy, intent),
	}
	if len(candidates) > 0 {
		rec.PrimaryTemplate = candidates[0]
	}
	if theme, err := s.getTheme(strategy.Theme); err == nil && theme != nil {
		rec.Theme = theme
	}
	return rec, nil
}

func explicitRequestedPageCount(query string) (int, bool) {
	for _, pattern := range explicitPageCountPatterns {
		match := pattern.FindStringSubmatch(strings.TrimSpace(query))
		if len(match) < 2 {
			continue
		}
		pageCount, err := strconv.Atoi(match[1])
		if err == nil && pageCount >= 1 && pageCount <= 40 {
			return pageCount, true
		}
	}
	return 0, false
}

func (s *Server) recommendTemplateStrategy(query string) (TemplateStrategy, *templates.TemplateInfo, error) {
	return s.recommendTemplateStrategyWithIntent(nil)
}

func (s *Server) recommendTemplateStrategyWithIntent(intent *agentintent.ClassificationResult) (TemplateStrategy, *templates.TemplateInfo, error) {
	presets := s.templateLoader.ListPresets()
	if len(presets) == 0 {
		return TemplateStrategy{}, nil, fmt.Errorf("no preset templates available")
	}
	var preset *templates.TemplateInfo
	if intent != nil {
		for _, name := range intent.SuggestedTemplates {
			if candidate := s.templateLoader.GetPreset(strings.TrimSpace(name)); candidate != nil {
				preset = candidate
				break
			}
		}
	}
	if preset == nil {
		preset = s.templateLoader.GetPreset("generic")
	}
	if preset == nil {
		preset = &presets[0]
	}

	theme := preset.DefaultPalette
	if intent != nil && strings.TrimSpace(intent.SuggestedTheme) != "" {
		theme = intent.SuggestedTheme
	}
	strategy := TemplateStrategy{
		Mode:      "recommended",
		Template:  preset.Name,
		PageCount: normalizedRecommendedPageCount(intent),
	}
	strategy.Theme = s.validThemeOrFallback(theme)
	if intent != nil && strings.TrimSpace(intent.IntentReasoning) != "" {
		strategy.Reason = strings.TrimSpace(intent.IntentReasoning)
	} else {
		strategy.Reason = "使用通用视觉风格，由主 Agent 根据主题动态规划内容"
	}
	strategy.Reason += "；专题视觉素材交由 Planner 使用图片搜索规划"
	return strategy, preset, nil
}

func (s *Server) rankTemplateCandidates(intent *agentintent.ClassificationResult, primary *templates.TemplateInfo) []TemplateCandidate {
	seen := map[string]bool{}
	var candidates []TemplateCandidate
	add := func(name, reason string) {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			return
		}
		preset := s.templateLoader.GetPreset(name)
		if preset == nil {
			return
		}
		seen[name] = true
		candidates = append(candidates, templateCandidateFromInfo(preset, reason))
	}
	if primary != nil {
		add(primary.Name, "主推荐模板，匹配当前任务类型和可用生成能力")
	}
	if intent != nil {
		for _, name := range intent.SuggestedTemplates {
			add(name, "意图识别建议的候选模板")
		}
		switch intent.Domain {
		case agentintent.DomainGovernment:
			add("current-affairs", "正式汇报和政策类内容适合稳健叙事")
			add("politics-ideology", "党政/思政场景适合庄重表达")
		case agentintent.DomainBusiness:
			add("project-proposal", "商务方案适合目标、路径、资源和收益结构")
			add("weekly-report", "汇报复盘适合进展、问题和计划结构")
		case agentintent.DomainAcademic:
			add("research-report", "研究报告适合问题、方法、发现和结论结构")
			add("course-module", "学术/教学内容适合按章节递进展开")
		case agentintent.DomainTechnical:
			add("tech-sharing", "技术分享适合概念、架构和实践案例")
			add("tech-intro", "技术介绍适合从概念到应用分层讲解")
		case agentintent.DomainCreative:
			add("product-intro", "产品说明适合卖点、场景和功能结构")
			add("product-launch", "发布类内容适合亮点、路线图和行动号召")
		case agentintent.DomainPersonal:
			add("personal-summary", "个人总结适合经历、成果和反思结构")
			add("short-class-talk", "短讲内容适合少页高密度表达")
		}
	}
	add("generic", "通用兜底模板，适合交由 Planner 自适应规划")
	if len(candidates) > 3 {
		candidates = candidates[:3]
	}
	return candidates
}

func templateCandidateFromInfo(preset *templates.TemplateInfo, reason string) TemplateCandidate {
	if preset == nil {
		return TemplateCandidate{}
	}
	return TemplateCandidate{
		Name:        preset.Name,
		DisplayName: preset.DisplayName,
		Description: preset.Description,
		Category:    preset.Category,
		Thumbnail:   preset.Thumbnail,
		SlideCount:  preset.SlideCount,
		Tags:        append([]string(nil), preset.Tags...),
		Reason:      reason,
	}
}

func visualPolicyText(strategy TemplateStrategy) string {
	_ = strategy
	return "采用组件化信息表面与专题图片搜索，Planner 会根据内容密度选择图文、表格、流程或长论述组件。"
}

func componentFocusForStrategy(strategy TemplateStrategy, intent *agentintent.ClassificationResult) []string {
	focus := []string{"观点-论据叙事", "图文混排", "章节分割"}
	if intent != nil {
		switch intent.Domain {
		case agentintent.DomainBusiness:
			focus = append(focus, "指标卡", "对比表格", "行动建议")
		case agentintent.DomainAcademic:
			focus = append(focus, "长段论述", "证据表格", "结论摘要")
		case agentintent.DomainTechnical:
			focus = append(focus, "架构图", "步骤讲解", "示例页")
		case agentintent.DomainCreative:
			focus = append(focus, "场景案例", "功能卡片", "路线图")
		default:
			focus = append(focus, "流程图", "重点卡片")
		}
	} else {
		focus = append(focus, "流程图", "重点卡片")
	}
	if strategy.PageCount > 0 && strategy.PageCount <= 4 {
		focus = append(focus, "少页高密度")
	}
	return dedupeStrings(focus)
}

func recommendationRisks(strategy TemplateStrategy, intent *agentintent.ClassificationResult) []string {
	var risks []string
	if strategy.PageCount > 24 {
		risks = append(risks, "页数较多，建议后续按章节审查生成质量")
	}
	if intent == nil {
		risks = append(risks, "未获得稳定意图识别结果，已使用通用推荐兜底")
	}
	return risks
}

func dedupeStrings(items []string) []string {
	seen := map[string]bool{}
	var result []string
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		result = append(result, item)
	}
	return result
}

func normalizedRecommendedPageCount(intent *agentintent.ClassificationResult) int {
	pageCount := 12
	if intent != nil {
		pageCount = intent.SuggestedPageCount
		if pageCount <= 0 {
			pageCount = intent.Complexity.PageCountEstimate
		}
	}
	if pageCount <= 0 {
		return 12
	}
	if pageCount > 40 {
		return 40
	}
	return pageCount
}

func (s *Server) validThemeOrFallback(name string) string {
	if _, err := s.getTheme(name); err == nil {
		return name
	}
	if _, err := s.getTheme("ocean_soft"); err == nil {
		return "ocean_soft"
	}
	themes := s.templateLoader.ListThemes()
	if len(themes) > 0 {
		return themes[0].Name
	}
	return ""
}

func (s *Server) outlineFromTemplate(query string, preset *templates.TemplateInfo, strategy TemplateStrategy) *deck.TaskOutline {
	slides := make([]deck.SlideOutline, 0, len(preset.DefaultSlides))
	for _, slide := range preset.DefaultSlides {
		slides = append(slides, deck.SlideOutline{
			Title: slide.Title, ContentType: slide.ContentType,
			Description: slide.Description,
		})
	}
	return &deck.TaskOutline{
		Template: preset.Name, Theme: strategy.Theme, Title: strings.TrimSpace(query),
		ContentMode:   deck.OutlineContentModeTemplateScaffold,
		UseBackground: strategy.UseBackground, RecommendedBackground: strategy.Background,
		RecommendationReason: strategy.Reason, Slides: slides,
	}
}

func (s *Server) recommendedOutlineFromTemplate(query string, preset *templates.TemplateInfo, strategy TemplateStrategy) *deck.TaskOutline {
	return &deck.TaskOutline{
		Template:              preset.Name,
		Theme:                 strategy.Theme,
		Title:                 strings.TrimSpace(query),
		ContentMode:           deck.OutlineContentModeRecommendedStyle,
		UseBackground:         strategy.UseBackground,
		RecommendedBackground: strategy.Background,
		RecommendationReason:  strategy.Reason,
		SuggestedPageCount:    strategy.PageCount,
		Slides:                []deck.SlideOutline{},
	}
}
