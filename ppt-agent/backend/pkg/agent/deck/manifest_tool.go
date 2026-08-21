package deck

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

var manifestTaskPatchSchema = map[string]*schema.ParameterInfo{
	"task_id":        {Type: schema.String, Desc: "任务 ID；patch 模式必填"},
	"page_index":     {Type: schema.Integer, Desc: "页码；initialize 模式必填"},
	"title":          {Type: schema.String, Desc: "页面标题"},
	"content_type":   {Type: schema.String, Desc: "合法的单页模板英文 ID"},
	"layout_variant": {Type: schema.String, Desc: "同一 content_type 下的版式变体 ID，可为空"},
	"description":    {Type: schema.String, Desc: "页面内容描述"},
	"output_file":    {Type: schema.String, Desc: "输出 PPTX 文件名"},
	"status":         {Type: schema.String, Desc: "pending、generating、done、qa_done、fixed 或 failed"},
	"qa_report":      {Type: schema.String, Desc: "可选的 QA 报告"},
	"fix_attempts":   {Type: schema.Integer, Desc: "修复尝试次数"},
	"background":     {Type: schema.String, Desc: "可用背景主题 ID，可为空"},
	"content_plan": {
		Type: schema.Object,
		Desc: "结构化页面内容规划",
		SubParams: map[string]*schema.ParameterInfo{
			"summary":        {Type: schema.String, Desc: "页面核心摘要"},
			"slide_intent":   {Type: schema.String, Desc: "本页在整套 PPT 中承担的语义目标"},
			"section_number": {Type: schema.String, Desc: "章节分隔页的大章节编号，如 01/02；仅 section_divider 使用，不得写页码"},
			"capacity_hint": {
				Type: schema.Object,
				Desc: "内容容量预估，只能表达密度和溢出风险，不得包含字号坐标",
				SubParams: map[string]*schema.ParameterInfo{
					"estimated_density": {Type: schema.String, Desc: "sparse、normal 或 dense"},
					"overflow_risk":     {Type: schema.String, Desc: "low、medium 或 high"},
					"component_count":   {Type: schema.Integer, Desc: "组件数量预估"},
				},
			},
			"reviewer_status": {
				Type: schema.Object,
				Desc: "规划审查状态，记录轮次和问题；锁定由系统校验后确认",
				SubParams: map[string]*schema.ParameterInfo{
					"planner_round": {Type: schema.Integer, Desc: "规划/润色轮次"},
					"locked":        {Type: schema.String, Desc: "是否已通过质量门，true 或 false"},
					"locked_at":     {Type: schema.String, Desc: "锁定时间，由系统或规划器填写"},
					"issues": {
						Type: schema.Array,
						ElemInfo: &schema.ParameterInfo{Type: schema.Object, SubParams: map[string]*schema.ParameterInfo{
							"code":         {Type: schema.String},
							"severity":     {Type: schema.String},
							"message":      {Type: schema.String},
							"page_index":   {Type: schema.Integer},
							"component_id": {Type: schema.String},
						}},
					},
				},
			},
			"visual_intent": {
				Type: schema.Object,
				Desc: "页面视觉意图，用于规划图片、图表、地图、卡片等视觉角色",
				SubParams: map[string]*schema.ParameterInfo{
					"role":              {Type: schema.String, Desc: "hero_photo、supporting_photo、map、chart、icon、cards 等"},
					"asset_purpose":     {Type: schema.String, Desc: "图片用途：background（整页氛围）/scene（真实场景）/evidence（事实或案例证据）/decorative（装饰）"},
					"asset_subject":     {Type: schema.String, Desc: "经过语义转换后的视觉主体或代理意象，不直接复制用户标题"},
					"asset_query":       {Type: schema.String, Desc: "经过视觉转换、可直接提交给图片 provider 的最终检索词"},
					"composition":       {Type: schema.String, Desc: "构图约束，如 wide landscape、clean negative space on left、subject on right"},
					"preferred_variant": {Type: schema.String, Desc: "偏好的 layout_variant"},
					"image_position":    {Type: schema.String, Desc: "background、left、right、strip、inline 等"},
					"caption":           {Type: schema.String, Desc: "图片说明或替代文本"},
					"asset_id":          {Type: schema.String, Desc: "已选择的本地素材 ID"},
				},
			},
			"elements": {
				Type: schema.Array,
				ElemInfo: &schema.ParameterInfo{Type: schema.Object, SubParams: map[string]*schema.ParameterInfo{
					"type":        {Type: schema.String},
					"items":       {Type: schema.Array, ElemInfo: &schema.ParameterInfo{Type: schema.String}},
					"text":        {Type: schema.String},
					"title":       {Type: schema.String},
					"description": {Type: schema.String},
					"layout_hint": {Type: schema.String},
				}},
			},
			"components": {
				Type: schema.Array,
				Desc: "页内语义组件计划；禁止写坐标、字号、颜色、边距等渲染参数",
				ElemInfo: &schema.ParameterInfo{Type: schema.Object, SubParams: map[string]*schema.ParameterInfo{
					"id":            {Type: schema.String},
					"type":          {Type: schema.String},
					"title":         {Type: schema.String},
					"text":          {Type: schema.String},
					"body":          {Type: schema.String},
					"items":         {Type: schema.Array, ElemInfo: &schema.ParameterInfo{Type: schema.String}},
					"emphasis":      {Type: schema.String},
					"role":          {Type: schema.String},
					"relation":      {Type: schema.String},
					"target":        {Type: schema.String, Desc: "可选语义目标组件 ID，用于 arrow 等关系组件；不得写坐标"},
					"icon":          {Type: schema.String, Desc: "可选图标语义或短 glyph，由生成器映射为实际图标"},
					"source":        {Type: schema.String},
					"asset_purpose": {Type: schema.String, Desc: "background、scene、evidence 或 decorative"},
					"asset_subject": {Type: schema.String, Desc: "图片要表现的具体视觉主体或代理意象"},
					"asset_query":   {Type: schema.String},
					"composition":   {Type: schema.String, Desc: "图片构图和留白方向，使用语义描述"},
					"caption":       {Type: schema.String},
					"asset_id":      {Type: schema.String},
					"description":   {Type: schema.String},
					"data":          {Type: schema.Object},
				}},
			},
		},
	},
}

var manifestToolInfo = &schema.ToolInfo{
	Name: "update_tasks_manifest",
	Desc: "以结构化、原子方式初始化或批量更新 DeckSpec。initialize/patch 只保存规划草稿，不做完整 DeckSpec 硬校验；审查完成后 commit 才校验并发布 tasks.json。不要使用 edit_file 修改 tasks.json。",
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"mode":     {Type: schema.String, Required: true, Desc: "initialize、patch 或 commit"},
		"title":    {Type: schema.String, Desc: "PPT 标题；initialize 模式建议填写，缺省时系统会从任务上下文推断"},
		"theme":    {Type: schema.String, Desc: "配色英文 ID"},
		"template": {Type: schema.String, Desc: "整套模板英文 ID"},
		"tasks": {
			Type: schema.Array, Desc: "要初始化或更新的页面列表；initialize/patch 必填，commit 可省略",
			ElemInfo: &schema.ParameterInfo{Type: schema.Object, SubParams: manifestTaskPatchSchema},
		},
	}),
}

type manifestTaskPatch struct {
	TaskID        string       `json:"task_id"`
	PageIndex     *int         `json:"page_index,omitempty"`
	Title         *string      `json:"title,omitempty"`
	ContentType   *string      `json:"content_type,omitempty"`
	LayoutVariant *string      `json:"layout_variant,omitempty"`
	Description   *string      `json:"description,omitempty"`
	OutputFile    *string      `json:"output_file,omitempty"`
	Status        *string      `json:"status,omitempty"`
	QAReport      *string      `json:"qa_report,omitempty"`
	FixAttempts   *int         `json:"fix_attempts,omitempty"`
	ContentPlan   *ContentPlan `json:"content_plan,omitempty"`
	Background    *string      `json:"background,omitempty"`
}

type manifestToolInput struct {
	Mode     string              `json:"mode"`
	Title    string              `json:"title,omitempty"`
	Theme    string              `json:"theme,omitempty"`
	Template string              `json:"template,omitempty"`
	Tasks    []manifestTaskPatch `json:"tasks"`
}

type manifestToolRawInput struct {
	Mode     string          `json:"mode"`
	Title    string          `json:"title,omitempty"`
	Theme    string          `json:"theme,omitempty"`
	Template string          `json:"template,omitempty"`
	Tasks    json.RawMessage `json:"tasks"`
}

type manifestTool struct {
	workDir               string
	backgroundRoot        string
	recommendedBackground string
	normalizeBackgrounds  bool
	fallbackTitle         string
	draftFirst            bool
}

func newManifestTool(workDir string) tool.InvokableTool {
	return &manifestTool{workDir: workDir}
}

func newConfiguredManifestTool(workDir, skillsDir string, outline *TaskOutline, query string) tool.InvokableTool {
	t := &manifestTool{
		workDir:        workDir,
		backgroundRoot: filepath.Join(skillsDir, "visual_designer", "background_templates"),
		fallbackTitle:  compactManifestTitle(query),
		draftFirst:     true,
	}
	if outline != nil && strings.TrimSpace(outline.Title) != "" {
		t.fallbackTitle = compactManifestTitle(outline.Title)
	}
	if outline != nil && outline.ContentMode == OutlineContentModeRecommendedStyle && outline.UseBackground {
		t.normalizeBackgrounds = true
		t.recommendedBackground = strings.TrimSpace(outline.RecommendedBackground)
	}
	return t
}

func (t *manifestTool) Info(context.Context) (*schema.ToolInfo, error) {
	return manifestToolInfo, nil
}

func (t *manifestTool) InvokableRun(_ context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	input, err := parseManifestToolInput(argumentsInJSON)
	if err != nil {
		return "", fmt.Errorf("invalid manifest tool input: %w", err)
	}

	var manifest *TasksManifest
	target := "final"
	switch strings.ToLower(strings.TrimSpace(input.Mode)) {
	case "initialize":
		if len(input.Tasks) == 0 {
			return "", fmt.Errorf("tasks must not be empty")
		}
		manifest, err = t.initializeManifest(input)
		target = t.writeTarget()
	case "patch":
		if len(input.Tasks) == 0 {
			return "", fmt.Errorf("tasks must not be empty")
		}
		manifest, err = t.patchManifest(input)
		target = t.writeTarget()
	case "commit":
		committed, ok, commitErr := CommitTasksDraftManifestIfPresent(t.workDir)
		if commitErr != nil {
			result, _ := json.Marshal(map[string]any{
				"ok": false, "mode": input.Mode, "target": "draft", "committed": false, "error": commitErr.Error(),
			})
			return string(result), nil
		}
		if !ok {
			committed, commitErr = ReadTasksManifest(t.workDir)
			if commitErr != nil {
				return "", fmt.Errorf("no draft manifest to commit and final tasks.json is unavailable: %w", commitErr)
			}
		}
		result, _ := json.Marshal(map[string]any{
			"ok": true, "mode": input.Mode, "target": "final", "committed": ok, "total": len(committed.Tasks),
		})
		return string(result), nil
	default:
		err = fmt.Errorf("unsupported mode %q", input.Mode)
	}
	if err != nil {
		return "", err
	}
	if err := t.normalizeManifestBackgrounds(manifest); err != nil {
		return "", err
	}
	normalizeManifestLayoutVariants(manifest)
	if !t.shouldDeferValidation() {
		if err := validateManifestForWrite(manifest); err != nil {
			return "", err
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

func (t *manifestTool) normalizeManifestBackgrounds(manifest *TasksManifest) error {
	if !t.normalizeBackgrounds || manifest == nil || len(manifest.Tasks) == 0 {
		return nil
	}
	if !t.validBackgroundReference(t.recommendedBackground) {
		return fmt.Errorf("recommended background %q is unavailable", t.recommendedBackground)
	}

	recommendedTheme := backgroundTheme(t.recommendedBackground)
	for _, item := range manifest.Tasks {
		if item == nil {
			continue
		}
		background := strings.TrimSpace(item.Background)
		if background == "" {
			continue
		}
		if !t.validBackgroundReference(background) || backgroundTheme(background) != recommendedTheme {
			item.Background = ""
		}
	}
	refs := t.backgroundImageRefs(t.recommendedBackground)
	previous := ""
	for _, item := range manifest.Tasks {
		if item == nil {
			continue
		}
		if strings.TrimSpace(item.Background) == "" || (len(refs) > 1 && item.Background == previous) {
			item.Background = rotatingBackgroundRef(t.recommendedBackground, refs, item.PageIndex, previous)
		}
		previous = item.Background
	}
	return nil
}

func (t *manifestTool) validBackgroundReference(background string) bool {
	background = strings.TrimSpace(strings.ReplaceAll(background, "\\", "/"))
	if background == "" || t.backgroundRoot == "" {
		return false
	}
	parts := strings.Split(background, "/")
	themeRoot := filepath.Join(t.backgroundRoot, parts[0])
	if info, err := os.Stat(themeRoot); err != nil || !info.IsDir() {
		return false
	}
	if len(parts) == 1 {
		return true
	}
	candidate := filepath.Clean(filepath.Join(t.backgroundRoot, filepath.FromSlash(background)))
	root := filepath.Clean(t.backgroundRoot) + string(filepath.Separator)
	if !strings.HasPrefix(candidate, root) {
		return false
	}
	info, err := os.Stat(candidate)
	return err == nil && !info.IsDir()
}

func (t *manifestTool) backgroundImageRefs(background string) []string {
	theme := strings.Split(strings.TrimSpace(strings.ReplaceAll(background, "\\", "/")), "/")[0]
	paths, _ := filepath.Glob(filepath.Join(t.backgroundRoot, theme, "images", "*"))
	refs := make([]string, 0, len(paths))
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		refs = append(refs, filepath.ToSlash(filepath.Join(theme, "images", filepath.Base(path))))
	}
	sort.Strings(refs)
	return refs
}

func rotatingBackgroundRef(theme string, refs []string, pageIndex int, previous string) string {
	if len(refs) == 0 {
		return theme
	}
	index := (pageIndex - 1) % len(refs)
	if refs[index] == previous && len(refs) > 1 {
		index = (index + 1) % len(refs)
	}
	return refs[index]
}

func backgroundTheme(background string) string {
	background = strings.TrimSpace(strings.ReplaceAll(background, "\\", "/"))
	if background == "" {
		return ""
	}
	return strings.Split(background, "/")[0]
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
	var raw manifestToolRawInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &raw); err != nil {
		return manifestToolInput{}, err
	}
	var tasks []manifestTaskPatch
	if len(raw.Tasks) > 0 && strings.TrimSpace(string(raw.Tasks)) != "" && string(raw.Tasks) != "null" {
		var err error
		tasks, err = parseManifestTaskPatches(raw.Tasks)
		if err != nil {
			return manifestToolInput{}, err
		}
	}
	return manifestToolInput{
		Mode: raw.Mode, Title: raw.Title, Theme: raw.Theme, Template: raw.Template, Tasks: tasks,
	}, nil
}

func parseManifestTaskPatches(raw json.RawMessage) ([]manifestTaskPatch, error) {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "" || string(raw) == "null" {
		return nil, fmt.Errorf("tasks must not be empty")
	}
	var tasks []manifestTaskPatch
	if err := json.Unmarshal(raw, &tasks); err == nil {
		return tasks, nil
	}
	var encoded string
	if err := json.Unmarshal(raw, &encoded); err != nil {
		return nil, fmt.Errorf("tasks must be an array or a JSON array string")
	}
	encoded = strings.TrimSpace(encoded)
	if encoded == "" {
		return nil, fmt.Errorf("tasks must not be empty")
	}
	if err := json.Unmarshal([]byte(encoded), &tasks); err != nil {
		if recovered, recoverErr := extractManifestTaskPatchObjects(encoded); recoverErr == nil && len(recovered) > 0 {
			return recovered, nil
		} else if recoverErr != nil {
			return nil, fmt.Errorf("tasks string must contain a JSON array or task objects: %v; recovery failed: %w", err, recoverErr)
		}
		return nil, fmt.Errorf("tasks string must contain a JSON array or task objects: %w", err)
	}
	return tasks, nil
}

func extractManifestTaskPatchObjects(value string) ([]manifestTaskPatch, error) {
	value = escapeBareQuotesInJSONStrings(value)
	var tasks []manifestTaskPatch
	inString := false
	escaped := false
	depth := 0
	startStack := make([]int, 0, 8)
	for i := 0; i < len(value); i++ {
		char := value[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if char == '\\' {
				escaped = true
				continue
			}
			if char == '"' {
				inString = false
			}
			continue
		}
		switch char {
		case '"':
			if depth == 0 {
				inString = true
				continue
			}
			inString = true
		case '{':
			startStack = append(startStack, i)
			depth++
		case '}':
			if depth == 0 {
				continue
			}
			depth--
			start := startStack[len(startStack)-1]
			startStack = startStack[:len(startStack)-1]
			var task manifestTaskPatch
			if err := json.Unmarshal([]byte(value[start:i+1]), &task); err == nil && looksLikeManifestTaskPatch(task) {
				tasks = append(tasks, task)
			}
		default:
			continue
		}
	}
	if depth != 0 {
		return nil, fmt.Errorf("unterminated task object")
	}
	if inString {
		return nil, fmt.Errorf("unterminated task string")
	}
	if len(tasks) == 0 {
		return nil, fmt.Errorf("no task objects found")
	}
	return tasks, nil
}

func looksLikeManifestTaskPatch(task manifestTaskPatch) bool {
	if strings.TrimSpace(task.TaskID) != "" {
		return true
	}
	if task.PageIndex != nil || task.ContentType != nil || task.OutputFile != nil || task.Status != nil {
		return true
	}
	if task.LayoutVariant != nil || task.ContentPlan != nil || task.Background != nil {
		return true
	}
	return task.QAReport != nil || task.FixAttempts != nil
}

func escapeBareQuotesInJSONStrings(value string) string {
	var normalized strings.Builder
	normalized.Grow(len(value))
	inString := false
	escaped := false
	for i := 0; i < len(value); i++ {
		char := value[i]
		if !inString {
			normalized.WriteByte(char)
			if char == '"' {
				inString = true
			}
			continue
		}
		if escaped {
			normalized.WriteByte(char)
			escaped = false
			continue
		}
		if char == '\\' {
			normalized.WriteByte(char)
			escaped = true
			continue
		}
		if char != '"' {
			normalized.WriteByte(char)
			continue
		}
		if closesJSONString(value, i) {
			normalized.WriteByte(char)
			inString = false
			continue
		}
		normalized.WriteString(`\"`)
	}
	return normalized.String()
}

func closesJSONString(value string, quoteIndex int) bool {
	for i := quoteIndex + 1; i < len(value); i++ {
		switch value[i] {
		case ' ', '\t', '\r', '\n':
			continue
		case ':', ',', '}', ']':
			return true
		default:
			return false
		}
	}
	return true
}

func isManifestTaskSeparator(char byte) bool {
	switch char {
	case ' ', '\t', '\r', '\n', '[', ']', ',':
		return true
	default:
		return false
	}
}

func (t *manifestTool) initializeManifest(input manifestToolInput) (*TasksManifest, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = inferManifestTitle(input.Tasks, t.fallbackTitle)
	}
	if title == "" {
		return nil, fmt.Errorf("title is required in initialize mode and could not be inferred from query or tasks")
	}
	manifest := &TasksManifest{Title: title, Theme: input.Theme, Template: input.Template}
	seen := make(map[string]bool, len(input.Tasks))
	for _, patch := range input.Tasks {
		if patch.PageIndex == nil || patch.Title == nil || patch.ContentType == nil || patch.Description == nil || patch.OutputFile == nil {
			return nil, fmt.Errorf("initialize task %q is missing required fields", patch.TaskID)
		}
		id := strings.TrimSpace(patch.TaskID)
		if id == "" {
			id = fmt.Sprint(*patch.PageIndex)
		}
		if seen[id] {
			return nil, fmt.Errorf("duplicate task_id %q", id)
		}
		seen[id] = true
		status := StatusPending
		if patch.Status != nil && *patch.Status != "" {
			status = *patch.Status
		}
		item := &TaskItem{
			TaskID: id, PageIndex: *patch.PageIndex, Title: *patch.Title,
			ContentType: *patch.ContentType, Description: *patch.Description,
			OutputFile: *patch.OutputFile, Status: status,
			ContentPlan: patch.ContentPlan,
		}
		if patch.LayoutVariant != nil {
			item.LayoutVariant = *patch.LayoutVariant
		}
		if patch.QAReport != nil {
			item.QAReport = *patch.QAReport
		}
		if patch.FixAttempts != nil {
			item.FixAttempts = *patch.FixAttempts
		}
		if patch.Background != nil {
			item.Background = *patch.Background
		}
		manifest.Tasks = append(manifest.Tasks, item)
	}
	return manifest, nil
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
	if input.Theme != "" {
		manifest.Theme = input.Theme
	}
	if input.Template != "" {
		manifest.Template = input.Template
	}
	for i, patch := range input.Tasks {
		item := manifest.GetTask(strings.TrimSpace(patch.TaskID))
		if item == nil && patch.PageIndex != nil {
			for _, existing := range manifest.Tasks {
				if existing != nil && existing.PageIndex == *patch.PageIndex {
					item = existing
					break
				}
			}
		}
		if item == nil && strings.TrimSpace(patch.TaskID) == "" && patch.PageIndex == nil && i < len(manifest.Tasks) {
			item = manifest.Tasks[i]
		}
		if item == nil {
			if strings.TrimSpace(patch.TaskID) == "" {
				return nil, fmt.Errorf("task_id or page_index is required in patch mode")
			}
			return nil, fmt.Errorf("task %q not found", patch.TaskID)
		}
		applyManifestPatch(item, patch)
	}
	return manifest, nil
}

func ReadTasksDraftManifest(workDir string) (*TasksManifest, error) {
	data, err := os.ReadFile(filepath.Join(workDir, "tasks.draft.json"))
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

	filePath := filepath.Join(workDir, "tasks.draft.json")
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

func CommitTasksDraftManifestIfPresent(workDir string) (*TasksManifest, bool, error) {
	manifest, err := ReadTasksDraftManifest(workDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if err := validateManifestForWrite(manifest); err != nil {
		return nil, false, fmt.Errorf("draft DeckSpec is invalid: %w", err)
	}
	if err := WriteTasksManifest(workDir, manifest); err != nil {
		return nil, false, err
	}
	if err := os.Remove(filepath.Join(workDir, "tasks.draft.json")); err != nil && !os.IsNotExist(err) {
		return nil, false, err
	}
	return manifest, true, nil
}

func applyManifestPatch(item *TaskItem, patch manifestTaskPatch) {
	if patch.PageIndex != nil {
		item.PageIndex = *patch.PageIndex
	}
	if patch.Title != nil {
		item.Title = *patch.Title
	}
	if patch.ContentType != nil {
		item.ContentType = *patch.ContentType
	}
	if patch.LayoutVariant != nil {
		item.LayoutVariant = *patch.LayoutVariant
	}
	if patch.Description != nil {
		item.Description = *patch.Description
	}
	if patch.OutputFile != nil {
		item.OutputFile = *patch.OutputFile
	}
	if patch.Status != nil {
		item.Status = *patch.Status
	}
	if patch.QAReport != nil {
		item.QAReport = *patch.QAReport
	}
	if patch.FixAttempts != nil {
		item.FixAttempts = *patch.FixAttempts
	}
	if patch.ContentPlan != nil {
		item.ContentPlan = patch.ContentPlan
	}
	if patch.Background != nil {
		item.Background = *patch.Background
	}
}

func validateManifestForWrite(manifest *TasksManifest) error {
	if manifest == nil || len(manifest.Tasks) == 0 {
		return fmt.Errorf("manifest must contain tasks")
	}
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
		if err := validateContentPlanContract(item); err != nil {
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
	if item == nil || item.ContentPlan == nil {
		return nil
	}
	plan := item.ContentPlan
	for i := range plan.Components {
		component := &plan.Components[i]
		component.Type = normalizePlanComponentType(component.Type)
		if !validPlanComponentType(component.Type) {
			return fmt.Errorf("component %q has unsupported type %q", component.ID, component.Type)
		}
		if strings.TrimSpace(component.Type) == "" {
			return fmt.Errorf("component %q missing type", component.ID)
		}
		if strings.TrimSpace(component.Title) == "" && strings.TrimSpace(component.Text) == "" &&
			strings.TrimSpace(component.Body) == "" && len(component.Items) == 0 && len(component.Data) == 0 &&
			strings.TrimSpace(component.Role) == "" && strings.TrimSpace(component.Relation) == "" &&
			strings.TrimSpace(component.Target) == "" && strings.TrimSpace(component.Icon) == "" &&
			strings.TrimSpace(component.AssetQuery) == "" && strings.TrimSpace(component.Caption) == "" &&
			strings.TrimSpace(component.AssetID) == "" {
			return fmt.Errorf("component %q has no content", component.ID)
		}
	}
	if len(plan.Components) > maxComponentsForContentType(item.ContentType) {
		return fmt.Errorf("too many components for %s: %d > %d", item.ContentType, len(plan.Components), maxComponentsForContentType(item.ContentType))
	}
	if plan.CapacityHint != nil {
		if !validDensity(plan.CapacityHint.EstimatedDensity) {
			return fmt.Errorf("invalid estimated_density %q", plan.CapacityHint.EstimatedDensity)
		}
		if !validOverflowRisk(plan.CapacityHint.OverflowRisk) {
			return fmt.Errorf("invalid overflow_risk %q", plan.CapacityHint.OverflowRisk)
		}
		if plan.CapacityHint.ComponentCount > 0 && plan.CapacityHint.ComponentCount != len(plan.Components) {
			plan.CapacityHint.ComponentCount = len(plan.Components)
		}
	}
	if plan.ReviewerStatus != nil && plan.ReviewerStatus.PlannerRound > 3 {
		return fmt.Errorf("planner_round exceeds max refinement rounds: %d", plan.ReviewerStatus.PlannerRound)
	}
	return nil
}

func normalizePlanComponentType(componentType string) string {
	switch strings.TrimSpace(componentType) {
	case "title", "heading", "main_title", "section_title":
		return "headline"
	case "subtitle", "sub_title", "section_subtitle":
		return "subheadline"
	case "section_number":
		return "section_marker"
	case "long_text", "long_paragraph", "argument", "narrative_block":
		return "argument_block"
	case "ordered_list", "steps":
		return "numbered_list"
	case "list_group":
		return "list"
	default:
		return strings.TrimSpace(componentType)
	}
}

func validPlanComponentType(componentType string) bool {
	switch strings.TrimSpace(componentType) {
	case "",
		"headline", "subheadline", "eyebrow", "deck_title", "section_marker",
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
	switch strings.TrimSpace(contentType) {
	case "title_slide", "section_divider", "quote_slide":
		return 4
	case "agenda", "summary_slide", "kpi_dashboard":
		return 6
	case "timeline", "process_flow", "card_grid", "icon_grid", "three_column":
		return 8
	case "chart_slide", "comparison_table", "two_column", "image_text", "case_study", "example_detail", "deep_dive":
		return 10
	case "swot_analysis", "brand_focus", "region_map":
		return 8
	case "kanban":
		return 10
	case "image_hero":
		return 4
	default:
		return 8
	}
}

func validDensity(value string) bool {
	switch strings.TrimSpace(value) {
	case "", "sparse", "normal", "dense":
		return true
	default:
		return false
	}
}

func validOverflowRisk(value string) bool {
	switch strings.TrimSpace(value) {
	case "", "low", "medium", "high":
		return true
	default:
		return false
	}
}
