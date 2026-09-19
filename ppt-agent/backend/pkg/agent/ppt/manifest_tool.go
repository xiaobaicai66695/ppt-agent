package ppt

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/ppt-agent/pkg/templates"
)

const tasksDraftFileName = "tasks.draft.json"

var manifestTaskPatchSchema = map[string]*schema.ParameterInfo{
	"page_index":     {Type: schema.Integer, Desc: "页码；initialize 和 patch 模式必填"},
	"section_id":     {Type: schema.String, Desc: "页面所属章节 ID，用于后续定点或整节修复"},
	"section_title":  {Type: schema.String, Desc: "页面所属章节标题"},
	"title":          {Type: schema.String, Desc: "页面标题"},
	"content_type":   {Type: schema.String, Desc: "合法的单页模板英文 ID"},
	"layout_variant": {Type: schema.String, Desc: "同一 content_type 下的版式变体 ID，可为空"},
	"page_intent":    {Type: schema.String, Desc: "该页在章节中的具体叙事任务"},
	"evidence_refs":  {Type: schema.Array, Desc: "引用顶层 content_bank 的事实/主题/实体 ID", ElemInfo: &schema.ParameterInfo{Type: schema.String}},
	"content_plan": {
		Type: schema.Object,
		Desc: "结构化页面内容规划",
		SubParams: map[string]*schema.ParameterInfo{
			"summary":        {Type: schema.String, Desc: "页面核心摘要"},
			"slide_intent":   {Type: schema.String, Desc: "本页在整套 PPT 中承担的语义目标"},
			"section_number": {Type: schema.String, Desc: "章节分隔页的大章节编号，如 01/02；仅 section_divider 使用，不得写页码"},
			"evidence_refs":  {Type: schema.Array, Desc: "引用顶层 content_bank 的事实/主题/实体 ID", ElemInfo: &schema.ParameterInfo{Type: schema.String}},
			"visual_intent": {
				Type: schema.Object,
				Desc: "页面视觉意图，用于规划图片、图表、地图、卡片等视觉角色",
				SubParams: map[string]*schema.ParameterInfo{
					"role":              {Type: schema.String, Desc: "hero_photo、supporting_photo、map、chart、icon、cards；用户明确要求无图时使用 clean_text_only"},
					"asset_purpose":     {Type: schema.String, Desc: "图片用途：background（整页氛围）/scene（真实场景）/evidence（事实或案例证据）/decorative（装饰）"},
					"asset_subject":     {Type: schema.String, Desc: "经过语义转换后的视觉主体或代理意象，不直接复制用户标题"},
					"asset_query":       {Type: schema.String, Desc: "经过视觉转换、可直接提交给图片 provider 的最终检索词"},
					"composition":       {Type: schema.String, Desc: "构图约束，如 wide landscape、clean negative space on left、subject on right"},
					"preferred_variant": {Type: schema.String, Desc: "偏好的 layout_variant"},
					"image_position":    {Type: schema.String, Desc: "background、left、right、strip、inline 等"},
					"caption":           {Type: schema.String, Desc: "图片说明或替代文本"},
					"local_path":        {Type: schema.String, Desc: "确定性图片物化阶段回填的本地图片路径，位于当前任务工作目录内"},
					"source_url":        {Type: schema.String, Desc: "图片来源页 URL，用于署名"},
					"attribution":       {Type: schema.String, Desc: "图片署名，例如 Photo by ... on Unsplash"},
				},
			},
			"components": {
				Type: schema.Array,
				Desc: "页内语义组件计划；每个组件只填写 component_contracts.json 中对应类型需要的字段，禁止写坐标、字号、颜色、边距等渲染参数",
				ElemInfo: &schema.ParameterInfo{Type: schema.Object, SubParams: map[string]*schema.ParameterInfo{
					"id":   {Type: schema.String, Desc: "组件稳定 ID，页内唯一"},
					"type": {Type: schema.String, Desc: "合法组件类型；其余字段按该类型契约填写"},
				}},
			},
		},
	},
}

var visualPolicySchema = map[string]*schema.ParameterInfo{
	"mode": {
		Type: schema.String,
		Desc: "视觉素材策略：required/optional/none；常规生成使用 required。required 时后端会要求所有页使用 background。",
	},
	"min_image_pages": {
		Type: schema.Integer,
		Desc: "required 模式下的最小图片页数；留空时后端按全部页面数补齐。",
	},
	"required_roles": {
		Type:     schema.Array,
		Desc:     "required 模式下必须包含 background；留空时后端补齐。",
		ElemInfo: &schema.ParameterInfo{Type: schema.String},
	},
	"user_declined_background": {
		Type: schema.Boolean,
		Desc: "仅 mode=none 且用户明确拒绝背景图时为 true。",
	},
	"decline_reason": {
		Type: schema.String,
		Desc: "仅 mode=none 时记录用户明确拒绝背景图的原因。",
	},
}

var plannerManifestToolInfo = &schema.ToolInfo{
	Name: "update_tasks_manifest",
	Desc: "一次性初始化完整 PPTSpec 规划草稿。Planner 只能使用 initialize，审查、修订和提交由后续阶段负责。",
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"mode":          {Type: schema.String, Required: true, Desc: "固定填写 initialize"},
		"title":         {Type: schema.String, Desc: "PPT 标题；缺省时系统会从任务上下文推断"},
		"visual_policy": {Type: schema.Object, Desc: "整套视觉素材策略；mode 为 required/optional/none，常规生成保持 required", SubParams: visualPolicySchema},
		"tasks": {
			Type:     schema.Array,
			Required: true,
			Desc:     "完整页面数组；每页只填 page_index/title/content_type/content_plan 等语义字段，运行字段由后端补齐",
			ElemInfo: &schema.ParameterInfo{Type: schema.Object, SubParams: manifestTaskPatchSchema},
		},
	}),
}

type manifestTaskPatch struct {
	PageIndex     *int         `json:"page_index,omitempty"`
	SectionID     *string      `json:"section_id,omitempty"`
	SectionTitle  *string      `json:"section_title,omitempty"`
	Title         *string      `json:"title,omitempty"`
	ContentType   *string      `json:"content_type,omitempty"`
	LayoutVariant *string      `json:"layout_variant,omitempty"`
	PageIntent    *string      `json:"page_intent,omitempty"`
	EvidenceRefs  []string     `json:"evidence_refs,omitempty"`
	ContentPlan   *ContentPlan `json:"content_plan,omitempty"`
}

type manifestToolInput struct {
	Mode         string              `json:"mode"`
	Title        string              `json:"title,omitempty"`
	ContentBank  map[string]any      `json:"content_bank,omitempty"`
	Sections     []PPTSection        `json:"sections,omitempty"`
	VisualPolicy *VisualPolicy       `json:"visual_policy,omitempty"`
	Tasks        []manifestTaskPatch `json:"tasks"`
}

type manifestToolRawInput struct {
	Mode         string          `json:"mode"`
	Title        string          `json:"title,omitempty"`
	ContentBank  map[string]any  `json:"content_bank,omitempty"`
	Sections     []PPTSection    `json:"sections,omitempty"`
	VisualPolicy *VisualPolicy   `json:"visual_policy,omitempty"`
	Tasks        json.RawMessage `json:"tasks"`
}

type manifestTool struct {
	workDir       string
	fallbackTitle string
	draftFirst    bool
	contract      *templates.ContractSummary
}

type plannerManifestTool struct{ inner *manifestTool }

func newPlannerManifestTool(workDir string, outline *TaskOutline, query string, skillsDirs ...string) tool.InvokableTool {
	contract := templates.BuiltInContractSummary()
	if len(skillsDirs) > 0 {
		contract = contractForSkills(skillsDirs[0])
	}
	return &plannerManifestTool{inner: newDraftManifestToolWithContract(workDir, outline, query, contract)}
}

func newDraftManifestTool(workDir string, outline *TaskOutline, query string) *manifestTool {
	return newDraftManifestToolWithContract(workDir, outline, query, templates.BuiltInContractSummary())
}

func newDraftManifestToolWithContract(workDir string, outline *TaskOutline, query string, contract *templates.ContractSummary) *manifestTool {
	inner := &manifestTool{
		workDir:       workDir,
		fallbackTitle: compactManifestTitle(query),
		draftFirst:    true,
		contract:      contract,
	}
	if outline != nil && strings.TrimSpace(outline.Title) != "" {
		inner.fallbackTitle = compactManifestTitle(outline.Title)
	}
	return inner
}

func (t *plannerManifestTool) Info(context.Context) (*schema.ToolInfo, error) {
	info := *plannerManifestToolInfo
	if t != nil && t.inner != nil && t.inner.contract != nil {
		info.Desc += "\n当前运行时容量契约：\n" + t.inner.contract.PromptText()
	}
	return &info, nil
}

func (t *plannerManifestTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	input, err := parseManifestToolInput(argumentsInJSON)
	if err != nil {
		return manifestToolRecoverableError("initialize", "draft", err), nil
	}
	if !strings.EqualFold(strings.TrimSpace(input.Mode), "initialize") {
		payload, _ := json.Marshal(map[string]any{
			"ok":          false,
			"mode":        input.Mode,
			"target":      "draft",
			"error":       "PPTPlanner only supports initialize",
			"next_action": "一次性提交完整页面数组并使用 mode=initialize",
		})
		return string(payload), nil
	}
	manifest, err := t.inner.initializeManifest(input)
	if err != nil {
		return manifestToolRecoverableError(input.Mode, "draft", err), nil
	}
	normalizeManifestLayoutVariants(manifest)
	normalizePlannerInitialManifestWithContract(manifest, t.inner.contract)
	if issues := plannerPreflightIssuesWithContract(manifest, t.inner.contract); len(issues) > 0 {
		if err := t.inner.writeManifest(manifest); err != nil {
			return "", err
		}
		payload, _ := json.Marshal(map[string]any{
			"ok":                  true,
			"mode":                "initialize",
			"target":              "draft",
			"quality_gate_passed": false,
			"issue_count":         len(issues),
			"issues":              issues,
			"capacity_contract":   contractMetadata(t.inner.contract),
			"next_action":         "草稿已写入 tasks.draft.json；停止全量重写，等待 Task Reviewer 按失败页/章节切片 patch。",
		})
		return string(payload), nil
	}
	if err := t.inner.writeManifest(manifest); err != nil {
		return "", err
	}
	result, _ := json.Marshal(map[string]any{
		"ok": true, "mode": "initialize", "target": "draft", "updated": len(manifest.Tasks), "total": len(manifest.Tasks),
		"capacity_contract": contractMetadata(t.inner.contract),
	})
	return string(result), nil
}

func (t *manifestTool) Info(context.Context) (*schema.ToolInfo, error) {
	info := *plannerManifestToolInfo
	if t != nil && t.contract != nil {
		info.Desc += "\n当前运行时容量契约：\n" + t.contract.PromptText()
	}
	return &info, nil
}

func (t *manifestTool) InvokableRun(_ context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	input, err := parseManifestToolInput(argumentsInJSON)
	if err != nil {
		return manifestToolRecoverableError("", "draft", err), nil
	}

	var manifest *TasksManifest
	target := "final"
	switch strings.ToLower(strings.TrimSpace(input.Mode)) {
	case "initialize":
		if len(input.Tasks) == 0 {
			return manifestToolRecoverableError(input.Mode, t.writeTarget(), fmt.Errorf("tasks must not be empty")), nil
		}
		manifest, err = t.initializeManifest(input)
		target = t.writeTarget()
	case "patch":
		if len(input.Tasks) == 0 && !hasManifestHeaderPatch(input) {
			return manifestToolRecoverableError(input.Mode, t.writeTarget(), fmt.Errorf("tasks must not be empty unless title/content_bank/sections is patched")), nil
		}
		manifest, err = t.patchManifest(input)
		target = t.writeTarget()
	default:
		err = fmt.Errorf("unsupported mode %q", input.Mode)
	}
	if err != nil {
		return manifestToolRecoverableError(input.Mode, target, err), nil
	}
	normalizeManifestLayoutVariants(manifest)
	if !t.shouldDeferValidation() {
		if err := validateManifestForWriteWithContract(manifest, t.contract); err != nil {
			return manifestToolRecoverableErrorWithManifest(input.Mode, target, err, manifest, t.contract), nil
		}
	}
	if err := t.writeManifest(manifest); err != nil {
		return "", err
	}
	result, _ := json.Marshal(map[string]any{
		"ok": true, "mode": input.Mode, "target": target, "updated": len(input.Tasks), "total": len(manifest.Tasks),
	})
	return string(result), nil
}

func hasManifestHeaderPatch(input manifestToolInput) bool {
	return strings.TrimSpace(input.Title) != "" ||
		input.ContentBank != nil ||
		input.Sections != nil ||
		input.VisualPolicy != nil
}

func manifestToolRecoverableError(mode, target string, err error) string {
	payload := map[string]any{
		"ok":          false,
		"mode":        strings.TrimSpace(mode),
		"target":      target,
		"error":       err.Error(),
		"next_action": "修正 update_tasks_manifest 的结构化 JSON 参数后重新调用；tasks 必须是完整对象数组，字符串形式必须是合法 JSON 数组且不能截断、混入未转义引号或非 JSON 字符。",
	}
	if payload["mode"] == "" {
		payload["mode"] = "unknown"
	}
	if payload["target"] == "" {
		payload["target"] = "draft"
	}
	result, _ := json.Marshal(payload)
	return string(result)
}

func manifestToolRecoverableErrorWithManifest(mode, target string, err error, manifest *TasksManifest, contract *templates.ContractSummary) string {
	var payload map[string]any
	_ = json.Unmarshal([]byte(manifestToolRecoverableError(mode, target, err)), &payload)
	var issues []PlanReviewIssue
	if manifest != nil {
		for _, task := range manifest.Tasks {
			if issue := componentCapacityIssue(task, contract); issue != nil {
				issues = append(issues, *issue)
			}
		}
	}
	if len(issues) > 0 {
		payload["issues"] = issues
		payload["capacity_contract"] = contractMetadata(contract)
	}
	result, _ := json.Marshal(payload)
	return string(result)
}

func (t *manifestTool) writeTarget() string {
	if t != nil && t.draftFirst {
		return "draft"
	}
	return "final"
}

func (t *manifestTool) shouldDeferValidation() bool {
	return t != nil && t.draftFirst
}

func (t *manifestTool) writeManifest(manifest *TasksManifest) error {
	if t != nil && t.draftFirst {
		return WriteTasksDraftManifest(t.workDir, manifest)
	}
	return WriteTasksManifest(t.workDir, manifest)
}

func normalizeManifestLayoutVariants(manifest *TasksManifest) {
	if manifest == nil {
		return
	}
	counts := map[string]int{}
	sectionCount := 0
	sectionVariant := preferredSectionDividerVariant(manifest)
	for _, item := range manifest.Tasks {
		if item == nil {
			continue
		}
		contentType := strings.TrimSpace(item.ContentType)
		if contentType == "" {
			continue
		}
		if contentType == "section_divider" {
			sectionCount++
			ensureSectionNumber(item, sectionCount)
			item.LayoutVariant = sectionVariant
			setPreferredVariant(item, sectionVariant)
			continue
		}
		if strings.TrimSpace(item.LayoutVariant) != "" {
			continue
		}
		if variants := supportedLayoutVariants(contentType); len(variants) > 0 {
			index := counts[contentType] % len(variants)
			item.LayoutVariant = variants[index]
			ensurePreferredVariant(item, item.LayoutVariant)
			counts[contentType]++
		}
	}
}

func normalizePlannerInitialManifest(manifest *TasksManifest) {
	normalizePlannerInitialManifestWithContract(manifest, templates.BuiltInContractSummary())
}

func normalizePlannerInitialManifestWithContract(manifest *TasksManifest, contract *templates.ContractSummary) {
	if manifest == nil {
		return
	}
	if contract == nil {
		contract = templates.BuiltInContractSummary()
	}
	normalizeVisualPolicy(manifest)
	for _, item := range manifest.Tasks {
		normalizePlannerInitialTaskWithContract(item, contract)
	}
}

// normalizeVisualPolicy keeps the Go-side plan contract aligned with the
// Python visual validator. Planner output may omit derived coverage fields,
// but a required policy cannot reach rendering without them.
func normalizeVisualPolicy(manifest *TasksManifest) {
	if manifest == nil {
		return
	}
	if manifest.VisualPolicy == nil {
		manifest.VisualPolicy = &VisualPolicy{}
	}
	policy := manifest.VisualPolicy
	policy.Mode = strings.ToLower(strings.TrimSpace(policy.Mode))
	if policy.Mode == "" {
		policy.Mode = "required"
	}
	if policy.Mode != "required" {
		return
	}
	roles := make([]string, 0, len(policy.RequiredRoles)+1)
	seen := make(map[string]struct{}, len(policy.RequiredRoles)+1)
	for _, role := range policy.RequiredRoles {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}
		roles = append(roles, role)
	}
	if _, ok := seen["background"]; !ok {
		roles = append(roles, "background")
	}
	policy.RequiredRoles = roles
	if policy.MinImagePages < len(manifest.Tasks) {
		policy.MinImagePages = len(manifest.Tasks)
	}
}

func normalizePlannerInitialTask(item *TaskItem) {
	normalizePlannerInitialTaskWithContract(item, templates.BuiltInContractSummary())
}

func normalizePlannerInitialTaskWithContract(item *TaskItem, contract *templates.ContractSummary) {
	if item == nil || item.ContentPlan == nil {
		return
	}
	if contract == nil {
		contract = templates.BuiltInContractSummary()
	}
	plan := item.ContentPlan
	limit := maxComponentsForContentTypeWithContract(item.ContentType, contract)
	if strings.TrimSpace(item.ContentType) == "agenda" {
		plan.Components = compactAgendaComponents(item, plan.Components, limit)
	}
	if strings.TrimSpace(item.ContentType) == "kpi_dashboard" && len(plan.Components) > 0 && !hasNarrativeAnchor(plan.Components) {
		plan.Components = appendInsightWithinLimit(plan.Components, limit, "insight_auto", "关键判断", summaryForAutoInsight(item))
	}
	for componentIndex := range plan.Components {
		component := &plan.Components[componentIndex]
		if strings.TrimSpace(component.Type) == "argument_block" &&
			runeLen(firstNonEmptyString(component.Body, component.Text)) < argumentBlockTargetMinChars {
			component.Type = "insight"
		}
	}
}

func compactAgendaComponents(item *TaskItem, components []PlanComponent, limit int) []PlanComponent {
	if limit <= 0 || len(components) <= limit {
		return components
	}
	anchor := firstNarrativeAnchorComponent(components)
	if strings.TrimSpace(anchor.ID) == "" {
		anchor = PlanComponent{
			ID:    "agenda_insight",
			Type:  "insight",
			Title: "阅读路径",
			Body:  summaryForAutoInsight(item),
		}
	}
	result := []PlanComponent{anchor}
	var overflow []string
	for _, component := range components {
		if sameComponentIdentity(component, anchor) {
			continue
		}
		if len(result) < limit {
			result = append(result, component)
			continue
		}
		label := firstNonEmptyString(component.Title, component.Text, component.Body)
		if strings.TrimSpace(label) != "" {
			overflow = append(overflow, compactManifestTitle(label))
		}
	}
	if len(overflow) > 0 && len(result) > 0 {
		last := &result[len(result)-1]
		last.Title = firstNonEmptyString(last.Title, "后续专题")
		last.Body = strings.TrimSpace(firstNonEmptyString(last.Body, last.Text) + "；" + strings.Join(overflow, "；"))
	}
	return result
}

func firstNarrativeAnchorComponent(components []PlanComponent) PlanComponent {
	for _, component := range components {
		switch strings.TrimSpace(component.Type) {
		case "headline", "ppt_title", "argument_block", "insight", "recommendation", "key_point", "quote_block":
			return component
		}
	}
	return PlanComponent{}
}

func appendInsightWithinLimit(components []PlanComponent, limit int, id, title, body string) []PlanComponent {
	insight := PlanComponent{ID: id, Type: "insight", Title: title, Body: body}
	if limit <= 0 || len(components) < limit {
		return append(components, insight)
	}
	if len(components) == 0 {
		return []PlanComponent{insight}
	}
	components[len(components)-1] = insight
	return components
}

func sameComponentIdentity(a, b PlanComponent) bool {
	if strings.TrimSpace(a.ID) != "" && strings.TrimSpace(a.ID) == strings.TrimSpace(b.ID) {
		return true
	}
	return strings.TrimSpace(a.Type) == strings.TrimSpace(b.Type) &&
		strings.TrimSpace(a.Title) == strings.TrimSpace(b.Title) &&
		strings.TrimSpace(a.Body) == strings.TrimSpace(b.Body) &&
		strings.TrimSpace(a.Text) == strings.TrimSpace(b.Text)
}

func summaryForAutoInsight(item *TaskItem) string {
	if item == nil {
		return "本页用于建立阅读顺序和核心判断，帮助观众理解后续内容之间的关系。"
	}
	if item.ContentPlan != nil {
		if text := strings.TrimSpace(item.ContentPlan.Summary); text != "" {
			return text
		}
		if text := strings.TrimSpace(item.ContentPlan.SlideIntent); text != "" {
			return text
		}
	}
	return "本页用于建立阅读顺序和核心判断，帮助观众理解后续内容之间的关系。"
}

func preferredSectionDividerVariant(manifest *TasksManifest) string {
	variants := supportedLayoutVariants("section_divider")
	if len(variants) == 0 {
		return ""
	}
	allowed := make(map[string]bool, len(variants))
	for _, variant := range variants {
		allowed[variant] = true
	}
	for _, item := range manifest.Tasks {
		if item == nil || strings.TrimSpace(item.ContentType) != "section_divider" {
			continue
		}
		variant := strings.TrimSpace(item.LayoutVariant)
		if allowed[variant] {
			return variant
		}
		if item.ContentPlan != nil && item.ContentPlan.VisualIntent != nil {
			variant = strings.TrimSpace(item.ContentPlan.VisualIntent.PreferredVariant)
			if allowed[variant] {
				return variant
			}
		}
	}
	return "number_sidebar"
}

func supportedLayoutVariants(contentType string) []string {
	switch strings.TrimSpace(contentType) {
	case "section_divider":
		return []string{"number_sidebar"}
	case "image_text":
		return []string{"image_left", "image_right", "image_top_band", "image_bottom_band"}
	default:
		return nil
	}
}

func ensurePreferredVariant(item *TaskItem, variant string) {
	if item == nil || strings.TrimSpace(variant) == "" || item.ContentPlan == nil || item.ContentPlan.VisualIntent == nil {
		return
	}
	if strings.TrimSpace(item.ContentPlan.VisualIntent.PreferredVariant) == "" {
		item.ContentPlan.VisualIntent.PreferredVariant = variant
	}
}

func setPreferredVariant(item *TaskItem, variant string) {
	if item == nil || strings.TrimSpace(variant) == "" || item.ContentPlan == nil || item.ContentPlan.VisualIntent == nil {
		return
	}
	item.ContentPlan.VisualIntent.PreferredVariant = variant
}

func ensureSectionNumber(item *TaskItem, sectionCount int) {
	if item == nil || sectionCount <= 0 {
		return
	}
	if item.ContentPlan == nil {
		item.ContentPlan = &ContentPlan{}
	}
	if strings.TrimSpace(item.ContentPlan.SectionNumber) == "" {
		item.ContentPlan.SectionNumber = fmt.Sprintf("%02d", sectionCount)
	}
}

func parseManifestToolInput(argumentsInJSON string) (manifestToolInput, error) {
	if err := rejectForbiddenManifestFields([]byte(argumentsInJSON)); err != nil {
		return manifestToolInput{}, err
	}
	var raw manifestToolRawInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &raw); err != nil {
		return manifestToolInput{}, err
	}
	if strings.TrimSpace(raw.Mode) == "" {
		return manifestToolInput{}, fmt.Errorf("mode must not be empty")
	}
	tasks, err := parseManifestTaskPatches(raw.Tasks)
	if err != nil {
		return manifestToolInput{}, err
	}
	input := manifestToolInput{
		Mode: raw.Mode, Title: raw.Title, ContentBank: raw.ContentBank,
		Sections: raw.Sections, VisualPolicy: raw.VisualPolicy, Tasks: tasks,
	}
	return input, nil
}

func rejectForbiddenManifestFields(raw []byte) error {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	root, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("tool arguments must be a JSON object")
	}
	for _, field := range []string{"theme", "template"} {
		if _, exists := root[field]; exists {
			return fmt.Errorf("%s is not part of the PPTSpec contract", field)
		}
	}
	if rawTasks, ok := root["tasks"]; ok {
		if tasksText, ok := rawTasks.(string); ok {
			var decoded any
			if err := json.Unmarshal([]byte(strings.TrimSpace(tasksText)), &decoded); err != nil {
				return fmt.Errorf("tasks JSON array string is invalid: %w", err)
			}
			rawTasks = decoded
		}
		tasks, ok := rawTasks.([]any)
		if !ok && rawTasks != nil {
			return fmt.Errorf("tasks must be an array")
		}
		for index, task := range tasks {
			if err := rejectForbiddenTaskFields(task, fmt.Sprintf("tasks[%d]", index)); err != nil {
				return err
			}
		}
	}
	return nil
}

func rejectForbiddenTaskFields(value any, path string) error {
	object, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("%s must be an object", path)
	}
	for _, field := range []string{"task_id", "output_file", "status", "description", "qa_report", "fix_attempts"} {
		if _, exists := object[field]; exists {
			return fmt.Errorf("%s.%s is not part of the LLM PPTSpec contract", path, field)
		}
	}
	if plan, ok := object["content_plan"]; ok {
		if err := rejectForbiddenContentPlanFields(plan, path+".content_plan"); err != nil {
			return err
		}
	}
	return nil
}

func rejectForbiddenContentPlanFields(value any, path string) error {
	object, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("%s must be an object", path)
	}
	for _, field := range []string{"capacity_hint", "reviewer_status", "description"} {
		if _, exists := object[field]; exists {
			return fmt.Errorf("%s.%s is not part of the PPTSpec contract", path, field)
		}
	}
	if components, ok := object["components"].([]any); ok {
		for i, component := range components {
			if err := rejectForbiddenComponentFields(component, fmt.Sprintf("%s.components[%d]", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func rejectForbiddenComponentFields(value any, path string) error {
	object, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("%s must be an object", path)
	}
	return rejectForbiddenFieldRecursive(object, "description", path)
}

func rejectForbiddenFieldRecursive(value any, field, path string) error {
	switch typed := value.(type) {
	case map[string]any:
		if _, exists := typed[field]; exists {
			return fmt.Errorf("%s.%s is not part of the component contract", path, field)
		}
		for key, child := range typed {
			if err := rejectForbiddenFieldRecursive(child, field, path+"."+key); err != nil {
				return err
			}
		}
	case []any:
		for i, child := range typed {
			if err := rejectForbiddenFieldRecursive(child, field, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func parseManifestTaskPatches(raw json.RawMessage) ([]manifestTaskPatch, error) {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "" || string(raw) == "null" {
		return nil, nil
	}
	var tasks []manifestTaskPatch
	if err := json.Unmarshal(raw, &tasks); err == nil {
		return tasks, nil
	}
	var encoded string
	if err := json.Unmarshal(raw, &encoded); err != nil {
		return nil, fmt.Errorf("tasks must be an array")
	}
	encoded = strings.TrimSpace(encoded)
	if encoded == "" {
		return nil, fmt.Errorf("tasks JSON array string must not be empty")
	}
	if err := json.Unmarshal([]byte(encoded), &tasks); err != nil {
		return nil, fmt.Errorf("tasks JSON array string is invalid: %w", err)
	}
	return tasks, nil
}

func (t *manifestTool) initializeManifest(input manifestToolInput) (*TasksManifest, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = inferManifestTitle(input.Tasks, t.fallbackTitle)
	}
	if title == "" {
		return nil, fmt.Errorf("title is required in initialize mode and could not be inferred from query or tasks")
	}
	manifest := &TasksManifest{Title: title, ContentBank: input.ContentBank, Sections: input.Sections, VisualPolicy: input.VisualPolicy}
	seen := make(map[string]bool, len(input.Tasks))
	for _, patch := range input.Tasks {
		if patch.PageIndex == nil || patch.Title == nil || patch.ContentType == nil || patch.ContentPlan == nil {
			return nil, fmt.Errorf("initialize task is missing required fields")
		}
		id := deterministicTaskID(*patch.PageIndex)
		if seen[id] {
			return nil, fmt.Errorf("duplicate task_id %q", id)
		}
		seen[id] = true
		item := &TaskItem{
			TaskID: id, PageIndex: *patch.PageIndex, Title: *patch.Title,
			ContentType: *patch.ContentType,
			OutputFile:  deterministicOutputFile(*patch.PageIndex),
			Status:      StatusPending,
			ContentPlan: patch.ContentPlan,
		}
		if patch.SectionID != nil {
			item.SectionID = *patch.SectionID
		}
		if patch.SectionTitle != nil {
			item.SectionTitle = *patch.SectionTitle
		}
		if patch.PageIntent != nil {
			item.PageIntent = *patch.PageIntent
		}
		if len(patch.EvidenceRefs) > 0 {
			item.EvidenceRefs = append([]string(nil), patch.EvidenceRefs...)
		}
		if patch.LayoutVariant != nil {
			item.LayoutVariant = *patch.LayoutVariant
		}
		manifest.Tasks = append(manifest.Tasks, item)
	}
	return manifest, nil
}

func deterministicTaskID(pageIndex int) string {
	if pageIndex <= 0 {
		pageIndex = 1
	}
	return fmt.Sprintf("slide-%d", pageIndex)
}

func deterministicOutputFile(pageIndex int) string {
	if pageIndex <= 0 {
		pageIndex = 1
	}
	return fmt.Sprintf("slide_%02d.pptx", pageIndex)
}

func (t *manifestTool) patchManifest(input manifestToolInput) (*TasksManifest, error) {
	if t != nil && t.draftFirst {
		if manifest, err := ReadTasksDraftManifest(t.workDir); err == nil {
			return applyManifestPatches(manifest, input)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return patchManifest(t.workDir, input)
}

func inferManifestTitle(tasks []manifestTaskPatch, fallback string) string {
	if title := compactManifestTitle(fallback); title != "" {
		return title
	}
	for _, task := range tasks {
		if task.Title == nil {
			continue
		}
		title := strings.TrimSpace(*task.Title)
		if title == "" || title == "封面" || title == "目录" || strings.EqualFold(title, "cover") {
			continue
		}
		return compactManifestTitle(title)
	}
	if len(tasks) > 0 && tasks[0].Title != nil {
		return compactManifestTitle(*tasks[0].Title)
	}
	return ""
}

func compactManifestTitle(title string) string {
	title = strings.TrimSpace(strings.ReplaceAll(title, "\r\n", "\n"))
	if title == "" {
		return ""
	}
	if idx := strings.IndexByte(title, '\n'); idx >= 0 {
		title = strings.TrimSpace(title[:idx])
	}
	const maxRunes = 80
	runes := []rune(title)
	if len(runes) > maxRunes {
		title = strings.TrimSpace(string(runes[:maxRunes]))
	}
	return title
}

func patchManifest(workDir string, input manifestToolInput) (*TasksManifest, error) {
	manifest, err := ReadTasksManifest(workDir)
	if err != nil {
		return nil, err
	}
	return applyManifestPatches(manifest, input)
}

func applyManifestPatches(manifest *TasksManifest, input manifestToolInput) (*TasksManifest, error) {
	if input.Title != "" {
		manifest.Title = input.Title
	}
	if input.ContentBank != nil {
		manifest.ContentBank = input.ContentBank
	}
	if input.Sections != nil {
		manifest.Sections = input.Sections
	}
	if input.VisualPolicy != nil {
		manifest.VisualPolicy = input.VisualPolicy
	}
	for _, patch := range input.Tasks {
		var item *TaskItem
		if patch.PageIndex != nil {
			for _, existing := range manifest.Tasks {
				if existing != nil && existing.PageIndex == *patch.PageIndex {
					item = existing
					break
				}
			}
		}
		if item == nil {
			if patch.PageIndex == nil {
				return nil, fmt.Errorf("page_index is required in patch mode")
			}
			return nil, fmt.Errorf("page_index %d not found", *patch.PageIndex)
		}
		applyManifestPatch(item, patch)
	}
	normalizeVisualPolicy(manifest)
	return manifest, nil
}

func ReadTasksDraftManifest(workDir string) (*TasksManifest, error) {
	data, err := os.ReadFile(filepath.Join(workDir, tasksDraftFileName))
	if err != nil {
		return nil, err
	}
	m := &TasksManifest{}
	if err := m.UnmarshalJSON(data); err != nil {
		return nil, err
	}
	return m, nil
}

func WriteTasksDraftManifest(workDir string, manifest *TasksManifest) error {
	tasksManifestMu.Lock()
	defer tasksManifestMu.Unlock()

	filePath := filepath.Join(workDir, tasksDraftFileName)
	tmpPath := filePath + ".tmp"
	content, err := manifest.MustMarshalJSON()
	if err != nil {
		return err
	}
	if err := json.Unmarshal(content, &TasksManifest{}); err != nil {
		return fmt.Errorf("draft tasks manifest produced invalid JSON: %w", err)
	}
	if err := os.WriteFile(tmpPath, content, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, filePath); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
}

func CommitReviewedTasksDraftManifestIfPresent(workDir string) (*TasksManifest, bool, error) {
	manifest, err := ReadTasksDraftManifest(workDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if err := RequirePassingPlanReview(workDir, manifest); err != nil {
		return nil, false, err
	}
	if err := validateManifestForWrite(manifest); err != nil {
		return nil, false, fmt.Errorf("draft PPTSpec is invalid: %w", err)
	}
	if err := WriteTasksManifest(workDir, manifest); err != nil {
		return nil, false, err
	}
	if err := os.Remove(filepath.Join(workDir, tasksDraftFileName)); err != nil && !os.IsNotExist(err) {
		return nil, false, err
	}
	return manifest, true, nil
}

func applyManifestPatch(item *TaskItem, patch manifestTaskPatch) {
	if patch.Title != nil {
		item.Title = *patch.Title
	}
	if patch.SectionID != nil {
		item.SectionID = *patch.SectionID
	}
	if patch.SectionTitle != nil {
		item.SectionTitle = *patch.SectionTitle
	}
	if patch.ContentType != nil {
		item.ContentType = *patch.ContentType
	}
	if patch.LayoutVariant != nil {
		item.LayoutVariant = *patch.LayoutVariant
	}
	if patch.PageIntent != nil {
		item.PageIntent = *patch.PageIntent
	}
	if len(patch.EvidenceRefs) > 0 {
		item.EvidenceRefs = append([]string(nil), patch.EvidenceRefs...)
	}
	if patch.ContentPlan != nil {
		item.ContentPlan = mergeContentPlanPatch(item.ContentPlan, patch.ContentPlan)
	}
}

func mergeContentPlanPatch(current, patch *ContentPlan) *ContentPlan {
	if current == nil {
		return patch
	}
	if patch == nil {
		return current
	}
	if strings.TrimSpace(patch.Summary) != "" {
		current.Summary = patch.Summary
	}
	if strings.TrimSpace(patch.SlideIntent) != "" {
		current.SlideIntent = patch.SlideIntent
	}
	if strings.TrimSpace(patch.SectionNumber) != "" {
		current.SectionNumber = patch.SectionNumber
	}
	if len(patch.EvidenceRefs) > 0 {
		current.EvidenceRefs = append([]string(nil), patch.EvidenceRefs...)
	}
	if patch.VisualIntent != nil {
		current.VisualIntent = mergeVisualIntentPatch(current.VisualIntent, patch.VisualIntent)
	}
	if len(patch.Components) > 0 {
		current.Components = patch.Components
	}
	return current
}

func mergeVisualIntentPatch(current, patch *VisualIntent) *VisualIntent {
	if current == nil {
		return patch
	}
	if patch == nil {
		return current
	}
	mergeString := func(target *string, value string) {
		if strings.TrimSpace(value) != "" {
			*target = value
		}
	}
	mergeString(&current.Role, patch.Role)
	mergeString(&current.AssetPurpose, patch.AssetPurpose)
	mergeString(&current.AssetSubject, patch.AssetSubject)
	mergeString(&current.AssetQuery, patch.AssetQuery)
	mergeString(&current.Composition, patch.Composition)
	mergeString(&current.PreferredVariant, patch.PreferredVariant)
	mergeString(&current.ImagePosition, patch.ImagePosition)
	mergeString(&current.Caption, patch.Caption)
	mergeString(&current.AssetID, patch.AssetID)
	mergeString(&current.LocalPath, patch.LocalPath)
	mergeString(&current.ImageURL, patch.ImageURL)
	mergeString(&current.PreviewURL, patch.PreviewURL)
	mergeString(&current.SourceURL, patch.SourceURL)
	mergeString(&current.Attribution, patch.Attribution)
	return current
}

func validateManifestForWrite(manifest *TasksManifest) error {
	return validateManifestForWriteWithContract(manifest, templates.BuiltInContractSummary())
}

func validateManifestForWriteWithContract(manifest *TasksManifest, contract *templates.ContractSummary) error {
	if contract == nil {
		contract = templates.BuiltInContractSummary()
	}
	if manifest == nil || len(manifest.Tasks) == 0 {
		return fmt.Errorf("manifest must contain tasks")
	}
	normalizeManifestComponentIDs(manifest)
	seenIDs := make(map[string]bool, len(manifest.Tasks))
	seenPages := make(map[int]bool, len(manifest.Tasks))
	for _, item := range manifest.Tasks {
		if item == nil || strings.TrimSpace(item.TaskID) == "" || item.PageIndex <= 0 {
			return fmt.Errorf("every task requires task_id and positive page_index")
		}
		if strings.TrimSpace(item.Title) == "" || strings.TrimSpace(item.ContentType) == "" || strings.TrimSpace(item.OutputFile) == "" {
			return fmt.Errorf("task %q requires title, content_type and output_file", item.TaskID)
		}
		if seenIDs[item.TaskID] || seenPages[item.PageIndex] {
			return fmt.Errorf("duplicate task identity at %q", item.TaskID)
		}
		seenIDs[item.TaskID], seenPages[item.PageIndex] = true, true
		if !validTaskStatus(item.Status) {
			return fmt.Errorf("task %q has invalid status %q", item.TaskID, item.Status)
		}
		if err := validateContentPlanContractWithContract(item, contract); err != nil {
			return fmt.Errorf("task %q has invalid content_plan: %w", item.TaskID, err)
		}
	}
	return nil
}

func validTaskStatus(status string) bool {
	switch status {
	case StatusPending, StatusGenerating, StatusDone, StatusQADone, StatusFixed, "failed":
		return true
	default:
		return false
	}
}

func validateContentPlanContract(item *TaskItem) error {
	return validateContentPlanContractWithContract(item, templates.BuiltInContractSummary())
}

func validateContentPlanContractWithContract(item *TaskItem, contract *templates.ContractSummary) error {
	if item == nil || item.ContentPlan == nil {
		return nil
	}
	if contract == nil {
		contract = templates.BuiltInContractSummary()
	}
	plan := item.ContentPlan
	for i := range plan.Components {
		component := &plan.Components[i]
		component.Type = strings.TrimSpace(component.Type)
		if !validPlanComponentType(component.Type) {
			return fmt.Errorf("component %q has unsupported type %q", component.ID, component.Type)
		}
		if strings.TrimSpace(component.Type) == "" {
			return fmt.Errorf("component %q missing type", component.ID)
		}
		if component.Type == "section_marker" && strings.TrimSpace(component.Text) == "" {
			return fmt.Errorf("component %q of type section_marker requires text such as 01", component.ID)
		}
		if component.Type == "image" && strings.TrimSpace(component.LocalPath) == "" && strings.TrimSpace(component.AssetQuery) == "" {
			return fmt.Errorf("image component %q requires local_path or asset_query before rendering", component.ID)
		}
		if strings.TrimSpace(component.Title) == "" && strings.TrimSpace(component.Text) == "" &&
			strings.TrimSpace(component.Body) == "" && len(component.Items) == 0 && len(component.Data) == 0 &&
			strings.TrimSpace(component.Role) == "" && strings.TrimSpace(component.Relation) == "" &&
			strings.TrimSpace(component.Target) == "" && strings.TrimSpace(component.Icon) == "" &&
			strings.TrimSpace(component.AssetQuery) == "" && strings.TrimSpace(component.Caption) == "" &&
			strings.TrimSpace(component.AssetID) == "" && strings.TrimSpace(component.LocalPath) == "" &&
			strings.TrimSpace(component.ImageURL) == "" && strings.TrimSpace(component.PreviewURL) == "" &&
			strings.TrimSpace(component.SourceURL) == "" && strings.TrimSpace(component.Attribution) == "" {
			return fmt.Errorf("component %q has no content", component.ID)
		}
	}
	capacity := contract.CapacityFor(item.ContentType)
	if len(plan.Components) > capacity.MaxComponents {
		return fmt.Errorf("too many components for %s: %d > %d (recommended %d-%d)", item.ContentType, len(plan.Components), capacity.MaxComponents, capacity.RecommendedMin, capacity.RecommendedMax)
	}
	return nil
}

func validPlanComponentType(componentType string) bool {
	switch strings.TrimSpace(componentType) {
	case "",
		"headline", "subheadline", "eyebrow", "ppt_title", "section_marker",
		"argument_block", "paragraph", "text_block", "bullet_list", "evidence_list", "list", "numbered_list",
		"feature_card", "fact_card", "key_point", "insight", "recommendation",
		"risk_item", "opportunity_item", "case_snapshot", "decision_item", "toc_item",
		"kpi_metric", "stat", "number_callout",
		"chart", "table", "comparison_matrix",
		"timeline_node", "process_step", "milestone",
		"image", "map", "diagram", "architecture_box",
		"divider", "icon", "tag", "shape", "arrow",
		"quote_block", "callout", "source_note":
		return true
	default:
		return false
	}
}

func maxComponentsForContentType(contentType string) int {
	return templates.BuiltInContractSummary().CapacityFor(contentType).MaxComponents
}

func maxComponentsForContentTypeWithContract(contentType string, contract *templates.ContractSummary) int {
	if contract == nil {
		contract = templates.BuiltInContractSummary()
	}
	return contract.CapacityFor(contentType).MaxComponents
}

func contractMetadata(contract *templates.ContractSummary) map[string]string {
	if contract == nil {
		contract = templates.BuiltInContractSummary()
	}
	return map[string]string{"version": fmt.Sprintf("%d", contract.Version), "sha256": contract.SHA256, "source": contract.Source}
}
