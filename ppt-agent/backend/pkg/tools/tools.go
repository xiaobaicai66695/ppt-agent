package tools

import (
	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/components/tool"

	"github.com/cloudwego/ppt-agent/pkg/tools/ppt"
	"github.com/cloudwego/ppt-agent/pkg/tools/search"
)

func NewSearchTool() tool.InvokableTool {
	return search.NewSearchTool()
}

func NewImageSearchTool() tool.InvokableTool {
	return search.NewImageSearchTool()
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
