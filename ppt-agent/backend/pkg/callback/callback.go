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

package callback

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/metrics"
)

// startTime 程序启动时间
var startTime = time.Now()

// maxToolOutputLen tool 输出截断阈值
const maxToolOutputLen = 500

// maxToolArgsLen tool 参数截断阈值
const maxToolArgsLen = 300

// keyToAgentName 是用于在 context 中存储 agent 名称的 key
const keyToAgentName = "eino.callback.agent.name"

// SetAgentName 将 agent 名称存入 context（供 wrapper/agent 调用）
func SetAgentName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, keyToAgentName, name)
}

// getAgentName 从 context 或 RunInfo 中获取 agent 名称
func getAgentName(ctx context.Context, info *callbacks.RunInfo) string {
	if ctx != nil {
		if name, ok := ctx.Value(keyToAgentName).(string); ok && name != "" {
			return name
		}
	}
	if info != nil && info.Name != "" {
		return info.Name
	}
	return "?"
}

// elapsed 返回程序启动以来的毫秒数
func elapsed() int64 {
	return time.Since(startTime).Milliseconds()
}

// httpErrorDetail 存储从 volcengine 错误链中提取的 HTTP 错误元数据
type httpErrorDetail struct {
	statusCode int
	requestID  string
}

// extractHTTPErrors 遍历错误链，查找 volcengine RequestFailure，提取状态码和请求ID。
// SDK 的 error 只暴露 Error() 方法，因此直接正则解析 Error() 字符串。
func extractHTTPErrors(err error) *httpErrorDetail {
	if err == nil {
		return nil
	}

	errStr := ""
	for e := err; e != nil; e = errors.Unwrap(e) {
		errStr = e.Error()
		if strings.Contains(errStr, "RequestError") {
			break
		}
	}

	if !strings.Contains(errStr, "RequestError") {
		return nil
	}

	detail := &httpErrorDetail{}

	// 解析 "RequestError code: 429, err: <nil>, request_id: xxx"
	if re := regexp.MustCompile(`code:\s*(\d+)`); re.MatchString(errStr) {
		if m := re.FindStringSubmatch(errStr); len(m) > 1 {
			if c, convErr := strconv.Atoi(m[1]); convErr == nil {
				detail.statusCode = c
			}
		}
	}

	// 解析 request_id: xxx（去除引号和逗号）
	if re := regexp.MustCompile(`request_id:\s*"?([^",\s]*)"?`); re.MatchString(errStr) {
		if m := re.FindStringSubmatch(errStr); len(m) > 1 {
			detail.requestID = strings.TrimSpace(m[1])
		}
	}

	if detail.statusCode == 0 {
		return nil
	}
	return detail
}

// truncate 截断过长字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + fmt.Sprintf("...[truncated %d chars]", len(s))
}

// extractToolArgs 提取工具参数字符串（带截断）
func extractToolArgs(input callbacks.CallbackInput) string {
	tci := tool.ConvCallbackInput(input)
	if tci == nil {
		return "(no args)"
	}
	if tci.ArgumentsInJSON == "" {
		return "(no args)"
	}
	return truncate(tci.ArgumentsInJSON, maxToolArgsLen)
}

// extractToolResult 提取工具结果字符串（带截断）
func extractToolResult(output callbacks.CallbackOutput) string {
	tco := tool.ConvCallbackOutput(output)
	if tco == nil {
		return "(empty)"
	}
	if tco.Response == "" {
		return "(empty)"
	}
	return truncate(tco.Response, maxToolOutputLen)
}

// StreamEvent 打印流式事件信息（从 MessageStream 中读取并消费）
func StreamEvent(event *adk.AgentEvent) {
	fields := []any{
		"agent", event.AgentName,
		"path", event.RunPath,
		"elapsed_ms", elapsed(),
	}

	if event.Output != nil && event.Output.MessageOutput != nil {
		if m := event.Output.MessageOutput.Message; m != nil {
			if len(m.Content) > 0 {
				fields = append(fields, "role", string(m.Role), "content_len", len(m.Content))
			}
			if len(m.ToolCalls) > 0 {
				toolNames := make([]string, len(m.ToolCalls))
				for i, tc := range m.ToolCalls {
					toolNames[i] = tc.Function.Name
				}
				fields = append(fields, "tool_calls", toolNames)
			}
		} else if s := event.Output.MessageOutput.MessageStream; s != nil {
			toolMap := map[int][]*schema.Message{}
			for {
				chunk, err := s.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}
					logger.Default().Error("stream recv error", "err", err)
					return
				}

				if len(chunk.ToolCalls) > 0 {
					for _, tc := range chunk.ToolCalls {
						index := tc.Index
						if index == nil {
							continue
						}
						toolMap[*index] = append(toolMap[*index], &schema.Message{
							Role: chunk.Role,
							ToolCalls: []schema.ToolCall{
								{
									ID:    tc.ID,
									Type:  tc.Type,
									Index: tc.Index,
									Function: schema.FunctionCall{
										Name:      tc.Function.Name,
										Arguments: tc.Function.Arguments,
									},
								},
							},
						})
					}
				}
			}

			for _, msgs := range toolMap {
				m, err := schema.ConcatMessages(msgs)
				if err != nil {
					continue
				}
				if len(m.ToolCalls) > 0 {
					fields = append(fields,
						"tool_name", m.ToolCalls[0].Function.Name,
						"tool_args_len", len(m.ToolCalls[0].Function.Arguments),
					)
				}
			}
		}
	}

	if event.Action != nil {
		if event.Action.TransferToAgent != nil {
			fields = append(fields, "action", "transfer", "dest_agent", event.Action.TransferToAgent.DestAgentName)
		}
		if event.Action.Interrupted != nil {
			fields = append(fields, "action", "interrupted")
		}
		if event.Action.Exit {
			fields = append(fields, "action", "exit")
		}
	}

	if event.Err != nil {
		fields = append(fields, "error", event.Err.Error())
	}

	logger.Default().Info("agent_stream_event", fields...)
}

// NewLogHandler 创建一个 slog-based 的日志 Handler。
func NewLogHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if info == nil {
				return ctx
			}

			agentName := getAgentName(ctx, info)
			ctx = SetAgentName(ctx, agentName)

			fields := []any{
				"elapsed_ms", elapsed(),
				"agent", agentName,
				"name", info.Name,
			}

			switch info.Component {
			case components.ComponentOfTool:
				args := extractToolArgs(input)
				if meta := utils.RuntimeMetaFromContext(ctx); meta != nil {
					meta.RecordToolStart(info.Name, args)
				}
				logger.Default().Info("tool_call_start", append(fields, "args", args)...)
			case components.ComponentOfChatModel:
				logger.Default().Info("llm_call_start", fields...)
			case adk.ComponentOfAgent:
				metrics.RecordAgentCall(agentName)
				logger.Default().Info("agent_start", fields...)
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if info == nil {
				return ctx
			}

			agentName := getAgentName(ctx, info)
			fields := []any{
				"elapsed_ms", elapsed(),
				"agent", agentName,
				"name", info.Name,
			}

			switch info.Component {
			case components.ComponentOfTool:
				result := extractToolResult(output)
				logger.Default().Info("tool_call_end", append(fields, "result_preview", truncate(result, 100))...)
				metrics.RecordToolCall(info.Name, "success")
			case components.ComponentOfChatModel:
				if mo := model.ConvCallbackOutput(output); mo != nil && mo.TokenUsage != nil {
					tu := mo.TokenUsage
					logger.Default().Info("llm_call_end",
						append(fields,
							"prompt_tokens", tu.PromptTokens,
							"completion_tokens", tu.CompletionTokens,
							"total_tokens", tu.TotalTokens,
						)...)
					metrics.RecordTokens(int64(tu.PromptTokens), int64(tu.CompletionTokens), int64(tu.TotalTokens))
					metrics.RecordLLMCall("success")
					if tt := utils.TokenTrackerFromContext(ctx); tt != nil {
						tt.Add(tu.PromptTokens, tu.CompletionTokens, tu.TotalTokens)
					}
					if meta := utils.RuntimeMetaFromContext(ctx); meta != nil {
						meta.RecordLLMTokens(int64(tu.PromptTokens), int64(tu.CompletionTokens), int64(tu.TotalTokens))
					}
				}
			case adk.ComponentOfAgent:
				logger.Default().Info("agent_end", fields...)
			}
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			if err == nil {
				return ctx
			}
			name := "?"
			infoName := "?"
			if info != nil {
				name = getAgentName(ctx, info)
				infoName = info.Name
			}
			logger.Default().Error("callback_error",
				"elapsed_ms", elapsed(),
				"agent", name,
				"name", infoName,
				"error", err.Error(),
			)
			if meta := utils.RuntimeMetaFromContext(ctx); meta != nil {
				meta.RecordToolError(infoName, err.Error())
			}

			// 提取 volcengine HTTP 错误的详细信息
			if httpErr := extractHTTPErrors(err); httpErr != nil {
				logger.Default().Error("callback_http_error",
					"elapsed_ms", elapsed(),
					"agent", name,
					"name", infoName,
					"http_status", httpErr.statusCode,
					"request_id", httpErr.requestID,
				)
			}

			if info != nil {
				switch info.Component {
				case components.ComponentOfChatModel:
					if utils.IsRateLimitError(err) {
						metrics.RecordLLMCall("rate_limit")
					} else if httpErr := extractHTTPErrors(err); httpErr != nil && httpErr.statusCode > 0 {
						metrics.RecordLLMCall(fmt.Sprintf("http_%d", httpErr.statusCode))
					} else {
						metrics.RecordLLMCall("error")
					}
				case components.ComponentOfTool:
					metrics.RecordToolCall(info.Name, "error")
				}
			}
			return ctx
		}).
		OnStartWithStreamInputFn(func(ctx context.Context, info *callbacks.RunInfo, input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
			return ctx
		}).
		OnEndWithStreamOutputFn(func(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
			return ctx
		}).
		Build()
}
