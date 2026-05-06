package web

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/cloudwego/ppt-agent/pkg/auth"
	"github.com/cloudwego/ppt-agent/pkg/task"
)

// ── Auth handlers ────────────────────────────────────────────────────────

func (s *Server) handleSendCode(c *gin.Context) {
	var req struct{ Email string `json:"email"` }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
		return
	}
	if err := auth.SendCode(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "验证码已发送"})
}

func (s *Server) handleLogin(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Code     string `json:"code"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
		return
	}

	if req.Code != "" {
		token, user, isNew, err := auth.LoginWithCode(req.Email, req.Code)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"token": token, "id": user.ID, "email": user.Email, "is_new": isNew,
		})
		return
	}

	if req.Password != "" {
		token, user, err := auth.LoginWithPassword(req.Email, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"token": token, "id": user.ID, "email": user.Email,
		})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "请提供验证码或密码"})
}

func (s *Server) handleSetPassword(c *gin.Context) {
	var req struct{ Password string `json:"password"` }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
		return
	}
	if err := auth.SetPassword(userIDGin(c), req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) handleLogout(c *gin.Context) {
	token := extractTokenFromGin(c)
	if token != "" {
		auth.Logout(token)
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) handleMe(c *gin.Context) {
	uid := userIDGin(c)
	email, _ := auth.UsernameFromContext(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"id": uid, "email": email})
}

// ── Task handlers ─────────────────────────────────────────────────────────

func (s *Server) handleCreateTask(c *gin.Context) {
	var req struct{ Query string `json:"query"` }
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Query) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}

	taskID := s.taskIDGen()
	cfg := s.makeTaskConfig(taskID)
	info, err := s.tasks.CreateTask(c.Request.Context(), req.Query, userIDGin(c), s.agentFactory, cfg)
	if err != nil {
		code := http.StatusInternalServerError
		if err == task.ErrTaskAlreadyRunning {
			code = http.StatusConflict
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, info)
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
	c.JSON(http.StatusOK, info)
}

func (s *Server) handleListTasks(c *gin.Context) {
	tasks := s.tasks.ListTasks(userIDGin(c))
	if tasks == nil {
		tasks = []task.TaskInfo{}
	}
	c.JSON(http.StatusOK, tasks)
}

func (s *Server) handleDownloadFile(c *gin.Context) {
	id := c.Param("id")
	filename := c.Param("filename")

	ts := s.tasks.GetTaskState(id)
	if ts == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	filePath := filepath.Join(ts.Info.WorkDir, filename)
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

	ts := s.tasks.GetTaskState(id)
	if ts == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	filePath := filepath.Join(ts.Info.WorkDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	jpeg, err := GenerateThumbnail(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

func (s *Server) handleDeleteTask(c *gin.Context) {
	id := c.Param("id")
	if err := s.tasks.DeleteTask(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
