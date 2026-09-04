package web

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cloudwego/ppt-agent/pkg/db"
	loganalysis "github.com/cloudwego/ppt-agent/pkg/log_analysis"
)

func (s *Server) handleListLogAnalyses(c *gin.Context) {
	if s.logAnalysis == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "log analysis service not enabled"})
		return
	}
	analyses, err := loganalysis.GetRecentAnalyses(50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"analyses": analyses})
}

// handleGetTaskLogAnalyses 返回特定任务的全部日志分析。

func (s *Server) handleGetTaskLogAnalyses(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id is required"})
		return
	}
	analyses, err := loganalysis.GetTaskAnalyses(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"analyses": analyses})
}

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

type adminTaskResponse struct {
	ID                   string     `json:"id"`
	UserID               uint       `json:"user_id"`
	UserEmail            string     `json:"user_email"`
	Query                string     `json:"query"`
	Status               string     `json:"status"`
	DoneCount            int        `json:"done_count"`
	TotalCount           int        `json:"total_count"`
	Duration             string     `json:"duration"`
	GenerationStartedAt  *time.Time `json:"generation_started_at,omitempty"`
	GenerationFinishedAt *time.Time `json:"generation_finished_at,omitempty"`
	GenerationDurationMS int64      `json:"generation_duration_ms"`
	FixerRunCount        int        `json:"fixer_run_count"`
	Error                string     `json:"error"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

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

func (s *Server) handleAdminLogAnalyses(c *gin.Context) {
	analyses, err := db.ListRecentErrorAnalyses(50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询日志分析失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"analyses": analyses})
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

func (s *Server) handleAdminDeleteLogAnalysis(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}
	if err := db.DeleteErrorAnalysis(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
