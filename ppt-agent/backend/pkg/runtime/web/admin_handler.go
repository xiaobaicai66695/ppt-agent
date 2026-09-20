package web

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/cloudwego/ppt-agent/pkg/db"
	webmodel "github.com/cloudwego/ppt-agent/pkg/runtime/web/model"
)

// onTaskContinue 任务完成后自动触发继续处理（TaskManager 通过回调调用）。
// 它从等待队列中取出消息，重新启动 SSE 流并处理继续逻辑。

func (s *Server) handleAdminUsers(c *gin.Context) {
	users, err := db.ListAdminUserMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户列表失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

type adminTaskResponse = webmodel.AdminTaskResponse
type adminExecutionRecordResponse = webmodel.AdminExecutionRecordResponse

func (s *Server) handleAdminTasks(c *gin.Context) {
	tasks, err := db.ListAllTaskRecords(100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询任务列表失败: " + err.Error()})
		return
	}

	users, err := db.ListAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询任务所属用户失败: " + err.Error()})
		return
	}
	emails := make(map[uint]string, len(users))
	for _, user := range users {
		emails[user.ID] = user.Email
	}

	result := make([]adminTaskResponse, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, adminTaskResponse{
			ID:                   task.ID,
			UserID:               task.UserID,
			UserEmail:            emails[task.UserID],
			Query:                task.Query,
			Status:               task.Status,
			DoneCount:            task.DoneCount,
			TotalCount:           task.TotalCount,
			Duration:             task.Duration,
			GenerationStartedAt:  task.GenerationStartedAt,
			GenerationFinishedAt: task.GenerationFinishedAt,
			GenerationDurationMS: task.GenerationDurationMS,
			FixerRunCount:        task.FixerRunCount,
			Error:                task.Error,
			CreatedAt:            task.CreatedAt,
			UpdatedAt:            task.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"tasks": result})
}

func (s *Server) handleAdminExecutionRecords(c *gin.Context) {
	page, pageSize, userID, err := parseAdminExecutionRecordQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	records, total, err := db.ListAdminTaskRecords(page, pageSize, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询执行记录失败: " + err.Error()})
		return
	}

	result := make([]adminExecutionRecordResponse, 0, len(records))
	for _, record := range records {
		result = append(result, adminExecutionRecordFromDB(record))
	}
	c.JSON(http.StatusOK, gin.H{
		"records": result,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

func parseAdminExecutionRecordQuery(c *gin.Context) (page, pageSize int, userID uint, err error) {
	page = 1
	pageSize = 50
	if raw := strings.TrimSpace(c.Query("page")); raw != "" {
		page, err = strconv.Atoi(raw)
		if err != nil || page < 1 {
			return 0, 0, 0, fmt.Errorf("page 必须是大于 0 的整数")
		}
	}
	if raw := strings.TrimSpace(c.Query("page_size")); raw != "" {
		pageSize, err = strconv.Atoi(raw)
		if err != nil || pageSize < 1 || pageSize > 200 {
			return 0, 0, 0, fmt.Errorf("page_size 必须是 1 到 200 的整数")
		}
	}
	if raw := strings.TrimSpace(c.Query("user_id")); raw != "" {
		parsed, parseErr := strconv.ParseUint(raw, 10, 64)
		if parseErr != nil || parsed == 0 {
			return 0, 0, 0, fmt.Errorf("user_id 必须是大于 0 的整数")
		}
		userID = uint(parsed)
	}
	return page, pageSize, userID, nil
}

func adminExecutionRecordFromDB(record db.AdminTaskRecord) adminExecutionRecordResponse {
	return adminExecutionRecordResponse{
		ID:                   record.ID,
		UserID:               record.UserID,
		UserEmail:            record.UserEmail,
		Query:                record.Query,
		Status:               record.Status,
		DoneCount:            record.DoneCount,
		TotalCount:           record.TotalCount,
		Duration:             record.Duration,
		Error:                record.Error,
		PromptTokens:         record.PromptTokens,
		CompletionTokens:     record.CompletionTokens,
		TotalTokens:          record.TotalTokens,
		Intent:               record.Intent,
		ConversationID:       record.ConversationID,
		SourceMessageID:      record.SourceMessageID,
		ParentTaskID:         record.ParentTaskID,
		GenerationStartedAt:  record.GenerationStartedAt,
		GenerationFinishedAt: record.GenerationFinishedAt,
		GenerationDurationMS: record.GenerationDurationMS,
		FixerRunCount:        record.FixerRunCount,
		CreatedAt:            record.CreatedAt,
		UpdatedAt:            record.UpdatedAt,
	}
}

func (s *Server) handleAdminStats(c *gin.Context) {
	metrics, err := db.GetAdminMetrics(rootAccountEmail())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询运营统计失败: " + err.Error()})
		return
	}
	var runningCount int64
	if db.DB != nil {
		db.DB.Model(&db.TaskRecord{}).Where("status = ?", "running").Count(&runningCount)
	}
	c.JSON(http.StatusOK, gin.H{
		"user_count":                     metrics.RegisteredUserCount,
		"task_count":                     metrics.PPTGenerationCount,
		"running_count":                  runningCount,
		"registered_user_count":          metrics.RegisteredUserCount,
		"non_root_registered_user_count": metrics.NonRootRegisteredUserCount,
		"ppt_active_user_count":          metrics.PPTActiveUserCount,
		"custom_api_key_user_count":      metrics.CustomAPIKeyUserCount,
		"ppt_generation_count":           metrics.PPTGenerationCount,
		"non_root_ppt_generation_count":  metrics.NonRootPPTGenerationCount,
		"feedback_count":                 metrics.FeedbackCount,
		"feedback_suggestion_count":      metrics.FeedbackSuggestionCount,
	})
}

func (s *Server) handleAdminFeedback(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	feedback, err := db.ListAdminTaskFeedback(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户反馈失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"feedback": feedback})
}

func rootAccountEmail() string {
	if email := strings.TrimSpace(os.Getenv("ROOT_EMAIL")); email != "" {
		return email
	}
	return "root@qq.com"
}
