// Package schema owns the database schema registration order.
package schema

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/cloudwego/ppt-agent/pkg/db/internal/model"
	"gorm.io/gorm"
)

// Migrate applies all durable-store schemas in dependency order.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.User{},
		&model.UserAPIKey{},
		&model.TaskRecord{},
		&model.TaskFeedback{},
		&model.PlanDraftRecord{},
		&model.ConversationMessage{},
		&model.RuntimeEventRecord{},
		&model.TaskErrorAnalysis{},
	); err != nil {
		return err
	}
	return migrateLegacyConversationStorage(db)
}

// migrateLegacyConversationStorage makes conversation_messages the only
// durable transcript source. It is deliberately forward-only: existing chunk
// rows and task-record copies are materialized into complete message rows
// before their obsolete schema is removed.
func migrateLegacyConversationStorage(db *gorm.DB) error {
	if db.Migrator().HasTable("conversation_message_chunks") {
		if err := mergeLegacyConversationChunks(db); err != nil {
			return err
		}
		if err := db.Migrator().DropTable("conversation_message_chunks"); err != nil {
			return err
		}
	}
	if err := backfillLegacyTaskMessages(db); err != nil {
		return err
	}
	for _, column := range []string{"conversation_content", "full_answer", "assistant_turns"} {
		if db.Migrator().HasColumn(&model.TaskRecord{}, column) {
			if err := db.Exec("ALTER TABLE task_records DROP COLUMN " + column).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func mergeLegacyConversationChunks(db *gorm.DB) error {
	type legacyChunk struct {
		MessageID uint
		Content   string
	}
	var chunks []legacyChunk
	if err := db.Table("conversation_message_chunks").Select("message_id, content").Order("message_id ASC").Order("sequence ASC").Find(&chunks).Error; err != nil {
		return err
	}
	merged := make(map[uint]*strings.Builder)
	for _, chunk := range chunks {
		if merged[chunk.MessageID] == nil {
			merged[chunk.MessageID] = &strings.Builder{}
		}
		merged[chunk.MessageID].WriteString(chunk.Content)
	}
	return db.Transaction(func(tx *gorm.DB) error {
		for messageID, content := range merged {
			if err := tx.Model(&model.ConversationMessage{}).Where("id = ?", messageID).Update("content", content.String()).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func backfillLegacyTaskMessages(db *gorm.DB) error {
	columns := []string{"id", "query", "created_at"}
	hasConversationContent := db.Migrator().HasColumn(&model.TaskRecord{}, "conversation_content")
	hasFullAnswer := db.Migrator().HasColumn(&model.TaskRecord{}, "full_answer")
	hasAssistantTurns := db.Migrator().HasColumn(&model.TaskRecord{}, "assistant_turns")
	if !hasConversationContent && !hasFullAnswer && !hasAssistantTurns {
		return nil
	}
	if hasConversationContent {
		columns = append(columns, "conversation_content")
	}
	if hasFullAnswer {
		columns = append(columns, "full_answer")
	}
	if hasAssistantTurns {
		columns = append(columns, "assistant_turns")
	}
	type legacyTask struct {
		ID                  string
		Query               string
		CreatedAt           time.Time
		ConversationContent string
		FullAnswer          string
		AssistantTurns      string
	}
	var tasks []legacyTask
	if err := db.Table("task_records").Select(strings.Join(columns, ", ")).Find(&tasks).Error; err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		for _, task := range tasks {
			var roles []struct {
				Role  string
				Count int64
			}
			if err := tx.Model(&model.ConversationMessage{}).Select("role, COUNT(*) AS count").Where("task_id = ?", task.ID).Group("role").Scan(&roles).Error; err != nil {
				return err
			}
			counts := map[string]int64{}
			for _, role := range roles {
				counts[role.Role] = role.Count
			}
			at := task.CreatedAt
			if at.IsZero() {
				at = time.Now()
			}
			if counts["user"] == 0 && strings.TrimSpace(task.Query) != "" {
				if err := tx.Create(&model.ConversationMessage{TaskID: task.ID, Role: "user", Content: task.Query, Timestamp: at}).Error; err != nil {
					return err
				}
			}
			if counts["assistant"] != 0 {
				continue
			}
			for index, content := range legacyAssistantMessages(task.AssistantTurns, task.FullAnswer, task.ConversationContent) {
				if err := tx.Create(&model.ConversationMessage{TaskID: task.ID, Role: "assistant", Content: content, Timestamp: at.Add(time.Duration(index+1) * time.Nanosecond)}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func legacyAssistantMessages(turnsJSON, fullAnswer, conversationContent string) []string {
	var turns []string
	if strings.TrimSpace(turnsJSON) != "" && json.Unmarshal([]byte(turnsJSON), &turns) == nil {
		messages := make([]string, 0, len(turns))
		for _, turn := range turns {
			if content := strings.TrimSpace(turn); content != "" {
				messages = append(messages, content)
			}
		}
		if len(messages) > 0 {
			return messages
		}
	}
	if content := strings.TrimSpace(fullAnswer); content != "" {
		return []string{content}
	}
	if content := strings.TrimSpace(conversationContent); content != "" {
		return []string{content}
	}
	return nil
}
