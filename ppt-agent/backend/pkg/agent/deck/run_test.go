package deck

import (
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"
)

func TestAssistantContentWithToolCallsIsEmittable(t *testing.T) {
	msg := schema.AssistantMessage("Thought: 需要读取模板\nAction: read_file", []schema.ToolCall{{
		ID:   "call-read-template",
		Type: "function",
		Function: schema.FunctionCall{
			Name:      "read_file",
			Arguments: `{"path":"/tmp/template.json"}`,
		},
	}})
	if !isChunkEmittable(msg) {
		t.Fatal("assistant content with tool_calls should remain visible")
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

func TestStreamTimeoutDefaultsToEightMinutes(t *testing.T) {
	t.Setenv("STREAM_TIMEOUT", "")

	if got := streamTimeout(); got != 8*time.Minute {
		t.Fatalf("streamTimeout() = %s, want 8m", got)
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

	if got := streamTimeout(); got != 8*time.Minute {
		t.Fatalf("streamTimeout() = %s, want 8m", got)
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

func TestADKStreamingEnabledDefaultsOff(t *testing.T) {
	t.Setenv("ADK_ENABLE_STREAMING", "")

	if adkStreamingEnabled() {
		t.Fatal("expected ADK streaming to be disabled by default")
	}
}

func TestADKStreamingEnabledOptIn(t *testing.T) {
	t.Setenv("ADK_ENABLE_STREAMING", "true")

	if !adkStreamingEnabled() {
		t.Fatal("expected ADK streaming to be enabled when ADK_ENABLE_STREAMING=true")
	}
}

func TestADKStreamingEnabledIgnoresOtherValues(t *testing.T) {
	for _, value := range []string{"false", "1", "yes"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("ADK_ENABLE_STREAMING", value)

			if adkStreamingEnabled() {
				t.Fatalf("expected ADK streaming to be disabled for %q", value)
			}
		})
	}
}
