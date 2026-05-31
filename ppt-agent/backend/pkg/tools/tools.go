package tools

import (
	"context"
	"regexp"
	"strings"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"

	"github.com/cloudwego/ppt-agent/pkg/tools/ppt"
	"github.com/cloudwego/ppt-agent/pkg/tools/qa"
	"github.com/cloudwego/ppt-agent/pkg/tools/search"
)

// SanitizeJSON 修复常见的 LLM JSON 错误后再进行反序列化。
// 处理：尾随逗号、多余的花括号、双重转义的引号。
func SanitizeJSON(raw string) string {
	raw = strings.TrimSpace(raw)

	// 1. 移除闭合花括号/方括号前的尾随逗号
	re := regexp.MustCompile(`,(\s*[}\]])`)
	raw = re.ReplaceAllString(raw, "$1")

	// 2. 修复 JSON 中的双重转义 JSON 字符串
	// LLM 有时输出：{"content":"{\"title\":...}"} 这是正确的，
	// 但有时：{"content":"{"title":...}"} 这是损坏的
	// 这由 edit_file.go 中的 fixContentObject 处理

	// 3. 移除第一个 { 或 [ 之前的任何文本以及最后一个 } 或 ] 之后的任何文本
	start := strings.IndexAny(raw, "{[")
	end := strings.LastIndexAny(raw, "}]")
	if start >= 0 && end > start {
		raw = raw[start : end+1]
	}

	return raw
}

func NewSearchTool() tool.InvokableTool {
	return search.NewSearchTool()
}

func NewPPTTool(op commandline.Operator) tool.InvokableTool {
	return ppt.NewPPTTool(op)
}

func NewToolSubmitResult() *ToolSubmitResult {
	return &ToolSubmitResult{}
}

// 代码生成工具 - 用于直接生成 PPT
func NewBashTool(op commandline.Operator) tool.InvokableTool {
	return NewBashToolImpl(op)
}

func NewEditFileTool(op commandline.Operator) tool.InvokableTool {
	return NewEditFileToolImpl(op)
}

func NewReadFileTool(op commandline.Operator) tool.InvokableTool {
	return NewReadFileToolImpl(op)
}

func NewPythonRunnerTool(op commandline.Operator) tool.InvokableTool {
	return NewPythonRunnerToolImpl(op)
}

// NewBatchPDFTool 创建一个批量 PDF QA 视觉审查工具。
// 将所有幻灯片合并为 PDF 后，一次发给多模态 LLM 审查（一次调用替代 N 次）。
func NewBatchPDFTool(op commandline.Operator, modelFn func(ctx context.Context) (model.ToolCallingChatModel, error)) tool.InvokableTool {
	return qa.NewBatchPDFTool(op, modelFn)
}
