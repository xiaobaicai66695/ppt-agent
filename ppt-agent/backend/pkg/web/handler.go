package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/agent/deep"
	"github.com/cloudwego/ppt-agent/pkg/auth"
	"github.com/cloudwego/ppt-agent/pkg/task"
	"github.com/cloudwego/ppt-agent/pkg/templates"
)

// safePath joins baseDir with filename and returns the cleaned absolute path.
// Returns an error if the result escapes baseDir (path traversal protection).
func safePath(baseDir, filename string) (string, error) {
	cleaned := filepath.Clean(filepath.Join(baseDir, filename))
	if !strings.HasPrefix(cleaned, filepath.Clean(baseDir)+string(filepath.Separator)) {
		return "", os.ErrPermission
	}
	return cleaned, nil
}

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
	var req struct {
		Query   string               `json:"query"`
		Outline *deep.TaskOutline   `json:"outline,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Query) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}

	taskID := s.taskIDGen()
	cfg := s.makeTaskConfig(taskID)

	// 如果有 outline，先写入 tasks.json
	if req.Outline != nil && len(req.Outline.Slides) > 0 {
		cfg.Outline = req.Outline
	}

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

	filePath, err := safePath(ts.Info.WorkDir, filename)
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

	ts := s.tasks.GetTaskState(id)
	if ts == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	filePath, err := safePath(ts.Info.WorkDir, filename)
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

// ── Template handlers ────────────────────────────────────────────────────────

func (s *Server) handleListTemplates(c *gin.Context) {
	list := s.templateLoader.ListPresets()
	c.JSON(http.StatusOK, gin.H{"presets": list})
}

func (s *Server) handleGetTemplate(c *gin.Context) {
	name := c.Param("name")
	t := s.templateLoader.GetPreset(name)
	if t == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模板不存在"})
		return
	}
	c.JSON(http.StatusOK, t)
}

func (s *Server) handleListLayouts(c *gin.Context) {
	layouts := s.templateLoader.ListLayouts()
	c.JSON(http.StatusOK, gin.H{"layouts": layouts})
}

func (s *Server) handleListThemes(c *gin.Context) {
	themes := s.templateLoader.ListThemes()
	c.JSON(http.StatusOK, gin.H{"themes": themes})
}

func (s *Server) handleAIExpand(c *gin.Context) {
	var req struct {
		Title       string `json:"title"`
		ContentType string `json:"content_type"`
		Description string `json:"description"`
		Theme       string `json:"theme"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	ctx := c.Request.Context()
	model, err := s.aiModelFactory(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "模型初始化失败: " + err.Error()})
		return
	}

	layout, _ := s.templateLoader.GetLayout(req.ContentType)
	layoutName := req.ContentType
	if layout != nil {
		layoutName = layout.DisplayName
	}

	theme, _ := s.getTheme(req.Theme)
	themeName := req.Theme
	if theme != nil {
		themeName = theme.DisplayName
	}

	prompt := fmt.Sprintf(`你是一个PPT内容生成专家。用户正在制作PPT，请根据以下信息，为一页幻灯片生成详细内容描述。

页面信息：
- 标题：%s
- 布局类型：%s
- 当前描述：%s
- 配色主题：%s

请根据这些信息，生成一段详细的内容描述，供AI生成PPT页面使用。要求：
1. 内容与标题紧密相关
2. 描述具体、充实，避免空洞
3. 包含具体的要点、数据或案例（如适用）
4. 适合该布局类型（%s）
5. 中文输出
6. 描述应该包含该页面的具体内容要点，供AI直接使用生成PPT内容

只返回内容描述，不要返回其他信息。字数控制在200-400字之间。`, req.Title, layoutName, req.Description, themeName, layoutName)

	resp, err := model.Generate(ctx, []*schema.Message{
		schema.UserMessage(prompt),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成失败: " + err.Error()})
		return
	}

	content := ""
	if resp != nil {
		content = resp.Content
	}

	c.JSON(http.StatusOK, gin.H{"description": content})
}

// getTheme returns the theme info by name
func (s *Server) getTheme(name string) (*templates.ThemeInfo, error) {
	themes := s.templateLoader.ListThemes()
	for i := range themes {
		if themes[i].Name == name {
			return &themes[i], nil
		}
	}
	return nil, fmt.Errorf("theme not found")
}

func (s *Server) handleDeleteTask(c *gin.Context) {
	id := c.Param("id")
	if err := s.tasks.DeleteTask(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
