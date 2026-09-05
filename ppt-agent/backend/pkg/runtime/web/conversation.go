package web

import (
	"strings"

	"github.com/cloudwego/ppt-agent/pkg/session"
)

// persistedConversationMessages prepares the rows loaded by task_id for an
// API response. The database is authoritative, so no task-record fallback or
// content-based de-duplication is applied: two identical user messages are
// still two distinct turns.
func persistedConversationMessages(messages []session.Message) []session.Message {
	result := make([]session.Message, 0, len(messages))
	for _, message := range messages {
		message.Content = strings.TrimSpace(message.Content)
		if message.Content == "" || message.Content == "..." || message.Content == "……" {
			continue
		}
		result = append(result, message)
	}
	return result
}
