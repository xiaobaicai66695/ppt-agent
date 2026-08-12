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
	"encoding/json"
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

// maxModelMessagePreviewLen limits observable model context shown in the UI.
const maxModelMessagePreviewLen = 600

// maxModelHistoryMessages keeps the latest model context compact enough for SSE.
const maxModelHistoryMessages = 14

// keyToAgentName 是用于在 context 中存储 agent 名称的 key
const keyToAgentName = "eino.callback.agent.name"

type callbackInputDetailsKey struct{}

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

func extractFullToolArgs(input callbacks.CallbackInput) string {
	tci := tool.ConvCallbackInput(input)
	if tci == nil || tci.ArgumentsInJSON == "" {
		return ""
	}
	return tci.ArgumentsInJSON
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

func extractFullToolResult(output callbacks.CallbackOutput) string {
	tco := tool.ConvCallbackOutput(output)
	if tco == nil {
		return ""
	}
	if tco.Response != "" {
		return tco.Response
	}
	if tco.ToolOutput != nil {
		data, err := json.Marshal(tco.ToolOutput)
		if err == nil {
			return string(data)
		}
	}
	return ""
}

func callbackInputDetailsFromContext(ctx context.Context) map[string]any {
	if ctx == nil {
		return nil
	}
	if details, ok := ctx.Value(callbackInputDetailsKey{}).(map[string]any); ok {
		return cloneMetadata(details)
	}
	return nil
}

func cloneMetadata(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func modelInputDetails(input callbacks.CallbackInput, agentName, runName string) map[string]any {
	mi := model.ConvCallbackInput(input)
	if mi == nil {
		return map[string]any{"agent": agentName, "name": runName}
	}
	toolNames := make([]string, 0, len(mi.Tools))
	for _, toolInfo := range mi.Tools {
		if toolInfo != nil && strings.TrimSpace(toolInfo.Name) != "" {
			toolNames = append(toolNames, toolInfo.Name)
		}
	}
	return map[string]any{
		"agent":             agentName,
		"name":              runName,
		"stage":             inferCallbackStage(agentName, runName),
		"message_count":     len(mi.Messages),
		"tool_names":        toolNames,
		"last_user_preview": lastMessagePreview(mi.Messages, string(schema.User), 180),
		"system_preview":    firstMessagePreview(mi.Messages, string(schema.System), 180),
		"history":           compactModelHistory(mi.Messages),
	}
}

func modelOutputDetails(output callbacks.CallbackOutput) map[string]any {
	mo := model.ConvCallbackOutput(output)
	if mo == nil {
		return nil
	}
	details := map[string]any{}
	if mo.Message != nil {
		details["output_preview"] = truncate(strings.TrimSpace(mo.Message.Content), 500)
		details["assistant_message"] = compactModelMessage(mo.Message, 0)
		if reasoning := compactReasoningPreview(mo.Message); reasoning != "" {
			details["reasoning_preview"] = reasoning
		}
		if len(mo.Message.ToolCalls) > 0 {
			toolNames := make([]string, 0, len(mo.Message.ToolCalls))
			toolCalls := make([]any, 0, len(mo.Message.ToolCalls))
			for _, tc := range mo.Message.ToolCalls {
				toolNames = append(toolNames, tc.Function.Name)
				toolCalls = append(toolCalls, compactToolCall(tc))
			}
			details["tool_calls"] = toolNames
			details["tool_call_details"] = toolCalls
		}
	}
	if mo.TokenUsage != nil {
		details["token_usage"] = map[string]any{
			"prompt_tokens":            mo.TokenUsage.PromptTokens,
			"prompt_token_details":     mo.TokenUsage.PromptTokenDetails,
			"completion_tokens":        mo.TokenUsage.CompletionTokens,
			"completion_token_details": mo.TokenUsage.CompletionTokensDetails,
			"total_tokens":             mo.TokenUsage.TotalTokens,
			"reasoning_tokens":         mo.TokenUsage.CompletionTokensDetails.ReasoningTokens,
			"cached_prompt_tokens":     mo.TokenUsage.PromptTokenDetails.CachedTokens,
		}
	}
	if mo.Config != nil {
		details["output_config"] = mo.Config
	}
	if mo.Extra != nil {
		if reasoning := compactReasoningFromAny(mo.Extra); reasoning != "" && details["reasoning_preview"] == nil {
			details["reasoning_preview"] = reasoning
		}
		details["output_extra"] = mo.Extra
	}
	return details
}

func compactModelHistory(messages []*schema.Message) []any {
	if len(messages) == 0 {
		return nil
	}
	start := 0
	if len(messages) > maxModelHistoryMessages {
		start = len(messages) - maxModelHistoryMessages
	}
	history := make([]any, 0, len(messages)-start)
	for i := start; i < len(messages); i++ {
		if compacted := compactModelMessage(messages[i], i); compacted != nil {
			history = append(history, compacted)
		}
	}
	return history
}

func compactModelMessage(message *schema.Message, index int) map[string]any {
	if message == nil {
		return nil
	}
	role := string(message.Role)
	result := map[string]any{
		"index": index,
		"role":  role,
	}
	if strings.TrimSpace(message.Name) != "" {
		result["name"] = message.Name
	}
	if role == string(schema.System) {
		result["content_preview"] = "[系统指令已隐藏，仅保留下方 system_preview 摘要]"
	} else if content := truncate(strings.TrimSpace(message.Content), maxModelMessagePreviewLen); content != "" {
		result["content_preview"] = content
	}
	if reasoning := compactReasoningPreview(message); reasoning != "" {
		result["reasoning_preview"] = reasoning
	}
	if len(message.ToolCalls) > 0 {
		toolNames := make([]string, 0, len(message.ToolCalls))
		toolCalls := make([]any, 0, len(message.ToolCalls))
		for _, tc := range message.ToolCalls {
			toolNames = append(toolNames, tc.Function.Name)
			toolCalls = append(toolCalls, compactToolCall(tc))
		}
		result["tool_calls"] = strings.Join(toolNames, ", ")
		result["tool_call_details"] = toolCalls
	}
	if strings.TrimSpace(message.ToolCallID) != "" {
		result["tool_call_id"] = truncate(strings.TrimSpace(message.ToolCallID), 120)
	}
	if strings.TrimSpace(message.ToolName) != "" {
		result["tool_name"] = truncate(strings.TrimSpace(message.ToolName), 120)
	}
	return result
}

func compactToolCall(tc schema.ToolCall) map[string]any {
	return map[string]any{
		"id":                truncate(strings.TrimSpace(tc.ID), 120),
		"name":              truncate(strings.TrimSpace(tc.Function.Name), 120),
		"arguments_preview": truncate(strings.TrimSpace(tc.Function.Arguments), maxToolArgsLen),
	}
}

func compactReasoningPreview(message *schema.Message) string {
	if message == nil {
		return ""
	}
	if reasoning := strings.TrimSpace(message.ReasoningContent); reasoning != "" {
		return truncate(reasoning, maxModelMessagePreviewLen)
	}
	return compactReasoningFromAny(message.Extra)
}

func compactReasoningFromAny(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return ""
	case map[string]any:
		for _, key := range []string{"reasoning_preview", "reasoning_content", "reasoning", "thinking", "thought"} {
			if text := strings.TrimSpace(fmt.Sprint(typed[key])); text != "" && text != "<nil>" {
				return truncate(text, maxModelMessagePreviewLen)
			}
		}
		for _, child := range typed {
			if text := compactReasoningFromAny(child); text != "" {
				return text
			}
		}
	case []any:
		for _, child := range typed {
			if text := compactReasoningFromAny(child); text != "" {
				return text
			}
		}
	}
	return ""
}

func inferCallbackStage(agentName, runName string) string {
	joined := strings.ToLower(agentName + " " + runName)
	switch {
	case strings.Contains(joined, "planner"):
		return "planner"
	case strings.Contains(joined, "workflow"), strings.Contains(joined, "render"):
		return "workflow"
	default:
		return "model"
	}
}

func firstMessagePreview(messages []*schema.Message, role string, maxLen int) string {
	for _, message := range messages {
		if message != nil && string(message.Role) == role {
			return truncate(strings.TrimSpace(message.Content), maxLen)
		}
	}
	return ""
}

func lastMessagePreview(messages []*schema.Message, role string, maxLen int) string {
	for i := len(messages) - 1; i >= 0; i-- {
		message := messages[i]
		if message != nil && string(message.Role) == role && !isAgentProgressMessage(message.Content) {
			return truncate(strings.TrimSpace(message.Content), maxLen)
		}
	}
	return ""
}

func isAgentProgressMessage(content string) bool {
	content = strings.TrimSpace(content)
	return strings.HasPrefix(content, "<agent_progress>") && strings.Contains(content, "</agent_progress>")
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
				fullArgs := extractFullToolArgs(input)
				args := truncate(fullArgs, maxToolArgsLen)
				if fullArgs == "" {
					args = extractToolArgs(input)
				}
				if meta := utils.RuntimeMetaFromContext(ctx); meta != nil {
					meta.RecordToolStart(info.Name, fullArgs)
				}
				logger.Default().Info("tool_call_start", append(fields, "args", args)...)
				ctx = context.WithValue(ctx, callbackInputDetailsKey{}, map[string]any{
					"kind":         "tool",
					"agent":        agentName,
					"name":         info.Name,
					"args":         fullArgs,
					"args_preview": args,
				})
			case components.ComponentOfChatModel:
				details := modelInputDetails(input, agentName, info.Name)
				if meta := utils.RuntimeMetaFromContext(ctx); meta != nil {
					meta.RecordLLMStartDetails(info.Name, details)
				}
				logger.Default().Info("llm_call_start", fields...)
				ctx = context.WithValue(ctx, callbackInputDetailsKey{}, details)
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
				inputDetails := callbackInputDetailsFromContext(ctx)
				fullArgs, _ := inputDetails["args"].(string)
				result := extractToolResult(output)
				fullResult := extractFullToolResult(output)
				if meta := utils.RuntimeMetaFromContext(ctx); meta != nil {
					meta.RecordToolEnd(info.Name, fullArgs, fullResult)
				}
				logger.Default().Info("tool_call_end", append(fields, "result_preview", truncate(result, 100))...)
				metrics.RecordToolCall(info.Name, "success")
			case components.ComponentOfChatModel:
				metadata := callbackInputDetailsFromContext(ctx)
				outputDetails := modelOutputDetails(output)
				for k, v := range outputDetails {
					if metadata == nil {
						metadata = map[string]any{}
					}
					metadata[k] = v
				}
				if mo := model.ConvCallbackOutput(output); mo != nil {
					var promptTokens, completionTokens, totalTokens int
					if mo.TokenUsage != nil {
						tu := mo.TokenUsage
						promptTokens = tu.PromptTokens
						completionTokens = tu.CompletionTokens
						totalTokens = tu.TotalTokens
						if tt := utils.TokenTrackerFromContext(ctx); tt != nil {
							tt.Add(tu.PromptTokens, tu.CompletionTokens, tu.TotalTokens)
						}
					}
					logger.Default().Info("llm_call_end",
						append(fields,
							"prompt_tokens", promptTokens,
							"completion_tokens", completionTokens,
							"total_tokens", totalTokens,
						)...)
					metrics.RecordTokens(int64(promptTokens), int64(completionTokens), int64(totalTokens))
					metrics.RecordLLMCall("success")
					if meta := utils.RuntimeMetaFromContext(ctx); meta != nil {
						meta.RecordLLMEndDetails(info.Name, int64(promptTokens), int64(completionTokens), int64(totalTokens), metadata)
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
				if info != nil && info.Component == components.ComponentOfChatModel {
					meta.RecordLLMErrorDetails(infoName, err.Error(), callbackInputDetailsFromContext(ctx))
				} else {
					meta.RecordToolErrorDetails(infoName, err.Error(), callbackInputDetailsFromContext(ctx))
				}
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
