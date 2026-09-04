package tools

import (
	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/components/tool"

	"github.com/cloudwego/ppt-agent/pkg/assets/unsplash"
	imagetool "github.com/cloudwego/ppt-agent/pkg/tools/image"
	"github.com/cloudwego/ppt-agent/pkg/tools/search"
)

func NewSearchTool(opts ...search.Option) tool.InvokableTool {
	return search.NewSearchTool(opts...)
}

func WithSearchContentSummarizer(summarizer search.ContentSummarizer) search.Option {
	return search.WithContentSummarizer(summarizer)
}

// NewImageSearchTool is for project agents that need to present image
// candidates to a user. It must not be added to the PPTPlanner tool set.
func NewImageSearchTool(client *unsplash.Client) tool.InvokableTool {
	return imagetool.NewImageSearchTool(client)
}

func NewReadFileTool(op commandline.Operator) tool.InvokableTool {
	return NewReadFileToolImpl(op)
}
