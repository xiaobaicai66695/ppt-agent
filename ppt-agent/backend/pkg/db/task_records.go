package db

func CreateTaskRecord(r *TaskRecord) error {
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Create(r).Error
}

// UpsertTaskRecord persists a task that may already have been created as a
// conversation. A conversation task is promoted in-place when the user later
// decides to generate a ppt, so inserting a second row with the same task ID
// would otherwise fail and leave the durable task state stale.
func UpsertTaskRecord(r *TaskRecord) error {
	if DB == nil {
		return nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Save(r).Error
}

// UpdateTaskRecord 更新任务元数据；会话正文由 conversation_messages 独立维护。
func UpdateTaskRecord(id string, updates map[string]any) error {
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Model(&TaskRecord{}).Where("id = ?", id).Updates(updates).Error
}

func GetTaskRecord(id string) (*TaskRecord, error) {
	var r TaskRecord
	ctx, cancel := withOperationTimeout()
	defer cancel()
	err := DB.WithContext(ctx).Where("id = ?", id).First(&r).Error
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func ListTaskRecordsByUser(userID uint) ([]TaskRecord, error) {
	var records []TaskRecord
	ctx, cancel := withOperationTimeout()
	defer cancel()
	err := DB.WithContext(ctx).Select("id", "user_id", "query", "status", "work_dir", "done_count", "total_count", "duration", "error", "prompt_tokens", "completion_tokens", "total_tokens", "files", "intent", "conversation_id", "source_message_id", "parent_task_id", "generation_started_at", "generation_finished_at", "generation_duration_ms", "fixer_run_count", "created_at", "updated_at").Where("user_id = ?", userID).Order("created_at DESC").Find(&records).Error
	return records, err
}

func ListAllTaskRecords(limit int) ([]TaskRecord, error) {
	var records []TaskRecord
	ctx, cancel := withOperationTimeout()
	defer cancel()
	err := DB.WithContext(ctx).Order("created_at DESC").Limit(limit).Find(&records).Error
	return records, err
}

func DeleteTaskRecord(id string) error {
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Where("id = ?", id).Delete(&TaskRecord{}).Error
}

// MarkZombieTasks 将所有状态为 "running" 的任务设置为 "failed"
// （在启动时调用 — 拥有这些任务的服务器进程已不存在）。
func MarkZombieTasks() error {
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Model(&TaskRecord{}).
		Where("status = ?", "running").
		Updates(map[string]any{
			"status": "failed",
			"error":  "服务器重启，任务中断",
		}).Error
}
