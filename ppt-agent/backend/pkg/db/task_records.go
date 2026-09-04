package db

func CreateTaskRecord(r *TaskRecord) error {
	return DB.Create(r).Error
}

// UpsertTaskRecord persists a task that may already have been created as a
// conversation. A conversation task is promoted in-place when the user later
// decides to generate a deck, so inserting a second row with the same task ID
// would otherwise fail and leave the durable task state stale.
func UpsertTaskRecord(r *TaskRecord) error {
	if DB == nil {
		return nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Save(r).Error
}

// UpdateTaskRecord 向可更新字段添加 conversation_content。
func UpdateTaskRecord(id string, updates map[string]any) error {
	return DB.Model(&TaskRecord{}).Where("id = ?", id).Updates(updates).Error
}

func GetTaskRecord(id string) (*TaskRecord, error) {
	var r TaskRecord
	err := DB.Where("id = ?", id).First(&r).Error
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func ListTaskRecordsByUser(userID uint) ([]TaskRecord, error) {
	var records []TaskRecord
	err := DB.Select("id", "user_id", "query", "status", "work_dir", "done_count", "total_count", "duration", "error", "prompt_tokens", "completion_tokens", "total_tokens", "files", "intent", "conversation_id", "source_message_id", "parent_task_id", "generation_started_at", "generation_finished_at", "generation_duration_ms", "fixer_run_count", "created_at", "updated_at").Where("user_id = ?", userID).Order("created_at DESC").Find(&records).Error
	return records, err
}

func ListAllTaskRecords(limit int) ([]TaskRecord, error) {
	var records []TaskRecord
	err := DB.Order("created_at DESC").Limit(limit).Find(&records).Error
	return records, err
}

func DeleteTaskRecord(id string) error {
	return DB.Where("id = ?", id).Delete(&TaskRecord{}).Error
}

// MarkZombieTasks 将所有状态为 "running" 的任务设置为 "failed"
// （在启动时调用 — 拥有这些任务的服务器进程已不存在）。
func MarkZombieTasks() error {
	return DB.Model(&TaskRecord{}).
		Where("status = ?", "running").
		Updates(map[string]any{
			"status": "failed",
			"error":  "服务器重启，任务中断",
		}).Error
}
