package deck

import (
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"
)

func TestAssistantContentWithToolCallsIsNotEmittable(t *testing.T) {
	msg := schema.AssistantMessage("Thought: 需要读取模板\nAction: read_file", []schema.ToolCall{{
		ID:   "call-read-template",
		Type: "function",
		Function: schema.FunctionCall{
			Name:      "read_file",
			Arguments: `{"path":"/tmp/template.json"}`,
		},
	}})
	if isChunkEmittable(msg) {
		t.Fatal("assistant content with tool_calls should stay telemetry-only")
	}
}

func TestToolMessageContentIsNotEmittable(t *testing.T) {
	msg := schema.ToolMessage(`{"ok":true}`, "call-read-template")
	if isChunkEmittable(msg) {
		t.Fatal("tool observation content should stay telemetry-only")
	}
}

func TestVisibleMessageContentKeepsRawPlannerOutput(t *testing.T) {
	raw := `{"thought":"规划 3 页延安介绍 PPT：\n1. 封面页 (title_slide)"}`
	if got := visibleMessageContent(raw); got != raw {
		t.Fatalf("visibleMessageContent() = %q, want raw output", got)
	}
}

func TestVisibleMessageContentHidesCompressionSummaryJSON(t *testing.T) {
	raw := `{
  "user_request_summary": "制作一份关于2025年国际局势纷争不断的PPT，共5325页",
  "progress_summary": "已生成部分 DeckSpec",
  "conversation_summary": "内部压缩摘要"
}`
	if got := visibleMessageContent(raw); got != "" {
		t.Fatalf("visibleMessageContent() = %q, want hidden compression summary", got)
	}
}

func TestVisibleMessageContentHidesFencedCompressionSummaryJSON(t *testing.T) {
	raw := "```json\n{\"user_request_summary\":\"x\",\"progress_summary\":\"y\"}\n```"
	if got := visibleMessageContent(raw); got != "" {
		t.Fatalf("visibleMessageContent() = %q, want hidden compression summary", got)
	}
}

func TestStreamTimeoutDefaultsToFifteenMinutes(t *testing.T) {
	t.Setenv("STREAM_TIMEOUT", "")

	if got := streamTimeout(); got != 15*time.Minute {
		t.Fatalf("streamTimeout() = %s, want 15m", got)
	}
}

func TestStreamTimeoutUsesEnvValue(t *testing.T) {
	t.Setenv("STREAM_TIMEOUT", "30s")

	if got := streamTimeout(); got != 30*time.Second {
		t.Fatalf("streamTimeout() = %s, want 30s", got)
	}
}

func TestStreamTimeoutInvalidEnvFallsBack(t *testing.T) {
	t.Setenv("STREAM_TIMEOUT", "soon")

	if got := streamTimeout(); got != 15*time.Minute {
		t.Fatalf("streamTimeout() = %s, want 15m", got)
	}
}

func TestStreamTimeoutCanBeDisabled(t *testing.T) {
	for _, value := range []string{"0", "-1s"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("STREAM_TIMEOUT", value)

			if got := streamTimeout(); got > 0 {
				t.Fatalf("streamTimeout() = %s, want disabled timeout", got)
			}
		})
	}
}

func TestADKStreamingEnabledDefaultsOn(t *testing.T) {
	t.Setenv("ADK_ENABLE_STREAMING", "")

	if !adkStreamingEnabled() {
		t.Fatal("expected ADK streaming to be enabled by default")
	}
}

func TestADKStreamingEnabledAcceptsTruthyValues(t *testing.T) {
	for _, value := range []string{"true", "1", "yes", "on"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("ADK_ENABLE_STREAMING", value)

			if !adkStreamingEnabled() {
				t.Fatalf("expected ADK streaming to be enabled for %q", value)
			}
		})
	}
}

func TestADKStreamingEnabledCanBeDisabled(t *testing.T) {
	for _, value := range []string{"false", "0", "no", "off", "disabled"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("ADK_ENABLE_STREAMING", value)

			if adkStreamingEnabled() {
				t.Fatalf("expected ADK streaming to be disabled for %q", value)
			}
		})
	}
}
