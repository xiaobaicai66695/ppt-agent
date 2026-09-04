package db

func CreateTaskErrorAnalysis(a *TaskErrorAnalysis) error {
	return DB.Create(a).Error
}

func GetTaskErrorAnalysis(taskID string) ([]TaskErrorAnalysis, error) {
	var records []TaskErrorAnalysis
	err := DB.Where("task_id = ?", taskID).Order("created_at DESC").Find(&records).Error
	return records, err
}

func ListRecentErrorAnalyses(limit int) ([]TaskErrorAnalysis, error) {
	var records []TaskErrorAnalysis
	err := DB.Order("created_at DESC").Limit(limit).Find(&records).Error
	return records, err
}

// DeleteErrorAnalysis 删除指定 ID 的日志分析记录。
func DeleteErrorAnalysis(id uint) error {
	return DB.Where("id = ?", id).Delete(&TaskErrorAnalysis{}).Error
}
