package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"

	"github.com/cloudwego/ppt-agent/pkg/agent/deep"
	agentlearning "github.com/cloudwego/ppt-agent/pkg/agent/learning"
	"github.com/cloudwego/ppt-agent/pkg/auth"
	"github.com/cloudwego/ppt-agent/pkg/db"
	loganalysis "github.com/cloudwego/ppt-agent/pkg/log_analysis"
	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/session"
	"github.com/cloudwego/ppt-agent/pkg/style"
	"github.com/cloudwego/ppt-agent/pkg/task"
	"github.com/cloudwego/ppt-agent/pkg/templates"
)

// safePath 将 baseDir 与 filename 连接并返回清理后的绝对路径。
// 如果结果逃逸出 baseDir，则返回错误（防止路径遍历攻击）。
func safePath(baseDir, filename string) (string, error) {
	cleaned := filepath.Clean(filepath.Join(baseDir, filename))
	if !strings.HasPrefix(cleaned, filepath.Clean(baseDir)+string(filepath.Separator)) {
		return "", os.ErrPermission
	}
	return cleaned, nil
}

// ── 认证处理器 ────────────────────────────────────────────────────────

func (s *Server) handleSendCode(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}
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
			"token": token, "id": user.ID, "email": user.Email, "is_new": isNew, "is_admin": user.IsAdmin,
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
			"is_admin": user.IsAdmin,
		})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "请提供验证码或密码"})
}

func (s *Server) handleSetPassword(c *gin.Context) {
	var req struct {
		Password string `json:"password"`
	}
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
	user, _ := auth.ValidateUser(uid)
	c.JSON(http.StatusOK, gin.H{
		"id":       uid,
		"email":    email,
		"is_admin": user != nil && user.IsAdmin,
	})
}

// ── 任务处理器 ─────────────────────────────────────────────────────────

func (s *Server) handleCreateTask(c *gin.Context) {
	var req struct {
		Query             string             `json:"query"`
		Outline           *deep.TaskOutline  `json:"outline,omitempty"`
		TemplateSelection *TemplateSelection `json:"template_selection,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Query) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}

	taskID := s.taskIDGen()
	cfg := s.makeTaskConfig(taskID)
	cfg.Query = req.Query

	// 如果有 outline，先做服务端兜底校验/补齐，再写入 tasks.json。
	if req.Outline != nil && len(req.Outline.Slides) > 0 {
		outline, err := s.prepareOutline(c.Request.Context(), req.Query, req.Outline)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "outline 处理失败: " + err.Error()})
			return
		}
		cfg.Outline = outline
	} else if req.TemplateSelection != nil {
		outline, _, err := s.resolveTemplateSelection(req.Query, req.TemplateSelection)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "模板选择失败: " + err.Error()})
			return
		}
		cfg.Outline = outline
	}

	// 注入用户风格偏好上下文
	uid := userIDGin(c)
	userProfile := s.styleStore.Get(uid)
	styleContext := userProfile.BuildStyleContext()
	if styleContext != "" {
		cfg.StyleContext = styleContext
	}

	info, err := s.tasks.CreateTask(c.Request.Context(), req.Query, uid, s.agentFactory, cfg)
	if err != nil {
		code := http.StatusInternalServerError
		if err == task.ErrTaskAlreadyRunning {
			code = http.StatusConflict
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}

	// 初始化会话（消息会自动写入数据库）
	sess := s.sessionManager.GetOrCreate(taskID, info.WorkDir)
	sess.AddUserMessage(req.Query)

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

func (s *Server) handleListBackgrounds(c *gin.Context) {
	backgrounds := s.templateLoader.ListBackgrounds()
	c.JSON(http.StatusOK, gin.H{"backgrounds": backgrounds})
}

func (s *Server) handleBackgroundPreview(c *gin.Context) {
	name := c.Param("name")
	backgrounds := s.templateLoader.ListBackgrounds()
	for _, bg := range backgrounds {
		if bg.Name == name {
			previewPath := filepath.Join(
				s.templateLoader.GetBackgroundTemplatesDir(),
				name,
				"preview.jpg",
			)
			if _, err := os.Stat(previewPath); err == nil {
				c.File(previewPath)
				return
			}
			// Fallback: try preview.png
			previewPath = filepath.Join(
				s.templateLoader.GetBackgroundTemplatesDir(),
				name,
				"preview.png",
			)
			if _, err := os.Stat(previewPath); err == nil {
				c.File(previewPath)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "preview not found"})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "background theme not found"})
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

	layout := s.templateLoader.GetLayout(req.ContentType)
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
- 容量要求：%s

请根据这些信息，生成一段详细的内容描述，供AI生成PPT页面使用。要求：
1. 内容与标题紧密相关
2. 描述具体、充实，避免空洞
3. 包含具体的要点、数据或案例（如适用）
4. 适合该布局类型（%s）
5. 中文输出
6. 描述应该包含该页面的具体内容要点，供AI直接使用生成PPT内容

只返回内容描述，不要返回其他信息。字数严格遵守容量要求。`, req.Title, layoutName, req.Description, themeName, layoutDescriptionTarget(req.ContentType), layoutName)

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

func (s *Server) prepareOutline(_ context.Context, query string, outline *deep.TaskOutline) (*deep.TaskOutline, error) {
	if outline == nil || len(outline.Slides) == 0 {
		return outline, nil
	}
	if strings.TrimSpace(outline.Theme) == "" {
		outline.Theme = "ocean_soft"
	}
	if _, err := s.getTheme(outline.Theme); err != nil {
		outline.Theme = "ocean_soft"
	}
	if strings.TrimSpace(outline.Title) == "" {
		outline.Title = strings.TrimSpace(query)
	}
	if outline.ContentMode != deep.OutlineContentModeTemplateScaffold {
		outline.ContentMode = deep.OutlineContentModeUserOutline
	}

	for i := range outline.Slides {
		slide := &outline.Slides[i]
		slide.Title = strings.TrimSpace(slide.Title)
		slide.ContentType = strings.TrimSpace(slide.ContentType)
		slide.Description = strings.TrimSpace(slide.Description)
		slide.Background = strings.TrimSpace(slide.Background)
		if s.templateLoader.GetLayout(slide.ContentType) == nil {
			return nil, fmt.Errorf("第%d页 content_type=%q 不存在", i+1, slide.ContentType)
		}
		if slide.Background != "" && !s.isValidBackground(slide.Background) {
			slide.Background = ""
		}
	}

	return outline, nil
}

// enrichOutline keeps the legacy explicit AI-fill endpoint compatible. Task
// creation deliberately avoids it; the main Agent completes template content.
func (s *Server) enrichOutline(ctx context.Context, query string, outline *deep.TaskOutline) (*deep.TaskOutline, error) {
	prepared, err := s.prepareOutline(ctx, query, outline)
	if err != nil {
		return nil, err
	}
	if !outlineNeedsEnrichment(prepared) {
		return prepared, nil
	}
	enriched, err := s.generateOutlineSlides(ctx, query, prepared)
	if err != nil {
		return nil, err
	}
	return s.mergeOutlineSlides(prepared, enriched), nil
}

func outlineNeedsEnrichment(outline *deep.TaskOutline) bool {
	for _, slide := range outline.Slides {
		if len([]rune(strings.TrimSpace(slide.Description))) < 30 || slide.ContentPlan == nil {
			return true
		}
	}
	return false
}

func (s *Server) mergeOutlineSlides(base *deep.TaskOutline, enriched []deep.SlideOutline) *deep.TaskOutline {
	for i := range base.Slides {
		if i >= len(enriched) {
			break
		}
		e := enriched[i]
		if strings.TrimSpace(e.Title) != "" {
			base.Slides[i].Title = strings.TrimSpace(e.Title)
		}
		if s.templateLoader.GetLayout(e.ContentType) != nil {
			base.Slides[i].ContentType = e.ContentType
		}
		if strings.TrimSpace(e.Description) != "" {
			base.Slides[i].Description = strings.TrimSpace(e.Description)
		}
		if e.ContentPlan != nil {
			base.Slides[i].ContentPlan = e.ContentPlan
		}
		if e.Background != "" && s.isValidBackground(e.Background) {
			base.Slides[i].Background = e.Background
		}
	}
	return base
}

func (s *Server) isValidBackground(name string) bool {
	for _, bg := range s.templateLoader.ListBackgrounds() {
		if bg.Name == name {
			return true
		}
	}
	return false
}

func (s *Server) backgroundCatalog() string {
	var lines []string
	for _, bg := range s.templateLoader.ListBackgrounds() {
		lines = append(lines, fmt.Sprintf("- `%s`：%s；适用场景：%s",
			bg.Name, bg.DisplayName, strings.Join(bg.Scenarios, "、")))
	}
	return strings.Join(lines, "\n")
}

func layoutDescriptionTarget(contentType string) string {
	switch contentType {
	case "title_slide", "section_divider":
		return "40-80字，给标题、副标题、讲述角度，不写长段落"
	case "content_slide":
		return "80-140字，拆成3-5条要点，每条不超过35字"
	case "two_column", "three_column", "card_grid", "process_flow", "summary_slide":
		return "100-180字，拆成短标题和短说明，严格控制单项长度"
	case "kpi_dashboard", "stat_slide", "chart_slide", "comparison_table":
		return "120-220字，必须包含可结构化的数据、指标、分类或表格字段"
	case "image_text", "case_study", "example_detail", "deep_dive":
		return "180-320字，允许稍长，但必须分清段落、案例、数据和结论"
	case "quote_slide":
		return "60-120字，必须包含 quote、attribution、kicker 字段含义"
	default:
		return "100-180字，内容密度匹配页面容量"
	}
}

func (s *Server) generateOutlineSlides(ctx context.Context, query string, outline *deep.TaskOutline) ([]deep.SlideOutline, error) {
	model, err := s.aiModelFactory(ctx)
	if err != nil {
		return nil, fmt.Errorf("模型初始化失败: %w", err)
	}

	theme, _ := s.getTheme(outline.Theme)
	themeName := outline.Theme
	if theme != nil {
		themeName = theme.DisplayName
	}

	var slideContexts strings.Builder
	for i, slide := range outline.Slides {
		layout := s.templateLoader.GetLayout(slide.ContentType)
		layoutName := slide.ContentType
		if layout != nil {
			layoutName = layout.DisplayName
		}
		slideContexts.WriteString(fmt.Sprintf(
			"\n第%d页：标题「%s」 | content_type=`%s` | 布局「%s」 | 容量要求：%s | 现有描述：%s",
			i+1, slide.Title, slide.ContentType, layoutName, layoutDescriptionTarget(slide.ContentType), slide.Description,
		))
	}

	prompt := fmt.Sprintf(`你是PPT内容规划专家。用户已编排好PPT结构，你的任务是根据用户主题为每一页生成可直接消费的结构化内容。

## 用户主题（所有内容必须围绕此展开）
%s

## 配色方案
%s

## 可用背景主题（background 只能从以下 id 中选择；如果页面不适合图片背景，可沿用空值）
%s

## 页面结构
%s

## 输出格式（严格返回JSON，不要任何解释）
{
  "slides": [
    {
      "title": "实际标题",
      "content_type": "content_slide",
      "background": "minimalist_blue",
      "description": "与布局容量匹配的内容描述",
      "content_plan": {
        "summary": "一句话概括",
        "elements": [
          {"type": "bullet_list", "items": ["短要点1：具体事实", "短要点2：具体数据"]}
        ]
      }
    }
  ]
}

## 强制要求
- 返回 slides 数量必须与页面结构数量一致，顺序也必须一致
- content_type 必须原样使用页面结构中的英文 id，禁止改成中文显示名或新名字
- background 只能使用可用背景主题中的 id；信息密集页可以留空，封面/章节页优先选择合适背景
- description 长度必须遵守每页容量要求，禁止统一生成300-400字长段落
- content_plan 要为后续生成器提供结构化字段：bullet_list、numbered_list、key_point_card、table、chart_placeholder、callout、quote 等
- 每页至少包含一个具体实体、数据、场景或案例；确需真实数据时写明需要搜索的数据项和来源方向
- 只输出JSON`, query, themeName, s.backgroundCatalog(), slideContexts.String())

	resp, err := model.Generate(ctx, []*schema.Message{
		schema.UserMessage(prompt),
	})
	if err != nil {
		return nil, fmt.Errorf("生成失败: %w", err)
	}

	content := ""
	if resp != nil {
		content = strings.TrimSpace(resp.Content)
	}
	if content == "" {
		return nil, fmt.Errorf("模型返回为空，请重试")
	}
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("模型返回格式为空，请重试")
	}

	var result struct {
		Slides []deep.SlideOutline `json:"slides"`
	}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("解析失败: %w", err)
	}
	if len(result.Slides) == 0 {
		return nil, fmt.Errorf("模型未返回有效内容")
	}
	if len(result.Slides) != len(outline.Slides) {
		return nil, fmt.Errorf("模型返回页数不匹配: got=%d want=%d", len(result.Slides), len(outline.Slides))
	}
	return result.Slides, nil
}

// handleAIGenerateOutline 接收带有空描述的部分大纲（幻灯片）和用户主题查询，
// 使用 AI 为每个空描述的幻灯片生成 content_plan。
// 返回填充后的幻灯片数组。
func (s *Server) handleAIGenerateOutline(c *gin.Context) {
	var req struct {
		Query   string            `json:"query"`
		Outline *deep.TaskOutline `json:"outline"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if req.Outline == nil || len(req.Outline.Slides) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "outline.slides is required"})
		return
	}

	outline, err := s.enrichOutline(c.Request.Context(), req.Query, req.Outline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"slides": outline.Slides})
}

// getTheme 根据名称返回主题信息
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

// ── 会话/继续处理器 ────────────────────────────────────────────────

// handleContinueTask 处理用户继续迭代现有任务的请求。
func (s *Server) handleContinueTask(c *gin.Context) {
	taskID := c.Param("id")

	var req struct {
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Message) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}

	ts := s.tasks.GetTaskState(taskID)
	if ts == nil {
		info := s.tasks.GetTask(taskID)
		if info == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		ts = s.tasks.NewColdTaskState(*info)
	}

	uid := userIDGin(c)

	// 将用户反馈写入结构化日志，供后续 LLM 分析
	logger.Info("user_feedback",
		"task_id", taskID,
		"user_id", uid,
		"message", req.Message,
		"task_status", string(ts.Info.Status),
		"page_count", fmt.Sprintf("%d/%d", ts.Info.DoneCount, ts.Info.TotalCount))

	// 任务运行中：加入等待队列，任务完成后自动处理
	if ts.Info.Status == task.TaskStatusRunning {
		afterEventID := ts.LatestEventID()
		firstQueued := ts.SetPendingContinueMsg(req.Message)

		statusMsg := "您的反馈已排队，将在当前任务完成后自动处理"
		if !firstQueued {
			statusMsg = "反馈已更新，将在当前任务完成后自动处理"
		}

		// 返回 202 Accepted，前端显示"已排队"
		c.JSON(http.StatusAccepted, gin.H{
			"status":         "queued",
			"message":        statusMsg,
			"task_id":        taskID,
			"after_event_id": afterEventID,
		})
		return
	}

	// 允许继续已完成的或已取消的任务
	afterEventID := ts.LatestEventID()
	sess := s.sessionManager.GetOrCreate(taskID, ts.Info.WorkDir)
	if err := sess.AddUserMessage(req.Message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存会话消息失败"})
		return
	}

	// 重新初始化任务为运行状态
	ts.Mu.Lock()
	ts.Info.Status = task.TaskStatusRunning
	ts.Mu.Unlock()
	ts.Persist()

	if s.continueStarter != nil {
		s.continueStarter(taskID, ts, req.Message, uid, sess)
	} else {
		s.startContinue(taskID, ts, req.Message, uid, sess)
	}
	c.JSON(http.StatusAccepted, gin.H{
		"status":         "accepted",
		"message":        "已收到修改请求，正在处理",
		"task_id":        taskID,
		"after_event_id": afterEventID,
	})
}

func (s *Server) startContinue(taskID string, ts *task.TaskState, message string, uid int, sess *session.ConversationSession) {
	ch := make(chan task.SSERichEvent, 64)
	go func() {
		for evt := range ch {
			ts.Broadcast(evt)
		}
		fullAnswer := ts.FullAnswer()
		ts.Mu.Lock()
		ts.Info.Status = task.TaskStatusCompleted
		ts.Info.FullAnswer = fullAnswer
		ts.Mu.Unlock()
		ts.Persist()
	}()
	go s.runContinue(taskID, ts, message, uid, sess, ch)
}

func (s *Server) runContinue(taskID string, ts *task.TaskState, message string, uid int, sess *session.ConversationSession, ch chan task.SSERichEvent) {
	defer close(ch)

	ctx := context.Background()

	// Step 1: Route intent - Fixer vs DeepAgent vs add_page
	ch <- task.SSERichEvent{Type: "answer", Content: "正在分析您的请求...\n"}

	route := s.routeContinueIntent(ctx, message, ts.Info.WorkDir)

	ch <- task.SSERichEvent{Type: "answer", Content: fmt.Sprintf("识别意图: %s (%s)\n", route.Intent, route.Reason)}

	switch route.Intent {
	case "fix":
		s.runFixerContinue(taskID, ts, &route, message, ch)
	case "regenerate", "regenerate_all", "add_page", "unknown":
		s.runDeepAgentContinue(taskID, ts, &route, message, ch)
	case "needs_clarification":
		question := route.ClarificationQuestion
		if question == "" {
			question = "您的反馈比较模糊，请问您是指哪一页？希望怎么改善？（比如：第2页字体太小、第3张换个颜色、第1页重新生成等）"
		}
		ch <- task.SSERichEvent{Type: "clarification", Content: question}
		ch <- task.SSERichEvent{Type: "continue_complete", Message: question}
		return
	}

	ch <- task.SSERichEvent{Type: "continue_complete", Message: "修改处理完成"}
}

// RouteResult 保存意图分类的结果。
type RouteResult struct {
	Intent string `json:"intent"` // "fix" | "regenerate" | "regenerate_all" | "add_page" | "needs_clarification" | "unknown"

	// Reason 描述选择此意图的原因。
	Reason string `json:"reason"`

	// TargetPages 包含用户提到的页面索引（从 1 开始）。
	TargetPages []int `json:"target_pages,omitempty"`

	// NeedsClarification 当用户意图模糊时为 true。
	NeedsClarification bool `json:"needs_clarification,omitempty"`

	// ClarificationQuestion 当 NeedsClarification 为 true 时要询问用户的问题。
	ClarificationQuestion string `json:"clarification_question,omitempty"`

	// FixDetails 当意图为 "fix" 时设置。描述要修复的内容。
	// 例如：{"aspect": "font_size", "value": "smaller", "pages": [2]}
	FixDetails *FixDetails `json:"fix_details,omitempty"`

	// RegenerateScope 为 "all" 或页面索引列表。
	RegenerateScope []int `json:"regenerate_scope,omitempty"`

	// SuggestFix 指示尽管需要澄清，但用户可能想要修复（而不是重新生成）。
	SuggestFix bool `json:"suggest_fix,omitempty"`
}

// FixDetails 描述用户想要调整的视觉属性。
type FixDetails struct {
	// Aspect 是要修复的视觉属性："font_size"、"color"、"alignment"、
	// "spacing"、"layout"、"position"、"text_content"、"style"、"other"。
	Aspect string `json:"aspect"`
	// Detail 是具体的调整："更大"、"红色"、"居中"、"加粗" 等。
	Detail string `json:"detail"`
	// TargetElements 描述页面上的哪些元素："标题"、"正文"、"所有文字"、"图表" 等。
	TargetElements string `json:"target_elements,omitempty"`
}

// routeContinueIntent 是意图路由的主入口点。
// 它始终委托给 LLM 分类以获得细致、结构化的结果。
// 只有少数高置信度的模式会绕过 LLM（例如"加一页"）。
func (s *Server) routeContinueIntent(ctx context.Context, message string, workDir string) RouteResult {
	// ── 规则快速路径 ─────────────────────────────────────────────────────
	addKeywords := []string{"再加", "添加", "新增", "加一页", "加两页", "再加一页", "再加几页"}
	for _, kw := range addKeywords {
		if strings.Contains(message, kw) {
			return RouteResult{
				Intent:      "add_page",
				Reason:      "用户明确要求新增页面",
				TargetPages: extractTargetPages(message),
			}
		}
	}

	// ── LLM 分类 ──────────────────────────────────────────────────
	var tasksSummary string
	manifest, err := deep.ReadTasksManifest(workDir)
	if err == nil && manifest != nil && len(manifest.Tasks) > 0 {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("当前 PPT 共 %d 页：\n", len(manifest.Tasks)))
		for _, t := range manifest.Tasks {
			sb.WriteString(fmt.Sprintf("  第%d页: %s (content_type=%s, status=%s)\n",
				t.PageIndex, t.Title, t.ContentType, t.Status))
		}
		tasksSummary = sb.String()
	} else {
		tasksSummary = "无法读取任务清单"
	}

	result := s.classifyIntentByLLM(ctx, message, tasksSummary)
	if result.Intent != "" && result.Intent != "unknown" {
		return result
	}

	return RouteResult{
		Intent:      "unknown",
		Reason:      "LLM 分类失败或返回 unknown，默认走深度处理",
		TargetPages: extractTargetPages(message),
	}
}

// classifyIntentByLLM 使用 AI 模型对用户的继续消息意图进行分类，
// 并提取结构化的修复详情。返回包含所有提取字段的 RouteResult。
// 为节省成本，使用 textModelFactory（ARK_TEXT_MODEL），必要时回退到 aiModelFactory。
func (s *Server) classifyIntentByLLM(ctx context.Context, message string, tasksSummary string) RouteResult {
	modelFactory := s.textModelFactory
	if modelFactory == nil {
		modelFactory = s.aiModelFactory
	}
	if modelFactory == nil {
		return RouteResult{Intent: "unknown", Reason: "无 AI 模型，跳过 LLM 分类"}
	}

	model, err := modelFactory(ctx)
	if err != nil {
		return RouteResult{Intent: "unknown", Reason: fmt.Sprintf("AI 模型初始化失败: %v", err)}
	}

	prompt := fmt.Sprintf(`你是一个PPT意图分析助手，根据用户的反馈消息判断其真实意图并提取结构化参数。

## 用户反馈
"%s"

## 当前 PPT 任务信息
%s

## 你的任务

分析用户反馈，判断其真实意图，共 5 种：

1. **fix**：用户要求调整样式、布局、颜色、字体、间距、位置等局部细节（如"第2页字体太小"、"换个颜色"、"对齐调整"、"第三张行距太紧"）
2. **regenerate**：用户要求重新生成某些指定页面，但并非全面不满意（如"第1页重新生成"、"第3张做得不好要重做"）
3. **regenerate_all**：用户对整体不满意，要求全部重做（如"全部重来"、"整体重新做"）
4. **add_page**：用户要求新增页面（如"再加一页"、"添加一个新页面"）
5. **needs_clarification**：用户反馈过于模糊，无法判断意图（如"不好"、"不太满意"、"难看"等，没有说明具体哪页、哪里有问题）

## 意图判断技巧

- "不好看/不好/不满意" 但没说哪里 → needs_clarification
- "不好，字太小" → fix
- "不满意，重新做" → regenerate
- "重新生成" → regenerate
- "全部重新做/重做" → regenerate_all
- 提到了具体调整词（颜色/字体/对齐等）→ fix
- 提到新增页面 → add_page
- 页码范围（如"1-3页"）→ regenerate，目标页为 1,2,3
- "从第4页开始" → regenerate，目标页为 4 到最后一页

## fix 意图必须提取的 FixDetails

当意图为 fix 时，必须在 fix_details 字段中填写：
- **aspect**：调整的视觉属性，取值范围：font_size（字号大小）、color（颜色配色）、alignment（对齐）、spacing（间距/行距）、position（位置移动）、style（整体风格）、text_content（文字内容）、layout（布局）、contrast（对比度）、other（其他）
- **detail**：具体调整描述，如"字号调大"、"换成蓝色"、"居中对齐"、"加粗"
- **target_elements**：目标元素，如"标题"、"正文"、"所有文字"、"图表"

## 输出格式（严格返回 JSON，不要任何其他内容）

{
  "intent": "fix|regenerate|regenerate_all|add_page|needs_clarification",
  "reason": "判断理由，1-2句话",
  "target_pages": [2, 3],
  "fix_details": {
    "aspect": "font_size",
    "detail": "调大",
    "target_elements": "标题"
  },
  "regenerate_scope": [1, 2, 3],
  "needs_clarification": false,
  "clarification_question": ""
}

注意事项：
- target_pages：只填用户明确提到的页码，没提到则空数组 []
- fix_details：只有 intent=fix 时才填写，其他情况填 null
- regenerate_scope：intent=regenerate 时填写具体页码数组，intent=regenerate_all 时填空数组 []（表示全部）
- needs_clarification：intent=needs_clarification 时为 true
- clarification_question：needs_clarification=true 时填写要问用户的问题（中文，简短具体），否则填空字符串
`, message, tasksSummary)

	resp, err := model.Generate(ctx, []*schema.Message{
		schema.UserMessage(prompt),
	})
	if err != nil {
		return RouteResult{Intent: "unknown", Reason: fmt.Sprintf("LLM 分类失败: %v", err)}
	}

	content := ""
	if resp != nil {
		content = strings.TrimSpace(resp.Content)
	}

	// Strip markdown code fences
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var raw struct {
		Intent      string `json:"intent"`
		Reason      string `json:"reason"`
		TargetPages []int  `json:"target_pages"`
		FixDetails  *struct {
			Aspect         string `json:"aspect"`
			Detail         string `json:"detail"`
			TargetElements string `json:"target_elements"`
		} `json:"fix_details"`
		RegenerateScope       []int  `json:"regenerate_scope"`
		NeedsClarification    bool   `json:"needs_clarification"`
		ClarificationQuestion string `json:"clarification_question"`
	}
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		return RouteResult{Intent: "unknown", Reason: fmt.Sprintf("LLM 返回格式解析失败: %v", err)}
	}

	if raw.Intent == "" {
		return RouteResult{Intent: "unknown", Reason: "LLM 未返回有效意图"}
	}

	result := RouteResult{
		Intent:                raw.Intent,
		Reason:                raw.Reason,
		TargetPages:           raw.TargetPages,
		NeedsClarification:    raw.NeedsClarification,
		ClarificationQuestion: raw.ClarificationQuestion,
	}

	if raw.FixDetails != nil {
		result.FixDetails = &FixDetails{
			Aspect:         raw.FixDetails.Aspect,
			Detail:         raw.FixDetails.Detail,
			TargetElements: raw.FixDetails.TargetElements,
		}
	}

	// For regenerate, use regenerate_scope if target_pages is empty
	if raw.Intent == "regenerate" && len(raw.RegenerateScope) > 0 {
		result.RegenerateScope = raw.RegenerateScope
		if len(result.TargetPages) == 0 {
			result.TargetPages = raw.RegenerateScope
		}
	}

	return result
}

// extractTargetPages 从用户消息中解析页面引用（从 1 开始）。
func extractTargetPages(message string) []int {
	var pages []int
	msg := strings.ToLower(message)
	for i := 1; i <= 20; i++ {
		patterns := []string{
			fmt.Sprintf("第%d页", i),
			fmt.Sprintf("第%d张", i),
			fmt.Sprintf("%d页", i),
			fmt.Sprintf("%d张", i),
		}
		for _, p := range patterns {
			if strings.Contains(msg, p) {
				pages = append(pages, i)
				break
			}
		}
	}
	return pages
}

func (s *Server) runFixerContinue(taskID string, ts *task.TaskState, route *RouteResult, fixMessage string, ch chan task.SSERichEvent) {
	ch <- task.SSERichEvent{Type: "answer", Content: "检测到是定点修复请求，将使用 Fixer 进行调整...\n"}

	// 读取 manifest 获取任务信息
	manifest, err := deep.ReadTasksManifest(ts.Info.WorkDir)
	if err != nil || manifest == nil {
		ch <- task.SSERichEvent{Type: "error", Error: "无法读取任务清单: " + err.Error()}
		return
	}

	// 使用 LLM 结果中的页面，fallback 到 inferPagesFromMessage
	pagesToFix := route.TargetPages
	if len(pagesToFix) == 0 {
		pagesToFix = inferPagesFromMessage(fixMessage, len(manifest.Tasks))
	}

	if len(pagesToFix) == 0 {
		// 如果没有指定具体页面，修复所有已完成的页面
		for _, t := range manifest.Tasks {
			if t.Status == deep.StatusDone || t.Status == deep.StatusQADone || t.Status == deep.StatusFixed {
				pagesToFix = append(pagesToFix, t.PageIndex)
			}
		}
	}

	if len(pagesToFix) == 0 {
		ch <- task.SSERichEvent{Type: "answer", Content: "未找到需要修复的页面，跳过。\n"}
		return
	}

	// 根据 LLM 修复详情构建 QA 报告
	var qaReport strings.Builder
	qaReport.WriteString("用户请求修复（LLM 意图分析）：")
	if route.FixDetails != nil {
		qaReport.WriteString(fmt.Sprintf("调整方面=%s, 具体要求=%s, 目标元素=%s",
			route.FixDetails.Aspect, route.FixDetails.Detail, route.FixDetails.TargetElements))
	} else {
		qaReport.WriteString(route.Reason)
	}

	// 处理每个页面
	for _, pageIdx := range pagesToFix {
		var targetTask *deep.TaskItem
		for _, t := range manifest.Tasks {
			if t.PageIndex == pageIdx {
				targetTask = t
				break
			}
		}

		if targetTask == nil {
			ch <- task.SSERichEvent{Type: "answer", Content: fmt.Sprintf("第%d页不存在，跳过\n", pageIdx)}
			continue
		}

		ch <- task.SSERichEvent{Type: "answer", Content: fmt.Sprintf("正在修复第%d页: %s (%s)\n", pageIdx, targetTask.Title, targetTask.ContentType)}

		// 读取输出文件以验证它存在
		outputFile := filepath.Join(ts.Info.WorkDir, targetTask.OutputFile)
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			ch <- task.SSERichEvent{Type: "error", Error: fmt.Sprintf("文件 %s 不存在，无法修复", targetTask.OutputFile)}
			continue
		}

		// 修复前更新任务状态
		targetTask.Status = deep.StatusQADone
		targetTask.QAReport = qaReport.String()
		targetTask.FixAttempts++
		deep.WriteTasksManifest(ts.Info.WorkDir, manifest)

		// 使用完整的 tasks.json 构建修复请求上下文
		manifestJSON, _ := json.MarshalIndent(manifest, "", "  ")
		fixRequest := fmt.Sprintf(`用户修复请求：%s

当前 PPT 的 tasks.json 完整内容如下，请从中找到对应页面进行修复：
%s

请读取 tasks.json 后，对第 %d 页（task_id=%s, output_file=%s, content_type=%s）执行修复。
`,
			fixMessage,
			string(manifestJSON),
			targetTask.PageIndex,
			targetTask.TaskID,
			targetTask.OutputFile,
			targetTask.ContentType,
		)

		// 使用 SSE 事件转发运行 Fixer Agent
		fixErr := deep.RunFixerAgentWithCallback(context.Background(),
			ts.Info.WorkDir,
			s.skillDir,
			s.operator,
			fixRequest,
			func(event deep.AgentEvent) {
				switch event.Type {
				case deep.AgentEventAnswer:
					ch <- task.SSERichEvent{Type: "answer", Content: event.Content}
				case deep.AgentEventError:
					ch <- task.SSERichEvent{Type: "error", Error: event.Error}
				}
			},
		)

		if fixErr != nil {
			ch <- task.SSERichEvent{Type: "answer", Content: fmt.Sprintf("第%d页 Fixer 执行出错: %v\n", pageIdx, fixErr)}
		}

		// 检查文件是否实际被更新了（fixer 会重写它）
		if _, err := os.Stat(outputFile); err == nil {
			targetTask.Status = deep.StatusFixed
		}
		deep.WriteTasksManifest(ts.Info.WorkDir, manifest)
	}

	// 刷新文件列表
	s.refreshFileList(ts, ch)

	// 更新最终文件列表
	manifest, _ = deep.ReadTasksManifest(ts.Info.WorkDir)
	var progressTasks []*deep.TaskItem
	if manifest != nil {
		progressTasks = manifest.Tasks
		ts.Mu.Lock()
		ts.Info.Files = task.ManifestOutputFiles(manifest)
		ts.Info.DoneCount = manifest.CompletedCount()
		ts.Info.TotalCount = len(manifest.Tasks)
		ts.Mu.Unlock()
	}

	ch <- task.SSERichEvent{
		Type:   "progress",
		Status: ts.Info.Status,
		Tasks:  progressTasks,
		Done:   ts.Info.DoneCount,
		Total:  ts.Info.TotalCount,
		Files:  ts.Info.Files,
	}
}

// FixResult 保存修复操作的结果。
type FixResult struct {
	Success bool
	Message string
}

func (s *Server) runFixerForTask(workDir string, task *deep.TaskItem, qaReport string) FixResult {
	if task.FixAttempts >= 2 {
		return FixResult{
			Success: false,
			Message: fmt.Sprintf("任务 %s 已达到最大修复次数（2次），跳过修复", task.TaskID),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	fixRequest := fmt.Sprintf("请修复 PPT 任务 %s（第 %d 页: %s）中的以下问题：\n%s\n工作目录: %s",
		task.TaskID, task.PageIndex+1, task.Title, qaReport, workDir)

	var fixErr error
	err := deep.RunFixerAgentWithCallback(ctx, workDir, s.skillDir, s.operator, fixRequest, func(event deep.AgentEvent) {
		// Fixer 事件可通过日志记录，暂不向客户端推送
		if event.Type == deep.AgentEventError {
			fixErr = fmt.Errorf("Fixer 错误: %s", event.Error)
		}
	})

	if err != nil {
		return FixResult{
			Success: false,
			Message: fmt.Sprintf("修复失败: %v", err),
		}
	}
	if fixErr != nil {
		return FixResult{
			Success: false,
			Message: fmt.Sprintf("修复执行出错: %v", fixErr),
		}
	}

	return FixResult{
		Success: true,
		Message: fmt.Sprintf("任务 %s 修复完成", task.TaskID),
	}
}

func (s *Server) runDeepAgentContinue(taskID string, ts *task.TaskState, route *RouteResult, continueMessage string, ch chan task.SSERichEvent) {
	ch <- task.SSERichEvent{Type: "answer", Content: "正在重新处理您的请求...\n"}

	manifest, err := deep.ReadTasksManifest(ts.Info.WorkDir)
	if err != nil || manifest == nil {
		ch <- task.SSERichEvent{Type: "error", Error: "无法读取任务清单"}
		return
	}

	var targetPages []int

	// 处理 add_page 意图
	if route.Intent == "add_page" {
		newTask := &deep.TaskItem{
			TaskID:      fmt.Sprintf("slide-%d", len(manifest.Tasks)+1),
			PageIndex:   len(manifest.Tasks) + 1,
			Title:       "新页面",
			ContentType: "content_slide",
			Description: fmt.Sprintf("用户新增页面（来自对话：%s）", route.Reason),
			OutputFile:  fmt.Sprintf("%d_%s.pptx", len(manifest.Tasks)+1, "新页面"),
			Status:      deep.StatusPending,
		}
		manifest.Tasks = append(manifest.Tasks, newTask)
		targetPages = []int{newTask.PageIndex}
		ch <- task.SSERichEvent{Type: "answer", Content: fmt.Sprintf("新增第%d页: %s\n", newTask.PageIndex, newTask.Title)}
	}

	// Handle regenerate intent (specific pages)
	if route.Intent == "regenerate" {
		pages := route.TargetPages
		if len(pages) == 0 && len(route.RegenerateScope) > 0 {
			pages = route.RegenerateScope
		}
		if len(pages) == 0 {
			ch <- task.SSERichEvent{Type: "answer", Content: "未找到需要重新生成的页面\n"}
		} else {
			for _, pageIdx := range pages {
				for _, t := range manifest.Tasks {
					if t.PageIndex == pageIdx {
						t.Status = deep.StatusPending
						ch <- task.SSERichEvent{Type: "answer", Content: fmt.Sprintf("第%d页 '%s' 已标记为待重新生成\n", pageIdx, t.Title)}
						break
					}
				}
			}
			targetPages = pages
		}
	}

	// 处理 regenerate_all 意图
	if route.Intent == "regenerate_all" {
		for _, t := range manifest.Tasks {
			t.Status = deep.StatusPending
		}
		ch <- task.SSERichEvent{Type: "answer", Content: "所有页面已标记为待重新生成\n"}
		targetPages = nil // nil 表示所有待处理页面
	}

	// 处理 unknown 意图 — 尝试重新生成所有页面
	if route.Intent == "unknown" {
		for _, t := range manifest.Tasks {
			t.Status = deep.StatusPending
		}
		ch <- task.SSERichEvent{Type: "answer", Content: "重新处理所有页面\n"}
		targetPages = nil
	}

	deep.WriteTasksManifest(ts.Info.WorkDir, manifest)

	// 调用 SlideExecutor 实际生成待处理的页面
	ch <- task.SSERichEvent{Type: "answer", Content: "正在调用幻灯片生成器...\n"}

	execErr := deep.RunSlideExecutorContinueWithCallback(context.Background(),
		ts.Info.WorkDir,
		s.skillDir,
		s.operator,
		continueMessage,
		targetPages,
		func(event deep.AgentEvent) {
			switch event.Type {
			case deep.AgentEventAnswer:
				ch <- task.SSERichEvent{Type: "answer", Content: event.Content}
			case deep.AgentEventError:
				ch <- task.SSERichEvent{Type: "error", Error: event.Error}
			}
		},
	)

	if execErr != nil {
		ch <- task.SSERichEvent{Type: "answer", Content: fmt.Sprintf("幻灯片生成出错: %v\n", execErr)}
	}

	// 刷新文件列表
	s.refreshFileList(ts, ch)

	// 更新任务状态：将成功生成的页面标记为完成
	manifest, _ = deep.ReadTasksManifest(ts.Info.WorkDir)
	var progressTasks []*deep.TaskItem
	if manifest != nil {
		progressTasks = manifest.Tasks
		ts.Mu.Lock()
		ts.Info.Files = task.ManifestOutputFiles(manifest)
		ts.Info.DoneCount = manifest.CompletedCount()
		ts.Info.TotalCount = len(manifest.Tasks)
		ts.Mu.Unlock()
	}

	ch <- task.SSERichEvent{
		Type:   "progress",
		Status: ts.Info.Status,
		Tasks:  progressTasks,
		Done:   ts.Info.DoneCount,
		Total:  ts.Info.TotalCount,
		Files:  ts.Info.Files,
	}
}

func inferPagesFromMessage(message string, maxPage int) []int {
	var pages []int
	msg := strings.ToLower(message)
	for i := 1; i <= maxPage && i <= 20; i++ {
		patterns := []string{
			fmt.Sprintf("第%d页", i),
			fmt.Sprintf("第%d张", i),
			fmt.Sprintf("%d页", i),
			fmt.Sprintf("%d张", i),
		}
		for _, p := range patterns {
			if strings.Contains(msg, p) {
				pages = append(pages, i)
				break
			}
		}
	}
	return pages
}

func (s *Server) refreshFileList(ts *task.TaskState, ch chan task.SSERichEvent) {
	entries, err := os.ReadDir(ts.Info.WorkDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".pptx") {
			ts.Mu.Lock()
			if !ts.ReportedFiles()[entry.Name()] {
				ts.SetReportedFile(entry.Name())
				ts.Mu.Unlock()
				evt := task.SSERichEvent{
					Type:     "file_ready",
					ToolName: entry.Name(),
					Files:    []string{task.CanonicalOutputFile(entry.Name())},
				}
				ch <- evt
			} else {
				ts.Mu.Unlock()
			}
		}
	}
}

// handleGetConversation 返回任务的对话历史记录。
// 如果任务在内存中，返回实时会话数据；如果不在（冷启动），从数据库重建完整历史。
func (s *Server) handleGetConversation(c *gin.Context) {
	taskID := c.Param("id")

	ts := s.tasks.GetTaskState(taskID)

	// 任务在内存中，直接返回实时会话数据（包含累计的 full_answer）。
	if ts != nil {
		fullAnswer := ts.FullAnswer()
		sess := s.sessionManager.GetOrCreate(taskID, ts.Info.WorkDir)
		snapshot := sess.Snapshot()
		info := ts.SnapshotInfo()
		messages := conversationMessagesWithFallback(snapshot.Messages, fullAnswer, info.ConversationContent, snapshot.UpdatedAt)
		c.JSON(http.StatusOK, gin.H{
			"task_id":              taskID,
			"messages":             messages,
			"full_answer":          fullAnswer,
			"conversation_content": info.ConversationContent,
			"status":               info.Status,
			"done_count":           info.DoneCount,
			"total_count":          info.TotalCount,
			"files":                task.DeduplicateOutputFiles(info.Files),
			"duration":             info.Duration,
			"prompt_tokens":        info.PromptTokens,
			"completion_tokens":    info.CompletionTokens,
			"total_tokens":         info.TotalTokens,
			"created_at":           snapshot.CreatedAt,
			"updated_at":           snapshot.UpdatedAt,
		})
		return
	}

	// 冷启动：从数据库重建完整对话历史。
	// 先从 task_records 获取 ConversationContent（已拼接的完整摘要）。
	info := s.tasks.GetTask(taskID)
	if info == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	// 追加 conversation_messages 中的助手消息（在 stream 已写入的部分）。
	dbMsgs, err := db.ListConversationMessages(taskID)
	if err != nil {
		dbMsgs = nil
	}

	// 按时间顺序合并：用户消息 + 助手消息。
	var messages []session.Message
	for _, m := range dbMsgs {
		messages = append(messages, session.Message{
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: m.Timestamp,
		})
	}
	messages = conversationMessagesWithFallback(messages, info.FullAnswer, info.ConversationContent, info.CreatedAt)

	c.JSON(http.StatusOK, gin.H{
		"task_id":              taskID,
		"messages":             messages,
		"conversation_content": info.ConversationContent,
		"full_answer":          info.FullAnswer,
		"status":               info.Status,
		"done_count":           info.DoneCount,
		"total_count":          info.TotalCount,
		"files":                task.DeduplicateOutputFiles(info.Files),
		"duration":             info.Duration,
		"prompt_tokens":        info.PromptTokens,
		"completion_tokens":    info.CompletionTokens,
		"total_tokens":         info.TotalTokens,
		"created_at":           info.CreatedAt,
		"updated_at":           info.CreatedAt,
	})
}

// ── 用户资料处理器 ───────────────────────────────────────────────────────

func (s *Server) handleGetUserProfile(c *gin.Context) {
	uid := userIDGin(c)
	p := s.styleStore.Get(uid)

	c.JSON(http.StatusOK, gin.H{
		"user_id":    p.UserID,
		"task_count": p.TaskCount,
		"profile":    p,
		"updated_at": p.UpdatedAt,
	})
}

func (s *Server) handleUpdateUserProfile(c *gin.Context) {
	uid := userIDGin(c)

	var req struct {
		PreferredThemes   []string `json:"preferred_themes"`
		PreferredColors   []string `json:"preferred_colors"`
		ContentPatterns   []string `json:"content_patterns"`
		LayoutPreferences []string `json:"layout_preferences"`
		LanguageTone      string   `json:"language_tone"`
		TypicalPageCount  int      `json:"typical_page_count"`
		SpecialNotes      []string `json:"special_notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	p := s.styleStore.Get(uid)

	if req.PreferredThemes != nil {
		p.PreferredThemes = req.PreferredThemes
	}
	if req.PreferredColors != nil {
		p.PreferredColors = req.PreferredColors
	}
	if req.ContentPatterns != nil {
		p.ContentPatterns = req.ContentPatterns
	}
	if req.LayoutPreferences != nil {
		p.LayoutPreferences = req.LayoutPreferences
	}
	if req.LanguageTone != "" {
		p.LanguageTone = req.LanguageTone
	}
	if req.TypicalPageCount > 0 {
		p.TypicalPageCount = req.TypicalPageCount
	}
	if req.SpecialNotes != nil {
		p.SpecialNotes = req.SpecialNotes
	}

	s.styleStore.Save(p)
	c.JSON(http.StatusOK, gin.H{"status": "ok", "profile": p})
}

func (s *Server) handleResetUserProfile(c *gin.Context) {
	uid := userIDGin(c)
	p := s.styleStore.Get(uid)

	// Reset to default
	p.PreferredThemes = nil
	p.PreferredColors = nil
	p.ContentPatterns = nil
	p.LayoutPreferences = nil
	p.LanguageTone = ""
	p.TypicalPageCount = 0
	p.ContentTypes = nil
	p.SpecialNotes = nil
	p.TaskCount = 0

	s.styleStore.Save(p)
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "风格偏好已重置"})
}

// handleSummarizeProfile returns the LLM-summarized user preferences
// (already stored in styleStore from previous task completions).
// User can review and manually edit before saving via PUT /me/profile.
func (s *Server) handleSummarizeProfile(c *gin.Context) {
	uid := userIDGin(c)
	p := s.styleStore.Get(uid)

	c.JSON(http.StatusOK, gin.H{
		"summary": gin.H{
			"preferred_themes":   p.PreferredThemes,
			"preferred_colors":   p.PreferredColors,
			"content_patterns":   p.ContentPatterns,
			"layout_preferences": p.LayoutPreferences,
			"language_tone":      p.LanguageTone,
			"typical_page_count": p.TypicalPageCount,
			"special_notes":      p.SpecialNotes,
		},
		"task_count": p.TaskCount,
		"updated_at": p.UpdatedAt,
	})
}

// updateUserStyleFromTask is called when a task completes to extract and save style preferences.
// 核心逻辑由 LLM 分析 PPTX 文本内容完成，fallback 到规则提取。
func (s *Server) updateUserStyleFromTask(userID int, workDir string, query string) {
	manifest, err := deep.ReadTasksManifest(workDir)
	if err != nil || manifest == nil {
		return
	}

	// 将 tasks 转换为 TaskItemInfo 用于回退
	taskInfos := make([]*style.TaskItemInfo, len(manifest.Tasks))
	for i, t := range manifest.Tasks {
		taskInfos[i] = &style.TaskItemInfo{
			ContentType: t.ContentType,
			Theme:       manifest.Theme,
		}
	}

	// 如果配置了 LLM 提取器，使用 LLM 分析
	if s.styleExtractor != nil {
		ctx := context.Background()
		extracted, err := s.styleExtractor.ExtractFromPPTX(ctx, workDir, query, manifest.Theme, taskInfos)
		if err == nil && extracted != nil {
			s.styleStore.UpdateWithTask(userID, extracted)
			return
		}
		// LLM 提取失败，fallback 到规则提取
	}

	// Fallback：使用规则提取
	extracted := style.ExtractFromTasks(taskInfos, manifest.Theme)
	s.styleStore.UpdateWithTask(userID, extracted)
}

// handleListLogAnalyses 返回最近的日志分析列表。
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

// ── Feedback & Learning handlers ──────────────────────────────────────────────

// FeedbackRequest 对应前端 FeedbackRequest 结构
type FeedbackRequest struct {
	Type      string  `json:"type"` // rating/edit/completion/abandon
	TaskID    string  `json:"task_id,omitempty"`
	Rating    float64 `json:"rating,omitempty"`
	PageIndex int     `json:"page_index,omitempty"`
	Before    string  `json:"before,omitempty"`
	After     string  `json:"after,omitempty"`
	Reason    string  `json:"reason,omitempty"`
	Progress  float64 `json:"progress,omitempty"`
}

func (s *Server) handleSubmitFeedback(c *gin.Context) {
	var req FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求: " + err.Error()})
		return
	}

	uid := userIDGin(c)

	engine := deep.GetLearningEngine()
	if engine == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "学习引擎未初始化"})
		return
	}

	feedback := &agentlearning.Feedback{
		Type:      req.Type,
		Rating:    req.Rating,
		PageIndex: req.PageIndex,
		Before:    req.Before,
		After:     req.After,
		Duration:  0,
		Reason:    req.Reason,
		Progress:  req.Progress,
		Data: map[string]interface{}{
			"task_id": req.TaskID,
		},
	}

	engine.RecordFeedback(uid, req.TaskID, feedback)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) handleGetUserInsights(c *gin.Context) {
	uid := userIDGin(c)

	engine := deep.GetLearningEngine()
	if engine == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "学习引擎未初始化"})
		return
	}

	insights := engine.GetUserInsights(uid)
	if insights == nil {
		c.JSON(http.StatusOK, gin.H{"insights": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"insights": insights})
}

func (s *Server) handleGetRecommendations(c *gin.Context) {
	uid := userIDGin(c)
	domain := c.Query("domain")

	engine := deep.GetLearningEngine()
	if engine == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "学习引擎未初始化"})
		return
	}

	rec := engine.GetRecommendations(uid, domain)
	c.JSON(http.StatusOK, gin.H{"recommendation": rec})
}

// onTaskContinue 任务完成后自动触发继续处理（TaskManager 通过回调调用）。
// 它从等待队列中取出消息，重新启动 SSE 流并处理继续逻辑。
func (s *Server) onTaskContinue(taskID string) {
	ts := s.tasks.GetTaskState(taskID)
	if ts == nil {
		return
	}

	pendingMsg := ts.GetPendingContinueMsg()
	if pendingMsg == "" {
		return
	}

	uid := ts.Info.UserID
	sess := s.sessionManager.GetOrCreate(taskID, ts.Info.WorkDir)
	sess.AddUserMessage(pendingMsg)

	// 重新标记为运行状态
	ts.Mu.Lock()
	ts.Info.Status = task.TaskStatusRunning
	ts.Mu.Unlock()
	ts.Persist()

	s.startContinue(taskID, ts, pendingMsg, uid, sess)
}

// ── 管理员 API ───────────────────────────────────────────────────────────

func (s *Server) handleAdminUsers(c *gin.Context) {
	users, err := db.ListAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户列表失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (s *Server) handleAdminTasks(c *gin.Context) {
	tasks, err := db.ListAllTaskRecords(100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询任务列表失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func (s *Server) handleAdminLogAnalyses(c *gin.Context) {
	analyses, err := db.ListRecentErrorAnalyses(50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询日志分析失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"analyses": analyses})
}

func (s *Server) handleAdminStyleProfiles(c *gin.Context) {
	profiles, err := db.ListAllStyleProfiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询风格偏好失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"profiles": profiles})
}

func (s *Server) handleAdminStats(c *gin.Context) {
	if db.DB == nil {
		c.JSON(http.StatusOK, gin.H{
			"user_count":    0,
			"task_count":    0,
			"running_count": 0,
		})
		return
	}

	var userCount, taskCount, runningCount int64
	db.DB.Model(&db.User{}).Count(&userCount)
	db.DB.Model(&db.TaskRecord{}).Count(&taskCount)
	db.DB.Model(&db.TaskRecord{}).Where("status = ?", "running").Count(&runningCount)

	c.JSON(http.StatusOK, gin.H{
		"user_count":    userCount,
		"task_count":    taskCount,
		"running_count": runningCount,
	})
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
