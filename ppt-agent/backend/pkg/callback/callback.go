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
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
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
	// 1. 优先从 context 中获取（最准确，由 agent wrapper 设置）
	if ctx != nil {
		if name, ok := ctx.Value(keyToAgentName).(string); ok && name != "" {
			return name
		}
	}

	// 2. 从 RunInfo.Name 中获取
	if info != nil && info.Name != "" {
		return info.Name
	}

	return "?"
}

// elapsed 返回程序启动以来的毫秒数
func elapsed() int64 {
	return time.Since(startTime).Milliseconds()
}

// truncate 截断过长字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + fmt.Sprintf(" ...[截断 %d 字符]", len(s))
}

// extractToolArgs 提取工具参数字符串（带截断）
func extractToolArgs(input callbacks.CallbackInput) string {
	tci := tool.ConvCallbackInput(input)
	if tci.ArgumentsInJSON == "" {
		return "(无参数)"
	}
	return truncate(tci.ArgumentsInJSON, maxToolArgsLen)
}

// extractToolResult 提取工具结果字符串（带截断）
func extractToolResult(output callbacks.CallbackOutput) string {
	tco := tool.ConvCallbackOutput(output)
	if tco.Response == "" {
		return "(空响应)"
	}
	return truncate(tco.Response, maxToolOutputLen)
}

// NewLogHandler 创建一个精简的日志 Handler。
// 通过 context 和 RunInfo 获取 agent 名称，不再依赖全局 agentStack。
func NewLogHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if info == nil {
				return ctx
			}

			agentName := getAgentName(ctx, info)

			switch info.Component {
			case components.ComponentOfTool:
				args := extractToolArgs(input)
				log.Printf("[%dms] [%s] → TOOL: %s | args: %s", elapsed(), agentName, info.Name, args)
			case components.ComponentOfChatModel:
				log.Printf("[%dms] [%s] → LLM: %s", elapsed(), agentName, info.Name)
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if info == nil {
				return ctx
			}

			agentName := getAgentName(ctx, info)

			switch info.Component {
			case components.ComponentOfTool:
				result := extractToolResult(output)
				log.Printf("[%dms] [%s] ← TOOL: %s | result: %s", elapsed(), agentName, info.Name, result)
			case components.ComponentOfChatModel:
				// 不输出 LLM 响应内容
			}
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			if err == nil {
				return ctx
			}
			name := "?"
			if info != nil {
				name = getAgentName(ctx, info)
			}
			log.Printf("[%dms] [ERROR] [%s] %s | %v", elapsed(), name, info.Name, err)
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
