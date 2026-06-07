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

package workflow

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
)

// AgentEventType mirrors deep.AgentEvent for workflow agents.
type AgentEventType = string

const (
	AgentEventAnswer   = "answer"
	AgentEventToolCall = "tool_call"
	AgentEventError    = "error"
	AgentEventProgress = "progress"
)

// AgentEvent mirrors the deep.AgentEvent structure.
type AgentEvent struct {
	Type        string `json:"type"`
	Content     string `json:"content,omitempty"`
	ToolName    string `json:"tool_name,omitempty"`
	ToolArgs    string `json:"tool_args,omitempty"`
	Error       string `json:"error,omitempty"`
	Phase       string `json:"phase,omitempty"`
	PhaseDetail string `json:"phase_detail,omitempty"`
}

// AgentEventCallback mirrors the deep.AgentEventCallback.
type AgentEventCallback func(event AgentEvent)

// PPTTaskResult mirrors the deep.PPTTaskResult.
type PPTTaskResult struct {
	Message     string
	TotalSlides int
	DoneSlides  int
	Files       []string
	Duration    time.Duration
}

// RunAgentWithCallback runs the workflow agent with event streaming.
// It uses the graph's Stream method directly instead of going through the adk.Runner,
// which allows the workflow agent to have its own event processing logic.
func RunAgentWithCallback(ctx context.Context, agent *Agent, workDir string, userQuery string, onEvent AgentEventCallback) (*PPTTaskResult, error) {
	startTime := time.Now()

	msgs := []*schema.Message{
		schema.UserMessage(userQuery),
	}

	stream, err := agent.runnable.Stream(ctx, msgs)
	if err != nil {
		onEvent(AgentEvent{Type: AgentEventError, Error: "启动工作流失败: " + err.Error()})
		return nil, err
	}
	defer stream.Close()

	var lastMsg string
	answerBuf := strings.Builder{}

	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			onEvent(AgentEvent{Type: AgentEventError, Error: err.Error()})
			break
		}

		if chunk == nil {
			continue
		}

		// Emit non-tool-call content as answer events.
		if isChunkEmittable(chunk) && chunk.Content != "" {
			answerBuf.WriteString(chunk.Content)
			onEvent(AgentEvent{
				Type:    AgentEventAnswer,
				Content: chunk.Content,
			})
		}

		// Emit tool calls.
		for _, tc := range chunk.ToolCalls {
			onEvent(AgentEvent{
				Type:     AgentEventToolCall,
				ToolName: tc.Function.Name,
				ToolArgs: tc.Function.Arguments,
			})
		}
	}

	lastMsg = answerBuf.String()

	manifest, err := ReadTasksManifest(workDir)
	result := &PPTTaskResult{
		Message:  lastMsg,
		Duration: time.Since(startTime),
	}

	if err == nil && manifest != nil {
		result.TotalSlides = len(manifest.Tasks)
		result.DoneSlides = manifest.CompletedCount()
		for _, t := range manifest.Tasks {
			if t.Status == StatusDone || t.Status == StatusQADone || t.Status == StatusFixed {
				result.Files = append(result.Files, filepath.Join(workDir, t.OutputFile))
			}
		}
	}

	return result, nil
}

// isChunkEmittable returns true if a message should be emitted as LLM text.
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
