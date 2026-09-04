package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"

	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
	agentrouter "github.com/cloudwego/ppt-agent/pkg/agent/router"
	"github.com/cloudwego/ppt-agent/pkg/auth"
	"github.com/cloudwego/ppt-agent/pkg/db"
	agentutils "github.com/cloudwego/ppt-agent/pkg/runtime/model"
	"github.com/cloudwego/ppt-agent/pkg/runtime/task"
	"github.com/cloudwego/ppt-agent/pkg/session"
	"github.com/cloudwego/ppt-agent/pkg/utils/logger"
)

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
	if resumed, err := s.resumeDraftCheckpoint(taskID, ts, ch); resumed {
		if err != nil {
			ch <- task.SSERichEvent{Type: "error", Error: err.Error()}
			ch <- task.SSERichEvent{Type: "continue_complete", Message: "恢复未完成"}
			return
		}
		ch <- task.SSERichEvent{Type: "continue_complete", Message: "已从规划审查检查点恢复并完成后续生成"}
		return
	}

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

// resumeDraftCheckpoint restores a task that failed after Planner wrote
// tasks.draft.json but before Reviewer committed tasks.json. The normal
// continuation workflow requires tasks.json, so without this branch a user
// entering "继续任务" would immediately get "无法读取任务清单".

func (s *Server) resumeDraftCheckpoint(taskID string, ts *task.TaskState, ch chan task.SSERichEvent) (bool, error) {
	if ts == nil || strings.TrimSpace(ts.Info.WorkDir) == "" {
		return false, nil
	}
	if manifest, err := deck.ReadTasksManifest(ts.Info.WorkDir); err == nil {
		// handleContinueTask switches the task back to running before opening
		// SSE. Preserve the retryable marker in Error so the render checkpoint
		// remains distinguishable from an ordinary completed-task edit.
		if !strings.HasPrefix(strings.TrimSpace(ts.Info.Error), "可恢复的 ") {
			return false, nil
		}
		ch <- task.SSERichEvent{Type: "answer", Content: "检测到可恢复的渲染检查点，正在继续未完成页面。\n"}
		credential := userModelCredential(ts.Info.UserID)
		cfg := &deck.PPTTaskConfig{
			WorkDir:       ts.Info.WorkDir,
			TaskID:        taskID,
			Query:         ts.Info.Query,
			SkillsDir:     s.skillDir,
			Operator:      s.operator,
			RuntimeMeta:   agentutils.NewRuntimeMeta(taskID, ts.Info.WorkDir),
			Concurrency:   5,
			UserID:        ts.Info.UserID,
			ModelAPIKey:   credential.APIKey,
			ModelProvider: credential.Provider,
		}
		if _, err := deck.RenderPPT(context.Background(), cfg, func(event deck.DeckRenderEvent) {
			ch <- deckRenderSSE(event)
		}); err != nil {
			markTaskFailed(ts, err.Error())
			return true, fmt.Errorf("从渲染检查点恢复失败: %w", err)
		}
		updated, readErr := deck.ReadTasksManifest(ts.Info.WorkDir)
		if readErr == nil && updated != nil {
			manifest = updated
		}
		s.refreshFileList(ts, ch)
		ts.Mu.Lock()
		ts.Info.Files = task.ManifestOutputFiles(manifest)
		ts.Info.DoneCount = manifest.CompletedCount()
		ts.Info.TotalCount = len(manifest.Tasks)
		ts.Mu.Unlock()
		return true, nil
	} else if !os.IsNotExist(err) {
		return true, fmt.Errorf("读取正式任务清单失败: %w", err)
	}
	draft, err := deck.ReadTasksDraftManifest(ts.Info.WorkDir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return true, fmt.Errorf("读取恢复草稿失败: %w", err)
	}
	if draft == nil || len(draft.Tasks) == 0 {
		return true, fmt.Errorf("恢复草稿为空")
	}

	ch <- task.SSERichEvent{Type: "answer", Content: "检测到规划审查中断，正在从已保存的草稿继续，不会重新规划。\n"}
	credential := userModelCredential(ts.Info.UserID)
	runtimeMeta := agentutils.NewRuntimeMeta(taskID, ts.Info.WorkDir)
	cfg := &deck.PPTTaskConfig{
		WorkDir:          ts.Info.WorkDir,
		TaskID:           taskID,
		Query:            ts.Info.Query,
		SkillsDir:        s.skillDir,
		Operator:         s.operator,
		RuntimeMeta:      runtimeMeta,
		Concurrency:      5,
		UserID:           ts.Info.UserID,
		OnFixerTriggered: ts.RecordFixerRun,
		ModelAPIKey:      credential.APIKey,
		ModelProvider:    credential.Provider,
	}
	if _, err := deck.ResumePPTPlannerFromDraftWithCallback(context.Background(), cfg, func(event deck.AgentEvent) {
		switch event.Type {
		case deck.AgentEventAnswer:
			ch <- task.SSERichEvent{Type: "answer", Content: event.Content}
		case deck.AgentEventProgress:
			ch <- task.SSERichEvent{Type: "progress", Phase: event.Phase, PhaseDetail: event.PhaseDetail}
		case deck.AgentEventError:
			ch <- task.SSERichEvent{Type: "error", Error: event.Error}
		}
	}); err != nil {
		markTaskFailed(ts, err.Error())
		return true, fmt.Errorf("从规划审查检查点恢复失败: %w", err)
	}

	if _, err := deck.RenderPPT(context.Background(), cfg, func(event deck.DeckRenderEvent) {
		ch <- deckRenderSSE(event)
	}); err != nil {
		markTaskFailed(ts, err.Error())
		return true, fmt.Errorf("恢复后的幻灯片生成失败: %w", err)
	}

	s.refreshFileList(ts, ch)
	manifest, err := deck.ReadTasksManifest(ts.Info.WorkDir)
	if err != nil || manifest == nil {
		return true, fmt.Errorf("恢复后读取任务清单失败")
	}
	ts.Mu.Lock()
	ts.Info.Files = task.ManifestOutputFiles(manifest)
	ts.Info.DoneCount = manifest.CompletedCount()
	ts.Info.TotalCount = len(manifest.Tasks)
	ts.Mu.Unlock()
	ch <- task.SSERichEvent{
		Type:   "progress",
		Status: ts.Info.Status,
		Tasks:  manifest.Tasks,
		Done:   ts.Info.DoneCount,
		Total:  ts.Info.TotalCount,
		Files:  ts.Info.Files,
	}
	return true, nil
}

// RouteResult 保存意图分类的结果。
type RouteResult struct {
	Intent string `json:"intent"` // "fix" | "regenerate" | "regenerate_all" | "add_page" | "needs_clarification" | "unknown"

	// Reason 描述选择此意图的原因。
	Reason string `json:"reason"`

	// TargetPages 包含用户提到的页面索引（从 1 开始）。
	TargetPages []int `json:"target_pages,omitempty"`

	// TargetTaskIDs 是服务端依据当前 manifest 从 TargetPages 解析出的稳定任务 ID。
	// 仅 fix 链路使用，作为 Fixer 的上下文切片和工具授权边界；LLM 不负责生成此字段。
	TargetTaskIDs []string `json:"target_task_ids,omitempty"`

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
		return resolveRouteTaskIDs(result, manifest)
	}

	return resolveRouteTaskIDs(RouteResult{
		Intent:      "unknown",
		Reason:      "LLM 分类失败或返回 unknown，默认走深度处理",
		TargetPages: extractTargetPages(message),
	}, manifest)
}

// resolveRouteTaskIDs 把 LLM 的展示层页码绑定到当前正式 manifest 的稳定 task_id。
// 只有 fix 使用该绑定，避免将模型生成的页码直接当作 Fixer 的授权依据。

func resolveRouteTaskIDs(route RouteResult, manifest *deck.TasksManifest) RouteResult {
	if route.Intent != "fix" {
		return route
	}
	_, route.TargetTaskIDs = resolveManifestTargets(manifest, route.TargetPages)
	return route
}

func resolveManifestTargets(manifest *deck.TasksManifest, pageIndexes []int) ([]int, []string) {
	if manifest == nil {
		return nil, nil
	}
	seenPages := make(map[int]struct{}, len(pageIndexes))
	pages := make([]int, 0, len(pageIndexes))
	taskIDs := make([]string, 0, len(pageIndexes))
	for _, pageIndex := range pageIndexes {
		if pageIndex <= 0 {
			continue
		}
		if _, seen := seenPages[pageIndex]; seen {
			continue
		}
		item := findManifestTaskByPage(manifest, pageIndex)
		if item == nil || strings.TrimSpace(item.TaskID) == "" {
			continue
		}
		seenPages[pageIndex] = struct{}{}
		pages = append(pages, item.PageIndex)
		taskIDs = append(taskIDs, item.TaskID)
	}
	return pages, taskIDs
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
	agent := agentrouter.NewAgent(func(routeCtx context.Context, messages []*schema.Message) (*schema.Message, error) {
		return model.Generate(routeCtx, messages)
	})
	routed, err := agent.RouteContinuation(ctx, agentrouter.ContinuationInput{Message: message, TasksSummary: tasksSummary})
	if err != nil {
		return RouteResult{Intent: "unknown", Reason: fmt.Sprintf("RouterAgent 分类失败: %v", err)}
	}
	result := RouteResult{
		Intent:                routed.Intent,
		Reason:                routed.Reason,
		TargetPages:           routed.TargetPages,
		RegenerateScope:       routed.RegenerateScope,
		NeedsClarification:    routed.NeedsClarification,
		ClarificationQuestion: routed.ClarificationQuestion,
	}
	if routed.FixDetails != nil {
		result.FixDetails = &FixDetails{
			Aspect:         routed.FixDetails.Aspect,
			Detail:         routed.FixDetails.Detail,
			TargetElements: routed.FixDetails.TargetElements,
		}
	}
	if result.Intent == "regenerate" && len(result.TargetPages) == 0 && len(result.RegenerateScope) > 0 {
		result.TargetPages = result.RegenerateScope
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
