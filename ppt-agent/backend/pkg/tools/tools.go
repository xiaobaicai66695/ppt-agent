package tools

import (
	"context"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"

	"github.com/cloudwego/ppt-agent/pkg/tools/ppt"
	"github.com/cloudwego/ppt-agent/pkg/tools/qa"
	"github.com/cloudwego/ppt-agent/pkg/tools/search"
)

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

// NewSingleQATool 创建一个单页 QA 视觉审查工具。
// modelFn 用于创建支持多模态（图片输入）的 LLM，通常是视觉模型。
func NewSingleQATool(op commandline.Operator, modelFn func(ctx context.Context) (model.ToolCallingChatModel, error)) tool.InvokableTool {
	return qa.NewSingleTool(op, modelFn)
}
