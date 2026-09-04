package db

import "github.com/cloudwego/ppt-agent/pkg/db/internal/model"

// The aliases keep the db package's public model contract stable while the
// persistence implementation is organized below internal/model.
type (
	User                     = model.User
	UserAPIKey               = model.UserAPIKey
	VerificationCode         = model.VerificationCode
	TaskRecord               = model.TaskRecord
	TaskFeedback             = model.TaskFeedback
	PlanDraftRecord          = model.PlanDraftRecord
	ConversationMessage      = model.ConversationMessage
	ConversationMessageChunk = model.ConversationMessageChunk
	TaskErrorAnalysis        = model.TaskErrorAnalysis
	RuntimeEventRecord       = model.RuntimeEventRecord
)
