package db

func CreateRuntimeEvent(r *RuntimeEventRecord) error {
	if DB == nil || r == nil {
		return nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Create(r).Error
}

func ListRuntimeEvents(taskID string) ([]RuntimeEventRecord, error) {
	if DB == nil {
		return nil, nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var events []RuntimeEventRecord
	err := DB.WithContext(ctx).Where("task_id = ?", taskID).Order("event_id ASC").Find(&events).Error
	return events, err
}

func ListRuntimeEventSummaries(taskID string) ([]RuntimeEventRecord, error) {
	if DB == nil {
		return nil, nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var events []RuntimeEventRecord
	// Metadata is needed here so the web layer can derive a bounded, redacted
	// public summary for inline tool previews. Raw payloads are never returned by
	// the conversation endpoint.
	err := DB.WithContext(ctx).Select("id", "task_id", "event_id", "timestamp", "elapsed_ms", "kind", "phase", "name", "status", "detail", "metadata", "created_at").
		Where("task_id = ?", taskID).
		Order("event_id ASC").
		Find(&events).Error
	return events, err
}

func GetRuntimeEvent(taskID string, eventID int64) (*RuntimeEventRecord, error) {
	if DB == nil {
		return nil, nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var event RuntimeEventRecord
	err := DB.WithContext(ctx).Where("task_id = ? AND event_id = ?", taskID, eventID).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func DeleteRuntimeEvents(taskID string) error {
	if DB == nil {
		return nil
	}
	return DB.Where("task_id = ?", taskID).Delete(&RuntimeEventRecord{}).Error
}
