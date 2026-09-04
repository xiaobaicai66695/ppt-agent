package web

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
	"github.com/cloudwego/ppt-agent/pkg/retry"
	agentutils "github.com/cloudwego/ppt-agent/pkg/runtime/model"
	"github.com/cloudwego/ppt-agent/pkg/runtime/task"
	"github.com/cloudwego/ppt-agent/pkg/utils/logger"
)

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
		allowedPageIndexes, allowedTaskIDs := resolveManifestTargets(manifest, pages)
		route.TargetPages = allowedPageIndexes
		route.TargetTaskIDs = allowedTaskIDs
		fixerApplied := false
		if len(allowedTaskIDs) > 0 {
			beforeFix, _ := manifest.MustMarshalJSON()
			fixerCfg := &deck.PPTTaskConfig{
				WorkDir:          ts.Info.WorkDir,
				TaskID:           taskID,
				Query:            ts.Info.Query,
				SkillsDir:        s.skillDir,
				Operator:         s.operator,
				UserID:           ts.Info.UserID,
				OnFixerTriggered: ts.RecordFixerRun,
				ModelAPIKey:      credential.APIKey,
				ModelProvider:    credential.Provider,
			}
			fixerCtx, cancelFixer := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancelFixer()
			fixer, fixerErr := deck.NewPPTFixerAgentForTasks(fixerCtx, fixerCfg, allowedTaskIDs)
			if fixerErr == nil {
				fixerInput := fmt.Sprintf("用户要求：%s\n允许修改的任务 ID：%v\n允许修改的页面：%v\n结构化修复提示：%s", continueMessage, allowedTaskIDs, allowedPageIndexes, instruction)
				fixerCfg.NotifyFixerTriggered()
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
		for _, pageIdx := range allowedPageIndexes {
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
	if _, err := deck.RenderPPT(context.Background(), cfg, func(event deck.DeckRenderEvent) {
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
	if retry.IsRetryable(fmt.Errorf("%s", message)) {
		ts.Info.Status = task.TaskStatusPausedRetryable
		ts.Info.Error = fmt.Sprintf("可恢复的 %s：%s。可在对话框输入“继续任务”从检查点恢复。", retry.ClassifyError(fmt.Errorf("%s", message)), message)
		return
	}
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
