package db

import (
	"strings"

	"gorm.io/gorm"
)

func CreateConversationMessage(m *ConversationMessage) error {
	if DB == nil {
		return nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	chunks := splitConversationContent(m.Content, 16*1024)
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(chunks) > 1 {
			m.Content = ""
		}
		if err := tx.Create(m).Error; err != nil {
			return err
		}
		if len(chunks) <= 1 {
			return nil
		}
		records := make([]ConversationMessageChunk, 0, len(chunks))
		for index, content := range chunks {
			records = append(records, ConversationMessageChunk{MessageID: m.ID, Sequence: index, Content: content})
		}
		return tx.Create(&records).Error
	})
}

func ListConversationMessages(taskID string) ([]ConversationMessage, error) {
	if DB == nil {
		return nil, nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var msgs []ConversationMessage
	err := DB.WithContext(ctx).Preload("Chunks", func(tx *gorm.DB) *gorm.DB { return tx.Order("sequence ASC") }).Where("task_id = ?", taskID).Order("timestamp ASC").Find(&msgs).Error
	for index := range msgs {
		if len(msgs[index].Chunks) == 0 {
			continue
		}
		var content strings.Builder
		for _, chunk := range msgs[index].Chunks {
			content.WriteString(chunk.Content)
		}
		msgs[index].Content = content.String()
	}
	return msgs, err
}

func DeleteConversationMessages(taskID string) error {
	if DB == nil {
		return nil
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("message_id IN (SELECT id FROM conversation_messages WHERE task_id = ?)", taskID).Delete(&ConversationMessageChunk{}).Error; err != nil {
			return err
		}
		return tx.Where("task_id = ?", taskID).Delete(&ConversationMessage{}).Error
	})
}

func splitConversationContent(content string, maxBytes int) []string {
	if maxBytes <= 0 || len(content) <= maxBytes {
		return []string{content}
	}
	chunks := make([]string, 0, len(content)/maxBytes+1)
	start, size := 0, 0
	for index, char := range content {
		width := len(string(char))
		if size > 0 && size+width > maxBytes {
			chunks = append(chunks, content[start:index])
			start, size = index, 0
		}
		size += width
	}
	if start < len(content) {
		chunks = append(chunks, content[start:])
	}
	return chunks
}
