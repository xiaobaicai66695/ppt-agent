package ppt

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/agent/command"
	"github.com/cloudwego/ppt-agent/pkg/params"
)

var pptToolInfo = &schema.ToolInfo{
	Name: "generate_ppt",
	Desc: `根据提供的内容一次性生成完整的PowerPoint文件。
支持完整的JSON设计计划，包括：
- 全局样式（配色方案、字体设置）
- 幻灯片列表（标题、内容、图片、表格等）
- 输出文件路径

底层通过 python3 工具执行 generators/ 中的模板脚本生成 PPT。`,
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"global_style": {
			Type:     "object",
			Desc:     "全局样式，包含幻灯片尺寸、配色方案、字体设置等",
			Required: true,
		},
		"slides": {
			Type:     "array",
			Desc:     "幻灯片列表，每个幻灯片包含布局类型、元素列表及样式",
			Required: true,
		},
		"output_path": {
			Type:     "string",
			Desc:     "输出文件路径",
			Required: true,
		},
	}),
}

type PPTRequest struct {
	GlobalStyle GlobalStyle `json:"global_style"`
	Slides      []Slide     `json:"slides"`
	OutputPath  string      `json:"output_path"`
}

type GlobalStyle struct {
	SlideWidth  int         `json:"slide_width"`
	SlideHeight int         `json:"slide_height"`
	ColorScheme ColorScheme `json:"color_scheme"`
	Typography  Typography  `json:"typography"`
}

type ColorScheme struct {
	Primary       string `json:"primary"`
	PrimaryLight  string `json:"primary_light"`
	Accent        string `json:"accent"`
	Background    string `json:"background"`
	TextMain      string `json:"text_main"`
	TextSecondary string `json:"text_secondary"`
}

type Typography struct {
	TitleFont string `json:"title_font"`
	BodyFont  string `json:"body_font"`
	TitleSize int    `json:"title_size"`
	BodySize  int    `json:"body_size"`
}

type Slide struct {
	Index          int                    `json:"index"`
	Type           string                 `json:"type"`
	Layout         string                 `json:"layout"`
	Content        map[string]interface{} `json:"content"`
	Style          map[string]interface{} `json:"style"`
	VisualElements []string               `json:"visual_elements"`
}

type PPTResponse struct {
	Success    bool     `json:"success"`
	FilePath   string   `json:"file_path"`
	SlideCount int      `json:"slide_count"`
	Message    string   `json:"message"`
	Warnings   []string `json:"warnings,omitempty"`
	Error      string   `json:"error,omitempty"`
}

func NewPPTTool(op commandline.Operator) tool.InvokableTool {
	return &pptTool{op: op}
}

type pptTool struct {
	op commandline.Operator
}

func (t *pptTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return pptToolInfo, nil
}

func (t *pptTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var req PPTRequest
	if err := json.Unmarshal([]byte(argumentsInJSON), &req); err != nil {
		return formatError("参数解析失败: " + err.Error()), nil
	}

	if len(req.Slides) == 0 {
		return formatError("至少需要一页幻灯片"), nil
	}

	if req.OutputPath == "" {
		return formatError("输出路径不能为空"), nil
	}

	// 获取工作目录
	workDir := t.getWorkDir(ctx)
	if workDir == "" {
		workDir, _ = os.Getwd()
	}

	// 准备输出路径
	outputPath := req.OutputPath
	if !filepath.IsAbs(outputPath) {
		outputPath = filepath.Join(workDir, outputPath)
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return formatError("创建输出目录失败: " + err.Error()), nil
	}

	// 构建 Python 代码并写入临时脚本
	pythonCode := buildPythonCode(workDir, outputPath, req)
	scriptFile := filepath.Join(workDir, "temp_ppt_script.py")
	if err := os.WriteFile(scriptFile, []byte(pythonCode), 0o644); err != nil {
		return formatError("写入脚本文件失败: " + err.Error()), nil
	}
	defer os.Remove(scriptFile)

	// 通过 commandline.Operator 执行（与 python_runner 相同的执行路径）
	cmd := []string{"/root/pptx_env/bin/python", scriptFile}
	result, err := t.op.RunCommand(ctx, cmd)
	if err != nil {
		return formatError("执行失败: " + err.Error()), nil
	}

	// 解析 stdout（Python 脚本应输出 JSON 结果）
	output := result.Stdout
	if result.Stderr != "" {
		output += "\n" + result.Stderr
	}

	var resp PPTResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		if strings.Contains(output, "Traceback") || strings.Contains(output, "Error") {
			return formatError("Python执行出错:\n" + output), nil
		}
		return formatError("解析结果失败: " + err.Error() + "\n原始输出:\n" + output), nil
	}

	if !resp.Success {
		return formatError(resp.Error), nil
	}

	return formatSuccess(resp), nil
}

// getWorkDir 获取工作目录（与 LocalOperator 逻辑一致）
func (t *pptTool) getWorkDir(ctx context.Context) string {
	if wd, ok := params.GetTypedContextParams[string](ctx, params.WorkDirSessionKey); ok {
		return wd
	}
	if op, ok := t.op.(*command.LocalOperator); ok {
		return op.GetWorkDir(ctx)
	}
	return ""
}

// buildPythonCode 根据请求构建完整的 Python 执行代码
func buildPythonCode(workDir string, outputPath string, req PPTRequest) string {
	// 将请求序列化为 JSON，注入 Python 代码
	specJSON, _ := json.Marshal(req.Slides)
	palette := "ocean_soft"

	if req.GlobalStyle.ColorScheme.Primary != "" {
		palette = "" // 用户提供了自定义配色
	}

	code := fmt.Sprintf(`
# -*- coding: utf-8 -*-
import sys
import json
import os
from pathlib import Path

# 添加 generators 包路径（work_dir=output/{taskID}/, 父目录是 skills/visual_designer）
work_dir = %q
generators_pkg_dir = (Path(work_dir) / ".." / ".." / "skills" / "visual_designer").resolve()
sys.path.insert(0, str(generators_pkg_dir))

from generators import (
    new_presentation,
    generate_title_slide,
    generate_section_divider,
    generate_content_slide,
    generate_stat_slide,
    generate_quote_slide,
    generate_card_grid,
    generate_timeline,
    generate_process_flow,
    generate_two_column,
    generate_three_column,
    generate_summary_slide,
    generate_image_text,
)

# 反序列化幻灯片数据
slides_data = json.loads(%q)

# 创建演示文稿
palette = %q
prs = new_presentation(palette=palette)

# 遍历每页数据，调用对应的生成器
for slide in slides_data:
    slide_type = slide.get("type", "content_slide")
    content = slide.get("content", {})
    style = slide.get("style", {})

    if slide_type == "title_slide":
        generate_title_slide(prs,
            palette=palette,
            title=content.get("title", ""),
            subtitle=content.get("subtitle", ""),
            author=content.get("author", ""),
            date=content.get("date", ""))
    elif slide_type == "section_divider":
        generate_section_divider(prs,
            palette=palette,
            number=content.get("number", "01"),
            title=content.get("title", ""),
            subtitle=content.get("subtitle", ""))
    elif slide_type == "content_slide":
        generate_content_slide(prs,
            palette=palette,
            title=content.get("title", ""),
            section_header=content.get("section_header", ""),
            bullets=content.get("bullets", []))
    elif slide_type == "stat_slide":
        generate_stat_slide(prs,
            palette=palette,
            title=content.get("title", ""),
            stats=content.get("stats", []))
    elif slide_type == "quote_slide":
        generate_quote_slide(prs,
            palette=palette,
            quote=content.get("quote", ""),
            attribution=content.get("attribution", ""))
    elif slide_type == "card_grid":
        generate_card_grid(prs,
            palette=palette,
            title=content.get("title", ""),
            layout=content.get("layout", "2x2"),
            cards=content.get("cards", []))
    elif slide_type == "timeline":
        generate_timeline(prs,
            palette=palette,
            title=content.get("title", ""),
            direction=content.get("direction", "horizontal"),
            nodes=content.get("nodes", []))
    elif slide_type == "process_flow":
        generate_process_flow(prs,
            palette=palette,
            title=content.get("title", ""),
            direction=content.get("direction", "horizontal_zigzag"),
            steps=content.get("steps", []))
    elif slide_type == "two_column":
        generate_two_column(prs,
            palette=palette,
            title=content.get("title", ""),
            left_header=content.get("left_header", ""),
            right_header=content.get("right_header", ""),
            left_bullets=content.get("left_bullets", []),
            right_bullets=content.get("right_bullets", []))
    elif slide_type == "three_column":
        generate_three_column(prs,
            palette=palette,
            title=content.get("title", ""),
            columns=content.get("columns", []))
    elif slide_type == "summary_slide":
        generate_summary_slide(prs,
            palette=palette,
            title=content.get("title", ""),
            key_points=content.get("key_points", []),
            thank_you=content.get("thank_you", ""),
            contact=content.get("contact", ""))
    elif slide_type == "image_text":
        generate_image_text(prs,
            palette=palette,
            title=content.get("title", ""),
            layout=content.get("layout", "right-image"),
            header=content.get("header", ""),
            bullets=content.get("bullets", []))
    else:
        # 兜底：默认使用 content_slide
        generate_content_slide(prs,
            palette=palette,
            title=content.get("title", ""),
            bullets=content.get("bullets", []))

# 保存
output_path = %q
os.makedirs(os.path.dirname(output_path), exist_ok=True)
prs.save(output_path)

# 输出结果 JSON
import json as _json
result = {
    "success": True,
    "file_path": output_path,
    "slide_count": len(prs.slides),
    "message": "PPT生成成功"
}
print(_json.dumps(result, ensure_ascii=False))
`, workDir, string(specJSON), palette, outputPath)

	return code
}

func formatSuccess(resp PPTResponse) string {
	resp.Message = fmt.Sprintf("PPT已生成，共%d页，保存至: %s", resp.SlideCount, resp.FilePath)
	data, _ := json.Marshal(resp)
	return string(data)
}

func formatError(err string) string {
	resp := PPTResponse{
		Success: false,
		Error:   err,
	}
	data, _ := json.Marshal(resp)
	return string(data)
}
