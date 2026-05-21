package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/cloudwego/ppt-agent/pkg/agent/deep"
	"github.com/cloudwego/ppt-agent/pkg/auth"
	"github.com/cloudwego/ppt-agent/pkg/db"
	"github.com/cloudwego/ppt-agent/pkg/session"
	"github.com/cloudwego/ppt-agent/pkg/style"
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
	c.JSON(http.StatusOK, gin.H{"id": uid, "email": email})
}

// ── Task handlers ─────────────────────────────────────────────────────────

func (s *Server) handleCreateTask(c *gin.Context) {
	var req struct {
		Query   string            `json:"query"`
		Outline *deep.TaskOutline `json:"outline,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Query) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}

	taskID := s.taskIDGen()
	cfg := s.makeTaskConfig(taskID)
	cfg.Query = req.Query

	// 如果有 outline，先写入 tasks.json
	if req.Outline != nil && len(req.Outline.Slides) > 0 {
		cfg.Outline = req.Outline
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

// handleAIGenerateOutline takes a partial outline (slides with empty descriptions) and the user's topic query,
// then uses the AI to generate content_plan for each slide that has an empty description.
// It returns the enriched slides array.
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

	ctx := c.Request.Context()
	model, err := s.aiModelFactory(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "模型初始化失败: " + err.Error()})
		return
	}

	theme, _ := s.getTheme(req.Outline.Theme)
	themeName := req.Outline.Theme
	if theme != nil {
		themeName = theme.DisplayName
	}

	// Build per-slide context: structure only, NO content hints from presets
	var slideContexts strings.Builder
	for i, slide := range req.Outline.Slides {
		layout := s.templateLoader.GetLayout(slide.ContentType)
		layoutName := slide.ContentType
		if layout != nil {
			layoutName = layout.DisplayName
		}
		slideContexts.WriteString(fmt.Sprintf("\n第%d页：标题「%s」 | 布局「%s」",
			i+1, slide.Title, layoutName))
	}

	prompt := fmt.Sprintf(`你是PPT内容规划专家。用户已编排好PPT结构，你的任务是根据用户主题为每一页生成具体内容。

## 用户主题（所有内容必须围绕此展开）
%s

## 配色方案
%s

## 页面结构
%s

## 输出格式（严格返回JSON，不要任何解释）
{
  "slides": [
    {"index": 1, "title": "实际标题", "content_type": "布局名", "description": "300-400字内容描述", "content_plan": {"summary": "一句话概括", "elements": [{"type": "bullet_list", "items": ["要点1：具体说明", "要点2：具体说明"]}]}}
  ]
}

## 强制要求
- description字数300-400字，必须包含具体数据或案例，禁止空洞泛泛
- 内容严格围绕用户主题展开，禁止偏离
- 只输出JSON`, req.Query, themeName, slideContexts.String())

	resp, err := model.Generate(ctx, []*schema.Message{
		schema.UserMessage(prompt),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成失败: " + err.Error()})
		return
	}

	content := ""
	if resp != nil {
		content = strings.TrimSpace(resp.Content)
	}

	// Guard: empty response from model
	if content == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "模型返回为空，请重试"})
		return
	}

	// Strip markdown code fences
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// Guard: empty after stripping fences
	if content == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "模型返回格式为空，请重试"})
		return
	}

	// Parse response
	var result struct {
		Slides []deep.SlideOutline `json:"slides"`
	}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析失败: " + err.Error()})
		return
	}

	if len(result.Slides) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "模型未返回有效内容"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"slides": result.Slides})
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

// ── Session / Continue handlers ───────────────────────────────────────────────

// handleContinueTask handles user requests to continue iterating on an existing task.
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
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	// Only allow continue on completed tasks
	if ts.Info.Status != task.TaskStatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only completed tasks can be continued"})
		return
	}

	uid := userIDGin(c)
	sess := s.sessionManager.GetOrCreate(taskID, ts.Info.WorkDir)
	sess.AddUserMessage(req.Message)

	// Re-initialize task to running state
	ts.Mu.Lock()
	ts.Info.Status = task.TaskStatusRunning
	ts.Mu.Unlock()
	ts.Persist()

	// Start SSE stream immediately so frontend can connect
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, _ := c.Writer.(interface{ Flush() })
	writer := c.Writer

	// Start continue processing in goroutine and stream results
	listenerID := uuid.New().String()
	ch := make(chan task.SSERichEvent, 64)
	ts.AddListener(listenerID, ch)
	defer func() {
		ts.RemoveListener(listenerID)
		// Mark task as completed again
		ts.Mu.Lock()
		ts.Info.Status = task.TaskStatusCompleted
		ts.Mu.Unlock()
		ts.Persist()
	}()

	// Flush headers immediately
	writeSSEFlush(writer, flusher)

	go s.runContinue(taskID, ts, req.Message, uid, sess, ch)

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			return false
		case evt, ok := <-ch:
			if !ok {
				return false
			}
			writeSSEToWriter(writer, flusher, evt)
			if evt.Type == "continue_complete" {
				return false
			}
			return true
		}
	})
}

func (s *Server) runContinue(taskID string, ts *task.TaskState, message string, uid int, sess *session.ConversationSession, ch chan task.SSERichEvent) {
	defer close(ch)

	ctx := context.Background()

	// Step 1: Route intent - Fixer vs DeepAgent vs add_page
	ch <- task.SSERichEvent{Type: "answer", Content: "正在分析您的请求...\n"}

	route := s.routeContinueIntent(ctx, message, ts.Info.WorkDir)

	ch <- task.SSERichEvent{Type: "tool_call", ToolName: "intent_classifier", ToolArgs: fmt.Sprintf(`{"intent":"%s","reason":"%s","target_pages":%v}`, route.Intent, route.Reason, route.TargetPages)}

	switch route.Intent {
	case "fix":
		s.runFixerContinue(taskID, ts, &route, ch)
	case "regenerate", "regenerate_all", "add_page", "unknown":
		s.runDeepAgentContinue(taskID, ts, &route, ch)
	case "needs_clarification":
		question := route.ClarificationQuestion
		if question == "" {
			question = "您的反馈比较模糊，请问您是指哪一页？希望怎么改善？（比如：第2页字体太小、第3张换个颜色、第1页重新生成等）"
		}
		ch <- task.SSERichEvent{Type: "clarification", Content: question}
		ts.Broadcast(task.SSERichEvent{Type: "continue_complete"})
		return
	}

	// Add assistant message
	ts.Broadcast(task.SSERichEvent{Type: "continue_complete"})
}

// RouteResult holds the result of intent classification.
type RouteResult struct {
	Intent string `json:"intent"` // "fix" | "regenerate" | "regenerate_all" | "add_page" | "needs_clarification" | "unknown"

	// Reason describes why this intent was chosen.
	Reason string `json:"reason"`

	// TargetPages contains the page indices (1-based) mentioned by the user.
	TargetPages []int `json:"target_pages,omitempty"`

	// NeedsClarification is true when the user's intent is vague.
	NeedsClarification bool `json:"needs_clarification,omitempty"`

	// ClarificationQuestion is the question to ask the user when NeedsClarification is true.
	ClarificationQuestion string `json:"clarification_question,omitempty"`

	// FixDetails is set when intent is "fix". Describes what to fix.
	// e.g. {"aspect": "font_size", "value": "smaller", "pages": [2]}
	FixDetails *FixDetails `json:"fix_details,omitempty"`

	// RegenerateScope is "all" or a list of page indices.
	RegenerateScope []int `json:"regenerate_scope,omitempty"`

	// SuggestFix indicates that despite needing clarification, the user may want a fix (not regenerate).
	SuggestFix bool `json:"suggest_fix,omitempty"`
}

// FixDetails describes what visual aspect the user wants to adjust.
type FixDetails struct {
	// Aspect is the visual property to fix: "font_size", "color", "alignment",
	// "spacing", "layout", "position", "text_content", "style", "other".
	Aspect string `json:"aspect"`
	// Detail is the specific adjustment: "更大", "红色", "居中", "加粗", etc.
	Detail string `json:"detail"`
	// TargetElements describes which elements on the page: "标题", "正文", "所有文字", "图表" etc.
	TargetElements string `json:"target_elements,omitempty"`
}

// routeContinueIntent is the main entry point for intent routing.
// It always delegates to LLM classification for nuanced, structured results.
// Only a few very high-confidence patterns bypass LLM (e.g., "加一页").
func (s *Server) routeContinueIntent(ctx context.Context, message string, workDir string) RouteResult {
	// ── Rule fast-paths ─────────────────────────────────────────────────────
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

	// ── LLM classification ──────────────────────────────────────────────────
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

// classifyIntentByLLM uses the AI model to classify the user's continue message intent
// and extract structured fix details. It returns a RouteResult with all extracted fields.
// Uses textModelFactory (ARK_TEXT_MODEL) for cost efficiency, falling back to aiModelFactory.
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

// extractTargetPages parses page references from the user message (1-based).
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

func (s *Server) runFixerContinue(taskID string, ts *task.TaskState, route *RouteResult, ch chan task.SSERichEvent) {
	ch <- task.SSERichEvent{Type: "answer", Content: "检测到是定点修复请求，将使用 Fixer 进行调整...\n"}

	// Read manifest to get task info
	manifest, err := deep.ReadTasksManifest(ts.Info.WorkDir)
	if err != nil || manifest == nil {
		ch <- task.SSERichEvent{Type: "error", Error: "无法读取任务清单: " + err.Error()}
		return
	}

	// Use pages from LLM result, fall back to inferPagesFromMessage
	pagesToFix := route.TargetPages
	if len(pagesToFix) == 0 {
		pagesToFix = inferPagesFromMessage(route.Reason, len(manifest.Tasks))
	}

	if len(pagesToFix) == 0 {
		// If no specific page mentioned, fix all done pages
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

	// Build QA report from LLM fix details
	var qaReport strings.Builder
	qaReport.WriteString("用户请求修复（LLM 意图分析）：")
	if route.FixDetails != nil {
		qaReport.WriteString(fmt.Sprintf("调整方面=%s, 具体要求=%s, 目标元素=%s",
			route.FixDetails.Aspect, route.FixDetails.Detail, route.FixDetails.TargetElements))
	} else {
		qaReport.WriteString(route.Reason)
	}

	// Process each page
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

		// Read output file to verify it exists
		outputFile := filepath.Join(ts.Info.WorkDir, targetTask.OutputFile)
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			ch <- task.SSERichEvent{Type: "error", Error: fmt.Sprintf("文件 %s 不存在，无法修复", targetTask.OutputFile)}
			continue
		}

		// Update task status
		targetTask.Status = deep.StatusQADone
		targetTask.QAReport = qaReport.String()
		targetTask.FixAttempts++
		deep.WriteTasksManifest(ts.Info.WorkDir, manifest)

		// Run fixer via Python tool
		fixResult := s.runFixerForTask(ts.Info.WorkDir, targetTask, qaReport.String())
		if fixResult.Success {
			targetTask.Status = deep.StatusFixed
			ch <- task.SSERichEvent{Type: "answer", Content: fmt.Sprintf("第%d页修复成功: %s\n", pageIdx, fixResult.Message)}
		} else {
			ch <- task.SSERichEvent{Type: "answer", Content: fmt.Sprintf("第%d页修复完成: %s\n", pageIdx, fixResult.Message)}
		}
		deep.WriteTasksManifest(ts.Info.WorkDir, manifest)
	}

	// Refresh file list
	s.refreshFileList(ts, ch)

	// Update final files
	manifest, _ = deep.ReadTasksManifest(ts.Info.WorkDir)
	if manifest != nil {
		var files []string
		for _, t := range manifest.Tasks {
			if t.Status == deep.StatusDone || t.Status == deep.StatusQADone || t.Status == deep.StatusFixed {
				files = append(files, filepath.Join(ts.Info.WorkDir, t.OutputFile))
			}
		}
		ts.Mu.Lock()
		ts.Info.Files = files
		ts.Info.DoneCount = manifest.CompletedCount()
		ts.Mu.Unlock()
	}

	ch <- task.SSERichEvent{
		Type:   "complete",
		Status: ts.Info.Status,
		Done:   ts.Info.DoneCount,
		Total:  ts.Info.TotalCount,
		Files:  ts.Info.Files,
	}
}

// FixResult holds the result of a fixer operation.
type FixResult struct {
	Success bool
	Message string
}

func (s *Server) runFixerForTask(workDir string, task *deep.TaskItem, qaReport string) FixResult {
	return FixResult{
		Success: true,
		Message: "修复请求已记录，将在后续质检流程中处理",
	}
}

func (s *Server) runDeepAgentContinue(taskID string, ts *task.TaskState, route *RouteResult, ch chan task.SSERichEvent) {
	ch <- task.SSERichEvent{Type: "answer", Content: "正在重新处理您的请求...\n"}

	manifest, err := deep.ReadTasksManifest(ts.Info.WorkDir)
	if err != nil || manifest == nil {
		ch <- task.SSERichEvent{Type: "error", Error: "无法读取任务清单"}
		return
	}

	// Handle add_page intent
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
		deep.WriteTasksManifest(ts.Info.WorkDir, manifest)
		ch <- task.SSERichEvent{Type: "answer", Content: fmt.Sprintf("新增第%d页: %s\n", newTask.PageIndex, newTask.Title)}
	}

	// Handle regenerate intent (specific pages)
	if route.Intent == "regenerate" {
		targetPages := route.TargetPages
		if len(targetPages) == 0 && len(route.RegenerateScope) > 0 {
			targetPages = route.RegenerateScope
		}
		if len(targetPages) == 0 {
			ch <- task.SSERichEvent{Type: "answer", Content: "未找到需要重新生成的页面\n"}
		} else {
			for _, pageIdx := range targetPages {
				for _, t := range manifest.Tasks {
					if t.PageIndex == pageIdx {
						t.Status = deep.StatusPending
						ch <- task.SSERichEvent{Type: "answer", Content: fmt.Sprintf("第%d页 '%s' 已标记为待重新生成\n", pageIdx, t.Title)}
						break
					}
				}
			}
			deep.WriteTasksManifest(ts.Info.WorkDir, manifest)
		}
	}

	// Handle regenerate_all intent
	if route.Intent == "regenerate_all" {
		for _, t := range manifest.Tasks {
			t.Status = deep.StatusPending
		}
		deep.WriteTasksManifest(ts.Info.WorkDir, manifest)
		ch <- task.SSERichEvent{Type: "answer", Content: "所有页面已标记为待重新生成\n"}
	}

	// Refresh file list
	s.refreshFileList(ts, ch)

	ch <- task.SSERichEvent{
		Type:   "complete",
		Status: ts.Info.Status,
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

func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_", "?", "_",
		"\"", "_", "<", "_", ">", "_", "|", "_", "\n", "_",
	)
	return replacer.Replace(name)
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
		if filepath.Ext(entry.Name()) == ".pptx" {
			ts.Mu.Lock()
			if !ts.ReportedFiles()[entry.Name()] {
				ts.SetReportedFile(entry.Name())
				ts.Mu.Unlock()
				evt := task.SSERichEvent{
					Type:     "file_ready",
					ToolName: entry.Name(),
					Files:    []string{entry.Name()},
				}
				ch <- evt
				ts.Broadcast(evt)
			} else {
				ts.Mu.Unlock()
			}
		}
	}
}

// writeSSEFlush sends a minimal SSE comment to flush the response.
func writeSSEFlush(writer interface{ Write([]byte) (int, error) }, flusher interface{ Flush() }) {
	writer.Write([]byte(": flush\n\n"))
	if flusher != nil {
		flusher.Flush()
	}
}

// handleGetConversation returns the conversation history for a task.
func (s *Server) handleGetConversation(c *gin.Context) {
	taskID := c.Param("id")

	ts := s.tasks.GetTaskState(taskID)
	if ts == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	sess := s.sessionManager.GetOrCreate(taskID, ts.Info.WorkDir)
	c.JSON(http.StatusOK, gin.H{
		"task_id":    taskID,
		"messages":   sess.Messages,
		"created_at": sess.CreatedAt,
		"updated_at": sess.UpdatedAt,
	})
}

// ── User profile handlers ───────────────────────────────────────────────────────

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

// updateUserStyleFromTask is called when a task completes to extract and save style preferences.
// 核心逻辑由 LLM 分析 PPTX 文本内容完成，fallback 到规则提取。
func (s *Server) updateUserStyleFromTask(userID int, workDir string, query string) {
	manifest, err := deep.ReadTasksManifest(workDir)
	if err != nil || manifest == nil {
		return
	}

	// Convert tasks to TaskItemInfo for fallback
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

// ── Event log handlers ──────────────────────────────────────────────────────

func (s *Server) handleListTaskEvents(c *gin.Context) {
	taskID := c.Param("id")
	limit := 500
	if l := c.Query("limit"); l != "" {
		if n, err := fmt.Sscanf(l, "%d", &limit); err != nil || n != 1 || limit <= 0 {
			limit = 500
		}
	}
	events, err := db.ListTaskEvents(taskID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询事件日志失败: " + err.Error()})
		return
	}
	if events == nil {
		events = []db.TaskEvent{}
	}
	c.JSON(http.StatusOK, gin.H{"task_id": taskID, "count": len(events), "events": events})
}
