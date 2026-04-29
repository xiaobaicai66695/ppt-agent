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
	"io"
	"log"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk"
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
	if tci == nil {
		return "(无参数)"
	}
	if tci.ArgumentsInJSON == "" {
		return "(无参数)"
	}
	return truncate(tci.ArgumentsInJSON, maxToolArgsLen)
}

// extractToolResult 提取工具结果字符串（带截断）
func extractToolResult(output callbacks.CallbackOutput) string {
	tco := tool.ConvCallbackOutput(output)
	if tco == nil {
		return "(空响应)"
	}
	if tco.Response == "" {
		return "(空响应)"
	}
	return truncate(tco.Response, maxToolOutputLen)
}

// Event 打印事件信息（与 eino-examples/adk/common/prints/util.go 保持一致）
// 流式消息通过事件循环中的 MessageStream.Copy(2) 模式处理，
// 此函数只处理非流式的 Message 或已经被复制消费后的流。
func Event(event *adk.AgentEvent) {
	fmt.Printf("name: %s\npath: %s", event.AgentName, event.RunPath)

	if event.Output != nil && event.Output.MessageOutput != nil {
		if m := event.Output.MessageOutput.Message; m != nil {
			if len(m.Content) > 0 {
				if m.Role == schema.Tool {
					fmt.Printf("\ntool response: %s", m.Content)
				} else {
					fmt.Printf("\nanswer: %s", m.Content)
				}
			}
			if len(m.ToolCalls) > 0 {
				for _, tc := range m.ToolCalls {
					fmt.Printf("\ntool name: %s", tc.Function.Name)
					fmt.Printf("\narguments: %s", tc.Function.Arguments)
				}
			}
		}
	}

	if event.Action != nil {
		if event.Action.TransferToAgent != nil {
			fmt.Printf("\naction: transfer to %v", event.Action.TransferToAgent.DestAgentName)
		}
		if event.Action.Interrupted != nil {
			for _, ic := range event.Action.Interrupted.InterruptContexts {
				str, ok := ic.Info.(fmt.Stringer)
				if ok {
					fmt.Printf("\n%s", str.String())
				} else {
					fmt.Printf("\n%v", ic.Info)
				}
			}
		}
		if event.Action.Exit {
			fmt.Printf("\naction: exit")
		}
	}

	if event.Err != nil {
		fmt.Printf("\nerror: %v", event.Err)
	}

	fmt.Println()
	fmt.Println()
}

// StreamEvent 打印流式事件信息（从 MessageStream 中读取并消费）
func StreamEvent(event *adk.AgentEvent) {
	fmt.Printf("name: %s\npath: %s", event.AgentName, event.RunPath)

	if event.Output != nil && event.Output.MessageOutput != nil {
		if m := event.Output.MessageOutput.Message; m != nil {
			if len(m.Content) > 0 {
				if m.Role == schema.Tool {
					fmt.Printf("\ntool response: %s", m.Content)
				} else {
					fmt.Printf("\nanswer: %s", m.Content)
				}
			}
			if len(m.ToolCalls) > 0 {
				for _, tc := range m.ToolCalls {
					fmt.Printf("\ntool name: %s", tc.Function.Name)
					fmt.Printf("\narguments: %s", tc.Function.Arguments)
				}
			}
		} else if s := event.Output.MessageOutput.MessageStream; s != nil {
			// 流式消息：读取并打印内容
			toolMap := map[int][]*schema.Message{}
			var contentStart bool
			charNumOfOneRow := 0
			maxCharNumOfOneRow := 120

			for {
				chunk, err := s.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}
					fmt.Printf("error: %v", err)
					return
				}

				if chunk.Content != "" {
					if !contentStart {
						contentStart = true
						if chunk.Role == schema.Tool {
							fmt.Printf("\ntool response: ")
						} else {
							fmt.Printf("\nanswer: ")
						}
					}

					charNumOfOneRow += len(chunk.Content)
					if strings.Contains(chunk.Content, "\n") {
						charNumOfOneRow = 0
					} else if charNumOfOneRow >= maxCharNumOfOneRow {
						fmt.Printf("\n")
						charNumOfOneRow = 0
					}
					fmt.Printf("%v", chunk.Content)
				}

				if len(chunk.ToolCalls) > 0 {
					for _, tc := range chunk.ToolCalls {
						index := tc.Index
						if index == nil {
							log.Printf("[WARN] tool call index is nil, skipping")
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

			// 打印工具调用
			for _, msgs := range toolMap {
				m, err := schema.ConcatMessages(msgs)
				if err != nil {
					log.Printf("ConcatMessage failed: %v", err)
					continue
				}
				if len(m.ToolCalls) > 0 {
					fmt.Printf("\ntool name: %s", m.ToolCalls[0].Function.Name)
					fmt.Printf("\narguments: %s", m.ToolCalls[0].Function.Arguments)
				}
			}
		}
	}

	if event.Action != nil {
		if event.Action.TransferToAgent != nil {
			fmt.Printf("\naction: transfer to %v", event.Action.TransferToAgent.DestAgentName)
		}
		if event.Action.Interrupted != nil {
			for _, ic := range event.Action.Interrupted.InterruptContexts {
				str, ok := ic.Info.(fmt.Stringer)
				if ok {
					fmt.Printf("\n%s", str.String())
				} else {
					fmt.Printf("\n%v", ic.Info)
				}
			}
		}
		if event.Action.Exit {
			fmt.Printf("\naction: exit")
		}
	}

	if event.Err != nil {
		fmt.Printf("\nerror: %v", event.Err)
	}

	fmt.Println()
	fmt.Println()
}

// NewLogHandler 创建一个精简的日志 Handler（参考 eino-examples/adk/common/prints）
// 注意：此 handler 只处理 OnStart 和 OnEnd，不处理流式输出。
// 流式输出通过事件循环中的 MessageStream.Copy(2) 模式在 StreamEvent 中处理。
func NewLogHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if info == nil {
				return ctx
			}

			agentName := getAgentName(ctx, info)
			ctx = SetAgentName(ctx, agentName)

			switch info.Component {
			case components.ComponentOfTool:
				args := extractToolArgs(input)
				log.Printf("[%dms] [%s] → TOOL: %s | args: %s", elapsed(), agentName, info.Name, args)
			case components.ComponentOfChatModel:
				log.Printf("[%dms] [%s] → LLM: %s", elapsed(), agentName, info.Name)
			case adk.ComponentOfAgent:
				log.Printf("[%dms] [%s] → AGENT: %s", elapsed(), agentName, info.Name)
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
				if resp := extractToolResult(output); resp != "(空响应)" {
					log.Printf("[%dms] [%s] ← LLM | result: %s", elapsed(), agentName, resp)
				}
			case adk.ComponentOfAgent:
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
