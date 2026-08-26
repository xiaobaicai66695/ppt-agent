package utils

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type fakeToolCallingChatModel struct {
	stream func(context.Context, []*schema.Message, ...model.Option) (*schema.StreamReader[*schema.Message], error)
}

func (f fakeToolCallingChatModel) Generate(context.Context, []*schema.Message, ...model.Option) (*schema.Message, error) {
	return nil, errors.New("Generate not implemented")
}

func (f fakeToolCallingChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return f.stream(ctx, messages, opts...)
}

func (f fakeToolCallingChatModel) WithTools([]*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	return f, nil
}

func streamWithError(err error) *schema.StreamReader[*schema.Message] {
	reader, writer := schema.Pipe[*schema.Message](1)
	go func() {
		defer writer.Close()
		writer.Send(nil, err)
	}()
	return reader
}

func streamWithChunkThenError(chunk *schema.Message, err error) *schema.StreamReader[*schema.Message] {
	reader, writer := schema.Pipe[*schema.Message](2)
	go func() {
		defer writer.Close()
		writer.Send(chunk, nil)
		writer.Send(nil, err)
	}()
	return reader
}

func TestModelAPIKeyFromConfigPrefersExplicitKey(t *testing.T) {
	t.Setenv("ARK_API_KEY", "env-key")

	cfg := &ChatModelConfig{}
	WithAPIKey(" user-key ")(cfg)

	if got := modelAPIKeyFromConfig(cfg); got != "user-key" {
		t.Fatalf("modelAPIKeyFromConfig() = %q, want user-key", got)
	}
}

func TestModelAPIKeyFromConfigFallsBackToEnv(t *testing.T) {
	t.Setenv("ARK_API_KEY", "env-key")

	if got := modelAPIKeyFromConfig(&ChatModelConfig{}); got != "env-key" {
		t.Fatalf("modelAPIKeyFromConfig() = %q, want env-key", got)
	}
}

func TestResolveModelAPIKeyPrefersAccountKey(t *testing.T) {
	if got := ResolveModelAPIKey(" account-key ", "env-key"); got != "account-key" {
		t.Fatalf("ResolveModelAPIKey() = %q, want account-key", got)
	}
}

func TestResolveModelAPIKeyFallsBackToEnvironment(t *testing.T) {
	if got := ResolveModelAPIKey("  ", " env-key "); got != "env-key" {
		t.Fatalf("ResolveModelAPIKey() = %q, want env-key", got)
	}
}

func TestFallbackStreamRetriesBackupWhenPrimaryReadFailsBeforeOutput(t *testing.T) {
	readErr := errors.New("context deadline exceeded while reading body")
	fallback := &FallbackChatModel{
		models: []model.ToolCallingChatModel{
			fakeToolCallingChatModel{stream: func(context.Context, []*schema.Message, ...model.Option) (*schema.StreamReader[*schema.Message], error) {
				return streamWithError(readErr), nil
			}},
			fakeToolCallingChatModel{stream: func(context.Context, []*schema.Message, ...model.Option) (*schema.StreamReader[*schema.Message], error) {
				return schema.StreamReaderFromArray([]*schema.Message{schema.AssistantMessage("backup ok", nil)}), nil
			}},
		},
		modelNames: []string{"primary", "backup"},
	}

	stream, err := fallback.Stream(context.Background(), []*schema.Message{schema.UserMessage("生成 PPT")})
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}
	defer stream.Close()

	msg, err := stream.Recv()
	if err != nil {
		t.Fatalf("Recv() error = %v", err)
	}
	if msg.Content != "backup ok" {
		t.Fatalf("content = %q, want backup ok", msg.Content)
	}
	if _, err := stream.Recv(); !errors.Is(err, io.EOF) {
		t.Fatalf("final Recv() error = %v, want EOF", err)
	}
}

func TestFallbackStreamDoesNotReplayAfterPartialOutput(t *testing.T) {
	readErr := errors.New("context deadline exceeded while reading body")
	fallback := &FallbackChatModel{
		models: []model.ToolCallingChatModel{
			fakeToolCallingChatModel{stream: func(context.Context, []*schema.Message, ...model.Option) (*schema.StreamReader[*schema.Message], error) {
				return streamWithChunkThenError(schema.AssistantMessage("partial", nil), readErr), nil
			}},
			fakeToolCallingChatModel{stream: func(context.Context, []*schema.Message, ...model.Option) (*schema.StreamReader[*schema.Message], error) {
				return schema.StreamReaderFromArray([]*schema.Message{schema.AssistantMessage("backup should not replay", nil)}), nil
			}},
		},
		modelNames: []string{"primary", "backup"},
	}

	stream, err := fallback.Stream(context.Background(), []*schema.Message{schema.UserMessage("生成 PPT")})
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}
	defer stream.Close()

	msg, err := stream.Recv()
	if err != nil {
		t.Fatalf("first Recv() error = %v", err)
	}
	if msg.Content != "partial" {
		t.Fatalf("content = %q, want partial", msg.Content)
	}
	if _, err := stream.Recv(); !errors.Is(err, readErr) {
		t.Fatalf("second Recv() error = %v, want readErr", err)
	}
}

func TestSanitizeStreamingToolCallDeltasKeepsTextChunksLive(t *testing.T) {
	input := schema.StreamReaderFromArray([]*schema.Message{
		schema.AssistantMessage("正在读取模板", nil),
		schema.AssistantMessage("并规划页面", nil),
	})
	stream := sanitizeStreamingToolCallDeltas(input)
	defer stream.Close()

	first, err := stream.Recv()
	if err != nil {
		t.Fatalf("first Recv() error = %v", err)
	}
	if first.Content != "正在读取模板" {
		t.Fatalf("first content = %q", first.Content)
	}

	second, err := stream.Recv()
	if err != nil {
		t.Fatalf("second Recv() error = %v", err)
	}
	if second.Content != "并规划页面" {
		t.Fatalf("second content = %q", second.Content)
	}

	if _, err := stream.Recv(); !errors.Is(err, io.EOF) {
		t.Fatalf("final Recv() error = %v, want EOF", err)
	}
}

func TestSanitizeStreamingToolCallDeltasDefersPartialToolCalls(t *testing.T) {
	idx := 0
	input := schema.StreamReaderFromArray([]*schema.Message{
		schema.AssistantMessage("准备写入规划", nil),
		{
			Role: schema.Assistant,
			ToolCalls: []schema.ToolCall{{
				Index: &idx,
				ID:    "call_1",
				Type:  "function",
				Function: schema.FunctionCall{
					Name:      "update_tasks_manifest",
					Arguments: `{"theme":"`,
				},
			}},
		},
		{
			ToolCalls: []schema.ToolCall{{
				Index: &idx,
				Function: schema.FunctionCall{
					Arguments: `ocean_soft"}`,
				},
			}},
		},
	})
	stream := sanitizeStreamingToolCallDeltas(input)
	defer stream.Close()

	first, err := stream.Recv()
	if err != nil {
		t.Fatalf("first Recv() error = %v", err)
	}
	if first.Content != "准备写入规划" || len(first.ToolCalls) != 0 {
		t.Fatalf("first chunk = %#v", first)
	}

	second, err := stream.Recv()
	if err != nil {
		t.Fatalf("second Recv() error = %v", err)
	}
	if second.Content != "" {
		t.Fatalf("deferred tool call chunk duplicated visible content: %q", second.Content)
	}
	if len(second.ToolCalls) != 1 {
		t.Fatalf("tool calls = %#v, want one merged call", second.ToolCalls)
	}
	tc := second.ToolCalls[0]
	if tc.Function.Name != "update_tasks_manifest" || tc.Function.Arguments != `{"theme":"ocean_soft"}` {
		t.Fatalf("merged tool call = %#v", tc)
	}

	if _, err := stream.Recv(); !errors.Is(err, io.EOF) {
		t.Fatalf("final Recv() error = %v, want EOF", err)
	}
}

func TestSanitizeStreamingToolCallDeltasFlushesCompletedToolCallBeforeLaterText(t *testing.T) {
	idx := 0
	input := schema.StreamReaderFromArray([]*schema.Message{
		schema.AssistantMessage("准备搜索图片", nil),
		{
			Role: schema.Assistant,
			ToolCalls: []schema.ToolCall{{
				Index: &idx,
				ID:    "call_1",
				Type:  "function",
				Function: schema.FunctionCall{
					Name:      "search_images",
					Arguments: `{"query":"world map"`,
				},
			}},
		},
		{
			ToolCalls: []schema.ToolCall{{
				Index: &idx,
				Function: schema.FunctionCall{
					Arguments: `}`,
				},
			}},
		},
		schema.AssistantMessage("继续说明下一步", nil),
	})
	stream := sanitizeStreamingToolCallDeltas(input)
	defer stream.Close()

	first, err := stream.Recv()
	if err != nil {
		t.Fatalf("first Recv() error = %v", err)
	}
	if first.Content != "准备搜索图片" {
		t.Fatalf("first content = %q", first.Content)
	}

	second, err := stream.Recv()
	if err != nil {
		t.Fatalf("second Recv() error = %v", err)
	}
	if len(second.ToolCalls) != 1 {
		t.Fatalf("second chunk tool calls = %#v, want one completed call", second.ToolCalls)
	}
	if second.ToolCalls[0].Function.Arguments != `{"query":"world map"}` {
		t.Fatalf("completed tool arguments = %q", second.ToolCalls[0].Function.Arguments)
	}

	third, err := stream.Recv()
	if err != nil {
		t.Fatalf("third Recv() error = %v", err)
	}
	if third.Content != "继续说明下一步" {
		t.Fatalf("third content = %q", third.Content)
	}

	if _, err := stream.Recv(); !errors.Is(err, io.EOF) {
		t.Fatalf("final Recv() error = %v, want EOF", err)
	}
}
