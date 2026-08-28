package deck

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// RunPPTPlannerDraftWithCallback runs only PPTPlanner and returns the first
// tasks.draft.json it produced. It intentionally skips TaskPlanReviewer and
// rendering so benchmarks can measure Planner first-draft quality directly.
func RunPPTPlannerDraftWithCallback(ctx context.Context, agent adk.Agent, cfg *PPTTaskConfig, userQuery string, onEvent AgentEventCallback) (*TasksManifest, *PPTTaskResult, error) {
	if onEvent == nil {
		onEvent = func(AgentEvent) {}
	}
	started := time.Now()
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: adkStreamingEnabled(),
	})
	iter := runner.Run(ctx, []adk.Message{schema.UserMessage(userQuery)})

	var answerBuf strings.Builder
	var lastMessage adk.Message
	for {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		event, ok, timedOut := nextWithTimeout(ctx, iter)
		if timedOut {
			return nil, nil, fmt.Errorf("LLM 流式输出超时（%s）", streamTimeout())
		}
		if !ok {
			break
		}
		if event.Output != nil && event.Output.MessageOutput != nil {
			if event.Output.MessageOutput.IsStreaming {
				cpStream := event.Output.MessageOutput.MessageStream.Copy(2)
				event.Output.MessageOutput.MessageStream = cpStream[0]
				msgStream := cpStream[1]
				processStreamingMessage(msgStream, onEvent, &answerBuf)
				msgStream.Close()
			} else if msg := event.Output.MessageOutput.Message; msg != nil {
				lastMessage = msg
				if isChunkEmittable(msg) && msg.Content != "" {
					if content := visibleMessageContent(msg.Content); content != "" {
						answerBuf.WriteString(content)
						onEvent(AgentEvent{Type: AgentEventAnswer, Content: content})
					}
				}
				for _, tc := range msg.ToolCalls {
					onEvent(AgentEvent{Type: AgentEventToolCall, ToolName: tc.Function.Name, ToolArgs: tc.Function.Arguments})
				}
			}
		}
		if event.Err != nil {
			onEvent(AgentEvent{Type: AgentEventError, Error: event.Err.Error()})
			return nil, nil, event.Err
		}
	}

	message := answerBuf.String()
	if message == "" && lastMessage != nil {
		message = lastMessage.Content
	}
	manifest, err := ensurePlannerDraft(cfg, userQuery, message, onEvent)
	result := &PPTTaskResult{Message: message, Duration: time.Since(started)}
	if manifest != nil {
		result.TotalSlides = len(manifest.Tasks)
		result.DoneSlides = manifest.CompletedCount()
		for _, task := range manifest.Tasks {
			if task != nil && task.OutputFile != "" {
				result.Files = append(result.Files, filepath.Join(cfg.WorkDir, task.OutputFile))
			}
		}
	}
	return manifest, result, err
}

// RunTaskPlanReviewerWithCallback runs one scoped reviewer repair pass against
// tasks.draft.json. The caller owns deterministic pre/post review and scoring.
func RunTaskPlanReviewerWithCallback(ctx context.Context, agent adk.Agent, userInput string, onEvent AgentEventCallback) error {
	return runAgentWithCallback(ctx, agent, userInput, onEvent)
}

// BuildPlanReviewRevisionInput exposes the production reviewer slice builder
// for standalone benchmark cases.
func BuildPlanReviewRevisionInput(workDir string, round int, report *PlanReviewReport) (string, []int, error) {
	return buildPlanReviewRevisionInput(workDir, round, report)
}
