/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package deck

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/human"
	"github.com/cloudwego/ppt-agent/pkg/logger"
)

// AgentEventType 流式事件类型常量
const (
	AgentEventAnswer   = "answer"
	AgentEventToolCall = "tool_call"
	AgentEventError    = "error"
	AgentEventProgress = "progress"
)

// streamTimeout 返回单次 iter.Next() 调用允许阻塞的最长时间
// 通过 STREAM_TIMEOUT 环境变量设置（如 "3m"）。零或负值禁用超时
func streamTimeout() time.Duration {
	v := os.Getenv("STREAM_TIMEOUT")
	if v == "" {
		return 15 * time.Minute
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 15 * time.Minute
	}
	return d
}

// adkStreamingEnabled controls ADK runner streaming. Streaming is enabled by
// default; set ADK_ENABLE_STREAMING=false/0/no/off to fall back to blocking
// model calls during provider incidents.
func adkStreamingEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("ADK_ENABLE_STREAMING"))) {
	case "", "true", "1", "yes", "on":
		return true
	default:
		return false
	}
}

// nextWithTimeout 调用 iter.Next() 但如果在配置的timeout时间内没有事件到达则放弃底层流
// 这可以防止代理在 LLM 流停滞时永久卡住（如格式错误的 JSON）
// 返回 (event, ok, timeout)
func nextWithTimeout(ctx context.Context, iter *adk.AsyncIterator[*adk.AgentEvent]) (*adk.AgentEvent, bool, bool) {
	timeout := streamTimeout()
	if timeout <= 0 {
		// 超时禁用 — 使用原始阻塞调用
		event, ok := iter.Next()
		return event, ok, false
	}

	done := make(chan *adk.AgentEvent, 1)
	okCh := make(chan bool, 1)

	go func() {
		event, ok := iter.Next()
		done <- event
		okCh <- ok
	}()

	select {
	case <-ctx.Done():
		return nil, false, false
	case <-time.After(timeout):
		// 超时 — 放弃停滞的流并返回合成错误事件
		// 调用者会将其作为超时错误传播
		return nil, false, true
	case event := <-done:
		return event, <-okCh, false
	}
}

// AgentEvent 代理执行期间发出的结构化事件
type AgentEvent struct {
	Type        string `json:"type"`
	Content     string `json:"content,omitempty"`
	ToolName    string `json:"tool_name,omitempty"`
	ToolArgs    string `json:"tool_args,omitempty"`
	Error       string `json:"error,omitempty"`
	Phase       string `json:"phase,omitempty"`        // 当前阶段
	PhaseDetail string `json:"phase_detail,omitempty"` // 阶段详情
}

// AgentEventCallback 在代理执行期间每个事件被调用
type AgentEventCallback func(event AgentEvent)

func StartPPTPlanner(ctx context.Context, agent adk.Agent, cfg *PPTTaskConfig, userQuery string) (*PPTTaskStart, error) {
	startTime := time.Now()

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: adkStreamingEnabled(),
	})

	iter := runner.Run(ctx, []adk.Message{
		schema.UserMessage(userQuery),
	})

	return &PPTTaskStart{
		Runner:       runner,
		Iter:         iter,
		CheckpointID: cfg.TaskID,
		StartTime:    startTime,
	}, nil
}

func RunPPTPlannerWithHuman(ctx context.Context, agent adk.Agent, cfg *PPTTaskConfig,
	userQuery string, hm *human.Manager) (*PPTTaskResult, error) {

	start, err := StartPPTPlanner(ctx, agent, cfg, userQuery)
	if err != nil {
		return nil, err
	}

	event, err := hm.RunWithApproval(ctx, start.Runner, start.CheckpointID, start.Iter)
	if err != nil {
		return nil, err
	}

	var lastMsg string
	if event != nil && event.Output != nil && event.Output.MessageOutput != nil {
		if msg, _, getErr := adk.GetMessage(event); getErr == nil && msg != nil {
			lastMsg = msg.Content
		}
	}

	manifest, err := ReconcileTasksManifestOutputFiles(cfg.WorkDir)
	result := &PPTTaskResult{
		Message:  lastMsg,
		Duration: time.Since(start.StartTime),
	}

	if err == nil && manifest != nil {
		result.TotalSlides = len(manifest.Tasks)
		result.DoneSlides = manifest.CompletedCount()
		for _, t := range manifest.Tasks {
			if t.Status == StatusDone || t.Status == StatusQADone || t.Status == StatusFixed {
				result.Files = append(result.Files, filepath.Join(cfg.WorkDir, t.OutputFile))
			}
		}
	}

	return result, nil
}

func RunPPTPlanner(ctx context.Context, agent adk.Agent, cfg *PPTTaskConfig, userQuery string) (*PPTTaskResult, error) {
	return runPPTPlannerInternal(ctx, agent, cfg, userQuery, makePrintCallback())
}

// RunPPTPlannerWithCallback 运行 Planner 并为每个流式事件调用 onEvent
// 回调是同步调用的 — 调用者应转发事件或快速缓冲
func RunPPTPlannerWithCallback(ctx context.Context, agent adk.Agent, cfg *PPTTaskConfig,
	userQuery string, onEvent AgentEventCallback) (*PPTTaskResult, error) {
	return runPPTPlannerInternal(ctx, agent, cfg, userQuery, onEvent)
}

func runPPTPlannerInternal(ctx context.Context, agent adk.Agent, cfg *PPTTaskConfig,
	userQuery string, onEvent AgentEventCallback) (result *PPTTaskResult, err error) {

	defer func() {
		if r := recover(); r != nil {
			onEvent(AgentEvent{Type: AgentEventError, Error: fmt.Sprintf("agent internal panic: %v", r)})
			err = fmt.Errorf("agent internal panic: %v", r)
		}
	}()

	start, err := StartPPTPlanner(ctx, agent, cfg, userQuery)
	if err != nil {
		return nil, err
	}

	iter := start.Iter

	var (
		lastMessage       adk.Message
		lastMessageStream *schema.StreamReader[adk.Message]
		answerBuf         strings.Builder
		plannerErr        error
	)

	for {
		if err := ctx.Err(); err != nil {
			if lastMessageStream != nil {
				lastMessageStream.Close()
			}
			return nil, err
		}

		event, ok, timedOut := nextWithTimeout(ctx, iter)
		if ctx.Err() != nil {
			if lastMessageStream != nil {
				lastMessageStream.Close()
			}
			return nil, ctx.Err()
		}

		// 流停滞 — 视为软错误并退出循环
		// 清单状态被保留以便任务可以恢复
		if timedOut {
			logger.Error("stream_timeout", "task_id", cfg.TaskID, "timeout", streamTimeout().String())
			if lastMessageStream != nil {
				lastMessageStream.Close()
			}
			return nil, fmt.Errorf("LLM 流式输出超时（%s），任务已暂停", streamTimeout())
		}

		if !ok {
			break
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			if lastMessageStream != nil {
				lastMessageStream.Close()
			}

			if event.Output.MessageOutput.IsStreaming {
				cpStream := event.Output.MessageOutput.MessageStream.Copy(2)
				event.Output.MessageOutput.MessageStream = cpStream[0]
				lastMessage = nil
				lastMessageStream = cpStream[1]
				processStreamingMessage(lastMessageStream, onEvent, &answerBuf)
			} else {
				lastMessage = event.Output.MessageOutput.Message
				lastMessageStream = nil
				if lastMessage != nil && isChunkEmittable(lastMessage) && lastMessage.Content != "" {
					content := visibleMessageContent(lastMessage.Content)
					if content != "" {
						onEvent(AgentEvent{
							Type:    AgentEventAnswer,
							Content: content,
						})
					}
				}
			}

			if m := event.Output.MessageOutput.Message; m != nil {
				for _, tc := range m.ToolCalls {
					onEvent(AgentEvent{
						Type:     AgentEventToolCall,
						ToolName: tc.Function.Name,
						ToolArgs: tc.Function.Arguments,
					})
				}
			}
		}

		if event.Err != nil {
			onEvent(AgentEvent{
				Type:  AgentEventError,
				Error: event.Err.Error(),
			})
			if plannerErr == nil {
				plannerErr = event.Err
			}
		}
	}

	if lastMessageStream != nil {
		lastMessageStream.Close()
	}

	var lastMsg string
	if lastMessage != nil {
		lastMsg = lastMessage.Content
	} else if answerBuf.Len() > 0 {
		lastMsg = answerBuf.String()
	}

	result = &PPTTaskResult{
		Message:  lastMsg,
		Duration: time.Since(start.StartTime),
	}

	if plannerErr != nil {
		return result, fmt.Errorf("Planner 执行失败: %w", plannerErr)
	}
	if _, err := ensurePlannerDraft(cfg, userQuery, lastMsg, onEvent); err != nil {
		return result, err
	}
	manifest, err := reviewAndCommitDeckSpec(ctx, cfg, onEvent)
	if err != nil {
		return result, err
	}
	if manifest == nil || len(manifest.Tasks) == 0 {
		return result, fmt.Errorf("Planner 生成的 DeckSpec 为空，无法进入渲染")
	}
	result.TotalSlides = len(manifest.Tasks)
	result.DoneSlides = manifest.CompletedCount()
	for _, t := range manifest.Tasks {
		if t.Status == StatusDone || t.Status == StatusQADone || t.Status == StatusFixed {
			result.Files = append(result.Files, filepath.Join(cfg.WorkDir, t.OutputFile))
		}
	}

	return result, nil
}

const maxPlanReviewRounds = 3

func ensurePlannerDraft(cfg *PPTTaskConfig, userQuery, plannerOutput string, onEvent AgentEventCallback) (*TasksManifest, error) {
	if draft, err := ReadTasksDraftManifest(cfg.WorkDir); err == nil && draft != nil && len(draft.Tasks) > 0 {
		return draft, nil
	} else if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("读取 Planner DeckSpec 草稿失败: %w", err)
	}

	// 兼容旧任务或提前写入的 outline：正式清单必须重新经过 Reviewer 质量门。
	if existing, err := ReadTasksManifest(cfg.WorkDir); err == nil && existing != nil && len(existing.Tasks) > 0 {
		if err := WriteTasksDraftManifest(cfg.WorkDir, existing); err != nil {
			return nil, fmt.Errorf("迁移现有 DeckSpec 到审查草稿失败: %w", err)
		}
		return existing, nil
	} else if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("读取现有 DeckSpec 失败: %w", err)
	}

	recovered, err := recoverMissingPlannerManifest(cfg, userQuery, plannerOutput)
	if err != nil {
		return nil, fmt.Errorf("Planner 未生成 DeckSpec 草稿，代码恢复失败: %w", err)
	}
	onEvent(AgentEvent{
		Type:    AgentEventAnswer,
		Content: fmt.Sprintf("Planner 未落下结构化草稿，系统已从现有规划输出恢复 %d 页，接下来仍会经过独立 Reviewer。\n", len(recovered.Tasks)),
	})
	if cfg.RuntimeMeta != nil {
		cfg.RuntimeMeta.RecordEvent("deck_spec_recovered", tasksDraftFileName, "warning", fmt.Sprintf("recovered %d slides from planner output", len(recovered.Tasks)), map[string]any{
			"slide_count": len(recovered.Tasks),
		})
	}
	return recovered, nil
}

func reviewAndCommitDeckSpec(ctx context.Context, cfg *PPTTaskConfig, onEvent AgentEventCallback) (*TasksManifest, error) {
	var latest *PlanReviewReport
	for round := 1; round <= maxPlanReviewRounds; round++ {
		report, err := ReviewTasksDraftManifest(cfg.WorkDir, round)
		if err != nil {
			return nil, fmt.Errorf("第 %d 轮 DeckSpec 硬校验失败: %w", round, err)
		}
		latest = report
		onEvent(AgentEvent{
			Type:        AgentEventProgress,
			Phase:       "reviewing",
			PhaseDetail: fmt.Sprintf("Task Reviewer 第 %d/%d 轮：%s", round, maxPlanReviewRounds, report.Summary),
		})
		if cfg.RuntimeMeta != nil {
			status := "needs_revision"
			if report.Passed {
				status = "passed"
			}
			cfg.RuntimeMeta.RecordEvent("deck_spec_review", "TaskPlanReviewer", status, report.Summary, map[string]any{
				"round":       round,
				"issue_count": report.IssueCount,
			})
		}
		if report.Passed {
			manifest, ok, commitErr := CommitReviewedTasksDraftManifestIfPresent(cfg.WorkDir)
			if commitErr != nil {
				return nil, fmt.Errorf("提交已审查 DeckSpec 失败: %w", commitErr)
			}
			if !ok || manifest == nil {
				return nil, fmt.Errorf("DeckSpec 审查通过但没有可提交的规划草稿")
			}
			onEvent(AgentEvent{Type: AgentEventAnswer, Content: fmt.Sprintf("DeckSpec 已通过 Task Reviewer 并提交，共 %d 页，开始进入并发渲染。\n", len(manifest.Tasks))})
			if cfg.RuntimeMeta != nil {
				cfg.RuntimeMeta.RecordEvent("deck_spec_committed", tasksDraftFileName, "success", fmt.Sprintf("committed %d slides from reviewed draft", len(manifest.Tasks)), map[string]any{
					"slide_count": len(manifest.Tasks),
				})
			}
			return manifest, nil
		}

		input, allowedPageIndexes, err := buildPlanReviewRevisionInput(cfg.WorkDir, round, report)
		if err != nil {
			return nil, fmt.Errorf("构建第 %d 轮 Reviewer 切片输入失败: %w", round, err)
		}
		if cfg.RuntimeMeta != nil {
			cfg.RuntimeMeta.RecordEvent("deck_spec_review_slice", "TaskPlanReviewer", "needs_revision", fmt.Sprintf("scoped review patch for %d pages", len(allowedPageIndexes)), map[string]any{
				"round":                round,
				"allowed_page_indexes": allowedPageIndexes,
				"allowed_count":        len(allowedPageIndexes),
				"review_summary":       report.Summary,
			})
		}
		reviewer, err := NewTaskPlanReviewerAgent(ctx, cfg, allowedPageIndexes)
		if err != nil {
			return nil, err
		}
		if err := runAgentWithCallback(ctx, reviewer, input, onEvent); err != nil {
			return nil, fmt.Errorf("Task Reviewer 第 %d 轮修正失败: %w", round, err)
		}
	}
	if latest == nil {
		return nil, fmt.Errorf("Task Reviewer 未产生审查结果")
	}
	return nil, fmt.Errorf("DeckSpec 在 %d 轮审查后仍未通过：%s", maxPlanReviewRounds, latest.Summary)
}

func makePrintCallback() AgentEventCallback {
	return func(event AgentEvent) {
		switch event.Type {
		case AgentEventAnswer:
			logger.Info("cli_event", "type", "answer", "content", event.Content)
		case AgentEventToolCall:
			logger.Info("cli_event", "type", "tool_call", "tool", event.ToolName, "args_len", len(event.ToolArgs))
		case AgentEventError:
			logger.Error("cli_event", "type", "error", "error", event.Error)
		}
	}
}

// isChunkEmittable 如果消息的 content 应作为 LLM 答案文本发出则返回 true。
// 工具结果块和携带 ToolCalls 的 assistant 内容都被跳过，避免把内部工具规划
// 或供应商格式碎片作为用户可见回答输出。ToolCalls 本身仍通过 tool_call 事件转发。
func isChunkEmittable(chunk *schema.Message) bool {
	if chunk == nil {
		return false
	}
	if chunk.Role == schema.Tool {
		return false
	}
	if chunk.ToolCallID != "" {
		return false
	}
	if len(chunk.ToolCalls) > 0 {
		return false
	}
	return true
}

func visibleMessageContent(content string) string {
	if isInternalCompressionSummaryContent(content) {
		return ""
	}
	return content
}

func isInternalCompressionSummaryContent(content string) bool {
	content = strings.TrimSpace(content)
	if content == "" {
		return false
	}
	content = strings.TrimSpace(strings.TrimPrefix(content, "\ufeff"))
	if strings.HasPrefix(content, "```") {
		lines := strings.Split(content, "\n")
		if len(lines) >= 2 && strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
			content = strings.TrimSpace(strings.Join(lines[1:len(lines)-1], "\n"))
		}
	}
	if !strings.HasPrefix(content, "{") {
		return false
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(content), &payload); err != nil {
		return false
	}
	if _, ok := payload["user_request_summary"]; !ok {
		return false
	}
	if _, ok := payload["progress_summary"]; ok {
		return true
	}
	if _, ok := payload["conversation_summary"]; ok {
		return true
	}
	return false
}

func shouldBufferForInternalSummaryProbe(buffer string) bool {
	trimmed := strings.TrimSpace(buffer)
	if trimmed == "" {
		return true
	}
	trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "\ufeff"))
	return strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "```")
}

func emitVisibleAnswer(content string, onEvent AgentEventCallback, buf *strings.Builder) {
	content = visibleMessageContent(content)
	if content == "" {
		return
	}
	if buf != nil {
		buf.WriteString(content)
	}
	onEvent(AgentEvent{
		Type:    AgentEventAnswer,
		Content: content,
	})
}

func processStreamingMessage(stream *schema.StreamReader[adk.Message], onEvent AgentEventCallback, buf *strings.Builder) {
	if stream == nil {
		return
	}
	var probe strings.Builder
	probing := true
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if probe.Len() > 0 {
				emitVisibleAnswer(probe.String(), onEvent, buf)
			}
			if err != io.EOF {
				onEvent(AgentEvent{Type: AgentEventError, Error: err.Error()})
			}
			return
		}
		if !isChunkEmittable(chunk) {
			continue
		}
		if chunk.Content == "" {
			continue
		}
		if probing {
			probe.WriteString(chunk.Content)
			if shouldBufferForInternalSummaryProbe(probe.String()) {
				continue
			}
			probing = false
			emitVisibleAnswer(probe.String(), onEvent, buf)
			probe.Reset()
			continue
		}
		emitVisibleAnswer(chunk.Content, onEvent, buf)
	}
}

// runAgentWithCallback 是通用代理运行器，通过 onEvent 流式传输事件
func runAgentWithCallback(ctx context.Context, agent adk.Agent, userInput string, onEvent AgentEventCallback) error {
	startTime := time.Now()

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: adkStreamingEnabled(),
	})

	var messages []adk.Message
	if userInput != "" {
		messages = []adk.Message{schema.UserMessage(userInput)}
	}
	iter := runner.Run(ctx, messages)

	err := streamAgentEvents(ctx, iter, onEvent)
	logger.Info("agent_stream_completed", "duration", time.Since(startTime).String())
	return err
}

// RunPPTFixerWithCallback 运行生成后定点修复 Agent。
func RunPPTFixerWithCallback(ctx context.Context, agent adk.Agent, userInput string, onEvent AgentEventCallback) error {
	return runAgentWithCallback(ctx, agent, userInput, onEvent)
}

// streamAgentEvents 消费所有代理事件并通过 onEvent 转发它们
// 使用流超时来避免在停滞的 LLM 流上永久阻塞
func streamAgentEvents(ctx context.Context, iter *adk.AsyncIterator[*adk.AgentEvent], onEvent AgentEventCallback) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		event, ok, timedOut := nextWithTimeout(ctx, iter)
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if timedOut {
			logger.Error("stream_agent_timeout", "timeout", streamTimeout().String())
			return fmt.Errorf("LLM 流式输出超时（%s）", streamTimeout())
		}

		if !ok {
			break
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			if event.Output.MessageOutput.IsStreaming {
				cpStream := event.Output.MessageOutput.MessageStream.Copy(2)
				event.Output.MessageOutput.MessageStream = cpStream[0]
				lastMsgStream := cpStream[1]
				defer lastMsgStream.Close()

				answerBuf := strings.Builder{}
				processStreamingMessage(lastMsgStream, onEvent, &answerBuf)
			} else {
				if msg := event.Output.MessageOutput.Message; msg != nil {
					if isChunkEmittable(msg) && msg.Content != "" {
						if content := visibleMessageContent(msg.Content); content != "" {
							onEvent(AgentEvent{Type: AgentEventAnswer, Content: content})
						}
					}
					for _, tc := range msg.ToolCalls {
						onEvent(AgentEvent{
							Type:     AgentEventToolCall,
							ToolName: tc.Function.Name,
							ToolArgs: tc.Function.Arguments,
						})
					}
				}
			}
		}

		if event.Err != nil {
			onEvent(AgentEvent{Type: AgentEventError, Error: event.Err.Error()})
		}
	}

	return nil
}
