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

	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
	agentlearning "github.com/cloudwego/ppt-agent/pkg/agent/learning"
	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
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
	c.JSON(http.StatusOK, gin.H{
		"id":       uid,
		"email":    email,
		"is_admin": isAdminGin(c),
	})
}

// ── 任务处理器 ─────────────────────────────────────────────────────────

func (s *Server) handleCreateTask(c *gin.Context) {
	var req struct {
		Query             string             `json:"query"`
		Outline           *deck.TaskOutline  `json:"outline,omitempty"`
		TemplateSelection *TemplateSelection `json:"template_selection,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Query) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}

	taskID := s.taskIDGen()
	cfg := s.makeTaskConfig(taskID)
	cfg.Query = req.Query
	uid := userIDGin(c)

	// 推荐模式需要先拿到结构化意图结果，模板选择与 TaskManager 共用这一次分类。
	if req.Outline == nil && req.TemplateSelection != nil && strings.EqualFold(strings.TrimSpace(req.TemplateSelection.Mode), "recommended") {
		routingCtx, cancel := context.WithTimeout(c.Request.Context(),
			time.Duration(agentutils.EnvInt("INTENT_ROUTE_TIMEOUT_SECONDS", 30))*time.Second)
		intentCfg, err := deck.ProcessUserIntent(routingCtx, req.Query, uid)
		cancel()
		if err != nil {
			logger.Warn("template_recommendation_intent_failed", "error", err.Error())
		} else if intentCfg != nil {
			mergeIntentTaskConfig(cfg, intentCfg)
		}
	}

	// 如果有 outline，先做服务端兜底校验/补齐，再写入 tasks.json。
	if req.Outline != nil && len(req.Outline.Slides) > 0 {
		outline, err := s.prepareOutline(c.Request.Context(), req.Query, req.Outline)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "outline 处理失败: " + err.Error()})
			return
		}
		cfg.Outline = outline
	} else if req.TemplateSelection != nil {
		outline, _, err := s.resolveTemplateSelectionWithIntent(req.Query, req.TemplateSelection, cfg.IntentResult)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "模板选择失败: " + err.Error()})
			return
		}
		cfg.Outline = outline
	}

	// 用户偏好由意图识别后的 domain-aware 路径统一注入，避免在未知领域时提前强绑定历史风格。

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

func mergeIntentTaskConfig(target, source *deck.PPTTaskConfig) {
	if target == nil || source == nil {
		return
	}
	target.UserID = source.UserID
	target.IntentResult = source.IntentResult
	target.RoutingDecision = source.RoutingDecision
	target.EnhancedProfile = source.EnhancedProfile
	if strings.TrimSpace(source.StyleContext) != "" {
		if strings.TrimSpace(target.StyleContext) != "" {
			target.StyleContext += "\n" + source.StyleContext
		} else {
			target.StyleContext = source.StyleContext
		}
	}
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
			if refs := s.backgroundImageRefs(name); len(refs) > 0 {
				c.File(filepath.Join(s.templateLoader.GetBackgroundTemplatesDir(), filepath.FromSlash(refs[0])))
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

输出只包含内容描述。容量要求表示正常信息密度目标，可根据事实完整性在可渲染范围内适度调整。`, req.Title, layoutName, req.Description, themeName, layoutDescriptionTarget(req.ContentType), layoutName)

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

func (s *Server) prepareOutline(_ context.Context, query string, outline *deck.TaskOutline) (*deck.TaskOutline, error) {
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
	if outline.ContentMode != deck.OutlineContentModeTemplateScaffold {
		outline.ContentMode = deck.OutlineContentModeUserOutline
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
func (s *Server) enrichOutline(ctx context.Context, query string, outline *deck.TaskOutline) (*deck.TaskOutline, error) {
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

func outlineNeedsEnrichment(outline *deck.TaskOutline) bool {
	for _, slide := range outline.Slides {
		if len([]rune(strings.TrimSpace(slide.Description))) < 30 || slide.ContentPlan == nil {
			return true
		}
	}
	return false
}

func (s *Server) mergeOutlineSlides(base *deck.TaskOutline, enriched []deck.SlideOutline) *deck.TaskOutline {
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
	name = strings.TrimSpace(name)
	for _, bg := range s.templateLoader.ListBackgrounds() {
		if bg.Name == name {
			return true
		}
	}
	return s.isValidBackgroundReference(name)
}

func (s *Server) isValidBackgroundReference(name string) bool {
	name = filepath.ToSlash(strings.TrimSpace(name))
	if name == "" || strings.HasPrefix(name, "/") || filepath.IsAbs(name) || strings.Contains(name, "..") {
		return false
	}
	ext := strings.ToLower(filepath.Ext(name))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return false
	}
	root, err := filepath.Abs(s.templateLoader.GetBackgroundTemplatesDir())
	if err != nil {
		return false
	}
	target, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(name)))
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(root, target)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return false
	}
	info, err := os.Stat(target)
	return err == nil && !info.IsDir()
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
		return "4-6条要点，每条目标35-60字，事实、解释和结论组合完整"
	case "two_column":
		return "左右各3-5条，每条目标30-55字，体现可比较的共同维度"
	case "three_column":
		return "每栏2-4条，每条目标24-45字，三栏粒度保持一致"
	case "card_grid":
		return "4-6张卡片，header 6-16字，body目标60-100字"
	case "process_flow":
		return "4-6个步骤，每步包含动作、对象和结果"
	case "summary_slide":
		return "3-5条总结，每条目标35-60字，可附行动建议或展望"
	case "kpi_dashboard", "stat_slide", "chart_slide", "comparison_table":
		return "120-220字，必须包含可结构化的数据、指标、分类或表格字段"
	case "image_text", "case_study", "example_detail", "deep_dive":
		return "180-320字，允许稍长，但必须分清段落、案例、数据和结论"
	case "quote_slide":
		return "quote目标35-110字，attribution目标8-32字，保留完整观点与出处"
	default:
		return "100-180字，内容密度匹配页面容量"
	}
}

func (s *Server) generateOutlineSlides(ctx context.Context, query string, outline *deck.TaskOutline) ([]deck.SlideOutline, error) {
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

## 可用背景主题
%s

## 页面结构
%s

## 输出格式
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

## 生成契约
- slides 数量和顺序与页面结构保持一致。
- content_type 使用页面结构中的稳定英文 id。
- background 从可用背景 id 中选择；视觉叙事页优先，数据密集页使用清晰信息表面。
- description 以各页容量目标为正常密度，内容完整性与版式容量共同决定最终长度。
- content_plan 使用 bullet_list、numbered_list、key_point_card、table、chart_placeholder、callout、quote 等结构化字段。
- 每页包含具体实体、数据、场景或案例；真实数据明确搜索项和来源方向。
- 输出为一个 JSON 对象。`, query, themeName, s.backgroundCatalog(), slideContexts.String())

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
		Slides []deck.SlideOutline `json:"slides"`
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
		Outline *deck.TaskOutline `json:"outline"`
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
		if ts.Info.Status == task.TaskStatusRunning {
			ts.Info.Status = task.TaskStatusCompleted
		}
		ts.Info.FullAnswer = fullAnswer
		ts.Mu.Unlock()
		ts.Persist()
	}()
	go s.runContinue(taskID, ts, message, uid, sess, ch)
}

func (s *Server) runContinue(taskID string, ts *task.TaskState, message string, uid int, sess *session.ConversationSession, ch chan task.SSERichEvent) {
	defer close(ch)

	ctx := context.Background()

	ch <- task.SSERichEvent{Type: "answer", Content: "正在分析您的请求...\n"}

	route := s.routeContinueIntent(ctx, message, ts.Info.WorkDir)

	ch <- task.SSERichEvent{Type: "answer", Content: fmt.Sprintf("识别意图: %s (%s)\n", route.Intent, route.Reason)}

	switch route.Intent {
	case "needs_clarification":
		question := route.ClarificationQuestion
		if question == "" {
			question = "您的反馈比较模糊，请问您是指哪一页？希望怎么改善？（比如：第2页字体太小、第3张换个颜色、第1页重新生成等）"
		}
		ch <- task.SSERichEvent{Type: "clarification", Content: question}
		ch <- task.SSERichEvent{Type: "continue_complete", Message: question}
		return
	default:
		s.runWorkflowContinue(taskID, ts, &route, message, ch)
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
	manifest, err := deck.ReadTasksManifest(workDir)
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

func (s *Server) runWorkflowContinue(taskID string, ts *task.TaskState, route *RouteResult, continueMessage string, ch chan task.SSERichEvent) {
	ch <- task.SSERichEvent{Type: "answer", Content: "正在更新页面计划并重新渲染...\n"}

	manifest, err := deck.ReadTasksManifest(ts.Info.WorkDir)
	if err != nil || manifest == nil {
		ch <- task.SSERichEvent{Type: "error", Error: "无法读取任务清单"}
		markTaskFailed(ts, "无法读取任务清单")
		return
	}

	switch route.Intent {
	case "add_page":
		nextPage := len(manifest.Tasks) + 1
		title := "新增页面"
		if strings.TrimSpace(route.Reason) != "" {
			title = "补充说明"
		}
		newTask := &deck.TaskItem{
			TaskID:      fmt.Sprint(nextPage),
			PageIndex:   nextPage,
			Title:       title,
			ContentType: "content_slide",
			Description: fmt.Sprintf("根据用户继续请求新增页面：%s", continueMessage),
			OutputFile:  fmt.Sprintf("%d_%s.pptx", nextPage, title),
			Status:      deck.StatusPending,
			ContentPlan: &deck.ContentPlan{
				Summary: fmt.Sprintf("补充说明用户要求：%s", continueMessage),
				Elements: []deck.ContentElement{{
					Type:  "bullet_list",
					Items: []string{continueMessage, "围绕原演示主题补充新的信息点", "保持与前后页面一致的叙事和视觉风格"},
				}},
			},
		}
		manifest.Tasks = append(manifest.Tasks, newTask)
		ch <- task.SSERichEvent{Type: "answer", Content: fmt.Sprintf("新增第%d页: %s\n", newTask.PageIndex, newTask.Title)}
	case "fix":
		pages := route.TargetPages
		if len(pages) == 0 {
			pages = inferPagesFromMessage(continueMessage, len(manifest.Tasks))
		}
		if len(pages) == 0 {
			for _, item := range manifest.Tasks {
				if item != nil && (item.Status == deck.StatusDone || item.Status == deck.StatusQADone || item.Status == deck.StatusFixed) {
					pages = append(pages, item.PageIndex)
				}
			}
		}
		if len(pages) == 0 {
			ch <- task.SSERichEvent{Type: "answer", Content: "未找到需要调整的页面，跳过。\n"}
			return
		}
		instruction := buildContinueInstruction(route, continueMessage)
		for _, pageIdx := range pages {
			if item := findManifestTaskByPage(manifest, pageIdx); item != nil {
				markTaskForRerender(ts.Info.WorkDir, item, instruction, true)
				ch <- task.SSERichEvent{Type: "answer", Content: fmt.Sprintf("第%d页 '%s' 已加入重渲染队列\n", pageIdx, item.Title)}
			}
		}
	case "regenerate":
		pages := route.TargetPages
		if len(pages) == 0 && len(route.RegenerateScope) > 0 {
			pages = route.RegenerateScope
		}
		if len(pages) == 0 {
			ch <- task.SSERichEvent{Type: "answer", Content: "未找到需要重新生成的页面\n"}
		} else {
			for _, pageIdx := range pages {
				if item := findManifestTaskByPage(manifest, pageIdx); item != nil {
					markTaskForRerender(ts.Info.WorkDir, item, continueMessage, false)
					ch <- task.SSERichEvent{Type: "answer", Content: fmt.Sprintf("第%d页 '%s' 已标记为待重新生成\n", pageIdx, item.Title)}
				}
			}
		}
	case "regenerate_all", "unknown":
		for _, item := range manifest.Tasks {
			markTaskForRerender(ts.Info.WorkDir, item, continueMessage, false)
		}
		ch <- task.SSERichEvent{Type: "answer", Content: "所有页面已标记为待重新生成\n"}
	}

	if err := deck.WriteTasksManifest(ts.Info.WorkDir, manifest); err != nil {
		ch <- task.SSERichEvent{Type: "error", Error: fmt.Sprintf("更新任务清单失败: %v", err)}
		markTaskFailed(ts, err.Error())
		return
	}

	runtimeMeta := agentutils.NewRuntimeMeta(taskID, ts.Info.WorkDir)
	runtimeMeta.RecordPhase("rendering", "继续请求已写入 DeckSpec，开始并发渲染")
	cfg := &deck.PPTTaskConfig{
		WorkDir:     ts.Info.WorkDir,
		TaskID:      taskID,
		SkillsDir:   s.skillDir,
		Operator:    s.operator,
		RuntimeMeta: runtimeMeta,
		Concurrency: 5,
	}
	if _, err := deck.RenderDeckByTaskIDWorkflow(context.Background(), cfg, func(event deck.DeckRenderEvent) {
		ch <- deckRenderSSE(event)
	}); err != nil {
		ch <- task.SSERichEvent{Type: "error", Error: fmt.Sprintf("幻灯片生成出错: %v", err)}
		markTaskFailed(ts, err.Error())
	}

	s.refreshFileList(ts, ch)

	manifest, _ = deck.ReadTasksManifest(ts.Info.WorkDir)
	var progressTasks []*deck.TaskItem
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

func findManifestTaskByPage(manifest *deck.TasksManifest, pageIdx int) *deck.TaskItem {
	if manifest == nil {
		return nil
	}
	for _, item := range manifest.Tasks {
		if item != nil && item.PageIndex == pageIdx {
			return item
		}
	}
	return nil
}

func markTaskForRerender(workDir string, item *deck.TaskItem, instruction string, isFix bool) {
	if item == nil {
		return
	}
	if strings.TrimSpace(instruction) != "" {
		item.QAReport = instruction
		if isFix {
			item.Description = strings.TrimSpace(item.Description + "\n继续修改要求：" + instruction)
			item.FixAttempts++
		}
	}
	item.Status = deck.StatusPending
	if item.OutputFile != "" {
		_ = os.Remove(filepath.Join(workDir, item.OutputFile))
	}
}

func buildContinueInstruction(route *RouteResult, message string) string {
	if route != nil && route.FixDetails != nil {
		return fmt.Sprintf("用户继续请求：%s；调整方面=%s；具体要求=%s；目标元素=%s",
			message, route.FixDetails.Aspect, route.FixDetails.Detail, route.FixDetails.TargetElements)
	}
	if route != nil && strings.TrimSpace(route.Reason) != "" {
		return fmt.Sprintf("用户继续请求：%s；意图判断：%s", message, route.Reason)
	}
	return "用户继续请求：" + message
}

func deckRenderSSE(event deck.DeckRenderEvent) task.SSERichEvent {
	switch event.Type {
	case "workflow_start":
		return task.SSERichEvent{Type: "progress", Phase: "rendering", PhaseDetail: event.Detail}
	case "slide_start":
		return task.SSERichEvent{Type: "progress", Phase: "rendering", PhaseDetail: fmt.Sprintf("开始生成第 %d 页：%s", event.PageIndex, event.Detail)}
	case "slide_done":
		return task.SSERichEvent{Type: "progress", Phase: "rendering", PhaseDetail: fmt.Sprintf("第 %d 页生成完成：%s", event.PageIndex, event.OutputFile)}
	case "slide_error":
		return task.SSERichEvent{Type: "error", Error: fmt.Sprintf("第 %d 页生成失败：%s", event.PageIndex, event.Error), Phase: "rendering", PhaseDetail: event.OutputFile}
	default:
		return task.SSERichEvent{Type: "progress", Phase: "rendering", PhaseDetail: event.Detail}
	}
}

func markTaskFailed(ts *task.TaskState, message string) {
	if ts == nil {
		return
	}
	ts.Mu.Lock()
	defer ts.Mu.Unlock()
	ts.Info.Status = task.TaskStatusFailed
	ts.Info.Error = message
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

	// 只有运行中的任务需要读取内存态实时会话。终态任务即使仍暂存在内存
	// map 中，也走持久化快照，避免被收尾 goroutine 的 TaskState.Mu 牵住。
	if ts != nil && ts.Info.Status == task.TaskStatusRunning {
		fullAnswer := ts.FullAnswer()
		sess := s.sessionManager.GetOrCreate(taskID, ts.Info.WorkDir)
		snapshot := sess.Snapshot()
		info := ts.SnapshotInfo()
		runtimeMeta := conversationRuntimeMeta(taskID, info.WorkDir)
		messages := conversationMessagesWithFallback(snapshot.Messages, fullAnswer, info.ConversationContent, snapshot.UpdatedAt)
		if info.Status == task.TaskStatusRunning {
			// The unfinished turn is replayed from replay_after_event_id via SSE.
			messages = conversationMessagesWithFallback(snapshot.Messages, "", "", snapshot.UpdatedAt)
		}
		latestEventID, replayAfterEventID := ts.EventBoundaries()
		c.JSON(http.StatusOK, gin.H{
			"task_id":               taskID,
			"latest_event_id":       latestEventID,
			"replay_after_event_id": replayAfterEventID,
			"messages":              messages,
			"full_answer":           fullAnswer,
			"conversation_content":  info.ConversationContent,
			"status":                info.Status,
			"done_count":            info.DoneCount,
			"total_count":           info.TotalCount,
			"files":                 task.DeduplicateOutputFiles(info.Files),
			"duration":              info.Duration,
			"prompt_tokens":         info.PromptTokens,
			"completion_tokens":     info.CompletionTokens,
			"total_tokens":          info.TotalTokens,
			"runtime_meta":          runtimeMeta,
			"created_at":            snapshot.CreatedAt,
			"updated_at":            snapshot.UpdatedAt,
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
	runtimeMeta := conversationRuntimeMeta(taskID, info.WorkDir)

	c.JSON(http.StatusOK, gin.H{
		"task_id":               taskID,
		"latest_event_id":       uint64(0),
		"replay_after_event_id": uint64(0),
		"messages":              messages,
		"conversation_content":  info.ConversationContent,
		"full_answer":           info.FullAnswer,
		"status":                info.Status,
		"done_count":            info.DoneCount,
		"total_count":           info.TotalCount,
		"files":                 task.DeduplicateOutputFiles(info.Files),
		"duration":              info.Duration,
		"prompt_tokens":         info.PromptTokens,
		"completion_tokens":     info.CompletionTokens,
		"total_tokens":          info.TotalTokens,
		"runtime_meta":          runtimeMeta,
		"created_at":            info.CreatedAt,
		"updated_at":            info.CreatedAt,
	})
}

var listRuntimeEvents = db.ListRuntimeEventSummaries
var getRuntimeEvent = db.GetRuntimeEvent

func conversationRuntimeMeta(taskID, workDir string) *agentutils.RuntimeMetaSnapshot {
	snapshot, _ := agentutils.LoadRuntimeMetaSnapshot(workDir)
	events := conversationRuntimeEvents(taskID)
	if snapshot == nil {
		if len(events) == 0 {
			return nil
		}
		snapshot = &agentutils.RuntimeMetaSnapshot{TaskID: taskID, WorkDir: workDir}
	}
	if len(events) > 0 {
		snapshot.RecentEvents = events
		snapshot.EventCounts = runtimeEventCounts(events)
	}
	if snapshot.TaskID == "" {
		snapshot.TaskID = taskID
	}
	if snapshot.WorkDir == "" {
		snapshot.WorkDir = workDir
	}
	return snapshot
}

func conversationRuntimeEvents(taskID string) []agentutils.RuntimeEvent {
	records, err := listRuntimeEvents(taskID)
	if err != nil || len(records) == 0 {
		return nil
	}
	events := make([]agentutils.RuntimeEvent, 0, len(records))
	for _, record := range records {
		events = append(events, runtimeEventFromRecord(record, false))
	}
	return events
}

func (s *Server) handleGetRuntimeEvent(c *gin.Context) {
	taskID := c.Param("id")
	eventID, err := strconv.ParseInt(c.Param("event_id"), 10, 64)
	if err != nil || eventID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event_id"})
		return
	}
	record, err := getRuntimeEvent(taskID, eventID)
	if err != nil || record == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "runtime event not found"})
		return
	}
	c.JSON(http.StatusOK, runtimeEventFromRecord(*record, true))
}

func runtimeEventFromRecord(record db.RuntimeEventRecord, includeMetadata bool) agentutils.RuntimeEvent {
	event := agentutils.RuntimeEvent{
		ID:        record.EventID,
		TaskID:    record.TaskID,
		Timestamp: record.Timestamp.Format(time.RFC3339Nano),
		ElapsedMS: record.ElapsedMS,
		Kind:      record.Kind,
		Phase:     record.Phase,
		Name:      record.Name,
		Status:    record.Status,
		Detail:    record.Detail,
	}
	if includeMetadata && strings.TrimSpace(record.Metadata) != "" {
		metadata := map[string]any{}
		if err := json.Unmarshal([]byte(record.Metadata), &metadata); err == nil {
			event.Metadata = metadata
		} else {
			event.Metadata = map[string]any{"raw": record.Metadata}
		}
	}
	return event
}

func runtimeEventCounts(events []agentutils.RuntimeEvent) map[string]int {
	if len(events) == 0 {
		return nil
	}
	counts := make(map[string]int)
	for _, event := range events {
		kind := strings.TrimSpace(event.Kind)
		if kind == "" {
			kind = "event"
		}
		counts[kind]++
	}
	return counts
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
		PreferredThemes   []string         `json:"preferred_themes"`
		PreferredColors   []string         `json:"preferred_colors"`
		ContentPatterns   []string         `json:"content_patterns"`
		LayoutPreferences []string         `json:"layout_preferences"`
		LanguageTone      string           `json:"language_tone"`
		TypicalPageCount  int              `json:"typical_page_count"`
		SpecialNotes      []string         `json:"special_notes"`
		UserFacts         *style.UserFacts `json:"user_facts"`
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
	if req.UserFacts != nil {
		p.UserFacts = *req.UserFacts
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
			"user_facts":         p.UserFacts,
		},
		"task_count": p.TaskCount,
		"updated_at": p.UpdatedAt,
	})
}

// updateUserStyleFromTask is called when a task completes to extract and save style preferences.
// 核心逻辑由 LLM 分析 PPTX 文本内容完成，fallback 到规则提取。
func (s *Server) updateUserStyleFromTask(userID int, workDir string, query string) {
	manifest, err := deck.ReadTasksManifest(workDir)
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

	engine := deck.GetLearningEngine()
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

	engine := deck.GetLearningEngine()
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

	engine := deck.GetLearningEngine()
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
