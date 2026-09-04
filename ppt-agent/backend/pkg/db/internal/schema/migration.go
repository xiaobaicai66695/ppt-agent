// Package schema owns the database schema registration order.
package schema

import (
	"github.com/cloudwego/ppt-agent/pkg/db/internal/model"
	"gorm.io/gorm"
)

// Migrate applies all durable-store schemas in dependency order.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.UserAPIKey{},
		&model.VerificationCode{},
		&model.TaskRecord{},
		&model.TaskFeedback{},
		&model.PlanDraftRecord{},
		&model.ConversationMessage{},
		&model.ConversationMessageChunk{},
		&model.RuntimeEventRecord{},
		&model.TaskErrorAnalysis{},
	)
}
