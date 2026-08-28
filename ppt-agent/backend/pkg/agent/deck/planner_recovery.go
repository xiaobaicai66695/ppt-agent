package deck

import (
	"encoding/json"
	"fmt"
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

	pageCount := 0
	if cfg.Outline != nil {
		pageCount = len(cfg.Outline.Slides)
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

	manifest := &TasksManifest{Title: title}
	for i, spec := range slideSpecs {
		page := i + 1
		contentType := normalizeRecoveredContentType(spec.ContentType)
		task := &TaskItem{
			TaskID:      deterministicTaskID(page),
			PageIndex:   page,
			Title:       fallbackString(spec.Title, fmt.Sprintf("第%d页", page)),
			ContentType: contentType,
			OutputFile:  deterministicOutputFile(page),
			Status:      StatusPending,
			ContentPlan: recoveredContentPlan(title, spec.Title, contentType),
		}
		manifest.Tasks = append(manifest.Tasks, task)
	}
	normalizeManifestLayoutVariants(manifest)
	if err := validateManifestForWrite(manifest); err != nil {
		return nil, err
	}
	if err := WriteTasksDraftManifest(cfg.WorkDir, manifest); err != nil {
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
	summary := recoveredSummary(topic, title, contentType)
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
		Summary:     summary,
		SlideIntent: recoveredSlideIntent(topic, title, contentType),
		VisualIntent: &VisualIntent{
			Role:          role,
			AssetQuery:    topic + " " + title,
			ImagePosition: position,
			Caption:       title,
		},
		Components: recoveredComponents(topic, title, contentType),
	}
}

func recoveredSlideIntent(topic, title, contentType string) string {
	switch contentType {
	case "title_slide":
		return fmt.Sprintf("建立“%s”的整体主题、演示语境和第一印象。", topic)
	case "agenda":
		return fmt.Sprintf("把%s的叙事拆成可跟随的章节路线，让观众知道后续如何展开。", topic)
	case "section_divider":
		return fmt.Sprintf("开启“%s”章节，并提示这一部分在整套 PPT 中承担的转场作用。", title)
	case "summary_slide":
		return fmt.Sprintf("收束%s的核心信息，给出可复述的结论和后续行动建议。", topic)
	default:
		return fmt.Sprintf("围绕“%s”展开%s相关事实、解释和结论，支撑整套 PPT 的主线。", title, topic)
	}
}

func recoveredComponents(topic, title, contentType string) []PlanComponent {
	switch contentType {
	case "title_slide":
		return []PlanComponent{
			{ID: "deck_title_1", Type: "deck_title", Text: topic},
			{ID: "subheadline_1", Type: "subheadline", Text: fmt.Sprintf("从背景定位、代表事实、核心亮点和可感知价值四个层面，建立观众对%s的完整第一印象。", topic)},
			{ID: "key_point_1", Type: "key_point", Title: "汇报主线", Body: fmt.Sprintf("本套 PPT 将避免停留在泛泛介绍，而是把%s放到具体场景中说明：先交代基本背景，再展开代表性事实与体验，最后形成可以复述的总结判断。", topic), Emphasis: "primary"},
		}
	case "agenda":
		return []PlanComponent{
			{ID: "toc_item_1", Type: "toc_item", Title: "背景定位", Body: fmt.Sprintf("先说明%s的基本坐标、主题范围和观众需要预先理解的背景信息。", topic)},
			{ID: "toc_item_2", Type: "toc_item", Title: "核心事实", Body: fmt.Sprintf("围绕%s的代表性事实、地点、人物、机制或案例展开，避免只给抽象评价。", topic)},
			{ID: "toc_item_3", Type: "toc_item", Title: "体验与价值", Body: fmt.Sprintf("把%s与真实体验、业务价值或学习启发连接起来，让内容更容易被听众带走。", topic)},
			{ID: "toc_item_4", Type: "toc_item", Title: "总结建议", Body: "最后沉淀成清晰结论、行动建议或后续关注方向，形成完整闭环。"},
		}
	case "section_divider":
		return []PlanComponent{
			{ID: "section_marker_1", Type: "section_marker", Text: "01"},
			{ID: "headline_1", Type: "headline", Text: title},
			{ID: "subheadline_1", Type: "subheadline", Text: fmt.Sprintf("这一章聚焦%s中最需要先建立共识的部分，为后续事实、案例和结论展开提供清晰入口。", topic)},
		}
	case "card_grid":
		return []PlanComponent{
			{ID: "fact_card_1", Type: "fact_card", Title: "背景坐标", Body: fmt.Sprintf("先把%s放到地理、历史、产业或使用场景中定位，说明它为什么值得被单独介绍，以及观众应该从哪个角度理解它。", topic), Emphasis: "primary"},
			{ID: "fact_card_2", Type: "fact_card", Title: "代表事实", Body: fmt.Sprintf("补充%s最有代表性的地点、事件、产品、机制或人物，用具体对象替代抽象形容，让页面具备可验证的信息密度。", topic)},
			{ID: "insight_1", Type: "insight", Title: "关键洞察", Body: fmt.Sprintf("从这些事实中提炼出%s的核心价值：它不仅是一个被介绍的对象，还能说明某种趋势、方法或体验变化。", topic)},
			{ID: "recommendation_1", Type: "recommendation", Title: "表达建议", Body: "在讲述时优先选择听众熟悉的入口，再逐步补充细节和判断，避免一开始堆砌名词导致理解成本过高。"},
		}
	case "two_column":
		return []PlanComponent{
			{ID: "fact_card_1", Type: "fact_card", Title: "认知维度", Body: fmt.Sprintf("从历史、地理、系统结构或业务背景解释%s为何形成今天的样貌，帮助观众建立稳定的理解框架。", topic), Emphasis: "primary"},
			{ID: "insight_1", Type: "insight", Title: "判断维度", Body: fmt.Sprintf("进一步说明%s背后的关键变化、代表意义或方法启发，让页面不只是罗列事实，而是给出可以带走的判断。", topic)},
			{ID: "fact_card_2", Type: "fact_card", Title: "体验维度", Body: fmt.Sprintf("从真实场景、典型路线、用户触点或工作过程说明%s如何被感知，增强内容的现场感和可讲述性。", topic)},
			{ID: "recommendation_1", Type: "recommendation", Title: "行动维度", Body: "将前面的分析收束为下一步建议、参观路径、学习方法或汇报重点，使结论可以直接转化为行动。"},
		}
	case "summary_slide":
		return []PlanComponent{
			{ID: "key_point_1", Type: "key_point", Title: "核心认知", Body: fmt.Sprintf("%s的介绍需要同时覆盖背景、事实和价值三个层次，只有把它们串成主线，观众才容易形成稳定记忆。", topic), Emphasis: "primary"},
			{ID: "insight_1", Type: "insight", Title: "主要收获", Body: fmt.Sprintf("通过前面的展开，可以看到%s不仅有可展示的表层亮点，也能反映更深层的历史脉络、系统逻辑或实践方法。", topic)},
			{ID: "recommendation_1", Type: "recommendation", Title: "后续建议", Body: "正式汇报时可根据听众背景补充数据、图片或案例来源，把介绍型内容进一步升级为可讨论、可追问的交流材料。"},
		}
	default:
		return []PlanComponent{
			{ID: "key_point_1", Type: "key_point", Title: "基本背景", Body: fmt.Sprintf("先说明%s中“%s”的基本定位，包括它出现的场景、涉及的对象以及为什么需要单独展开。", topic, title), Emphasis: "primary"},
			{ID: "fact_card_1", Type: "fact_card", Title: "具体事实", Body: fmt.Sprintf("补充与%s直接相关的时间、地点、人物、数据、流程或案例，让页面从概念说明变成具备证据支撑的信息页。", title)},
			{ID: "insight_1", Type: "insight", Title: "解释判断", Body: fmt.Sprintf("说明这些事实对理解%s有什么帮助，并提炼出观众应该记住的一句话结论。", topic)},
			{ID: "recommendation_1", Type: "recommendation", Title: "讲述落点", Body: "收束到一个可行动或可复述的表达落点，例如下一步关注方向、实践建议、体验路线或汇报中的承接问题。"},
		}
	}
}

func recoveredSummary(topic, title, contentType string) string {
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

func fallbackString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
