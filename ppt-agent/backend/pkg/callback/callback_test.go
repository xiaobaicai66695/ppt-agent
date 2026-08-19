package callback

import (
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestLastMessagePreviewSkipsAgentProgressMessages(t *testing.T) {
	messages := []*schema.Message{
		{Role: schema.User, Content: "介绍延安"},
		{Role: schema.Assistant, Content: "开始规划"},
		{Role: schema.User, Content: "<agent_progress>\ngenerated_slides: 0\ntotal_slides: 0\n</agent_progress>"},
	}

	if got := lastMessagePreview(messages, string(schema.User), 180); got != "介绍延安" {
		t.Fatalf("lastMessagePreview() = %q, want real user query", got)
	}
}

func TestCompactModelMessageKeepsAuditableContentAndToolArguments(t *testing.T) {
	message := schema.AssistantMessage("Thought: 先读取模板\nAction: read_file", []schema.ToolCall{{
		ID:   "call-read-template",
		Type: "function",
		Function: schema.FunctionCall{
			Name:      "read_file",
			Arguments: `{"path":"/tmp/template.json","reason":"inspect contract"}`,
		},
	}})

	got := compactModelMessage(message, 0)
	if got["content"] != "Thought: 先读取模板\nAction: read_file" {
		t.Fatalf("assistant content missing: %#v", got)
	}
	calls, ok := got["tool_call_details"].([]any)
	if !ok || len(calls) != 1 {
		t.Fatalf("tool calls missing: %#v", got)
	}
	call, ok := calls[0].(map[string]any)
	if !ok || call["arguments"] != `{"path":"/tmp/template.json","reason":"inspect contract"}` {
		t.Fatalf("tool arguments missing: %#v", calls[0])
	}
}

func TestCompactModelMessageHidesSystemContent(t *testing.T) {
	got := compactModelMessage(schema.SystemMessage("private system prompt"), 0)
	if got["content"] != nil {
		t.Fatalf("system content should not be exposed: %#v", got)
	}
	if got["content_preview"] == "" {
		t.Fatalf("system preview marker missing: %#v", got)
	}
}
