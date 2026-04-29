package ppt

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/agent/command"
)

var pptToolInfo = &schema.ToolInfo{
	Name: "generate_ppt",
	Desc: `根据提供的内容一次性生成完整的PowerPoint文件。
支持完整的JSON设计计划，包括：
- 全局样式（配色方案、字体设置）
- 幻灯片列表（标题、内容、图片、表格等）
- 输出文件路径`,
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

	// 验证必要参数
	if len(req.Slides) == 0 {
		return formatError("至少需要一页幻灯片"), nil
	}

	if req.OutputPath == "" {
		return formatError("输出路径不能为空"), nil
	}

	// 获取工作目录
	var workDir string
	if op, ok := t.op.(*command.LocalOperator); ok {
		workDir = op.GetWorkDir(ctx)
	}
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

	// 构建完整的生成计划
	plan := map[string]interface{}{
		"global_style": req.GlobalStyle,
		"slides":       req.Slides,
		"output_path":  outputPath,
	}

	// 设置默认样式
	if req.GlobalStyle.SlideWidth == 0 {
		plan["global_style"] = GlobalStyle{
			SlideWidth:  960,
			SlideHeight: 540,
			ColorScheme: ColorScheme{
				Primary:       "#1B2A3A",
				PrimaryLight:  "#2A3B4D",
				Accent:        "#5899A8",
				Background:    "#FAF0E6",
				TextMain:      "#0F172A",
				TextSecondary: "#5A6A7A",
			},
			Typography: Typography{
				TitleFont: "Microsoft YaHei",
				BodyFont:  "Microsoft YaHei",
				TitleSize: 32,
				BodySize:  18,
			},
		}
	}

	// 写入临时 JSON 文件
	specFile := filepath.Join(workDir, "pptx_spec.json")
	specData, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return formatError("JSON序列化失败: " + err.Error()), nil
	}
	if err := os.WriteFile(specFile, specData, 0o644); err != nil {
		return formatError("写入规格文件失败: " + err.Error()), nil
	}

	// 查找 pptx_writer.py 脚本
	scriptPath := filepath.Join(workDir, "..", "..", "..", "skills", "pptx_writer", "scripts", "pptx_writer.py")

	// 执行 Python 脚本生成 PPT
	cmd := exec.Command("/root/pptx_env/bin/python", scriptPath, "--spec", specFile, "--output", outputPath)
	output, err := cmd.CombinedOutput()

	// 清理临时文件
	os.Remove(specFile)

	if err != nil {
		return formatError("生成PPT失败: " + string(output)), nil
	}

	// 解析结果
	var result PPTResponse
	if err := json.Unmarshal(output, &result); err != nil {
		return formatError("解析结果失败: " + err.Error()), nil
	}

	if !result.Success {
		return formatError(result.Error), nil
	}

	return formatSuccess(result), nil
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
