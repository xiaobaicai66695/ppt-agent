package deck

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const planReviewFileName = "tasks.review.json"

var planReviewToolInfo = &schema.ToolInfo{
	Name: "review_tasks_manifest",
	Desc: "读取当前 DeckSpec 草稿并执行确定性规划审核。审核覆盖结构合法性、内容密度、组件容量、背景图片规划和图片本地化。通过后会锁定每页 reviewer_status 并写入 review artifact；草稿变更后必须重新审核再 commit。",
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"round": {
			Type: "integer",
			Desc: "当前 Reflexion 审核轮次，首次审核为 1，修订后递增",
		},
		"focus": {
			Type: "string",
			Desc: "可选审核重点，例如 background、density、schema；留空则执行完整审核",
		},
	}),
}

type planReviewTool struct {
	workDir string
}

type planReviewInput struct {
	Round int    `json:"round,omitempty"`
	Focus string `json:"focus,omitempty"`
}

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

func newPlanReviewTool(workDir string) tool.InvokableTool {
	return &planReviewTool{workDir: workDir}
}

func (t *planReviewTool) Info(context.Context) (*schema.ToolInfo, error) {
	return planReviewToolInfo, nil
}

func (t *planReviewTool) InvokableRun(_ context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	input := planReviewInput{}
	if strings.TrimSpace(argumentsInJSON) != "" {
		if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
			return "", fmt.Errorf("invalid plan review input: %w", err)
		}
	}
	report, err := ReviewTasksDraftManifest(t.workDir, input.Round)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(report)
	if err != nil {
		return "", fmt.Errorf("marshal plan review report: %w", err)
	}
	return string(data), nil
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
		round = inferNextReviewRound(manifest)
	}
	report := ReviewTasksManifest(manifest, target, round)
	if target == "tasks.draft.json" {
		markManifestReviewStatus(manifest, report)
		if err := WriteTasksDraftManifest(workDir, manifest); err != nil {
			return nil, fmt.Errorf("write reviewed draft manifest: %w", err)
		}
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
	if strings.TrimSpace(manifest.Theme) == "" {
		report.Issues = append(report.Issues, PlanReviewIssue{Code: "layout_mismatch", Severity: "error", Message: "缺少整套 theme，生成器无法稳定选择配色。"})
	}
	if err := validateManifestForWrite(manifest); err != nil {
		report.Issues = append(report.Issues, PlanReviewIssue{
			Code:     "invalid_component_schema",
			Severity: "error",
			Message:  err.Error(),
		})
	}
	for _, task := range manifest.Tasks {
		reviewTaskPlan(task, report)
	}
	return finalizePlanReviewReport(report, manifest)
}

func reviewTaskPlan(task *TaskItem, report *PlanReviewReport) {
	if task == nil {
		report.Issues = append(report.Issues, PlanReviewIssue{Code: "invalid_component_schema", Severity: "error", Message: "存在空页面任务。"})
		return
	}
	page := task.PageIndex
	if (strings.TrimSpace(task.Description) == "" || runeLen(task.Description) < 8) && !isSimpleTitleLikeSlide(task.ContentType) {
		report.Issues = append(report.Issues, PlanReviewIssue{
			Code:      "low_information_density",
			Severity:  "error",
			PageIndex: page,
			Message:   "页面 description 过短，不能支撑稳定生成。",
		})
	}
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
	if strings.TrimSpace(plan.Summary) == "" {
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
	if !hasUsableBackgroundPlan(task) {
		report.Issues = append(report.Issues, PlanReviewIssue{
			Code:      "missing_background_image",
			Severity:  "error",
			PageIndex: page,
			Message:   "页面缺少可执行图片计划：应在 visual_intent 或 image 组件中使用 asset_purpose=background，并写入已下载 local_path。",
		})
	} else {
		report.BackgroundPages++
	}
	for i := range plan.Components {
		component := &plan.Components[i]
		if isExternalBackgroundComponent(component) && strings.TrimSpace(component.LocalPath) == "" {
			report.Issues = append(report.Issues, PlanReviewIssue{
				Code:        "missing_background_image",
				Severity:    "error",
				PageIndex:   page,
				ComponentID: component.ID,
				Message:     "背景图片组件缺少 local_path；需要调用 search_images(download=true) 并写回下载路径。",
			})
		}
		if component.Type == "argument_block" && runeLen(firstNonEmptyString(component.Body, component.Text, component.Description)) < 220 {
			report.Issues = append(report.Issues, PlanReviewIssue{
				Code:        "low_information_density",
				Severity:    "warning",
				PageIndex:   page,
				ComponentID: component.ID,
				Message:     "argument_block 字数偏少，无法承担大段论述页的深度表达。",
			})
		}
	}
}

func finalizePlanReviewReport(report *PlanReviewReport, manifest *TasksManifest) *PlanReviewReport {
	report.IssueCount = len(report.Issues)
	report.Passed = !hasBlockingPlanReviewIssue(report.Issues)
	if report.Passed {
		report.Summary = fmt.Sprintf("规划审核通过：%d 页均满足结构、组件容量和背景图质量门。", report.TotalSlides)
		report.NextActions = []string{"调用 update_tasks_manifest(mode=\"commit\") 发布正式 tasks.json"}
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

func markManifestReviewStatus(manifest *TasksManifest, report *PlanReviewReport) {
	if manifest == nil || report == nil {
		return
	}
	issuesByPage := map[int][]PlanReviewIssue{}
	for _, issue := range report.Issues {
		if issue.PageIndex > 0 {
			issuesByPage[issue.PageIndex] = append(issuesByPage[issue.PageIndex], issue)
		}
	}
	for _, task := range manifest.Tasks {
		if task == nil {
			continue
		}
		if task.ContentPlan == nil {
			task.ContentPlan = &ContentPlan{}
		}
		task.ContentPlan.ReviewerStatus = &PlanReviewerStatus{
			PlannerRound: report.Round,
			Locked:       report.Passed,
			LockedAt:     "",
			Issues:       issuesByPage[task.PageIndex],
		}
		if report.Passed {
			task.ContentPlan.ReviewerStatus.LockedAt = report.ReviewedAt
			task.ContentPlan.ReviewerStatus.Issues = nil
		}
	}
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
		actions = append(actions, "修订草稿后重新调用 review_tasks_manifest")
	}
	return actions
}

func planReviewActionForCode(code string) string {
	switch strings.TrimSpace(code) {
	case "missing_background_image":
		return "为缺失页面补充背景图：优先 search_images(download=true)，并把 local_path 写回 visual_intent 或 image 组件"
	case "low_information_density":
		return "补足页面事实、解释、案例和结论，避免只有泛化短句"
	case "weak_narrative":
		return "补充页面观点锚点和 slide_intent，形成观点-论据结构"
	case "invalid_component_schema":
		return "按 component_contracts.json 修正 content_type、components 和容量字段"
	case "layout_mismatch":
		return "补齐 theme/template/layout_variant，并确保页面类型和组件匹配"
	default:
		return "修订草稿后重新调用 review_tasks_manifest"
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
			return fmt.Errorf("draft DeckSpec has not been reviewed; call review_tasks_manifest before commit")
		}
		return fmt.Errorf("read plan review report: %w", err)
	}
	if !report.Passed {
		return fmt.Errorf("latest DeckSpec review did not pass: %s", report.Summary)
	}
	fingerprint := fingerprintTasksManifest(manifest)
	if report.Fingerprint != fingerprint {
		return fmt.Errorf("DeckSpec changed after latest review; call review_tasks_manifest again before commit")
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

func inferNextReviewRound(manifest *TasksManifest) int {
	maxRound := 0
	if manifest != nil {
		for _, task := range manifest.Tasks {
			if task != nil && task.ContentPlan != nil && task.ContentPlan.ReviewerStatus != nil && task.ContentPlan.ReviewerStatus.PlannerRound > maxRound {
				maxRound = task.ContentPlan.ReviewerStatus.PlannerRound
			}
		}
	}
	if maxRound >= 3 {
		return 3
	}
	return maxRound + 1
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

func hasUsableBackgroundPlan(task *TaskItem) bool {
	if task == nil {
		return false
	}
	if task.ContentPlan == nil {
		return false
	}
	if isExternalBackgroundIntent(task.ContentPlan.VisualIntent) {
		return strings.TrimSpace(task.ContentPlan.VisualIntent.LocalPath) != ""
	}
	for i := range task.ContentPlan.Components {
		component := &task.ContentPlan.Components[i]
		if isExternalBackgroundComponent(component) {
			return strings.TrimSpace(component.LocalPath) != ""
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
