package db

import (
	"strings"
	"time"
)

// AdminMetrics is the aggregate used by the management dashboard. Counts of
// PPT generation deliberately exclude workbench-only conversation records.
type AdminMetrics struct {
	RegisteredUserCount        int64 `json:"registered_user_count"`
	NonRootRegisteredUserCount int64 `json:"non_root_registered_user_count"`
	PPTActiveUserCount         int64 `json:"ppt_active_user_count"`
	CustomAPIKeyUserCount      int64 `json:"custom_api_key_user_count"`
	PPTGenerationCount         int64 `json:"ppt_generation_count"`
	NonRootPPTGenerationCount  int64 `json:"non_root_ppt_generation_count"`
	FeedbackCount              int64 `json:"feedback_count"`
	FeedbackSuggestionCount    int64 `json:"feedback_suggestion_count"`
}

// AdminUserMetric enriches a user row with management-only usage data.
type AdminUserMetric struct {
	User
	PPTGenerationCount     int64 `json:"ppt_generation_count"`
	CustomAPIKeyConfigured bool  `json:"custom_api_key_configured"`
}

// AdminTaskFeedback joins a delivery feedback item to safe task and user
// metadata for administrators. API key values are never selected or exposed.
type AdminTaskFeedback struct {
	TaskFeedback
	UserEmail string `json:"user_email"`
	TaskQuery string `json:"task_query"`
}

// AdminTaskRecord is the safe, durable execution metadata available to
// administrators. It deliberately omits work_dir and files because those are
// server-internal paths and delivery details, not operational audit data.
type AdminTaskRecord struct {
	ID                   string     `json:"id"`
	UserID               uint       `json:"user_id"`
	UserEmail            string     `json:"user_email"`
	Query                string     `json:"query"`
	Status               string     `json:"status"`
	DoneCount            int        `json:"done_count"`
	TotalCount           int        `json:"total_count"`
	Duration             string     `json:"duration"`
	Error                string     `json:"error"`
	PromptTokens         int64      `json:"prompt_tokens"`
	CompletionTokens     int64      `json:"completion_tokens"`
	TotalTokens          int64      `json:"total_tokens"`
	Intent               string     `json:"intent"`
	ConversationID       string     `json:"conversation_id"`
	SourceMessageID      string     `json:"source_message_id"`
	ParentTaskID         string     `json:"parent_task_id"`
	GenerationStartedAt  *time.Time `json:"generation_started_at"`
	GenerationFinishedAt *time.Time `json:"generation_finished_at"`
	GenerationDurationMS int64      `json:"generation_duration_ms"`
	FixerRunCount        int        `json:"fixer_run_count"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// ListAllUsers 返回所有用户（供管理员查看）。
func ListAllUsers() ([]User, error) {
	var users []User
	err := DB.Order("created_at DESC").Find(&users).Error
	return users, err
}

// ListAdminUserMetrics returns users with their PPT-generation count and API
// key configuration state. Conversation-only workbench records are excluded.
func ListAdminUserMetrics() ([]AdminUserMetric, error) {
	if DB == nil {
		return []AdminUserMetric{}, nil
	}
	var users []User
	if err := DB.Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, err
	}
	var generationCounts []struct {
		UserID uint
		Count  int64
	}
	if err := DB.Model(&TaskRecord{}).
		Select("user_id, COUNT(*) AS count").
		Where("intent = ?", "create").
		Group("user_id").
		Scan(&generationCounts).Error; err != nil {
		return nil, err
	}
	counts := make(map[uint]int64, len(generationCounts))
	for _, row := range generationCounts {
		counts[row.UserID] = row.Count
	}
	var keyUserIDs []uint
	if err := DB.Model(&UserAPIKey{}).Pluck("user_id", &keyUserIDs).Error; err != nil {
		return nil, err
	}
	configured := make(map[uint]bool, len(keyUserIDs))
	for _, userID := range keyUserIDs {
		configured[userID] = true
	}
	metrics := make([]AdminUserMetric, 0, len(users))
	for _, user := range users {
		metrics = append(metrics, AdminUserMetric{
			User:                   user,
			PPTGenerationCount:     counts[user.ID],
			CustomAPIKeyConfigured: configured[user.ID],
		})
	}
	return metrics, nil
}

// GetAdminMetrics calculates management counters using durable records only.
// rootEmail identifies the deployment's seeded root account so other admins
// are still included in normal user and usage metrics.
func GetAdminMetrics(rootEmail string) (*AdminMetrics, error) {
	if DB == nil {
		return &AdminMetrics{}, nil
	}
	rootEmail = strings.TrimSpace(rootEmail)
	metrics := &AdminMetrics{}
	normalUsers := DB.Model(&User{}).Where("email NOT LIKE ?", "guest-%@guest.local")
	if err := normalUsers.Count(&metrics.RegisteredUserCount).Error; err != nil {
		return nil, err
	}
	nonRootUsers := DB.Model(&User{}).Where("email NOT LIKE ?", "guest-%@guest.local")
	if rootEmail != "" {
		nonRootUsers = nonRootUsers.Where("email <> ?", rootEmail)
	}
	if err := nonRootUsers.Count(&metrics.NonRootRegisteredUserCount).Error; err != nil {
		return nil, err
	}
	generationTasks := DB.Model(&TaskRecord{}).Where("intent = ?", "create")
	if err := generationTasks.Count(&metrics.PPTGenerationCount).Error; err != nil {
		return nil, err
	}
	// Registered accounts are stable identities. Guest/default accounts are
	// intentionally ephemeral, so use their one-way IP fingerprint instead of
	// their generated account id for active-user aggregation. Legacy guests
	// without a fingerprint remain individually countable.
	var activeUserCount struct {
		Count int64
	}
	if err := DB.Table("task_records").
		Select(`COUNT(DISTINCT CASE WHEN users.email LIKE ? THEN CONCAT('guest:', COALESCE(NULLIF(users.guest_ip_hash, ''), CONCAT('legacy:', task_records.user_id))) ELSE CONCAT('user:', task_records.user_id) END) AS count`, "guest-%@guest.local").
		Joins("LEFT JOIN users ON users.id = task_records.user_id").
		Where("task_records.intent = ?", "create").
		Scan(&activeUserCount).Error; err != nil {
		return nil, err
	}
	metrics.PPTActiveUserCount = activeUserCount.Count
	if err := DB.Model(&UserAPIKey{}).Distinct("user_id").Count(&metrics.CustomAPIKeyUserCount).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&TaskFeedback{}).Count(&metrics.FeedbackCount).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&TaskFeedback{}).Where("suggestion <> ?", "").Count(&metrics.FeedbackSuggestionCount).Error; err != nil {
		return nil, err
	}
	nonRootTasks := DB.Model(&TaskRecord{}).Where("intent = ?", "create")
	if rootEmail != "" {
		nonRootTasks = nonRootTasks.Where("user_id NOT IN (?)", DB.Model(&User{}).Select("id").Where("email = ?", rootEmail))
	}
	if err := nonRootTasks.Count(&metrics.NonRootPPTGenerationCount).Error; err != nil {
		return nil, err
	}
	return metrics, nil
}

// ListAdminTaskFeedback returns recent delivery feedback for the management
// page. Suggestions remain task-owner content and are therefore admin-only.
func ListAdminTaskFeedback(limit int) ([]AdminTaskFeedback, error) {
	if DB == nil {
		return []AdminTaskFeedback{}, nil
	}
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	var records []AdminTaskFeedback
	err := DB.Table("task_feedbacks").
		Select("task_feedbacks.id, task_feedbacks.task_id, task_feedbacks.user_id, task_feedbacks.rating, task_feedbacks.suggestion, task_feedbacks.created_at, task_feedbacks.updated_at, users.email AS user_email, task_records.query AS task_query").
		Joins("LEFT JOIN users ON users.id = task_feedbacks.user_id").
		Joins("LEFT JOIN task_records ON task_records.id = task_feedbacks.task_id").
		Order("task_feedbacks.updated_at DESC").
		Limit(limit).
		Scan(&records).Error
	return records, err
}

// ListAdminTaskRecords returns a paginated, cross-user task execution audit
// trail. The caller must enforce administrator authorization before invoking
// it. An optional userID narrows results to a single account.
func ListAdminTaskRecords(page, pageSize int, userID uint) ([]AdminTaskRecord, int64, error) {
	if DB == nil {
		return []AdminTaskRecord{}, 0, nil
	}
	query := DB.Table("task_records").Joins("LEFT JOIN users ON users.id = task_records.user_id")
	if userID > 0 {
		query = query.Where("task_records.user_id = ?", userID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	records := make([]AdminTaskRecord, 0, pageSize)
	err := query.Select(`task_records.id, task_records.user_id, users.email AS user_email,
		task_records.query, task_records.status, task_records.done_count, task_records.total_count,
		task_records.duration, task_records.error, task_records.prompt_tokens,
		task_records.completion_tokens, task_records.total_tokens, task_records.intent,
		task_records.conversation_id, task_records.source_message_id, task_records.parent_task_id,
		task_records.generation_started_at, task_records.generation_finished_at,
		task_records.generation_duration_ms, task_records.fixer_run_count,
		task_records.created_at, task_records.updated_at`).
		Order("task_records.created_at DESC, task_records.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&records).Error
	return records, total, err
}
