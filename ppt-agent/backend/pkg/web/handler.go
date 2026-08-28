package web

import (
	"bytes"
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
	"github.com/cloudwego/ppt-agent/pkg/agent/modelcompat"
	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/auth"
	"github.com/cloudwego/ppt-agent/pkg/db"
	loganalysis "github.com/cloudwego/ppt-agent/pkg/log_analysis"
	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/session"
	"github.com/cloudwego/ppt-agent/pkg/task"
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
		Query   string            `json:"query"`
		Outline *deck.TaskOutline `json:"outline,omitempty"`
	}
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

	taskID := s.taskIDGen()
	cfg := s.makeTaskConfig(taskID)
	cfg.Query = req.Query
	cfg.UserID = uid
	cfg.ModelAPIKey = credential.APIKey
	cfg.ModelProvider = credential.Provider

	// 如果有 outline，先做服务端兜底校验/补齐；TaskManager 只将其写入规划草稿。
	if req.Outline != nil && len(req.Outline.Slides) > 0 {
		outline, err := s.prepareOutline(c.Request.Context(), req.Query, req.Outline)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "outline 处理失败: " + err.Error()})
			return
		}
		cfg.Outline = outline
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
	var tasks []task.TaskInfo
	if isAdminGin(c) {
		tasks = s.tasks.ListAllTasks()
	} else {
		tasks = s.tasks.ListTasks(userIDGin(c))
	}
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

func (s *Server) handleListLayouts(c *gin.Context) {
	layouts := s.templateLoader.ListLayouts()
	c.JSON(http.StatusOK, gin.H{"layouts": layouts})
}

func (s *Server) prepareOutline(_ context.Context, query string, outline *deck.TaskOutline) (*deck.TaskOutline, error) {
	if outline == nil || len(outline.Slides) == 0 {
		return outline, nil
	}
	if strings.TrimSpace(outline.Title) == "" {
		outline.Title = strings.TrimSpace(query)
	}
	outline.ContentMode = deck.OutlineContentModeUserOutline

	for i := range outline.Slides {
		slide := &outline.Slides[i]
		slide.Title = strings.TrimSpace(slide.Title)
		slide.ContentType = strings.TrimSpace(slide.ContentType)
		if s.templateLoader.GetLayout(slide.ContentType) == nil {
			return nil, fmt.Errorf("第%d页 content_type=%q 不存在", i+1, slide.ContentType)
		}
	}

	return outline, nil
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

	ctx := auth.WithUser(context.Background(), &db.User{ID: uint(uid)})

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
	credential := userModelCredential(ts.Info.UserID)

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
			OutputFile:  fmt.Sprintf("%d_%s.pptx", nextPage, title),
			Status:      deck.StatusPending,
			ContentPlan: &deck.ContentPlan{
				Summary: fmt.Sprintf("补充说明用户要求：%s", continueMessage),
				Components: []deck.PlanComponent{{
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
		allowedPageIndexes := make([]int, 0, len(pages))
		for _, pageIdx := range pages {
			if item := findManifestTaskByPage(manifest, pageIdx); item != nil {
				allowedPageIndexes = append(allowedPageIndexes, item.PageIndex)
			}
		}
		fixerApplied := false
		if len(allowedPageIndexes) > 0 {
			beforeFix, _ := manifest.MustMarshalJSON()
			fixerCfg := &deck.PPTTaskConfig{
				WorkDir:       ts.Info.WorkDir,
				TaskID:        taskID,
				Query:         ts.Info.Query,
				SkillsDir:     s.skillDir,
				Operator:      s.operator,
				UserID:        ts.Info.UserID,
				ModelAPIKey:   credential.APIKey,
				ModelProvider: credential.Provider,
			}
			fixerCtx, cancelFixer := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancelFixer()
			fixer, fixerErr := deck.NewPPTFixerAgent(fixerCtx, fixerCfg, allowedPageIndexes)
			if fixerErr == nil {
				fixerInput := fmt.Sprintf("用户要求：%s\n允许修改的页面：%v\n结构化修复提示：%s", continueMessage, pages, instruction)
				fixerErr = deck.RunPPTFixerWithCallback(fixerCtx, fixer, fixerInput, func(event deck.AgentEvent) {
					switch event.Type {
					case deck.AgentEventAnswer:
						ch <- task.SSERichEvent{Type: "answer", Content: event.Content}
					case deck.AgentEventProgress:
						ch <- task.SSERichEvent{Type: "progress", Phase: "fixing", PhaseDetail: event.PhaseDetail}
					}
				})
			}
			if fixerErr == nil {
				if updated, readErr := deck.ReadTasksManifest(ts.Info.WorkDir); readErr == nil && updated != nil {
					afterFix, _ := updated.MustMarshalJSON()
					if !bytes.Equal(beforeFix, afterFix) {
						manifest = updated
						fixerApplied = true
					}
				}
			}
			if !fixerApplied {
				if fixerErr == nil {
					fixerErr = fmt.Errorf("Fixer 未产生结构化页面修改")
				}
				logger.Warn("ppt_fixer_failed_using_code_fallback", "task_id", taskID, "error", fixerErr.Error())
				ch <- task.SSERichEvent{Type: "answer", Content: "定点修复规划未完成，系统已使用现有页面计划进入重渲染。\n"}
			}
		}
		for _, pageIdx := range pages {
			if item := findManifestTaskByPage(manifest, pageIdx); item != nil {
				if fixerApplied {
					markTaskForFixRerender(ts.Info.WorkDir, item, instruction)
				} else {
					markTaskForRerender(ts.Info.WorkDir, item, instruction, true)
				}
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
		UserID:      ts.Info.UserID,
	}
	cfg.ModelAPIKey = credential.APIKey
	cfg.ModelProvider = credential.Provider
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
	item.Status = deck.StatusPending
	if item.OutputFile != "" {
		_ = os.Remove(filepath.Join(workDir, item.OutputFile))
	}
}

func markTaskForFixRerender(workDir string, item *deck.TaskItem, instruction string) {
	if item == nil {
		return
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

// ── Account model key handlers ────────────────────────────────────────────────

func (s *Server) handleGetUserAPIKey(c *gin.Context) {
	uid := userIDGin(c)
	key, err := db.GetUserAPIKey(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询 API Key 配置失败: " + err.Error()})
		return
	}
	if key == nil {
		c.JSON(http.StatusOK, gin.H{
			"configured":         false,
			"provider":           string(defaultModelProvider()),
			"default_provider":   string(defaultModelProvider()),
			"masked_key":         "",
			"default_configured": systemProviderKeyConfigured(defaultModelProvider()),
		})
		return
	}
	provider := modelcompat.NormalizeProvider(key.Provider)
	c.JSON(http.StatusOK, gin.H{
		"configured":         true,
		"provider":           string(provider),
		"default_provider":   string(defaultModelProvider()),
		"masked_key":         maskAPIKey(key.APIKey),
		"default_configured": systemProviderKeyConfigured(provider),
		"updated_at":         key.UpdatedAt,
	})
}

func (s *Server) handleUpdateUserAPIKey(c *gin.Context) {
	var req struct {
		Provider string `json:"provider"`
		APIKey   string `json:"api_key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
		return
	}
	apiKey := strings.TrimSpace(req.APIKey)
	if len(apiKey) < 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API Key 长度过短"})
		return
	}
	provider := modelcompat.NormalizeProvider(req.Provider)
	if !isSupportedAccountProvider(provider) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的模型厂商: " + strings.TrimSpace(req.Provider)})
		return
	}
	if err := db.UpsertUserAPIKey(uint(userIDGin(c)), string(provider), apiKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存 API Key 失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"configured":         true,
		"provider":           string(provider),
		"default_provider":   string(defaultModelProvider()),
		"masked_key":         maskAPIKey(apiKey),
		"default_configured": systemProviderKeyConfigured(provider),
		"updated_at":         time.Now(),
	})
}

func (s *Server) handleDeleteUserAPIKey(c *gin.Context) {
	if err := db.DeleteUserAPIKey(uint(userIDGin(c))); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除 API Key 配置失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":             "ok",
		"configured":         false,
		"provider":           string(defaultModelProvider()),
		"default_provider":   string(defaultModelProvider()),
		"default_configured": systemProviderKeyConfigured(defaultModelProvider()),
	})
}

type modelCredential struct {
	Provider string
	APIKey   string
}

func userModelCredential(userID int) modelCredential {
	provider := defaultModelProvider()
	accountKey := ""
	if userID > 0 && db.DB != nil {
		key, err := db.GetUserAPIKey(uint(userID))
		if err != nil {
			logger.Warn("user_api_key_lookup_failed", "user_id", userID, "error", err.Error())
		} else if key != nil {
			provider = modelcompat.NormalizeProvider(key.Provider)
			accountKey = key.APIKey
		}
	}
	return modelCredential{
		Provider: string(provider),
		APIKey:   modelcompat.ResolveProviderAPIKey(provider, accountKey),
	}
}

func defaultModelProvider() modelcompat.Provider {
	for _, entry := range strings.Split(os.Getenv("MODEL_CHAIN"), ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		entryKey := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(entry, "-", "_"), " ", "_"))
		if provider := strings.TrimSpace(os.Getenv("MODEL_" + entryKey + "_PROVIDER")); provider != "" {
			return modelcompat.NormalizeProvider(provider)
		}
	}
	if provider := strings.TrimSpace(os.Getenv("MODEL_PRIMARY_PROVIDER")); provider != "" {
		return modelcompat.NormalizeProvider(provider)
	}
	return modelcompat.NormalizeProvider(os.Getenv("MODEL_PROVIDER"))
}

func systemProviderKeyConfigured(provider modelcompat.Provider) bool {
	provider = modelcompat.NormalizeProvider(string(provider))
	if modelcompat.ResolveProviderAPIKey(provider, "") != "" {
		return true
	}
	for _, entry := range append(strings.Split(os.Getenv("MODEL_CHAIN"), ","), "primary", "text", "qa") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		entryKey := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(entry, "-", "_"), " ", "_"))
		entryProvider := modelcompat.NormalizeProvider(firstNonEmpty(
			os.Getenv("MODEL_"+entryKey+"_PROVIDER"),
			os.Getenv("MODEL_PROVIDER"),
		))
		if entryProvider != provider {
			continue
		}
		if strings.TrimSpace(os.Getenv("MODEL_"+entryKey+"_API_KEY")) != "" {
			return true
		}
		if keyEnv := strings.TrimSpace(os.Getenv("MODEL_" + entryKey + "_API_KEY_ENV")); keyEnv != "" && strings.TrimSpace(os.Getenv(keyEnv)) != "" {
			return true
		}
	}
	return false
}

func isSupportedAccountProvider(provider modelcompat.Provider) bool {
	switch modelcompat.NormalizeProvider(string(provider)) {
	case modelcompat.ProviderArk, modelcompat.ProviderOpenAI, modelcompat.ProviderOpenAICompat,
		modelcompat.ProviderSiliconFlow, modelcompat.ProviderDeepSeek, modelcompat.ProviderQwen:
		return true
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func maskAPIKey(apiKey string) string {
	apiKey = strings.TrimSpace(apiKey)
	runes := []rune(apiKey)
	if len(runes) <= 8 {
		return strings.Repeat("*", len(runes))
	}
	prefix := string(runes[:4])
	suffix := string(runes[len(runes)-4:])
	return prefix + strings.Repeat("*", 8) + suffix
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
	if strings.TrimSpace(record.Metadata) != "" {
		metadata := map[string]any{}
		if err := json.Unmarshal([]byte(record.Metadata), &metadata); err == nil {
			event.Metadata = metadata
		} else {
			event.Metadata = map[string]any{"raw": record.Metadata}
		}
	}
	if !includeMetadata {
		return agentutils.RuntimeEventSummary(event)
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

type adminTaskResponse struct {
	ID         string    `json:"id"`
	UserID     uint      `json:"user_id"`
	UserEmail  string    `json:"user_email"`
	Query      string    `json:"query"`
	Status     string    `json:"status"`
	DoneCount  int       `json:"done_count"`
	TotalCount int       `json:"total_count"`
	Duration   string    `json:"duration"`
	Error      string    `json:"error"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
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
			ID:         task.ID,
			UserID:     task.UserID,
			UserEmail:  emails[task.UserID],
			Query:      task.Query,
			Status:     task.Status,
			DoneCount:  task.DoneCount,
			TotalCount: task.TotalCount,
			Duration:   task.Duration,
			Error:      task.Error,
			CreatedAt:  task.CreatedAt,
			UpdatedAt:  task.UpdatedAt,
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
