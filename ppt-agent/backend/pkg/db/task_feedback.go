package db

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func UpsertTaskFeedback(taskID string, userID uint, rating int, suggestion string) (*TaskFeedback, error) {
	if DB == nil {
		return nil, fmt.Errorf("database unavailable")
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	now := time.Now()
	record := TaskFeedback{TaskID: taskID, UserID: userID, Rating: rating, Suggestion: suggestion, CreatedAt: now, UpdatedAt: now}
	if err := DB.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "task_id"}, {Name: "user_id"}}, DoUpdates: clause.Assignments(map[string]any{"rating": rating, "suggestion": suggestion, "updated_at": now})}).Create(&record).Error; err != nil {
		return nil, err
	}
	return GetTaskFeedback(taskID, userID)
}

func GetTaskFeedback(taskID string, userID uint) (*TaskFeedback, error) {
	if DB == nil {
		return nil, nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var record TaskFeedback
	err := DB.WithContext(ctx).Where("task_id = ? AND user_id = ?", taskID, userID).First(&record).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ListTaskFeedbackByUserAndTaskIDs avoids a feedback query per task row.
func ListTaskFeedbackByUserAndTaskIDs(userID uint, taskIDs []string) (map[string]TaskFeedback, error) {
	result := make(map[string]TaskFeedback)
	if DB == nil || userID == 0 || len(taskIDs) == 0 {
		return result, nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var records []TaskFeedback
	if err := DB.WithContext(ctx).Where("user_id = ? AND task_id IN ?", userID, taskIDs).Find(&records).Error; err != nil {
		return nil, err
	}
	for _, record := range records {
		result[record.TaskID] = record
	}
	return result, nil
}
