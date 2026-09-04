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
		result = appendConversationMessage(result, message)
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

func appendConversationMessage(messages []session.Message, incoming session.Message) []session.Message {
	incoming.Content = strings.TrimSpace(incoming.Content)
	if incoming.Content == "" {
		return messages
	}
	incomingNormalized := normalizeConversationContent(incoming.Content)
	for i := range messages {
		existing := messages[i]
		if existing.Role != incoming.Role {
			continue
		}
		relation := duplicateContentRelation(normalizeConversationContent(existing.Content), incomingNormalized)
		if relation == duplicateSame || relation == duplicateExistingContainsIncoming {
			return messages
		}
		if relation == duplicateIncomingContainsExisting {
			if suffix, ok := cumulativeConversationSuffix(existing.Content, incoming.Content); ok {
				incoming.Content = suffix
				if incoming.Content == "" {
					return messages
				}
				incomingNormalized = normalizeConversationContent(incoming.Content)
				continue
			}
			messages[i] = incoming
			return messages
		}
	}
	return append(messages, incoming)
}

func normalizeConversationContent(content string) string {
	return strings.ToLower(strings.Join(strings.Fields(content), ""))
}

type duplicateRelation int

const (
	duplicateNone duplicateRelation = iota
	duplicateSame
	duplicateExistingContainsIncoming
	duplicateIncomingContainsExisting
)

func duplicateContentRelation(existing, incoming string) duplicateRelation {
	if existing == "" || incoming == "" {
		return duplicateNone
	}
	if existing == incoming {
		return duplicateSame
	}
	minLength := len(existing)
	maxLength := len(incoming)
	if minLength > maxLength {
		minLength, maxLength = maxLength, minLength
	}
	if minLength < 20 || float64(minLength)/float64(maxLength) < 0.18 {
		return duplicateNone
	}
	if strings.Contains(existing, incoming) {
		return duplicateExistingContainsIncoming
	}
	if strings.Contains(incoming, existing) {
		return duplicateIncomingContainsExisting
	}
	return duplicateNone
}

func cumulativeConversationSuffix(existing, incoming string) (string, bool) {
	existing = strings.TrimSpace(existing)
	incoming = strings.TrimSpace(incoming)
	if existing == "" || incoming == "" || !strings.HasPrefix(incoming, existing) {
		return "", false
	}
	return strings.TrimSpace(strings.TrimPrefix(incoming, existing)), true
}
