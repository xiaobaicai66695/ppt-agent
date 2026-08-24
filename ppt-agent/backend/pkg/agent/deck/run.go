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

// adkStreamingEnabled controls ADK runner streaming.
// It is disabled by default because streaming tool-call arguments can be
// delivered in partial chunks and trigger intermittent JSON parse failures in ToolNode.
func adkStreamingEnabled() bool {
	return strings.EqualFold(os.Getenv("ADK_ENABLE_STREAMING"), "true")
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
					onEvent(AgentEvent{
						Type:    AgentEventAnswer,
						Content: content,
					})
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

	// Delivery readiness is owned by TaskManager metadata. The runner only
	// reports the manifest state and never performs a final filesystem scan.
	manifest, manifestErr := ReadTasksManifest(cfg.WorkDir)
	result = &PPTTaskResult{
		Message:  lastMsg,
		Duration: time.Since(start.StartTime),
	}

	if plannerErr != nil {
		return result, fmt.Errorf("Planner 执行失败: %w", plannerErr)
	}
	if committed, ok, commitErr := CommitReviewedTasksDraftManifestIfPresent(cfg.WorkDir); commitErr != nil {
		return result, fmt.Errorf("提交 DeckSpec 草稿失败: %w", commitErr)
	} else if ok {
		manifest = committed
		manifestErr = nil
		onEvent(AgentEvent{
			Type:    AgentEventAnswer,
			Content: fmt.Sprintf("DeckSpec 规划草稿已通过硬校验并提交，共 %d 页，开始进入并发渲染。\n", len(manifest.Tasks)),
		})
		if cfg.RuntimeMeta != nil {
			cfg.RuntimeMeta.RecordEvent("deck_spec_committed", "tasks.draft.json", "success", fmt.Sprintf("committed %d slides from draft", len(manifest.Tasks)), map[string]any{
				"slide_count": len(manifest.Tasks),
				"template":    manifest.Template,
				"theme":       manifest.Theme,
			})
		}
	}
	if manifestErr != nil {
		if os.IsNotExist(manifestErr) {
			recovered, recoverErr := recoverMissingPlannerManifest(cfg, userQuery, lastMsg)
			if recoverErr != nil {
				return result, fmt.Errorf("Planner 未生成 DeckSpec/tasks.json，恢复失败: %w", recoverErr)
			}
			manifest = recovered
			onEvent(AgentEvent{
				Type:    AgentEventAnswer,
				Content: fmt.Sprintf("Planner 已完成结构判断但未写入 DeckSpec，系统已根据现有规划摘要恢复 %d 页页面计划，继续进入并发渲染。\n", len(manifest.Tasks)),
			})
			if cfg.RuntimeMeta != nil {
				cfg.RuntimeMeta.RecordEvent("deck_spec_recovered", "tasks.json", "warning", fmt.Sprintf("recovered %d slides from planner output", len(manifest.Tasks)), map[string]any{
					"slide_count": len(manifest.Tasks),
					"template":    manifest.Template,
					"theme":       manifest.Theme,
				})
			}
		} else {
			return result, fmt.Errorf("读取 Planner DeckSpec/tasks.json 失败: %w", manifestErr)
		}
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
// 工具结果块（Role=="tool" 或 ToolCallID!=""）被跳过，因为它们是执行元数据。
// assistant 消息即使同时携带 ToolCalls，也可能包含 ReAct 风格的 Thought/Action 文本；
// 这部分 content 需要展示给用户，ToolCalls 本身仍通过 tool_call 事件单独转发。
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
	return true
}

func visibleMessageContent(content string) string {
	return content
}

func processStreamingMessage(stream *schema.StreamReader[adk.Message], onEvent AgentEventCallback, buf *strings.Builder) {
	if stream == nil {
		return
	}
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err != io.EOF {
				onEvent(AgentEvent{Type: AgentEventError, Error: err.Error()})
			}
			return
		}
		// 跳过工具结果块；assistant content 即使伴随 tool_calls 也应作为可见文本发送。
		if !isChunkEmittable(chunk) {
			continue
		}
		if chunk.Content == "" {
			continue
		}
		content := visibleMessageContent(chunk.Content)
		buf.WriteString(content)
		onEvent(AgentEvent{
			Type:    AgentEventAnswer,
			Content: content,
		})
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
						onEvent(AgentEvent{Type: AgentEventAnswer, Content: visibleMessageContent(msg.Content)})
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
