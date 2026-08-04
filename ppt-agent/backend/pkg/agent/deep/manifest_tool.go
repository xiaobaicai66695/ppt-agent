package deep

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

var manifestTaskPatchSchema = map[string]*schema.ParameterInfo{
	"task_id":      {Type: schema.String, Desc: "任务 ID；patch 模式必填"},
	"page_index":   {Type: schema.Integer, Desc: "页码；initialize 模式必填"},
	"title":        {Type: schema.String, Desc: "页面标题"},
	"content_type": {Type: schema.String, Desc: "合法的单页模板英文 ID"},
	"description":  {Type: schema.String, Desc: "页面内容描述"},
	"output_file":  {Type: schema.String, Desc: "输出 PPTX 文件名"},
	"status":       {Type: schema.String, Desc: "pending、generating、done、qa_done、fixed 或 failed"},
	"qa_report":    {Type: schema.String, Desc: "可选的 QA 报告"},
	"fix_attempts": {Type: schema.Integer, Desc: "修复尝试次数"},
	"background":   {Type: schema.String, Desc: "可用背景主题 ID，可为空"},
	"content_plan": {
		Type: schema.Object,
		Desc: "结构化页面内容规划",
		SubParams: map[string]*schema.ParameterInfo{
			"summary": {Type: schema.String, Desc: "页面核心摘要"},
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
		},
	},
}

var manifestToolInfo = &schema.ToolInfo{
	Name: "update_tasks_manifest",
	Desc: "以结构化、原子方式初始化或批量更新 tasks.json。一次提交全部页面规划；不要使用 edit_file 修改 tasks.json。",
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"mode":     {Type: schema.String, Required: true, Desc: "initialize 或 patch"},
		"title":    {Type: schema.String, Desc: "PPT 标题，initialize 模式必填"},
		"theme":    {Type: schema.String, Desc: "配色英文 ID"},
		"template": {Type: schema.String, Desc: "整套模板英文 ID"},
		"tasks": {
			Type: schema.Array, Required: true, Desc: "要初始化或更新的页面列表",
			ElemInfo: &schema.ParameterInfo{Type: schema.Object, SubParams: manifestTaskPatchSchema},
		},
	}),
}

type manifestTaskPatch struct {
	TaskID      string       `json:"task_id"`
	PageIndex   *int         `json:"page_index,omitempty"`
	Title       *string      `json:"title,omitempty"`
	ContentType *string      `json:"content_type,omitempty"`
	Description *string      `json:"description,omitempty"`
	OutputFile  *string      `json:"output_file,omitempty"`
	Status      *string      `json:"status,omitempty"`
	QAReport    *string      `json:"qa_report,omitempty"`
	FixAttempts *int         `json:"fix_attempts,omitempty"`
	ContentPlan *ContentPlan `json:"content_plan,omitempty"`
	Background  *string      `json:"background,omitempty"`
}

type manifestToolInput struct {
	Mode     string              `json:"mode"`
	Title    string              `json:"title,omitempty"`
	Theme    string              `json:"theme,omitempty"`
	Template string              `json:"template,omitempty"`
	Tasks    []manifestTaskPatch `json:"tasks"`
}

type manifestTool struct{ workDir string }

func newManifestTool(workDir string) tool.InvokableTool {
	return &manifestTool{workDir: workDir}
}

func (t *manifestTool) Info(context.Context) (*schema.ToolInfo, error) {
	return manifestToolInfo, nil
}

func (t *manifestTool) InvokableRun(_ context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	var input manifestToolInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("invalid manifest tool input: %w", err)
	}
	if len(input.Tasks) == 0 {
		return "", fmt.Errorf("tasks must not be empty")
	}

	var manifest *TasksManifest
	var err error
	switch strings.ToLower(strings.TrimSpace(input.Mode)) {
	case "initialize":
		manifest, err = initializeManifest(input)
	case "patch":
		manifest, err = patchManifest(t.workDir, input)
	default:
		err = fmt.Errorf("unsupported mode %q", input.Mode)
	}
	if err != nil {
		return "", err
	}
	if err := validateManifestForWrite(manifest); err != nil {
		return "", err
	}
	if err := WriteTasksManifest(t.workDir, manifest); err != nil {
		return "", err
	}
	result, _ := json.Marshal(map[string]any{
		"ok": true, "mode": input.Mode, "updated": len(input.Tasks), "total": len(manifest.Tasks),
	})
	return string(result), nil
}

func initializeManifest(input manifestToolInput) (*TasksManifest, error) {
	if strings.TrimSpace(input.Title) == "" {
		return nil, fmt.Errorf("title is required in initialize mode")
	}
	manifest := &TasksManifest{Title: input.Title, Theme: input.Theme, Template: input.Template}
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

func patchManifest(workDir string, input manifestToolInput) (*TasksManifest, error) {
	manifest, err := ReadTasksManifest(workDir)
	if err != nil {
		return nil, err
	}
	if input.Title != "" {
		manifest.Title = input.Title
	}
	if input.Theme != "" {
		manifest.Theme = input.Theme
	}
	if input.Template != "" {
		manifest.Template = input.Template
	}
	for _, patch := range input.Tasks {
		if strings.TrimSpace(patch.TaskID) == "" {
			return nil, fmt.Errorf("task_id is required in patch mode")
		}
		item := manifest.GetTask(patch.TaskID)
		if item == nil {
			return nil, fmt.Errorf("task %q not found", patch.TaskID)
		}
		applyManifestPatch(item, patch)
	}
	return manifest, nil
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
