package utils

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

func TestChatModelCompressorEmitsVisibleRuntimePhase(t *testing.T) {
	inner := fakeToolCallingChatModel{generate: func(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
		if len(messages) >= 4 {
			t.Fatalf("inner received uncompressed messages: %d", len(messages))
		}
		return schema.AssistantMessage("ok", nil), nil
	}}
	summarizer := fakeToolCallingChatModel{generate: func(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
		return schema.AssistantMessage(`{"user_intent_summary":"生成PPT","preserved_requirements":["保持最新要求"],"progress_summary":"已压缩","conversation_summary":"旧上下文"}`, nil), nil
	}}
	meta := NewRuntimeMeta("task-compress", t.TempDir())
	var stages []string
	compressor := NewChatModelCompressor(inner, summarizer,
		WithCompressThreshold(1),
		WithTokenThreshold(999999),
		WithPreserveCount(1),
	)
	compressor.SetRuntimeMeta(meta)
	compressor.SetCompressionEventCallback(func(event CompressionEvent) {
		stages = append(stages, event.Stage)
	})
	_, err := compressor.Generate(context.Background(), []*schema.Message{
		schema.SystemMessage("system"),
		schema.UserMessage("用户最初需求"),
		schema.AssistantMessage("旧回复", nil),
		schema.UserMessage("用户最新问题"),
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(stages) != 2 || stages[0] != "start" || stages[1] != "success" {
		t.Fatalf("stages = %#v, want start/success", stages)
	}
	snap := meta.Snapshot()
	if snap.Phase != "compressing_context" || snap.CompressionBeforeMessages == 0 || snap.CompressionAfterMessages == 0 {
		t.Fatalf("snapshot did not record compression phase/details: %#v", snap)
	}
}
