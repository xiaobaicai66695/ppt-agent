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
