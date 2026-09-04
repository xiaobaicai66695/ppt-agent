package task

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/cloudwego/ppt-agent/pkg/db"
)

const MaxDeliveryFeedbackSuggestionRunes = 1000

func ValidateDeliveryFeedback(rating int, suggestion string) (string, error) {
	if rating < 1 || rating > 5 {
		return "", fmt.Errorf("评分必须为 1 到 5 分")
	}
	suggestion = strings.TrimSpace(suggestion)
	if utf8.RuneCountInString(suggestion) > MaxDeliveryFeedbackSuggestionRunes {
		return "", fmt.Errorf("建议不能超过 %d 个字符", MaxDeliveryFeedbackSuggestionRunes)
	}
	return suggestion, nil
}

func (tm *TaskManager) GetDeliveryFeedback(taskID string, userID int) (*DeliveryFeedback, error) {
	if userID <= 0 {
		return nil, nil
	}
	record, err := db.GetTaskFeedback(taskID, uint(userID))
	if err != nil || record == nil {
		return nil, err
	}
	return &DeliveryFeedback{Rating: record.Rating, Suggestion: record.Suggestion, UpdatedAt: record.UpdatedAt}, nil
}

func (tm *TaskManager) SaveDeliveryFeedback(taskID string, userID, rating int, suggestion string) (*DeliveryFeedback, error) {
	info := tm.GetTask(taskID)
	if info == nil {
		return nil, fmt.Errorf("task not found")
	}
	if info.Status != TaskStatusCompleted {
		return nil, fmt.Errorf("PPT 尚未交付，暂不能评分")
	}
	suggestion, err := ValidateDeliveryFeedback(rating, suggestion)
	if err != nil {
		return nil, err
	}
	record, err := db.UpsertTaskFeedback(taskID, uint(userID), rating, suggestion)
	if err != nil {
		return nil, err
	}
	return &DeliveryFeedback{Rating: record.Rating, Suggestion: record.Suggestion, UpdatedAt: record.UpdatedAt}, nil
}
