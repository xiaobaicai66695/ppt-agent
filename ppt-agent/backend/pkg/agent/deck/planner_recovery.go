package deck

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var plannerSlideLinePattern = regexp.MustCompile(`(?m)^\s*(\d{1,2})[.、]\s*([^(\n\r]+?)\s*\(([-_a-zA-Z0-9]+)\)`)

type recoveredThought struct {
	Thought string `json:"thought"`
}

func recoverMissingPlannerManifest(cfg *PPTTaskConfig, userQuery, plannerOutput string) (*TasksManifest, error) {
	if cfg == nil {
		return nil, fmt.Errorf("missing task config")
	}
	title := strings.TrimSpace(userQuery)
	if title == "" {
		title = strings.TrimSpace(cfg.Query)
	}
	if title == "" && cfg.Outline != nil {
		title = strings.TrimSpace(cfg.Outline.Title)
	}
	if title == "" {
		title = "演示文稿"
	}

	template := "generic"
	theme := "ocean_soft"
	background := ""
	pageCount := 0
	if cfg.Outline != nil {
		template = fallbackString(cfg.Outline.Template, template)
		theme = fallbackString(cfg.Outline.Theme, theme)
		background = strings.TrimSpace(cfg.Outline.RecommendedBackground)
		if cfg.Outline.UseBackground && background == "" {
			background = firstBackgroundTheme(cfg.SkillsDir)
		}
		pageCount = cfg.Outline.SuggestedPageCount
	}
	if cfg.IntentResult != nil {
		theme = fallbackString(cfg.IntentResult.SuggestedTheme, theme)
		if cfg.IntentResult.SuggestedPageCount > 0 && pageCount <= 0 {
			pageCount = cfg.IntentResult.SuggestedPageCount
		}
	}
	if pageCount <= 0 {
		pageCount = 8
	}
	if pageCount > 18 {
		pageCount = 18
	}

	slideSpecs := parseRecoveredSlideSpecs(plannerOutput)
	if len(slideSpecs) == 0 {
		slideSpecs = defaultRecoveredSlideSpecs(title, pageCount)
	}
	if len(slideSpecs) > pageCount && pageCount >= 4 {
		slideSpecs = slideSpecs[:pageCount]
	}
	for len(slideSpecs) < pageCount {
		slideSpecs = append(slideSpecs, recoveredSlideSpec{
			Title:       fmt.Sprintf("延展内容 %02d", len(slideSpecs)+1),
			ContentType: "content_slide",
		})
	}

	manifest := &TasksManifest{Title: title, Theme: theme, Template: template}
	refs := backgroundImageRefsFromSkills(cfg.SkillsDir, background)
	for i, spec := range slideSpecs {
		page := i + 1
		contentType := normalizeRecoveredContentType(spec.ContentType)
		task := &TaskItem{
			TaskID:      strconv.Itoa(page),
			PageIndex:   page,
			Title:       fallbackString(spec.Title, fmt.Sprintf("第%d页", page)),
			ContentType: contentType,
			Description: recoveredDescription(title, spec.Title, contentType),
			OutputFile:  fmt.Sprintf("%02d_%s.pptx", page, safeOutputStem(spec.Title)),
			Status:      StatusPending,
			ContentPlan: recoveredContentPlan(title, spec.Title, contentType),
			Background:  rotatingBackgroundRef(background, refs, page, previousTaskBackground(manifest)),
		}
		manifest.Tasks = append(manifest.Tasks, task)
	}
	normalizeManifestLayoutVariants(manifest)
	if err := validateManifestForWrite(manifest); err != nil {
		return nil, err
	}
	if err := WriteTasksManifest(cfg.WorkDir, manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}

type recoveredSlideSpec struct {
	Title       string
	ContentType string
}

func parseRecoveredSlideSpecs(output string) []recoveredSlideSpec {
	output = plannerScratchText(output)
	matches := plannerSlideLinePattern.FindAllStringSubmatch(output, -1)
	if len(matches) == 0 {
		return nil
	}
	specs := make([]recoveredSlideSpec, 0, len(matches))
	seenPage := map[int]bool{}
	for _, match := range matches {
		page, _ := strconv.Atoi(match[1])
		if page <= 0 || seenPage[page] {
			continue
		}
		seenPage[page] = true
		specs = append(specs, recoveredSlideSpec{
			Title:       strings.TrimSpace(match[2]),
			ContentType: strings.TrimSpace(match[3]),
		})
	}
	return specs
}

func plannerScratchText(output string) string {
	output = strings.TrimSpace(output)
	if output == "" {
		return ""
	}
	var thought recoveredThought
	if err := json.Unmarshal([]byte(output), &thought); err == nil && strings.TrimSpace(thought.Thought) != "" {
		return thought.Thought
	}
	return output
}

func plannerVisibleThought(content string) string {
	thought := plannerScratchText(content)
	if thought == "" || thought == strings.TrimSpace(content) && !isPlannerScratchThought(content) {
		return ""
	}
	lines := strings.Split(thought, "\n")
	var out []string
	out = append(out, "规划草案：")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if plannerSlideLinePattern.MatchString(line) {
			out = append(out, line)
			continue
		}
		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "•") {
			out = append(out, line)
			continue
		}
		out = append(out, "- "+line)
	}
	out = append(out, "正在写入 DeckSpec，校验通过后会进入并发渲染。")
	return strings.Join(out, "\n") + "\n"
}

func isPlannerScratchThought(content string) bool {
	content = strings.TrimSpace(content)
	if content == "" || !strings.HasPrefix(content, "{") {
		return false
	}
	var thought recoveredThought
	return json.Unmarshal([]byte(content), &thought) == nil && strings.TrimSpace(thought.Thought) != ""
}

func defaultRecoveredSlideSpecs(topic string, pageCount int) []recoveredSlideSpec {
	base := []recoveredSlideSpec{
		{Title: topic, ContentType: "title_slide"},
		{Title: "内容概览", ContentType: "agenda"},
		{Title: "主题背景", ContentType: "section_divider"},
		{Title: "基本概况", ContentType: "image_text"},
		{Title: "核心亮点", ContentType: "card_grid"},
		{Title: "重点解读", ContentType: "two_column"},
		{Title: "实践建议", ContentType: "content_slide"},
		{Title: "总结展望", ContentType: "summary_slide"},
	}
	if pageCount <= 0 || pageCount >= len(base) {
		return base
	}
	return base[:pageCount]
}

func recoveredContentPlan(topic, title, contentType string) *ContentPlan {
	summary := recoveredDescription(topic, title, contentType)
	role := "cards"
	position := "inline"
	switch contentType {
	case "title_slide", "section_divider":
		role = "hero_photo"
		position = "background"
	case "image_text":
		role = "supporting_photo"
		position = "right"
	case "chart_slide", "kpi_dashboard":
		role = "chart"
	}
	return &ContentPlan{
		Summary: summary,
		VisualIntent: &VisualIntent{
			Role:          role,
			AssetQuery:    topic + " " + title,
			ImagePosition: position,
			Caption:       title,
		},
		Elements: recoveredElements(topic, title, contentType),
	}
}

func recoveredElements(topic, title, contentType string) []ContentElement {
	switch contentType {
	case "agenda", "title_slide", "section_divider":
		return nil
	case "card_grid":
		return []ContentElement{
			{Type: "key_point_card", Title: "自然资源", Description: fmt.Sprintf("围绕%s的地理环境、景观资源和城市识别度展开，说明其成为主题核心的原因。", topic)},
			{Type: "key_point_card", Title: "历史文化", Description: fmt.Sprintf("补充%s的历史沿革、地方文化与代表性符号，让内容从风景介绍延伸到城市气质。", topic)},
			{Type: "key_point_card", Title: "体验场景", Description: "整理游客或听众最容易感知的场景，包括路线、活动、消费和公共服务等具体触点。"},
			{Type: "key_point_card", Title: "总结价值", Description: fmt.Sprintf("提炼%s对外传播或学习汇报中的核心价值，形成可收束的演示结论。", topic)},
		}
	case "two_column":
		return []ContentElement{
			{Type: "point", Title: "认知维度", Items: []string{fmt.Sprintf("从历史、地理和产业视角解释%s为何值得介绍。", topic), fmt.Sprintf("用具体地点、时间和场景支撑%s的主题表达。", topic)}},
			{Type: "point", Title: "体验维度", Items: []string{fmt.Sprintf("从游览、文化和生活体验角度组织%s的叙事材料。", topic), "把亮点转化为可被观众快速理解的行动建议。"}},
		}
	default:
		return []ContentElement{
			{Type: "bullet_list", Items: []string{
				fmt.Sprintf("先说明%s的基本背景和定位，帮助观众快速建立主题坐标。", topic),
				fmt.Sprintf("再补充%s的代表性事实、地点或案例，让内容避免停留在泛泛介绍。", topic),
				fmt.Sprintf("最后提炼%s带来的观察结论或行动建议，形成完整收束。", topic),
			}},
		}
	}
}

func recoveredDescription(topic, title, contentType string) string {
	switch contentType {
	case "title_slide":
		return fmt.Sprintf("以“%s”为主题的开场页，点明演示对象、整体调性和核心看点。", topic)
	case "agenda":
		return fmt.Sprintf("列出围绕%s展开的主要章节，帮助观众理解后续内容顺序。", topic)
	case "section_divider":
		return fmt.Sprintf("章节过渡页，开启“%s”相关板块，并用一句话提示本章关注点。", title)
	case "summary_slide":
		return fmt.Sprintf("总结%s的核心认知、代表亮点和后续行动建议。", topic)
	default:
		return fmt.Sprintf("围绕%s中的“%s”展开，包含背景解释、具体事实和可被观众记住的结论。", topic, title)
	}
}

func normalizeRecoveredContentType(contentType string) string {
	contentType = strings.TrimSpace(contentType)
	if contentType == "" {
		return "content_slide"
	}
	switch contentType {
	case "title_slide", "agenda", "section_divider", "content_slide", "image_text", "card_grid",
		"two_column", "three_column", "summary_slide", "chart_slide", "kpi_dashboard", "process_flow",
		"timeline", "quote_slide", "stat_slide", "case_study", "comparison_table", "icon_grid":
		return contentType
	default:
		return "content_slide"
	}
}

func backgroundImageRefsFromSkills(skillsDir, background string) []string {
	theme := backgroundTheme(background)
	if theme == "" || skillsDir == "" {
		return nil
	}
	root := filepath.Join(skillsDir, "visual_designer", "background_templates", theme, "images")
	paths, _ := filepath.Glob(filepath.Join(root, "*"))
	refs := make([]string, 0, len(paths))
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		refs = append(refs, filepath.ToSlash(filepath.Join(theme, "images", filepath.Base(path))))
	}
	return refs
}

func firstBackgroundTheme(skillsDir string) string {
	root := filepath.Join(skillsDir, "visual_designer", "background_templates")
	entries, err := os.ReadDir(root)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return entry.Name()
		}
	}
	return ""
}

func previousTaskBackground(manifest *TasksManifest) string {
	if manifest == nil || len(manifest.Tasks) == 0 {
		return ""
	}
	return manifest.Tasks[len(manifest.Tasks)-1].Background
}

func fallbackString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func safeOutputStem(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return "slide"
	}
	var b strings.Builder
	for _, r := range title {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r >= '\u4e00' && r <= '\u9fff':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
		if b.Len() >= 48 {
			break
		}
	}
	stem := strings.Trim(b.String(), "_")
	if stem == "" {
		return "slide"
	}
	return stem
}
