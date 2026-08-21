package web

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
	agentintent "github.com/cloudwego/ppt-agent/pkg/agent/intent"
	"github.com/cloudwego/ppt-agent/pkg/assets/unsplash"
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

var backgroundSuppressKeywords = []string{"不要背景", "无背景", "纯色", "极简", "数据密集", "财务报表", "表格为主"}

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
		if s.localBackgroundsEnabled() && !containsAny(strings.ToLower(query), backgroundSuppressKeywords) {
			strategy.UseBackground = true
			strategy.Background = s.defaultBackgroundTheme(query, preset)
			if strategy.Background != "" {
				strategy.Reason += "，并让整套页面使用同一背景风格"
			}
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

	useBackground := true
	if intent != nil && intent.UseBackground != nil {
		useBackground = *intent.UseBackground
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
	if useBackground && s.localBackgroundsEnabled() {
		strategy.UseBackground = true
		if intent != nil && s.isValidBackground(intent.SuggestedBackground) {
			strategy.Background = strings.TrimSpace(intent.SuggestedBackground)
		} else {
			strategy.Background = s.defaultBackgroundForPreset(preset)
		}
	}
	if palette := s.recommendedPaletteForBackground(strategy.Background); palette != "" {
		theme = palette
	}
	strategy.Theme = s.validThemeOrFallback(theme)
	if intent != nil && strings.TrimSpace(intent.IntentReasoning) != "" {
		strategy.Reason = strings.TrimSpace(intent.IntentReasoning)
	} else {
		strategy.Reason = "使用通用视觉风格，由主 Agent 根据主题动态规划内容"
	}
	if strategy.UseBackground {
		strategy.Reason += "；整套页面使用同一背景主题并轮换同目录图片"
		if palette := s.recommendedPaletteForBackground(strategy.Background); palette != "" {
			strategy.Reason += "；配色优先匹配当前背景主题"
		}
	} else if useBackground && !s.localBackgroundsEnabled() {
		strategy.Reason += "；专题背景图交由 Planner 使用图片搜索规划"
	} else {
		strategy.Reason += "；整套采用清晰的纯色信息表面"
	}
	return strategy, preset, nil
}

func (s *Server) localBackgroundsEnabled() bool {
	return !unsplash.IsConfigured()
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

func (s *Server) recommendedPaletteForBackground(background string) string {
	theme := backgroundTheme(background)
	if theme == "" {
		return ""
	}
	for _, candidate := range s.templateLoader.ListBackgrounds() {
		if candidate.Name == theme && strings.TrimSpace(candidate.RecommendedPalette) != "" {
			return candidate.RecommendedPalette
		}
	}
	return ""
}

func (s *Server) outlineFromTemplate(query string, preset *templates.TemplateInfo, strategy TemplateStrategy) *deck.TaskOutline {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	slides := make([]deck.SlideOutline, 0, len(preset.DefaultSlides))
	previousBackground := ""
	for _, slide := range preset.DefaultSlides {
		background := strings.TrimSpace(slide.Background)
		if strategy.UseBackground && strategy.Background != "" {
			background = strategy.Background
		}
		if background != "" {
			background = s.randomBackgroundReference(background, previousBackground, rng)
			previousBackground = background
		}
		slides = append(slides, deck.SlideOutline{
			Title: slide.Title, ContentType: slide.ContentType,
			Description: slide.Description, Background: background,
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

func (s *Server) defaultBackgroundForPreset(preset *templates.TemplateInfo) string {
	if s.isValidBackground("minimalist_blue") {
		return "minimalist_blue"
	}
	if preset != nil && preset.BackgroundOpts != nil {
		for _, theme := range preset.BackgroundOpts.Themes {
			theme = strings.TrimSpace(theme)
			if theme != "" && s.isValidBackground(theme) {
				return theme
			}
		}
	}
	backgrounds := s.templateLoader.ListBackgrounds()
	if len(backgrounds) > 0 {
		return backgrounds[0].Name
	}
	return ""
}

func (s *Server) defaultBackgroundTheme(query string, preset *templates.TemplateInfo) string {
	if background, score := scoreBackgrounds(query, s.templateLoader.ListBackgrounds()); score > 0 {
		return background
	}
	if s.isValidBackground("minimalist_blue") {
		return "minimalist_blue"
	}
	if preset != nil && preset.BackgroundOpts != nil {
		for _, theme := range preset.BackgroundOpts.Themes {
			theme = strings.TrimSpace(theme)
			if theme != "" && s.isValidBackground(theme) {
				return theme
			}
		}
	}
	backgrounds := s.templateLoader.ListBackgrounds()
	if len(backgrounds) > 0 {
		return backgrounds[0].Name
	}
	return ""
}

func (s *Server) randomBackgroundReference(background string, previous string, rng *rand.Rand) string {
	theme := backgroundTheme(background)
	refs := s.backgroundImageRefs(theme)
	if len(refs) == 0 {
		return background
	}
	candidates := append([]string{}, refs...)
	if len(candidates) > 1 && backgroundTheme(previous) == theme {
		filtered := candidates[:0]
		for _, candidate := range candidates {
			if candidate != previous {
				filtered = append(filtered, candidate)
			}
		}
		if len(filtered) > 0 {
			candidates = filtered
		}
	}
	return candidates[rng.Intn(len(candidates))]
}

func (s *Server) backgroundImageRefs(theme string) []string {
	theme = strings.TrimSpace(theme)
	if theme == "" || strings.Contains(theme, "/") || strings.Contains(theme, "\\") {
		return nil
	}
	root := filepath.Join(s.templateLoader.GetBackgroundTemplatesDir(), theme)
	var refs []string
	for _, pattern := range []string{
		filepath.Join(root, "images", "*.jpg"),
		filepath.Join(root, "images", "*.jpeg"),
		filepath.Join(root, "images", "*.png"),
	} {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			if rel, err := filepath.Rel(s.templateLoader.GetBackgroundTemplatesDir(), match); err == nil {
				refs = append(refs, filepath.ToSlash(rel))
			}
		}
	}
	sort.Strings(refs)
	return refs
}

func backgroundTheme(background string) string {
	background = strings.TrimSpace(strings.ReplaceAll(background, "\\", "/"))
	if background == "" {
		return ""
	}
	return strings.Split(background, "/")[0]
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

func containsAny(value string, candidates []string) bool {
	for _, candidate := range candidates {
		if strings.Contains(value, strings.ToLower(candidate)) {
			return true
		}
	}
	return false
}
