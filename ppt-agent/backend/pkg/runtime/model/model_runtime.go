package utils

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// BuildFallbackModel creates the configured model chain. The implementation
// remains in model.go for compatibility; this focused entry is the preferred
// name for new runtime code.
func BuildFallbackModel(ctx context.Context, opts ...ChatModelOption) (model.ToolCallingChatModel, error) {
	return NewFallbackToolCallingChatModel(ctx, opts...)
}

// NewTextModel creates the lightweight QA/text model used by helper agents.
func NewTextModel(ctx context.Context, opts ...ChatModelOption) (model.ToolCallingChatModel, error) {
	return NewQAModel(ctx, opts...)
}

// SanitizeToolCallStream keeps partial tool-call JSON safe for consumers.
func SanitizeToolCallStream(stream *schema.StreamReader[*schema.Message]) *schema.StreamReader[*schema.Message] {
	return sanitizeStreamingToolCallDeltas(stream)
}
