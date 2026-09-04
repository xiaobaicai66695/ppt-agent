package web

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/cloudwego/ppt-agent/pkg/agent/ppt"
	"github.com/cloudwego/ppt-agent/pkg/db"
	"github.com/cloudwego/ppt-agent/pkg/runtime/task"
	webmodel "github.com/cloudwego/ppt-agent/pkg/runtime/web/model"
	webservice "github.com/cloudwego/ppt-agent/pkg/runtime/web/service"
	"github.com/cloudwego/ppt-agent/pkg/utils/logger"
)

func (s *Server) handleCreateTask(c *gin.Context) {
	var req webmodel.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Query) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}

	uid := userIDGin(c)
	credential := userModelCredential(uid)
	route := s.routeCreateRequest(c.Request.Context(), req.Query, req.Outline != nil && len(req.Outline.Slides) > 0, credential)
	switch route.Intent {
	case createIntentFixExisting:
		c.JSON(http.StatusConflict, gin.H{
			"error":                  firstCreateRouteText(route.ClarificationQuestion, "这像是对已有 PPT 的修改请求，请先选择对应任务后继续修改。"),
			"intent":                 route.Intent,
			"reason":                 route.Reason,
			"clarification_question": route.ClarificationQuestion,
		})
		return
	case createIntentClarifyTopic, createIntentChat:
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":                  firstCreateRouteText(route.ClarificationQuestion, "请补充要制作的 PPT 主题、受众或使用场景。"),
			"intent":                 route.Intent,
			"reason":                 route.Reason,
			"clarification_question": route.ClarificationQuestion,
		})
		return
	}
	if duplicate := s.findRecentDuplicateTask(uid, req.Query); duplicate != nil {
		c.JSON(http.StatusOK, duplicate)
		return
	}

	taskID := s.taskIDGen()
	info, err := s.taskApplication().Create(c.Request.Context(), webmodel.TaskCreateInput{
		TaskID:         taskID,
		Query:          req.Query,
		UserID:         uid,
		Credential:     credential,
		Intent:         messageIntentCreate,
		ConversationID: taskID,
		Outline:        req.Outline,
	})
	if err != nil {
		if errors.Is(err, webservice.ErrInvalidOutline) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "outline 处理失败: " + err.Error()})
			return
		}
		code := http.StatusInternalServerError
		if err == task.ErrTaskAlreadyRunning {
			code = http.StatusConflict
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, info)
}

func (s *Server) findRecentDuplicateTask(uid int, query string) *task.TaskInfo {
	return s.taskApplication().FindRecentDuplicate(uid, query)
}

func (s *Server) handleStartConversationTask(c *gin.Context) {
	taskID := c.Param("id")
	uid := userIDGin(c)
	info := s.tasks.GetTask(taskID)
	if info == nil || info.UserID != uid {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	if info.Status == task.TaskStatusRunning {
		c.JSON(http.StatusConflict, gin.H{"error": task.ErrTaskAlreadyRunning.Error()})
		return
	}

	sess := s.sessionManager.GetOrCreate(taskID, info.WorkDir)
	query := taskGenerationQuery(sess, info.Query)
	credential := userModelCredential(uid)
	started, err := s.taskApplication().StartConversation(c.Request.Context(), webmodel.StartConversationInput{
		TaskID:     taskID,
		Query:      query,
		UserID:     uid,
		Credential: credential,
	})
	if err != nil {
		code := http.StatusInternalServerError
		if err == task.ErrTaskAlreadyRunning {
			code = http.StatusConflict
		} else if err == os.ErrNotExist {
			code = http.StatusNotFound
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, started)
}

func (s *Server) handleGetTask(c *gin.Context) {
	id := c.Param("id")
	info := s.tasks.GetTask(id)
	if info == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	if info.Status == task.TaskStatusRunning {
		if m, err := s.tasks.ReadTasksManifestFile(id); err == nil && m != nil {
			info.DoneCount = m.CompletedCount()
			info.TotalCount = len(m.Tasks)
		}
	}
	s.attachTaskFeedback(info, userIDGin(c))
	c.JSON(http.StatusOK, info)
}

func (s *Server) handleListTasks(c *gin.Context) {
	var tasks []task.TaskInfo
	if isAdminGin(c) {
		tasks = s.tasks.ListAllTasks()
	} else {
		tasks = s.tasks.ListTasks(userIDGin(c))
	}
	if tasks == nil {
		tasks = []task.TaskInfo{}
	}
	if !isAdminGin(c) {
		userID := userIDGin(c)
		ids := make([]string, 0, len(tasks))
		for _, info := range tasks {
			if info.UserID == userID {
				ids = append(ids, info.ID)
			}
		}
		feedbackByTask, err := db.ListTaskFeedbackByUserAndTaskIDs(uint(userID), ids)
		if err != nil {
			logger.Warn("task_feedback_batch_load_failed", "user_id", userID, "error", err.Error())
		}
		for index := range tasks {
			if feedback, ok := feedbackByTask[tasks[index].ID]; ok {
				tasks[index].Feedback = &task.DeliveryFeedback{Rating: feedback.Rating, Suggestion: feedback.Suggestion, UpdatedAt: feedback.UpdatedAt}
			}
		}
	}
	c.JSON(http.StatusOK, tasks)
}

func (s *Server) attachTaskFeedback(info *task.TaskInfo, userID int) {
	if info == nil || info.UserID != userID || userID <= 0 {
		return
	}
	feedback, err := s.tasks.GetDeliveryFeedback(info.ID, userID)
	if err != nil {
		logger.Warn("task_feedback_load_failed", "task_id", info.ID, "error", err.Error())
		return
	}
	info.Feedback = feedback
}

func (s *Server) handleSaveTaskFeedback(c *gin.Context) {
	var req webmodel.TaskFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的反馈请求"})
		return
	}
	if _, err := task.ValidateDeliveryFeedback(req.Rating, req.Suggestion); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	info := s.tasks.GetTask(c.Param("id"))
	if info == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	if info.Status != task.TaskStatusCompleted {
		c.JSON(http.StatusConflict, gin.H{"error": "PPT 尚未交付，暂不能评分"})
		return
	}
	feedback, err := s.tasks.SaveDeliveryFeedback(info.ID, userIDGin(c), req.Rating, req.Suggestion)
	if err != nil {
		logger.Error("task_feedback_save_failed", "task_id", info.ID, "error", err.Error())
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "反馈暂时无法保存，请稍后重试"})
		return
	}
	info.Feedback = feedback
	c.JSON(http.StatusOK, info)
}

func (s *Server) handleDownloadFile(c *gin.Context) {
	id := c.Param("id")
	filename := c.Param("filename")

	// 优先从内存获取 workDir，冷加载时从 DB 兜底
	workDir := s.tasks.GetWorkDir(id)
	if workDir == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	filePath, err := safePath(workDir, filename)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "非法的文件路径"})
		return
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	c.File(filePath)
}

func (s *Server) handleThumbnail(c *gin.Context) {
	id := c.Param("id")
	filename := c.Param("filename")

	// 优先从内存获取 workDir，冷加载时从 DB 兜底
	workDir := s.tasks.GetWorkDir(id)
	if workDir == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	filePath, err := safePath(workDir, filename)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "非法的文件路径"})
		return
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	jpeg, err := GenerateThumbnail(filePath)
	if err != nil {
		c.Header("Cache-Control", "no-store")
		c.Header("Retry-After", "10")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Type", "image/jpeg")
	c.Header("Cache-Control", "public, max-age=3600")
	c.Data(http.StatusOK, "image/jpeg", jpeg)
}

func (s *Server) handleCancelTask(c *gin.Context) {
	id := c.Param("id")
	if s.tasks.CancelTask(id) {
		c.JSON(http.StatusOK, s.tasks.GetTask(id))
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "task not found or not running"})
}

// ── Template handlers ────────────────────────────────────────────────────────

func (s *Server) handleListLayouts(c *gin.Context) {
	layouts := s.templateLoader.ListLayouts()
	c.JSON(http.StatusOK, gin.H{"layouts": layouts})
}

func (s *Server) prepareOutline(_ context.Context, query string, outline *ppt.TaskOutline) (*ppt.TaskOutline, error) {
	return s.taskApplication().PrepareOutline(query, outline)
}

func (s *Server) handleDeleteTask(c *gin.Context) {
	id := c.Param("id")
	if err := s.tasks.DeleteTask(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ── 会话/继续处理器 ────────────────────────────────────────────────

// handleContinueTask 处理用户继续迭代现有任务的请求。
