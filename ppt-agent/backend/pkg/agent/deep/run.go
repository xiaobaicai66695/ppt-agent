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

package deep

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/human"
)

// AgentEventType constants for streaming events.
const (
	AgentEventAnswer   = "answer"
	AgentEventToolCall = "tool_call"
	AgentEventError    = "error"
)

// AgentEvent is a structured event emitted during agent execution.
type AgentEvent struct {
	Type     string `json:"type"`
	Content  string `json:"content,omitempty"`
	ToolName string `json:"tool_name,omitempty"`
	ToolArgs string `json:"tool_args,omitempty"`
	Error    string `json:"error,omitempty"`
}

// AgentEventCallback is called for each event during agent execution.
type AgentEventCallback func(event AgentEvent)

func StartPPTTaskDeepAgent(ctx context.Context, agent adk.Agent, cfg *PPTTaskConfig, userQuery string) (*PPTTaskStart, error) {
	startTime := time.Now()

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: true,
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

func RunPPTTaskDeepAgentWithHuman(ctx context.Context, agent adk.Agent, cfg *PPTTaskConfig,
	userQuery string, hm *human.Manager) (*PPTTaskResult, error) {

	start, err := StartPPTTaskDeepAgent(ctx, agent, cfg, userQuery)
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

	manifest, err := ReadTasksManifest(cfg.WorkDir)
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

func RunPPTTaskDeepAgent(ctx context.Context, agent adk.Agent, cfg *PPTTaskConfig, userQuery string) (*PPTTaskResult, error) {
	return runPPTTaskDeepAgentInternal(ctx, agent, cfg, userQuery, makePrintCallback())
}

// RunPPTTaskDeepAgentWithCallback runs the agent and calls onEvent for each streaming event.
// The callback is called synchronously — the caller should forward events or buffer them quickly.
func RunPPTTaskDeepAgentWithCallback(ctx context.Context, agent adk.Agent, cfg *PPTTaskConfig,
	userQuery string, onEvent AgentEventCallback) (*PPTTaskResult, error) {
	return runPPTTaskDeepAgentInternal(ctx, agent, cfg, userQuery, onEvent)
}

func runPPTTaskDeepAgentInternal(ctx context.Context, agent adk.Agent, cfg *PPTTaskConfig,
	userQuery string, onEvent AgentEventCallback) (*PPTTaskResult, error) {

	start, err := StartPPTTaskDeepAgent(ctx, agent, cfg, userQuery)
	if err != nil {
		return nil, err
	}

	iter := start.Iter

	var (
		lastMessage       adk.Message
		lastMessageStream *schema.StreamReader[adk.Message]
		answerBuf         strings.Builder
	)

	for {
		event, ok := iter.Next()
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
				if lastMessage != nil && lastMessage.Content != "" {
					onEvent(AgentEvent{
						Type:    AgentEventAnswer,
						Content: lastMessage.Content,
					})
				}
			}
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
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

	manifest, err := ReadTasksManifest(cfg.WorkDir)
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

func makePrintCallback() AgentEventCallback {
	return func(event AgentEvent) {
		switch event.Type {
		case AgentEventAnswer:
			fmt.Printf("\nanswer: %s\n", event.Content)
		case AgentEventToolCall:
			fmt.Printf("\ntool name: %s", event.ToolName)
			fmt.Printf("\narguments: %s", event.ToolArgs)
		case AgentEventError:
			fmt.Printf("\nerror: %s\n", event.Error)
		}
	}
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
		if chunk.Content == "" {
			continue
		}
		buf.WriteString(chunk.Content)
		onEvent(AgentEvent{
			Type:    AgentEventAnswer,
			Content: chunk.Content,
		})
	}
}

