package db

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UpsertUserAPIKey stores or replaces the current user's model key override.
func UpsertUserAPIKey(userID uint, provider, apiKey string) error {
	if DB == nil {
		return fmt.Errorf("database unavailable")
	}
	record, updates := buildUserAPIKeyUpsert(userID, provider, apiKey, time.Now())
	return DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(updates),
	}).Create(&record).Error
}

func buildUserAPIKeyUpsert(userID uint, provider, apiKey string, now time.Time) (UserAPIKey, map[string]any) {
	provider = strings.TrimSpace(provider)
	if provider == "" {
		provider = "ark"
	}
	apiKey = strings.TrimSpace(apiKey)
	record := UserAPIKey{
		UserID:    userID,
		Provider:  provider,
		APIKey:    apiKey,
		CreatedAt: now,
		UpdatedAt: now,
	}
	updates := map[string]any{
		"provider":   provider,
		"api_key":    apiKey,
		"updated_at": now,
	}
	return record, updates
}

// GetUserAPIKey returns the configured key override, or nil if the account uses defaults.
func GetUserAPIKey(userID uint) (*UserAPIKey, error) {
	if DB == nil {
		return nil, nil
	}
	var key UserAPIKey
	err := DB.Where("user_id = ?", userID).First(&key).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &key, err
}

// DeleteUserAPIKey removes the user's model key override.
func DeleteUserAPIKey(userID uint) error {
	if DB == nil {
		return nil
	}
	return DB.Where("user_id = ?", userID).Delete(&UserAPIKey{}).Error
}
