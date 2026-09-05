package db

func CreateConversationMessage(m *ConversationMessage) error {
	if DB == nil {
		return nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Create(m).Error
}

func ListConversationMessages(taskID string) ([]ConversationMessage, error) {
	if DB == nil {
		return nil, nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var msgs []ConversationMessage
	// Timestamp is useful for chronology, but can collide for messages created
	// in the same clock tick. ID is the monotonic tie-breaker.
	err := DB.WithContext(ctx).Where("task_id = ?", taskID).Order("timestamp ASC").Order("id ASC").Find(&msgs).Error
	return msgs, err
}

func DeleteConversationMessages(taskID string) error {
	if DB == nil {
		return nil
	}
	return DB.Where("task_id = ?", taskID).Delete(&ConversationMessage{}).Error
}
