package web

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cloudwego/ppt-agent/pkg/agent/deep"
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
}

var templateKeywordRules = map[string][]string{
	"activity-plan":      {"活动", "团建", "晚会", "节日", "策划"},
	"course-module":      {"课程", "课件", "教学", "课堂", "知识点"},
	"current-affairs":    {"时政", "政策", "国际形势", "形势分析"},
	"design-defense":     {"答辩", "毕业设计", "课程设计", "论文"},
	"innovation-compete": {"大创", "挑战杯", "创新创业", "竞赛"},
	"meeting-minutes":    {"会议纪要", "例会", "评审会", "行动项"},
	"personal-summary":   {"述职", "个人总结", "年终总结", "年度总结"},
	"pitch-deck":         {"路演", "融资", "创业", "投资人", "商业计划"},
	"politics-ideology":  {"思政", "团课", "爱国主义", "理想信念"},
	"product-intro":      {"产品介绍", "产品演示", "客户演示", "功能介绍"},
	"product-launch":     {"产品发布", "发布会", "新品", "上市发布"},
	"project-proposal":   {"项目立项", "项目申请", "资源申请", "项目提案"},
	"research-report":    {"调研", "研究报告", "市场分析", "行业分析", "可行性"},
	"short-class-talk":   {"短讲", "五分钟", "十分钟", "课堂分享"},
	"tech-intro":         {"技术介绍", "技术科普", "新技术", "入门介绍"},
	"tech-sharing":       {"技术分享", "架构", "开发实践", "代码", "工程"},
	"training-course":    {"培训", "新人", "内训", "技能训练"},
	"weekly-report":      {"周报", "月报", "工作汇报", "项目复盘", "季度复盘"},
}

var backgroundSuppressKeywords = []string{"不要背景", "无背景", "纯色", "极简", "数据密集", "财务报表", "表格为主"}

func (s *Server) resolveTemplateSelection(query string, selection *TemplateSelection) (*deep.TaskOutline, TemplateStrategy, error) {
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
		return outlineFromTemplate(query, preset, strategy), strategy, nil
	case "recommended":
		strategy, preset, err := s.recommendTemplateStrategy(query)
		if err != nil {
			return nil, TemplateStrategy{}, err
		}
		return outlineFromTemplate(query, preset, strategy), strategy, nil
	default:
		return nil, TemplateStrategy{}, fmt.Errorf("template_selection.mode must be preset or recommended")
	}
}

func (s *Server) recommendTemplateStrategy(query string) (TemplateStrategy, *templates.TemplateInfo, error) {
	presets := s.templateLoader.ListPresets()
	if len(presets) == 0 {
		return TemplateStrategy{}, nil, fmt.Errorf("no preset templates available")
	}
	preset, score, matches := scorePresets(query, presets)
	if score == 0 {
		if generic := s.templateLoader.GetPreset("generic"); generic != nil {
			preset = generic
		}
	}

	strategy := TemplateStrategy{
		Mode:     "recommended",
		Template: preset.Name,
		Theme:    s.validThemeOrFallback(preset.DefaultPalette),
	}
	background, backgroundScore := scoreBackgrounds(query, s.templateLoader.ListBackgrounds())
	if backgroundScore > 0 && !containsAny(strings.ToLower(query), backgroundSuppressKeywords) {
		strategy.UseBackground = true
		strategy.Background = background
	}
	if len(matches) > 0 {
		strategy.Reason = fmt.Sprintf("主题命中 %s，推荐%s", strings.Join(matches, "、"), preset.DisplayName)
	} else {
		strategy.Reason = "主题没有明显专用场景，使用通用模板和保守视觉策略"
	}
	if strategy.UseBackground {
		strategy.Reason += "，并在视觉页使用匹配背景"
	} else {
		strategy.Reason += "，信息页保持纯色背景"
	}
	return strategy, preset, nil
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

func outlineFromTemplate(query string, preset *templates.TemplateInfo, strategy TemplateStrategy) *deep.TaskOutline {
	slides := make([]deep.SlideOutline, 0, len(preset.DefaultSlides))
	for _, slide := range preset.DefaultSlides {
		background := strings.TrimSpace(slide.Background)
		if strategy.UseBackground && strategy.Background != "" && isVisualBackgroundPage(slide.ContentType) {
			background = strategy.Background
		}
		slides = append(slides, deep.SlideOutline{
			Title: slide.Title, ContentType: slide.ContentType,
			Description: slide.Description, Background: background,
		})
	}
	return &deep.TaskOutline{
		Template: preset.Name, Theme: strategy.Theme, Title: strings.TrimSpace(query),
		ContentMode:   deep.OutlineContentModeTemplateScaffold,
		UseBackground: strategy.UseBackground, RecommendedBackground: strategy.Background,
		RecommendationReason: strategy.Reason, Slides: slides,
	}
}

func scorePresets(query string, presets []templates.TemplateInfo) (*templates.TemplateInfo, int, []string) {
	normalized := strings.ToLower(strings.TrimSpace(query))
	type candidate struct {
		preset  *templates.TemplateInfo
		score   int
		matches []string
	}
	candidates := make([]candidate, 0, len(presets))
	for i := range presets {
		preset := &presets[i]
		current := candidate{preset: preset}
		for _, keyword := range templateKeywordRules[preset.Name] {
			if strings.Contains(normalized, strings.ToLower(keyword)) {
				current.score += 12
				current.matches = append(current.matches, keyword)
			}
		}
		for _, token := range append(append([]string{}, preset.Tags...), preset.DisplayName, preset.Description) {
			token = strings.ToLower(strings.TrimSpace(token))
			if len([]rune(token)) >= 2 && strings.Contains(normalized, token) {
				current.score += 4
				current.matches = append(current.matches, token)
			}
		}
		candidates = append(candidates, current)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].preset.Name < candidates[j].preset.Name
		}
		return candidates[i].score > candidates[j].score
	})
	best := candidates[0]
	return best.preset, best.score, uniqueStrings(best.matches)
}

func scoreBackgrounds(query string, backgrounds []templates.BackgroundThemeInfo) (string, int) {
	normalized := strings.ToLower(strings.TrimSpace(query))
	bestName, bestScore := "", 0
	for _, background := range backgrounds {
		score := 0
		for _, scenario := range background.Scenarios {
			if strings.Contains(normalized, strings.ToLower(scenario)) {
				score += 8
			}
		}
		if score > bestScore || (score == bestScore && score > 0 && background.Name < bestName) {
			bestName, bestScore = background.Name, score
		}
	}
	return bestName, bestScore
}

func isVisualBackgroundPage(contentType string) bool {
	switch contentType {
	case "title_slide", "section_divider", "quote_slide", "summary_slide", "brand_focus":
		return true
	default:
		return false
	}
}

func containsAny(value string, candidates []string) bool {
	for _, candidate := range candidates {
		if strings.Contains(value, strings.ToLower(candidate)) {
			return true
		}
	}
	return false
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok || value == "" {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
