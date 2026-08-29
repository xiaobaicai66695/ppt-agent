package main

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"time"

	agentcommand "github.com/cloudwego/ppt-agent/pkg/agent/command"
	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
	"github.com/cloudwego/ppt-agent/pkg/web"
)

func executeAgent(ctx context.Context, suite string, c benchCase, opt options, caseDir string) agentOutput {
	started := time.Now()
	out := agentOutput{CaseID: c.ID, Suite: suite, StartedAt: started.Format(time.RFC3339)}
	var input caseInput
	if err := json.Unmarshal(c.Input, &input); err != nil {
		out.Error = "invalid case input: " + err.Error()
		return out
	}

	switch suite {
	case "router":
		runRouterCase(ctx, &out, input)
	case "planner":
		runPlannerCase(ctx, &out, c, input, opt, caseDir)
	case "reviewer":
		runReviewerCase(ctx, &out, c, input, opt, caseDir)
	case "fixer":
		runFixerCase(ctx, &out, c, input, opt, caseDir)
	}
	out.DurationMS = time.Since(started).Milliseconds()
	return out
}

func runRouterCase(ctx context.Context, out *agentOutput, input caseInput) {
	message := firstNonEmpty(input.UserMessage, input.UserRequest)
	if len(input.ConversationContext) > 0 {
		out.Output = web.ClassifyTaskMessageForBenchmark(ctx, message, "benchmark-conversation-task", strings.Join(input.ConversationContext, "\n"), benchmarkAPIKey())
		return
	}
	if input.HasExistingTask {
		out.Output = web.ClassifyContinueIntentForBenchmark(ctx, message, input.TasksSummary, benchmarkAPIKey())
		return
	}
	out.Output = web.ClassifyCreateRequestForBenchmark(ctx, message, input.HasOutline, benchmarkAPIKey())
}

func runPlannerCase(ctx context.Context, out *agentOutput, c benchCase, input caseInput, opt options, caseDir string) {
	query := plannerQuery(input)
	events := []deck.AgentEvent{}
	cfg := taskConfig(caseDir, c.ID, query, opt)
	cfg.DisableImageSearch = true
	agent, err := deck.NewPPTPlannerAgent(ctx, cfg)
	if err != nil {
		out.Error = err.Error()
		return
	}
	manifest, result, err := deck.RunPPTPlannerDraftWithCallback(ctx, agent, cfg, query, func(e deck.AgentEvent) { events = append(events, e) })
	out.Events = events
	out.Error = firstEventError(events)
	out.Output = manifest
	if manifest != nil {
		out.DeterministicReview = deck.ReviewTasksManifest(manifest, "planner_draft", 1)
	}
	if result != nil && out.DurationMS == 0 {
		out.DurationMS = result.Duration.Milliseconds()
	}
	if err != nil {
		out.Error = err.Error()
	}
}

func runReviewerCase(ctx context.Context, out *agentOutput, c benchCase, input caseInput, opt options, caseDir string) {
	if input.DraftTasks == nil {
		out.Error = "reviewer case missing input.draft_tasks"
		return
	}
	if err := deck.WriteTasksDraftManifest(caseDir, input.DraftTasks); err != nil {
		out.Error = err.Error()
		return
	}
	before := cloneManifest(input.DraftTasks)
	report := deck.ReviewTasksManifest(input.DraftTasks, "case_draft", 1)
	if len(input.ReviewIssues) > 0 {
		report.Issues = input.ReviewIssues
		report.IssueCount = len(input.ReviewIssues)
		report.Passed = false
		report.Summary = "benchmark case supplied review issues"
	}
	inputText, allowed, err := deck.BuildPlanReviewRevisionInput(caseDir, 1, report)
	if err != nil {
		out.Error = err.Error()
		return
	}
	events := []deck.AgentEvent{}
	cfg := taskConfig(caseDir, c.ID, firstNonEmpty(input.UserRequest, input.UserMessage), opt)
	agent, err := deck.NewTaskPlanReviewerAgent(ctx, cfg, allowed)
	if err != nil {
		out.Error = err.Error()
		return
	}
	if err := deck.RunTaskPlanReviewerWithCallback(ctx, agent, inputText, func(e deck.AgentEvent) { events = append(events, e) }); err != nil {
		out.Error = err.Error()
	}
	if out.Error == "" {
		out.Error = firstEventError(events)
	}
	after, err := deck.ReadTasksDraftManifest(caseDir)
	if err != nil && out.Error == "" {
		out.Error = err.Error()
	}
	out.Before = before
	out.After = after
	out.Events = events
	if after != nil {
		out.DeterministicReview = deck.ReviewTasksManifest(after, "reviewer_after", 2)
	}
}

func runFixerCase(ctx context.Context, out *agentOutput, c benchCase, input caseInput, opt options, caseDir string) {
	if input.BaseTasks == nil {
		out.Error = "fixer case missing input.base_tasks"
		return
	}
	if err := deck.WriteTasksManifest(caseDir, input.BaseTasks); err != nil {
		out.Error = err.Error()
		return
	}
	allowed := input.AllowedPageIndexes
	if len(allowed) == 0 {
		allowed = inferAllowedPages(firstNonEmpty(input.UserRequest, input.UserMessage))
	}
	events := []deck.AgentEvent{}
	cfg := taskConfig(caseDir, c.ID, firstNonEmpty(input.UserRequest, input.UserMessage), opt)
	agent, err := deck.NewPPTFixerAgent(ctx, cfg, allowed)
	if err != nil {
		out.Error = err.Error()
		return
	}
	before := cloneManifest(input.BaseTasks)
	if err := deck.RunPPTFixerWithCallback(ctx, agent, firstNonEmpty(input.UserRequest, input.UserMessage), func(e deck.AgentEvent) { events = append(events, e) }); err != nil {
		out.Error = err.Error()
	}
	if out.Error == "" {
		out.Error = firstEventError(events)
	}
	after, err := deck.ReadTasksManifest(caseDir)
	if err != nil && out.Error == "" {
		out.Error = err.Error()
	}
	out.Before = before
	out.After = after
	out.Events = events
}

func taskConfig(workDir, caseID, query string, opt options) *deck.PPTTaskConfig {
	operator := &agentcommand.LocalOperator{}
	ctx := operator.SetWorkDir(context.Background(), workDir)
	_ = ctx
	return &deck.PPTTaskConfig{
		WorkDir:       workDir,
		TaskID:        "bench-" + safePathName(caseID),
		Query:         query,
		Concurrency:   1,
		Operator:      operator,
		SkillsDir:     filepath.Join(projectRoot(), "skills"),
		ModelAPIKey:   benchmarkAPIKey(),
		ModelProvider: benchmarkModelProvider,
	}
}

func plannerQuery(input caseInput) string {
	parts := []string{firstNonEmpty(input.UserRequest, input.UserMessage)}
	if len(input.SourceMaterials) > 0 {
		data, _ := json.MarshalIndent(input.SourceMaterials, "", "  ")
		parts = append(parts, "\n参考素材：\n"+string(data))
	}
	if len(input.Requirements) > 0 {
		parts = append(parts, "\n硬性要求：\n- "+strings.Join(input.Requirements, "\n- "))
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}
