package web

import (
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
