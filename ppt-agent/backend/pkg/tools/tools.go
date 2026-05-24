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

// SanitizeJSON fixes common LLM JSON errors before unmarshaling.
// Handles: trailing commas, extra braces, double-escaped quotes.
func SanitizeJSON(raw string) string {
	raw = strings.TrimSpace(raw)

	// 1. Remove trailing commas before closing braces/brackets
	re := regexp.MustCompile(`,(\s*[}\]])`)
	raw = re.ReplaceAllString(raw, "$1")

	// 2. Fix double-escaped JSON strings inside JSON
	// The LLM sometimes outputs: {"content":"{\"title\":...}"} which is correct,
	// but sometimes: {"content":"{"title":...}"} which is broken
	// This is handled by fixContentObject in edit_file.go

	// 3. Remove any text before first { or [ and after last } or ]
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
