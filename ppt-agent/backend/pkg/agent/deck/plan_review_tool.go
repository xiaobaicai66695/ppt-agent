package deck

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const planReviewFileName = "tasks.review.json"

const argumentBlockTargetMinChars = 440
const imageTextNarrativeMinChars = 240
const informationPageNarrativeMinChars = 220
const componentBodyMinChars = 18
const minVisualMixedSlidesForDeck = 2

type PlanReviewReport struct {
	OK              bool              `json:"ok"`
	Passed          bool              `json:"passed"`
	Target          string            `json:"target"`
	Fingerprint     string            `json:"fingerprint"`
	Round           int               `json:"round"`
	TotalSlides     int               `json:"total_slides"`
	IssueCount      int               `json:"issue_count"`
	Issues          []PlanReviewIssue `json:"issues,omitempty"`
	Summary         string            `json:"summary"`
	NextActions     []string          `json:"next_actions,omitempty"`
	ReviewedAt      string            `json:"reviewed_at"`
	BackgroundPages int               `json:"background_pages"`
}

func ReviewTasksDraftManifest(workDir string, round int) (*PlanReviewReport, error) {
	manifest, err := ReadTasksDraftManifest(workDir)
	target := "tasks.draft.json"
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		manifest, err = ReadTasksManifest(workDir)
		target = "tasks.json"
		if err != nil {
			return nil, fmt.Errorf("read manifest for review: %w", err)
		}
	}
	if round <= 0 {
		round = 1
	}
	report := ReviewTasksManifest(manifest, target, round)
	if target == "tasks.draft.json" {
		report.Fingerprint = fingerprintTasksManifest(manifest)
	}
	if err := WritePlanReviewReport(workDir, report); err != nil {
		return nil, err
	}
	return report, nil
}

func ReviewTasksManifest(manifest *TasksManifest, target string, round int) *PlanReviewReport {
	report := &PlanReviewReport{
		OK:         true,
		Target:     target,
		Round:      round,
		ReviewedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if manifest == nil {
		report.Issues = append(report.Issues, PlanReviewIssue{
			Code:     "invalid_component_schema",
			Severity: "error",
			Message:  "DeckSpec 为空，无法进入渲染。",
		})
		return finalizePlanReviewReport(report, nil)
	}
	report.TotalSlides = len(manifest.Tasks)
	report.Fingerprint = fingerprintTasksManifest(manifest)
	if strings.TrimSpace(manifest.Title) == "" {
		report.Issues = append(report.Issues, PlanReviewIssue{Code: "weak_narrative", Severity: "error", Message: "缺少整套 PPT 标题。"})
	}
	if err := validateManifestForWrite(manifest); err != nil {
		report.Issues = append(report.Issues, PlanReviewIssue{
			Code:     "invalid_component_schema",
			Severity: "error",
			Message:  err.Error(),
		})
	}
	backgroundMode := manifestBackgroundMode(manifest)
	for _, task := range manifest.Tasks {
		reviewTaskPlan(task, report, backgroundMode)
	}
	reviewDeckBackgroundVariety(manifest, report)
	reviewDeckVisualMix(manifest, report)
	return finalizePlanReviewReport(report, manifest)
}

func reviewTaskPlan(task *TaskItem, report *PlanReviewReport, backgroundMode string) {
	if task == nil {
		report.Issues = append(report.Issues, PlanReviewIssue{Code: "invalid_component_schema", Severity: "error", Message: "存在空页面任务。"})
		return
	}
	page := task.PageIndex
	if task.ContentPlan == nil {
		report.Issues = append(report.Issues, PlanReviewIssue{
			Code:      "invalid_component_schema",
			Severity:  "error",
			PageIndex: page,
			Message:   "缺少 content_plan，无法进行组件级渲染。",
		})
		return
	}
	plan := task.ContentPlan
	if !hasPlanNarrativeSummary(plan) {
		report.Issues = append(report.Issues, PlanReviewIssue{Code: "weak_narrative", Severity: "error", PageIndex: page, Message: "content_plan.summary 为空。"})
	}
	if strings.TrimSpace(plan.SlideIntent) == "" && !isSimpleTitleLikeSlide(task.ContentType) {
		report.Issues = append(report.Issues, PlanReviewIssue{Code: "weak_narrative", Severity: "error", PageIndex: page, Message: "缺少 slide_intent，页面在整套叙事中的作用不明确。"})
	}
	if len(plan.Components) == 0 {
		report.Issues = append(report.Issues, PlanReviewIssue{Code: "invalid_component_schema", Severity: "error", PageIndex: page, Message: "缺少 content_plan.components，当前主流程要求组件优先。"})
	}
	if len(plan.Components) > 0 && !hasNarrativeAnchor(plan.Components) {
		report.Issues = append(report.Issues, PlanReviewIssue{Code: "weak_narrative", Severity: "warning", PageIndex: page, Message: "组件缺少 headline、insight、recommendation 或 argument_block 等观点锚点，容易退化为罗列。"})
	}
	if !hasUsableBackgroundPlan(task) && backgroundMode != "none" {
		severity := "error"
		if backgroundMode == "optional" {
			severity = "warning"
		}
		report.Issues = append(report.Issues, PlanReviewIssue{
			Code:      "missing_background_image",
			Severity:  severity,
			PageIndex: page,
			Message:   "页面缺少可执行背景图片计划；必须在 visual_intent 或 image 组件中使用 asset_purpose=background，并填写 asset_query 或已下载 local_path。仅显式 clean_text_only 页面可豁免。",
		})
	} else {
		report.BackgroundPages++
	}
	for i := range plan.Components {
		component := &plan.Components[i]
		reviewComponentContentQuality(task.ContentType, page, component, report)
		if isExternalBackgroundComponent(component) && strings.TrimSpace(component.LocalPath) == "" && strings.TrimSpace(component.AssetQuery) == "" {
			report.Issues = append(report.Issues, PlanReviewIssue{
				Code:        "missing_background_image",
				Severity:    "error",
				PageIndex:   page,
				ComponentID: component.ID,
				Message:     "背景图片组件缺少 local_path 和 asset_query，无法执行后续素材规划。",
			})
		}
		if component.Type == "argument_block" && runeLen(firstNonEmptyString(component.Body, component.Text)) < argumentBlockTargetMinChars {
			currentChars := runeLen(firstNonEmptyString(component.Body, component.Text))
			report.Issues = append(report.Issues, PlanReviewIssue{
				Code:        "low_information_density",
				Severity:    "warning",
				PageIndex:   page,
				ComponentID: component.ID,
				Message:     fmt.Sprintf("argument_block 当前约 %d 字，少于 %d 字；需要完整论述时补足背景、证据、推理和结论，否则改用 paragraph、insight 或 key_point。", currentChars, argumentBlockTargetMinChars),
			})
		}
	}
	if task.ContentType == "image_text" && !hasForegroundImageComponent(plan.Components) {
		report.Issues = append(report.Issues, PlanReviewIssue{
			Code:      "layout_mismatch",
			Severity:  "warning",
			PageIndex: page,
			Message:   "image_text 页面缺少 scene/evidence 图片组件；背景图只承担浅色氛围，页内示例图片应作为 image 组件参与图文混排。",
		})
	}
	if task.ContentType == "image_text" {
		narrativeChars := imageTextNarrativeChars(plan.Components)
		if narrativeChars > 0 && narrativeChars < imageTextNarrativeMinChars {
			report.Issues = append(report.Issues, PlanReviewIssue{
				Code:      "low_information_density",
				Severity:  "warning",
				PageIndex: page,
				Message:   fmt.Sprintf("image_text 正文当前约 %d 字，少于 %d 字；图文页需要完整解释场景、事实、影响和结论，避免大面积文字面板空洞。", narrativeChars, imageTextNarrativeMinChars),
			})
		}
	}
	if requiresInformationDensity(task.ContentType) {
		narrativeChars := informationNarrativeChars(plan.Components)
		if narrativeChars > 0 && narrativeChars < informationPageNarrativeMinChars {
			report.Issues = append(report.Issues, PlanReviewIssue{
				Code:      "low_information_density",
				Severity:  "warning",
				PageIndex: page,
				Message:   fmt.Sprintf("%s 页面正文组件总量当前约 %d 字，少于 %d 字；需要补足背景、事实、影响和结论，避免大面积文本面板空洞。", task.ContentType, narrativeChars, informationPageNarrativeMinChars),
			})
		}
	}
}

func manifestBackgroundMode(manifest *TasksManifest) string {
	if manifest == nil || manifest.VisualPolicy == nil {
		return "required"
	}
	mode := strings.ToLower(strings.TrimSpace(manifest.VisualPolicy.Mode))
	switch mode {
	case "none", "optional", "required":
		return mode
	default:
		return "required"
	}
}

// plannerPreflightIssues returns quality problems that the Planner can and
// should prevent before a draft reaches the independent Reviewer. Keeping this
// based on ReviewTasksManifest avoids a second set of thresholds drifting from
// the review gate.
func plannerPreflightIssues(manifest *TasksManifest) []PlanReviewIssue {
	report := ReviewTasksManifest(manifest, "planner_initialize", 0)
	if report == nil || len(report.Issues) == 0 {
		return nil
	}
	issues := make([]PlanReviewIssue, 0, len(report.Issues))
	for _, issue := range report.Issues {
		switch strings.TrimSpace(issue.Code) {
		case "low_information_density":
			if !strings.EqualFold(strings.TrimSpace(issue.Severity), "warning") {
				issues = append(issues, issue)
			}
		case "invalid_component_schema", "missing_background_image", "weak_narrative", "overload_capacity", "layout_mismatch":
			issues = append(issues, issue)
		}
	}
	return issues
}

func finalizePlanReviewReport(report *PlanReviewReport, manifest *TasksManifest) *PlanReviewReport {
	report.IssueCount = len(report.Issues)
	report.Passed = !hasBlockingPlanReviewIssue(report.Issues)
	if report.Passed {
		report.Summary = fmt.Sprintf("规划审核通过：%d 页均满足结构、组件容量和阻塞性质量门。", report.TotalSlides)
		report.NextActions = []string{"由 Go workflow 原子提交正式 tasks.json"}
	} else {
		report.Summary = fmt.Sprintf("规划审核未通过：发现 %d 个问题，需要按页修订后重新 review。", report.IssueCount)
		report.NextActions = summarizePlanReviewActions(report.Issues)
	}
	sort.SliceStable(report.Issues, func(i, j int) bool {
		if report.Issues[i].PageIndex == report.Issues[j].PageIndex {
			return report.Issues[i].Code < report.Issues[j].Code
		}
		return report.Issues[i].PageIndex < report.Issues[j].PageIndex
	})
	if manifest != nil && report.Fingerprint == "" {
		report.Fingerprint = fingerprintTasksManifest(manifest)
	}
	return report
}

func hasBlockingPlanReviewIssue(issues []PlanReviewIssue) bool {
	for _, issue := range issues {
		if strings.EqualFold(strings.TrimSpace(issue.Severity), "warning") {
			continue
		}
		return true
	}
	return false
}

func summarizePlanReviewActions(issues []PlanReviewIssue) []string {
	seen := map[string]bool{}
	var actions []string
	for _, issue := range issues {
		action := planReviewActionForCode(issue.Code)
		if action == "" || seen[action] {
			continue
		}
		seen[action] = true
		actions = append(actions, action)
	}
	if len(actions) == 0 {
		actions = append(actions, "修订草稿后由 Go workflow 重新校验")
	}
	return actions
}

func planReviewActionForCode(code string) string {
	switch strings.TrimSpace(code) {
	case "missing_background_image":
		return "按 skill 为缺失页面补充背景图片计划；图片工具可用时下载 local_path，否则至少填写可执行 asset_query"
	case "low_information_density":
		return "补足页面事实、解释、案例和结论，避免只有泛化短句"
	case "weak_narrative":
		return "补充页面观点锚点和 slide_intent，形成观点-论据结构"
	case "invalid_component_schema":
		return "按 component_contracts.json 修正 content_type、components 和容量字段"
	case "layout_mismatch":
		return "修正 layout_variant、页面类型和组件匹配关系"
	default:
		return "修订草稿后由 Go workflow 重新校验"
	}
}

func reviewDeckBackgroundVariety(manifest *TasksManifest, report *PlanReviewReport) {
	if manifest == nil || report == nil {
		return
	}
	queriesByType := map[string]map[string]bool{}
	verbosePages := map[int]bool{}
	for _, task := range manifest.Tasks {
		if task == nil || task.ContentPlan == nil {
			continue
		}
		typeKey := backgroundContentTypeKey(task.ContentType)
		addQuery := func(subject, query string) {
			sourceQuery := strings.TrimSpace(query)
			query = firstNonEmptyString(subject, query)
			query = strings.TrimSpace(query)
			if query == "" {
				return
			}
			compact := normalizeAssetQuery(query, true)
			if compact == "" {
				return
			}
			if queriesByType[typeKey] == nil {
				queriesByType[typeKey] = map[string]bool{}
			}
			queriesByType[typeKey][compact] = true
			if isVerboseBackgroundQuery(sourceQuery, compact) {
				verbosePages[task.PageIndex] = true
			}
		}
		if visual := task.ContentPlan.VisualIntent; visual != nil && isExternalBackgroundIntent(visual) {
			addQuery(visual.AssetSubject, visual.AssetQuery)
		}
		for i := range task.ContentPlan.Components {
			component := &task.ContentPlan.Components[i]
			if isExternalBackgroundComponent(component) {
				addQuery(component.AssetSubject, component.AssetQuery)
			}
		}
	}
	for contentType, queries := range queriesByType {
		if len(queries) <= 1 {
			continue
		}
		report.Issues = append(report.Issues, PlanReviewIssue{
			Code:     "layout_mismatch",
			Severity: "warning",
			Message:  fmt.Sprintf("同一页面类型 %s 存在多个背景关键词，容易造成同类幻灯片视觉跳变；同一 content_type 应复用同一个背景关键词。", contentType),
		})
	}
	for page := range verbosePages {
		report.Issues = append(report.Issues, PlanReviewIssue{
			Code:      "layout_mismatch",
			Severity:  "warning",
			PageIndex: page,
			Message:   "背景 asset_query 过长；背景搜索应尽量只写一个英文关键词，构图、明暗和留白交给生成器处理，避免长搜索词搜到不相关图片。",
		})
	}
}

func isVerboseBackgroundQuery(query, compact string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	compact = strings.ToLower(strings.TrimSpace(compact))
	if query == "" || compact == "" || query == compact {
		return false
	}
	return len(strings.Fields(query)) > 2 ||
		strings.Contains(query, "wide landscape") ||
		strings.Contains(query, "clean negative space")
}

func reviewDeckVisualMix(manifest *TasksManifest, report *PlanReviewReport) {
	if manifest == nil || report == nil {
		return
	}
	contentSlides := 0
	visualMixedSlides := 0
	imageTextSlides := 0
	imageTextVariants := map[string]bool{}
	for _, task := range manifest.Tasks {
		if task == nil {
			continue
		}
		contentType := strings.TrimSpace(task.ContentType)
		if isNarrativeContentSlide(contentType) {
			contentSlides++
		}
		if isVisualMixedContentType(contentType) {
			visualMixedSlides++
			if contentType == "image_text" {
				imageTextSlides++
				if variant := strings.TrimSpace(task.LayoutVariant); variant != "" {
					imageTextVariants[variant] = true
				}
			}
		}
	}
	if contentSlides >= 4 && visualMixedSlides < minVisualMixedSlidesForDeck {
		report.Issues = append(report.Issues, PlanReviewIssue{
			Code:     "layout_mismatch",
			Severity: "warning",
			Message:  fmt.Sprintf("整套 PPT 图文混排不足：%d 个叙事内容页中仅 %d 页使用 image_text；应多用 scene/evidence 图片组件承载示例素材。", contentSlides, visualMixedSlides),
		})
	}
	if imageTextSlides >= 3 && len(imageTextVariants) == 1 {
		report.Issues = append(report.Issues, PlanReviewIssue{
			Code:     "layout_mismatch",
			Severity: "warning",
			Message:  "连续图文混排页的 layout_variant 过于单一；image_text 应在 image_left、image_right、image_top_band、image_bottom_band 之间轮换。未显式填写时渲染器会按页序自动轮换。",
		})
	}
}

func isNarrativeContentSlide(contentType string) bool {
	switch strings.TrimSpace(contentType) {
	case "content_slide", "card_grid", "image_text", "brand_focus":
		return true
	default:
		return false
	}
}

func isVisualMixedContentType(contentType string) bool {
	switch strings.TrimSpace(contentType) {
	case "image_text":
		return true
	default:
		return false
	}
}

func WritePlanReviewReport(workDir string, report *PlanReviewReport) error {
	if report == nil {
		return nil
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal plan review report: %w", err)
	}
	return os.WriteFile(filepath.Join(workDir, planReviewFileName), data, 0644)
}

func ReadPlanReviewReport(workDir string) (*PlanReviewReport, error) {
	data, err := os.ReadFile(filepath.Join(workDir, planReviewFileName))
	if err != nil {
		return nil, err
	}
	report := &PlanReviewReport{}
	if err := json.Unmarshal(data, report); err != nil {
		return nil, err
	}
	return report, nil
}

func RequirePassingPlanReview(workDir string, manifest *TasksManifest) error {
	report, err := ReadPlanReviewReport(workDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("draft DeckSpec has not been reviewed by the Go review workflow")
		}
		return fmt.Errorf("read plan review report: %w", err)
	}
	if !report.Passed {
		return fmt.Errorf("latest DeckSpec review did not pass: %s", report.Summary)
	}
	fingerprint := fingerprintTasksManifest(manifest)
	if report.Fingerprint != fingerprint {
		return fmt.Errorf("DeckSpec changed after latest review; run the Go review workflow again before commit")
	}
	return nil
}

func fingerprintTasksManifest(manifest *TasksManifest) string {
	if manifest == nil {
		return ""
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func isSimpleTitleLikeSlide(contentType string) bool {
	switch strings.TrimSpace(contentType) {
	case "title_slide", "section_divider", "quote_slide":
		return true
	default:
		return false
	}
}

func hasNarrativeAnchor(components []PlanComponent) bool {
	for _, component := range components {
		switch strings.TrimSpace(component.Type) {
		case "headline", "deck_title", "argument_block", "insight", "recommendation", "key_point", "quote_block":
			return true
		}
	}
	return false
}

func hasForegroundImageComponent(components []PlanComponent) bool {
	for i := range components {
		component := &components[i]
		if strings.TrimSpace(component.Type) != "image" || isExternalBackgroundComponent(component) {
			continue
		}
		purpose := strings.TrimSpace(component.AssetPurpose)
		if purpose != "" && !strings.EqualFold(purpose, "scene") && !strings.EqualFold(purpose, "evidence") && !strings.EqualFold(purpose, "decorative") {
			continue
		}
		if strings.TrimSpace(component.LocalPath) != "" || strings.TrimSpace(component.AssetQuery) != "" || strings.TrimSpace(component.AssetID) != "" {
			return true
		}
	}
	return false
}

func imageTextNarrativeChars(components []PlanComponent) int {
	maxChars := 0
	for _, component := range components {
		switch strings.TrimSpace(component.Type) {
		case "argument_block", "paragraph", "text_block":
			maxChars = max(maxChars, runeLen(firstNonEmptyString(component.Body, component.Text)))
		case "list", "numbered_list", "bullet_list", "evidence_list":
			maxChars = max(maxChars, runeLen(strings.Join(component.Items, "")))
		}
	}
	return maxChars
}

func requiresInformationDensity(contentType string) bool {
	switch strings.TrimSpace(contentType) {
	case "content_slide", "card_grid", "comparison_table", "swot_analysis":
		return true
	default:
		return false
	}
}

func informationNarrativeChars(components []PlanComponent) int {
	total := 0
	for _, component := range components {
		if isExternalBackgroundComponent(&component) {
			continue
		}
		switch strings.TrimSpace(component.Type) {
		case "headline", "deck_title", "subheadline", "section_marker", "source_note", "image", "map", "diagram", "divider", "shape", "arrow", "icon", "tag":
			continue
		case "list", "numbered_list", "bullet_list", "evidence_list":
			for _, item := range component.Items {
				total += runeLen(item)
			}
			if len(component.Items) == 0 {
				total += runeLen(firstNonEmptyString(component.Body, component.Text))
			}
		default:
			componentChars := runeLen(firstNonEmptyString(component.Body, component.Text))
			if componentChars == 0 && len(component.Items) > 0 {
				for _, item := range component.Items {
					componentChars += runeLen(item)
				}
			}
			total += componentChars
		}
	}
	return total
}

func reviewComponentContentQuality(contentType string, page int, component *PlanComponent, report *PlanReviewReport) {
	if component == nil || report == nil || isExternalBackgroundComponent(component) {
		return
	}
	componentType := strings.TrimSpace(component.Type)
	if isStructuralComponentType(componentType) {
		return
	}
	body := firstNonEmptyString(component.Body, component.Text)
	if isPlaceholderLikeText(body) {
		report.Issues = append(report.Issues, PlanReviewIssue{
			Code:        "low_information_density",
			Severity:    "error",
			PageIndex:   page,
			ComponentID: component.ID,
			Message:     fmt.Sprintf("组件 %s 仍是占位/大纲文本 %q；必须改成包含具体对象、事实、时间、数字、影响或结论的可上屏正文。", component.ID, body),
		})
	}
	for _, item := range component.Items {
		if isPlaceholderLikeText(item) {
			report.Issues = append(report.Issues, PlanReviewIssue{
				Code:        "low_information_density",
				Severity:    "error",
				PageIndex:   page,
				ComponentID: component.ID,
				Message:     fmt.Sprintf("组件 %s 的列表项 %q 仍是占位/大纲文本；必须补成完整事实、解释或判断。", component.ID, item),
			})
		}
	}
	if requiresInformationDensity(contentType) && requiresComponentBody(componentType) {
		bodyChars := runeLen(body)
		itemChars := 0
		for _, item := range component.Items {
			itemChars += runeLen(item)
		}
		if bodyChars == 0 && itemChars == 0 {
			report.Issues = append(report.Issues, PlanReviewIssue{
				Code:        "low_information_density",
				Severity:    "error",
				PageIndex:   page,
				ComponentID: component.ID,
				Message:     fmt.Sprintf("组件 %s 缺少可上屏正文；信息页不能只保留标题、栏目名或空卡片。", component.ID),
			})
		}
		if bodyChars > 0 && bodyChars < componentBodyMinChars {
			report.Issues = append(report.Issues, PlanReviewIssue{
				Code:        "low_information_density",
				Severity:    "error",
				PageIndex:   page,
				ComponentID: component.ID,
				Message:     fmt.Sprintf("组件 %s 正文约 %d 字，少于 %d 字；不能只写栏目名或大纲短语。", component.ID, bodyChars, componentBodyMinChars),
			})
		}
		for _, item := range component.Items {
			if chars := runeLen(item); chars > 0 && chars < componentBodyMinChars {
				report.Issues = append(report.Issues, PlanReviewIssue{
					Code:        "low_information_density",
					Severity:    "error",
					PageIndex:   page,
					ComponentID: component.ID,
					Message:     fmt.Sprintf("组件 %s 的列表项约 %d 字，少于 %d 字；需要写成完整要点而不是大纲词。", component.ID, chars, componentBodyMinChars),
				})
			}
		}
	}
}

func isStructuralComponentType(componentType string) bool {
	switch strings.TrimSpace(componentType) {
	case "headline", "deck_title", "subheadline", "section_marker", "source_note", "image", "map", "diagram", "divider", "shape", "arrow", "icon", "tag":
		return true
	default:
		return false
	}
}

func requiresComponentBody(componentType string) bool {
	switch strings.TrimSpace(componentType) {
	case "feature_card", "fact_card", "key_point", "insight", "recommendation", "risk_item", "opportunity_item", "decision_item", "case_snapshot", "paragraph", "text_block", "list", "numbered_list", "bullet_list", "evidence_list":
		return true
	default:
		return false
	}
}

func isPlaceholderLikeText(text string) bool {
	value := strings.TrimSpace(text)
	if value == "" {
		return false
	}
	lower := strings.ToLower(value)
	for _, marker := range []string{"todo", "tbd", "xxx", "占位", "待补充", "待填写", "待定", "{", "}", "示例文本", "这里填写", "若干数据", "某公司"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	compact := strings.NewReplacer(" ", "", "　", "", "：", "", ":", "", "-", "", "_", "").Replace(value)
	if runeLen(compact) <= 8 {
		generic := map[string]bool{
			"左栏成就": true, "左栏成就一": true, "左栏成就二": true, "左栏成就三": true,
			"右栏短板": true, "右栏短板一": true, "右栏短板二": true, "右栏短板三": true,
			"左栏事实": true, "右栏事实": true, "左栏风险": true, "右栏风险": true,
			"卡片一": true, "卡片二": true, "卡片三": true, "卡片四": true,
			"要点一": true, "要点二": true, "要点三": true, "要点四": true,
			"观点一": true, "观点二": true, "观点三": true, "观点四": true,
			"内容一": true, "内容二": true, "内容三": true, "内容四": true,
		}
		if generic[compact] {
			return true
		}
	}
	switch compact {
	case "对比后的判断", "未来展望洞察", "本页洞察", "核心洞察", "核心观点", "关键结论", "补充说明":
		return true
	default:
		return false
	}
}

func hasPlanNarrativeSummary(plan *ContentPlan) bool {
	if plan == nil {
		return false
	}
	for _, value := range []string{plan.Summary, plan.SlideIntent} {
		if runeLen(value) >= 8 {
			return true
		}
	}
	for _, component := range plan.Components {
		if runeLen(firstNonEmptyString(component.Body, component.Text)) >= 16 {
			return true
		}
	}
	return false
}

func hasUsableBackgroundPlan(task *TaskItem) bool {
	if task == nil {
		return false
	}
	if task.ContentPlan == nil {
		return false
	}
	if task.ContentPlan.VisualIntent != nil && strings.EqualFold(strings.TrimSpace(task.ContentPlan.VisualIntent.Role), "clean_text_only") {
		return true
	}
	if isExternalBackgroundIntent(task.ContentPlan.VisualIntent) {
		return strings.TrimSpace(task.ContentPlan.VisualIntent.LocalPath) != "" ||
			strings.TrimSpace(task.ContentPlan.VisualIntent.AssetQuery) != ""
	}
	for i := range task.ContentPlan.Components {
		component := &task.ContentPlan.Components[i]
		if isExternalBackgroundComponent(component) {
			return strings.TrimSpace(component.LocalPath) != "" || strings.TrimSpace(component.AssetQuery) != ""
		}
	}
	return false
}

func isExternalBackgroundIntent(intent *VisualIntent) bool {
	if intent == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(intent.AssetPurpose), "background") ||
		strings.EqualFold(strings.TrimSpace(intent.ImagePosition), "background")
}

func isExternalBackgroundComponent(component *PlanComponent) bool {
	if component == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(component.AssetPurpose), "background") ||
		strings.EqualFold(strings.TrimSpace(component.Role), "background") ||
		strings.EqualFold(strings.TrimSpace(component.ImagePosition()), "background")
}

func (c *PlanComponent) ImagePosition() string {
	if c == nil || c.Data == nil {
		return ""
	}
	if value, ok := c.Data["image_position"].(string); ok {
		return value
	}
	return ""
}

func runeLen(value string) int {
	return len([]rune(strings.TrimSpace(value)))
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
