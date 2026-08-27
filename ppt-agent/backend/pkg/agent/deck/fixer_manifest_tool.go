package deck

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

var selectedTasksPatchToolInfo = &schema.ToolInfo{
	Name: "patch_selected_tasks",
	Desc: "只修改后端授权的已生成页面。一次调用合并所有目标页修正，不允许新增、删除、重排页面或修改运行状态。",
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"tasks": {
			Type:     schema.Array,
			Required: true,
			Desc:     "目标页面 patch 列表，每项必须携带原 task_id",
			ElemInfo: &schema.ParameterInfo{Type: schema.Object, SubParams: manifestTaskPatchSchema},
		},
	}),
}

var draftTasksPatchToolInfo = &schema.ToolInfo{
	Name: "patch_tasks_draft",
	Desc: "批量修正当前 DeckSpec 草稿中已有页面。只允许 patch，不负责初始化、审查或提交。",
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"title":    {Type: schema.String, Desc: "仅在审查报告指出标题问题时修改整套标题"},
		"theme":    {Type: schema.String, Desc: "仅在审查报告指出主题不合法时修改"},
		"template": {Type: schema.String, Desc: "仅在审查报告指出模板不合法时修改"},
		"tasks": {
			Type:     schema.Array,
			Required: true,
			Desc:     "需要修正的已有页面 patch 列表",
			ElemInfo: &schema.ParameterInfo{Type: schema.Object, SubParams: manifestTaskPatchSchema},
		},
	}),
}

type draftTasksPatchTool struct {
	workDir string
	scoped  bool
	allowed map[string]bool
}

func newDraftTasksPatchTool(workDir string) tool.InvokableTool {
	return &draftTasksPatchTool{workDir: workDir}
}

func newScopedDraftTasksPatchTool(workDir string, allowedTaskIDs []string) tool.InvokableTool {
	allowed := make(map[string]bool, len(allowedTaskIDs))
	for _, taskID := range allowedTaskIDs {
		if taskID = strings.TrimSpace(taskID); taskID != "" {
			allowed[taskID] = true
		}
	}
	return &draftTasksPatchTool{workDir: workDir, scoped: true, allowed: allowed}
}

func (t *draftTasksPatchTool) Info(context.Context) (*schema.ToolInfo, error) {
	return draftTasksPatchToolInfo, nil
}

func (t *draftTasksPatchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	input, err := parseManifestPatchToolInput(argumentsInJSON)
	if err != nil {
		return fixerToolError(fmt.Errorf("invalid draft patch: %w", err)), nil
	}
	if len(input.Tasks) == 0 && !hasManifestHeaderPatch(input) {
		return fixerToolError(fmt.Errorf("tasks must not be empty unless title, theme or template is patched")), nil
	}
	if t.scoped {
		for _, patch := range input.Tasks {
			taskID := strings.TrimSpace(patch.TaskID)
			if taskID == "" {
				return fixerToolError(fmt.Errorf("scoped draft patch requires task_id")), nil
			}
			if !t.allowed[taskID] {
				return fixerToolError(fmt.Errorf("task_id %q is not authorized for this review slice", taskID)), nil
			}
			if patch.PageIndex != nil || patch.OutputFile != nil || patch.Status != nil || patch.FixAttempts != nil {
				return fixerToolError(fmt.Errorf("task_id %q cannot change page_index, output_file, status or fix_attempts", taskID)), nil
			}
		}
	}
	input.Mode = "patch"
	payload, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	return (&manifestTool{workDir: t.workDir, draftFirst: true}).InvokableRun(ctx, string(payload), opts...)
}

type selectedTasksPatchTool struct {
	workDir string
	allowed map[string]bool
}

func newSelectedTasksPatchTool(workDir string, allowedTaskIDs []string) tool.InvokableTool {
	allowed := make(map[string]bool, len(allowedTaskIDs))
	for _, taskID := range allowedTaskIDs {
		if taskID = strings.TrimSpace(taskID); taskID != "" {
			allowed[taskID] = true
		}
	}
	return &selectedTasksPatchTool{workDir: workDir, allowed: allowed}
}

func (t *selectedTasksPatchTool) Info(context.Context) (*schema.ToolInfo, error) {
	return selectedTasksPatchToolInfo, nil
}

func (t *selectedTasksPatchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	input, err := parseManifestPatchToolInput(argumentsInJSON)
	if err != nil {
		return fixerToolError(fmt.Errorf("invalid selected task patch: %w", err)), nil
	}
	if len(input.Tasks) == 0 {
		return fixerToolError(fmt.Errorf("tasks must not be empty")), nil
	}
	for _, patch := range input.Tasks {
		taskID := strings.TrimSpace(patch.TaskID)
		if !t.allowed[taskID] {
			return fixerToolError(fmt.Errorf("task_id %q is not authorized for this fix", taskID)), nil
		}
		if patch.PageIndex != nil || patch.OutputFile != nil || patch.Status != nil || patch.FixAttempts != nil {
			return fixerToolError(fmt.Errorf("task_id %q cannot change page_index, output_file, status or fix_attempts", taskID)), nil
		}
	}
	payload, err := json.Marshal(manifestToolInput{Mode: "patch", Tasks: input.Tasks})
	if err != nil {
		return "", err
	}
	return (&manifestTool{workDir: t.workDir}).InvokableRun(ctx, string(payload), opts...)
}

func parseManifestPatchToolInput(argumentsInJSON string) (manifestToolInput, error) {
	var raw manifestToolRawInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &raw); err != nil {
		return manifestToolInput{}, err
	}
	tasks, err := parseManifestTaskPatches(raw.Tasks)
	if err != nil {
		return manifestToolInput{}, err
	}
	return manifestToolInput{
		Title: raw.Title, Theme: raw.Theme, Template: raw.Template, Tasks: tasks,
	}, nil
}

func fixerToolError(err error) string {
	payload, _ := json.Marshal(map[string]any{"ok": false, "error": err.Error()})
	return string(payload)
}
