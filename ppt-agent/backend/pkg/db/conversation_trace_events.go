package db

func CreateConversationTraceEvent(event *ConversationTraceEvent) error {
	if DB == nil || event == nil {
		return nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Create(event).Error
}

func ListConversationTraceEvents(taskID string) ([]ConversationTraceEvent, error) {
	if DB == nil {
		return nil, nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var events []ConversationTraceEvent
	err := DB.WithContext(ctx).Where("task_id = ?", taskID).Order("timestamp ASC").Order("id ASC").Find(&events).Error
	return events, err
}

func DeleteConversationTraceEvents(taskID string) error {
	if DB == nil {
		return nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Where("task_id = ?", taskID).Delete(&ConversationTraceEvent{}).Error
}
