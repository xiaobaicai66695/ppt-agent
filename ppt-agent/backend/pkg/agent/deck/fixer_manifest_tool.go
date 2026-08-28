package deck

import (
	"context"
	"encoding/json"
	"fmt"

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
			Desc:     "目标页面 patch 列表，每项必须携带 page_index；运行字段由后端维护",
			ElemInfo: &schema.ParameterInfo{Type: schema.Object, SubParams: manifestTaskPatchSchema},
		},
	}),
}

var draftTasksPatchToolInfo = &schema.ToolInfo{
	Name: "patch_tasks_draft",
	Desc: "批量修正当前 DeckSpec 草稿中已有页面。只允许 patch，不负责初始化、审查或提交。",
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"title": {Type: schema.String, Desc: "仅在审查报告指出标题问题时修改整套标题"},
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
	allowed map[int]bool
}

func newDraftTasksPatchTool(workDir string) tool.InvokableTool {
	return &draftTasksPatchTool{workDir: workDir}
}

func newScopedDraftTasksPatchTool(workDir string, allowedPageIndexes []int) tool.InvokableTool {
	allowed := make(map[int]bool, len(allowedPageIndexes))
	for _, pageIndex := range allowedPageIndexes {
		if pageIndex > 0 {
			allowed[pageIndex] = true
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
		return fixerToolError(fmt.Errorf("tasks must not be empty unless title/content_bank/sections is patched")), nil
	}
	if t.scoped {
		for _, patch := range input.Tasks {
			if patch.PageIndex == nil || *patch.PageIndex <= 0 {
				return fixerToolError(fmt.Errorf("scoped draft patch requires page_index")), nil
			}
			if !t.allowed[*patch.PageIndex] {
				return fixerToolError(fmt.Errorf("page_index %d is not authorized for this review slice", *patch.PageIndex)), nil
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
	allowed map[int]bool
}

func newSelectedTasksPatchTool(workDir string, allowedPageIndexes []int) tool.InvokableTool {
	allowed := make(map[int]bool, len(allowedPageIndexes))
	for _, pageIndex := range allowedPageIndexes {
		if pageIndex > 0 {
			allowed[pageIndex] = true
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
		if patch.PageIndex == nil || *patch.PageIndex <= 0 {
			return fixerToolError(fmt.Errorf("selected task patch requires page_index")), nil
		}
		if !t.allowed[*patch.PageIndex] {
			return fixerToolError(fmt.Errorf("page_index %d is not authorized for this fix", *patch.PageIndex)), nil
		}
	}
	payload, err := json.Marshal(manifestToolInput{Mode: "patch", Tasks: input.Tasks})
	if err != nil {
		return "", err
	}
	return (&manifestTool{workDir: t.workDir}).InvokableRun(ctx, string(payload), opts...)
}

func parseManifestPatchToolInput(argumentsInJSON string) (manifestToolInput, error) {
	if err := rejectForbiddenManifestFields([]byte(argumentsInJSON)); err != nil {
		return manifestToolInput{}, err
	}
	var root map[string]any
	if err := json.Unmarshal([]byte(argumentsInJSON), &root); err != nil {
		return manifestToolInput{}, err
	}
	if _, exists := root["mode"]; exists {
		return manifestToolInput{}, fmt.Errorf("mode is not accepted by patch tools")
	}
	var raw manifestToolRawInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &raw); err != nil {
		return manifestToolInput{}, err
	}
	tasks, err := parseManifestTaskPatches(raw.Tasks)
	if err != nil {
		return manifestToolInput{}, err
	}
	return manifestToolInput{
		Title: raw.Title, ContentBank: raw.ContentBank, Sections: raw.Sections, Tasks: tasks,
	}, nil
}

func fixerToolError(err error) string {
	payload, _ := json.Marshal(map[string]any{"ok": false, "error": err.Error()})
	return string(payload)
}
