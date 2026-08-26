package utils

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/agent/modelcompat"
)

type fakeToolCallingChatModel struct {
	generate func(context.Context, []*schema.Message, ...model.Option) (*schema.Message, error)
	stream   func(context.Context, []*schema.Message, ...model.Option) (*schema.StreamReader[*schema.Message], error)
}

func (f fakeToolCallingChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	if f.generate != nil {
		return f.generate(ctx, messages, opts...)
	}
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

func TestResolveFallbackModelSpecsUsesLegacyArkConfig(t *testing.T) {
	clearModelEnv(t)
	t.Setenv("ARK_MODEL", "ark-primary")
	t.Setenv("ARK_MODEL_BACKUP1", "ark-backup")
	t.Setenv("ARK_API_KEY", "ark-key")
	t.Setenv("ARK_BASE_URL", "https://ark.example/v3")
	t.Setenv("ARK_REGION", "cn-test")

	specs, err := resolveFallbackModelSpecs(&ChatModelConfig{})
	if err != nil {
		t.Fatalf("resolveFallbackModelSpecs() error = %v", err)
	}
	if len(specs) != 2 {
		t.Fatalf("len(specs) = %d, want 2", len(specs))
	}
	if specs[0].Provider != modelcompat.ProviderArk || specs[0].Model != "ark-primary" {
		t.Fatalf("primary spec = %#v", specs[0])
	}
	if specs[0].APIKey != "ark-key" || specs[0].BaseURL != "https://ark.example/v3" || specs[0].Region != "cn-test" {
		t.Fatalf("primary env fields = %#v", specs[0])
	}
	if specs[1].Model != "ark-backup" {
		t.Fatalf("backup model = %q", specs[1].Model)
	}
}

func TestResolveFallbackModelSpecsUsesProviderAwareChain(t *testing.T) {
	clearModelEnv(t)
	t.Setenv("MODEL_CHAIN", "primary, backup1")
	t.Setenv("MODEL_PRIMARY_PROVIDER", "siliconflow")
	t.Setenv("MODEL_PRIMARY_NAME", "sf-model")
	t.Setenv("SILICONFLOW_API_KEY", "sf-key")
	t.Setenv("MODEL_SILICONFLOW_TIMEOUT_SECONDS", "35")
	t.Setenv("MODEL_BACKUP1_PROVIDER", "openai-compatible")
	t.Setenv("MODEL_BACKUP1_NAME", "compat-model")
	t.Setenv("MODEL_BACKUP1_API_KEY_ENV", "BACKUP_KEY")
	t.Setenv("BACKUP_KEY", "backup-key")
	t.Setenv("MODEL_BACKUP1_BASE_URL", "https://compat.example/v1")
	t.Setenv("MODEL_BACKUP1_TIMEOUT_SECONDS", "45")

	specs, err := resolveFallbackModelSpecs(&ChatModelConfig{})
	if err != nil {
		t.Fatalf("resolveFallbackModelSpecs() error = %v", err)
	}
	if len(specs) != 2 {
		t.Fatalf("len(specs) = %d, want 2", len(specs))
	}
	if specs[0].Provider != modelcompat.ProviderSiliconFlow || specs[0].BaseURL != modelcompat.SiliconFlowBaseURL {
		t.Fatalf("siliconflow spec = %#v", specs[0])
	}
	if specs[0].APIKey != "sf-key" || specs[0].Timeout != 35*time.Second {
		t.Fatalf("siliconflow key/timeout = %#v", specs[0])
	}
	if specs[1].Provider != modelcompat.ProviderOpenAICompat || specs[1].APIKey != "backup-key" {
		t.Fatalf("backup spec = %#v", specs[1])
	}
	if specs[1].BaseURL != "https://compat.example/v1" || specs[1].Timeout != 45*time.Second {
		t.Fatalf("backup base/timeout = %#v", specs[1])
	}
}

func TestAccountAPIKeyOnlyAppliesToMatchingProvider(t *testing.T) {
	clearModelEnv(t)
	t.Setenv("MODEL_CHAIN", "primary, backup1")
	t.Setenv("MODEL_PRIMARY_PROVIDER", "ark")
	t.Setenv("MODEL_PRIMARY_NAME", "ark-model")
	t.Setenv("ARK_API_KEY", "ark-env-key")
	t.Setenv("MODEL_BACKUP1_PROVIDER", "siliconflow")
	t.Setenv("MODEL_BACKUP1_NAME", "sf-model")
	t.Setenv("SILICONFLOW_API_KEY", "sf-env-key")

	cfg := &ChatModelConfig{}
	WithAPIKeyForProvider("siliconflow", " account-sf-key ")(cfg)
	specs, err := resolveFallbackModelSpecs(cfg)
	if err != nil {
		t.Fatalf("resolveFallbackModelSpecs() error = %v", err)
	}
	if specs[0].Provider != modelcompat.ProviderArk || specs[0].APIKey != "ark-env-key" {
		t.Fatalf("ark spec should keep env key, got %#v", specs[0])
	}
	if specs[1].Provider != modelcompat.ProviderSiliconFlow || specs[1].APIKey != "account-sf-key" {
		t.Fatalf("siliconflow spec should use account key, got %#v", specs[1])
	}
}

func TestResolveFallbackModelSpecsWithTextModelUsesProviderAwareText(t *testing.T) {
	clearModelEnv(t)
	t.Setenv("MODEL_TEXT_PROVIDER", "siliconflow")
	t.Setenv("MODEL_TEXT_NAME", "sf-text-model")
	t.Setenv("SILICONFLOW_API_KEY", "sf-key")

	cfg := &ChatModelConfig{}
	WithTextModel()(cfg)

	specs, err := resolveFallbackModelSpecs(cfg)
	if err != nil {
		t.Fatalf("resolveFallbackModelSpecs() error = %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("len(specs) = %d, want 1", len(specs))
	}
	if specs[0].Provider != modelcompat.ProviderSiliconFlow || specs[0].Model != "sf-text-model" {
		t.Fatalf("text spec = %#v", specs[0])
	}
}

func TestFallbackGlobalTrackerKeyUsesProviderIdentity(t *testing.T) {
	fallback := &FallbackChatModel{
		modelNames:   []string{"display"},
		rawNames:     []string{"same-model"},
		trackerNames: []string{"siliconflow:same-model"},
	}
	if got := fallback.globalTrackerKey(0); got != "siliconflow:same-model" {
		t.Fatalf("globalTrackerKey() = %q", got)
	}
}

func TestModelCallLimiterSerializesSameResource(t *testing.T) {
	limiter := &modelCallLimiter{slots: make(map[string]chan struct{})}
	ctx := context.Background()
	releaseFirst, err := limiter.acquire(ctx, "ark:model:key:one", 1)
	if err != nil {
		t.Fatalf("first acquire error = %v", err)
	}

	acquiredSecond := make(chan func(), 1)
	go func() {
		release, err := limiter.acquire(ctx, "ark:model:key:one", 1)
		if err != nil {
			t.Errorf("second acquire error = %v", err)
			return
		}
		acquiredSecond <- release
	}()

	select {
	case release := <-acquiredSecond:
		release()
		t.Fatal("second acquire should block while first slot is held")
	case <-time.After(50 * time.Millisecond):
	}

	releaseFirst()
	select {
	case release := <-acquiredSecond:
		release()
	case <-time.After(time.Second):
		t.Fatal("second acquire did not proceed after release")
	}
}

func TestModelCallLimiterSeparatesResources(t *testing.T) {
	limiter := &modelCallLimiter{slots: make(map[string]chan struct{})}
	ctx := context.Background()
	releaseFirst, err := limiter.acquire(ctx, "ark:model:key:one", 1)
	if err != nil {
		t.Fatalf("first acquire error = %v", err)
	}
	defer releaseFirst()

	releaseSecond, err := limiter.acquire(ctx, "siliconflow:model:key:two", 1)
	if err != nil {
		t.Fatalf("different resource acquire error = %v", err)
	}
	releaseSecond()
}

func TestFallbackGenerateRecordsAuditableModelRequestContext(t *testing.T) {
	meta := NewRuntimeMeta("task-1", t.TempDir())
	ctx := WithRuntimeMeta(context.Background(), meta)
	fallback := &FallbackChatModel{
		models: []model.ToolCallingChatModel{
			fakeToolCallingChatModel{generate: func(context.Context, []*schema.Message, ...model.Option) (*schema.Message, error) {
				return schema.AssistantMessage("ok", nil), nil
			}},
		},
		modelNames: []string{"siliconflow/sf-model"},
		profiles: []modelCallProfile{{
			Provider: "siliconflow",
			Model:    "sf-model",
			Timeout:  15 * time.Minute,
		}},
	}
	messages := []*schema.Message{
		schema.SystemMessage("system rules"),
		schema.UserMessage("生成一份 PPT"),
		schema.AssistantMessage("准备搜索", []schema.ToolCall{{
			ID:   "call-search",
			Type: "function",
			Function: schema.FunctionCall{
				Name:      "search",
				Arguments: `{"query":"test"}`,
			},
		}}),
		{Role: schema.Tool, ToolName: "search", ToolCallID: "call-search", Content: "result"},
	}

	if _, err := fallback.Generate(ctx, messages); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(meta.RecentEvents) == 0 {
		t.Fatal("expected runtime event")
	}
	event := meta.RecentEvents[len(meta.RecentEvents)-1]
	if event.Kind != "model_request" {
		t.Fatalf("event kind = %q, want model_request", event.Kind)
	}
	if event.Metadata["provider"] != "siliconflow" || event.Metadata["model"] != "sf-model" {
		t.Fatalf("provider/model metadata = %#v", event.Metadata)
	}
	counts, ok := event.Metadata["role_counts"].(map[string]any)
	if !ok {
		t.Fatalf("role_counts missing or wrong type: %#v", event.Metadata["role_counts"])
	}
	for _, role := range []string{"system", "user", "assistant", "tool"} {
		if counts[role] == nil {
			t.Fatalf("role %q missing from counts: %#v", role, counts)
		}
	}
	history, ok := event.Metadata["history"].([]any)
	if !ok || len(history) != 4 {
		t.Fatalf("history = %#v, want 4 messages", event.Metadata["history"])
	}
	if first, ok := history[0].(map[string]any); !ok || first["content"] != nil {
		t.Fatalf("system message should hide raw content: %#v", history[0])
	}
}

func clearModelEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"MODEL_CHAIN",
		"MODEL_PROVIDER",
		"MODEL_TIMEOUT_SECONDS",
		"MODEL_SILICONFLOW_TIMEOUT_SECONDS",
		"MODEL_PRIMARY_PROVIDER",
		"MODEL_PRIMARY_NAME",
		"MODEL_PRIMARY_MODEL",
		"MODEL_PRIMARY_API_KEY",
		"MODEL_PRIMARY_API_KEY_ENV",
		"MODEL_PRIMARY_BASE_URL",
		"MODEL_PRIMARY_REGION",
		"MODEL_PRIMARY_TIMEOUT_SECONDS",
		"MODEL_BACKUP1_PROVIDER",
		"MODEL_BACKUP1_NAME",
		"MODEL_BACKUP1_MODEL",
		"MODEL_BACKUP1_API_KEY",
		"MODEL_BACKUP1_API_KEY_ENV",
		"MODEL_BACKUP1_BASE_URL",
		"MODEL_BACKUP1_REGION",
		"MODEL_BACKUP1_TIMEOUT_SECONDS",
		"MODEL_TEXT_PROVIDER",
		"MODEL_TEXT_NAME",
		"MODEL_TEXT_MODEL",
		"MODEL_TEXT_API_KEY",
		"MODEL_TEXT_API_KEY_ENV",
		"MODEL_TEXT_BASE_URL",
		"MODEL_TEXT_REGION",
		"MODEL_TEXT_TIMEOUT_SECONDS",
		"ARK_MODEL",
		"ARK_MODEL_BACKUP1",
		"ARK_MODEL_BACKUP2",
		"ARK_MODEL_BACKUP3",
		"ARK_MODEL_BACKUP4",
		"ARK_TEXT_MODEL",
		"ARK_QA_MODEL",
		"ARK_API_KEY",
		"ARK_BASE_URL",
		"ARK_REGION",
		"OPENAI_API_KEY",
		"OPENAI_BASE_URL",
		"SILICONFLOW_API_KEY",
		"SILICONFLOW_BASE_URL",
		"BACKUP_KEY",
	}
	for _, key := range keys {
		t.Setenv(key, "")
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
