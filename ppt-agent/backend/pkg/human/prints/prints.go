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
 * distributed under the License is distributed on an "AS IS" BASIS
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package prints

import (
	"fmt"
	"io"
	"strings"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
	Bold   = "\033[1m"
)

func formatAgentName(name string) string {
	if name == "" {
		return "Assistant"
	}
	return name
}

func formatRunPath(path []adk.RunStep) string {
	return fmt.Sprintf("%v", path)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// PrintSeparator 打印分隔线
func PrintSeparator() {
	fmt.Println(strings.Repeat("─", 80))
}

func printHeader(title string) {
	fmt.Printf("%s%s%s\n", Bold, title, Reset)
}

// Event 打印事件信息
func Event(event *adk.AgentEvent) {
	PrintSeparator()

	// 打印 Agent 名称和路径
	agentName := formatAgentName(event.AgentName)
	runPath := formatRunPath(event.RunPath)
	fmt.Printf("%s[%s]%s  Path: %s\n", Cyan, agentName, Reset, runPath)

	// 打印输出内容
	if event.Output != nil && event.Output.MessageOutput != nil {
		if m := event.Output.MessageOutput.Message; m != nil {
			if len(m.Content) > 0 {
				content := m.Content
				if len(content) > 500 {
					content = content[:500] + "..."
				}
				if m.Role == schema.Tool {
					fmt.Printf("\n%s[Tool Response]%s\n%s\n", Yellow, Reset, content)
				} else {
					fmt.Printf("\n%s[Answer]%s\n%s\n", Green, Reset, content)
				}
			}
			if len(m.ToolCalls) > 0 {
				for _, tc := range m.ToolCalls {
					fmt.Printf("\n%s[Tool Call]%s\n", Purple, Reset)
					fmt.Printf("  %sTool:%s %s\n", Bold, Reset, tc.Function.Name)
					args := tc.Function.Arguments
					if len(args) > 300 {
						args = args[:300] + "..."
					}
					fmt.Printf("  %sArgs:%s %s\n", Bold, Reset, args)
				}
			}
		} else if s := event.Output.MessageOutput.MessageStream; s != nil {
			toolMap := map[int][]*schema.Message{}
			var contentStart bool

			for {
				chunk, err := s.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}
					fmt.Printf("%s[Error]%s %v\n", Red, Reset, err)
					return
				}
				if chunk.Content != "" {
					if !contentStart {
						contentStart = true
						if chunk.Role == schema.Tool {
							fmt.Printf("\n%s[Tool Response]%s\n", Yellow, Reset)
						} else {
							fmt.Printf("\n%s[Answer]%s\n", Green, Reset)
						}
					}
					fmt.Printf("%s", chunk.Content)
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

			// 打印工具调用
			for _, msgs := range toolMap {
				m, err := schema.ConcatMessages(msgs)
				if err != nil {
					continue
				}
				if len(m.ToolCalls) > 0 {
					fmt.Printf("\n%s[Tool Call]%s\n", Purple, Reset)
					fmt.Printf("  %sTool:%s %s\n", Bold, Reset, m.ToolCalls[0].Function.Name)
					args := m.ToolCalls[0].Function.Arguments
					if len(args) > 300 {
						args = args[:300] + "..."
					}
					fmt.Printf("  %sArgs:%s %s\n", Bold, Reset, args)
				}
			}
		}
	}

	// 打印 Action
	if event.Action != nil {
		if event.Action.TransferToAgent != nil {
			fmt.Printf("\n%s[Action]%s Transfer to: %s\n", Cyan, Reset, event.Action.TransferToAgent.DestAgentName)
		}
		if event.Action.Interrupted != nil {
			fmt.Printf("\n%s[Interrupted]%s\n", Yellow, Reset)
			for i, ic := range event.Action.Interrupted.InterruptContexts {
				fmt.Printf("  [%d] ID: %s, Type: %T\n", i, ic.ID, ic.Info)
				if str, ok := ic.Info.(fmt.Stringer); ok {
					fmt.Printf("      Info: %s\n", str.String())
				}
			}
		}
		if event.Action.Exit {
			fmt.Printf("\n%s[Action]%s Exit\n", Green, Reset)
		}
	}

	// 打印错误
	if event.Err != nil {
		fmt.Printf("\n%s[Error]%s %v\n", Red, Reset, event.Err)
	}

	fmt.Println()
}

// Summary 打印汇总信息
func Summary(agentName string, totalEvents int, success bool) {
	PrintSeparator()
	if success {
		fmt.Printf("%s%s[Summary]%s Agent: %s, Total Events: %d, Status: Success%s\n",
			Bold, Green, Reset, agentName, totalEvents, Reset)
	} else {
		fmt.Printf("%s%s[Summary]%s Agent: %s, Total Events: %d, Status: Failed%s\n",
			Bold, Red, Reset, agentName, totalEvents, Reset)
	}
	PrintSeparator()
}

// ToolCall 打印工具调用信息
func ToolCall(toolName string, args string) {
	fmt.Printf("%s[Tool Call]%s %s(%s...)\n", Purple, Reset, toolName, truncate(args, 100))
}

// ToolResponse 打印工具响应
func ToolResponse(toolName string, response string) {
	if len(response) > 500 {
		response = response[:500] + "..."
	}
	fmt.Printf("%s[Tool Response]%s %s: %s\n", Yellow, Reset, toolName, response)
}

// NodeInfo 打印节点信息
func NodeInfo(nodePath string, nodeType string) {
	fmt.Printf("%s[Node]%s Path: %s, Type: %s\n", Blue, Reset, nodePath, nodeType)
}

// StreamChunk 打印流式输出块
func StreamChunk(role schema.RoleType, content string) {
	if role == schema.Tool {
		fmt.Printf("%s", content)
	} else {
		fmt.Printf("%s", content)
	}
}
