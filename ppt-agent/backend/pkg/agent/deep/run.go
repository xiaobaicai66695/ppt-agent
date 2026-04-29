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
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/human"
)

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
		StartTime:   startTime,
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
	start, err := StartPPTTaskDeepAgent(ctx, agent, cfg, userQuery)
	if err != nil {
		return nil, err
	}

	iter := start.Iter

	var (
		lastMessage       adk.Message
		lastMessageStream *schema.StreamReader[adk.Message]
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
				printStreamingMessage(lastMessageStream)
			} else {
				lastMessage = event.Output.MessageOutput.Message
				lastMessageStream = nil
				if lastMessage != nil && lastMessage.Content != "" {
					fmt.Printf("\nanswer: %s\n", lastMessage.Content)
				}
			}
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			if m := event.Output.MessageOutput.Message; m != nil {
				for _, tc := range m.ToolCalls {
					fmt.Printf("\ntool name: %s", tc.Function.Name)
					fmt.Printf("\narguments: %s", tc.Function.Arguments)
				}
			}
		}

		if event.Err != nil {
			fmt.Printf("\nerror: %v\n", event.Err)
		}

		if event.Action != nil {
			if event.Action.Exit {
				fmt.Printf("\naction: exit\n")
			}
		}
	}

	if lastMessageStream != nil {
		lastMessageStream.Close()
	}

	var lastMsg string
	if lastMessage != nil {
		lastMsg = lastMessage.Content
	} else if lastMessageStream != nil {
		if msg, err := schema.ConcatMessageStream(lastMessageStream); err == nil {
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

func printStreamingMessage(stream *schema.StreamReader[adk.Message]) {
	if stream == nil {
		return
	}

	answerPrinted := false
	toolPrinted := false

	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("error: %v", err)
			return
		}

		if chunk.Content == "" {
			continue
		}

		if chunk.Role == schema.Tool {
			if !toolPrinted {
				toolPrinted = true
				fmt.Printf("\ntool response: ")
			}
			fmt.Print(chunk.Content)
		} else {
			if !answerPrinted {
				answerPrinted = true
				fmt.Printf("\nanswer: ")
			}
			fmt.Print(chunk.Content)
		}
	}
	fmt.Println()
}
