package tools

import (
	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/components/tool"

	"github.com/cloudwego/ppt-agent/pkg/assets/unsplash"
	imagetool "github.com/cloudwego/ppt-agent/pkg/tools/image"
	"github.com/cloudwego/ppt-agent/pkg/tools/search"
)

func NewSearchTool() tool.InvokableTool {
	return search.NewSearchTool()
}

// NewImageSearchTool creates the optional Unsplash-backed image search tool.
// The caller decides whether the provider is configured before registering it
// with an Agent.
func NewImageSearchTool(client *unsplash.Client, workDir string) tool.InvokableTool {
	return imagetool.NewImageSearchTool(client, workDir)
}

func NewReadFileTool(op commandline.Operator) tool.InvokableTool {
	return NewReadFileToolImpl(op)
}
