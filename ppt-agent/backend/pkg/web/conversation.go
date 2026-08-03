package web

import (
	"strings"
	"time"

	"github.com/cloudwego/ppt-agent/pkg/session"
)

func conversationMessagesWithFallback(messages []session.Message, fullAnswer, conversationContent string, fallbackAt time.Time) []session.Message {
	result := make([]session.Message, 0, len(messages)+1)
	hasAssistant := false
	for _, message := range messages {
		message.Content = strings.TrimSpace(message.Content)
		if message.Content == "" || message.Content == "..." || message.Content == "……" {
			continue
		}
		if message.Role == "assistant" {
			hasAssistant = true
		}
		if len(result) > 0 {
			last := result[len(result)-1]
			if last.Role == message.Role && last.Content == message.Content {
				continue
			}
		}
		result = append(result, message)
	}
	if hasAssistant {
		return result
	}

	legacy := strings.TrimSpace(fullAnswer)
	if legacy == "" {
		legacy = strings.TrimSpace(conversationContent)
	}
	if legacy == "" {
		return result
	}
	if fallbackAt.IsZero() {
		fallbackAt = time.Now()
	}
	result = append(result, session.Message{
		Role: "assistant", Content: legacy, Timestamp: fallbackAt,
	})
	return result
}
