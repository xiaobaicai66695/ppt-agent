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
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	agentintent "github.com/cloudwego/ppt-agent/pkg/agent/intent"
	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/style"
)

// AgentFactory creates a workflow agent from task config.
type AgentFactory func(ctx context.Context, cfg *Config) (adk.Agent, error)

// Config is the configuration for the workflow agent.
type Config struct {
	WorkDir           string
	TaskID            string
	Query             string
	Concurrency       int
	Operator          Operator
	QAModelFn         func(ctx context.Context) (model.ToolCallingChatModel, error)
	Skills            string
	SkillsDir         string
	CompressorTracker *agentutils.TokenTracker
	Outline           *TaskOutline
	StyleContext      string
	EnableQA          bool

	UserID          int
	IntentResult    *agentintent.ClassificationResult
	RoutingDecision *agentintent.RoutingDecision
	EnhancedProfile *style.EnhancedProfile
}

// Operator is the interface for file/bash operations.
type Operator interface {
	Execute(ctx context.Context, cmd string) (string, error)
}

// Agent represents the workflow-based PPT generation agent.
// It implements the adk.Agent interface using a compose.Graph under the hood.
type Agent struct {
	runnable compose.Runnable[[]*schema.Message, *schema.Message]
	name     string
}

// state holds the mutable workflow state shared across all graph nodes.
type workflowState struct {
	// Manifest is the central tasks manifest. It starts empty (Plan fills it).
	Manifest *TasksManifest

	// UserQuery is the original user request.
	UserQuery string

	// Skills content for prompts.
	Skills string

	// SkillsDir for constructing file paths.
	SkillsDir string

	// StyleContext for个性化.
	StyleContext string

	// EnhancedProfile for个性化.
	EnhancedProfile *style.EnhancedProfile

	// WorkDir for file operations.
	WorkDir string

	// EnableQA controls whether QA review is performed.
	EnableQA bool

	// Concurrency per batch.
	Concurrency int

	// QAModel factory (not stored in state to avoid serialization issues).
	QAModelFn func(ctx context.Context) (model.ToolCallingChatModel, error)

	// Operator for executing shell commands.
	Operator Operator

	// OutputMessages accumulates LLM text output for streaming.
	OutputMessages []string

	// mu protects mutable fields.
	mu sync.Mutex
}

// Node key constants.
const (
	nodeKeyPlan             = "plan"
	nodeKeyPlannerLLM       = "planner_llm"
	nodeKeyPlannerToOutline = "planner_to_outline"

	nodeKeyExecute        = "execute"
	nodeKeyExecutorLLM    = "executor_llm"
	nodeKeyExecutorTools  = "executor_tools"
	nodeKeyExecutorToList = "executor_to_list"

	nodeKeyQA       = "qa"
	nodeKeyQALLM    = "qa_llm"
	nodeKeyQATools  = "qa_tools"
	nodeKeyQAToList = "qa_to_list"

	nodeKeyFix       = "fix"
	nodeKeyFixLLM    = "fix_llm"
	nodeKeyFixTools  = "fix_tools"
	nodeKeyFixToList = "fix_to_list"

	nodeKeyToMessage = "to_message"
)

// defaultMaxStep is the default maximum workflow execution steps.
const defaultMaxStep = 200

// NewAgent creates a workflow-based PPT generation agent.
func NewAgent(ctx context.Context, cfg *Config) (*Agent, error) {
	return newAgent(ctx, cfg, "PPTWorkflowAgent")
}

// newAgent creates an agent with a custom name (for testing).
func newAgent(ctx context.Context, cfg *Config, name string) (*Agent, error) {
	enableQA := cfg.EnableQA
	if cfg.RoutingDecision != nil && cfg.RoutingDecision.SkipQA {
		enableQA = false
	}
	if enableQA {
		enableQA = isQAEnabled()
	}

	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = 5
	}

	state := &workflowState{
		UserQuery:       cfg.Query,
		Skills:          cfg.Skills,
		SkillsDir:       cfg.SkillsDir,
		StyleContext:    cfg.StyleContext,
		EnhancedProfile: cfg.EnhancedProfile,
		WorkDir:         cfg.WorkDir,
		EnableQA:        enableQA,
		Concurrency:     concurrency,
		QAModelFn:       cfg.QAModelFn,
		Operator:        cfg.Operator,
	}

	graph, err := buildWorkflowGraph(ctx, state)
	if err != nil {
		return nil, err
	}

	runnable, err := graph.Compile(ctx,
		compose.WithNodeTriggerMode(compose.AnyPredecessor),
		compose.WithMaxRunSteps(agentutils.EnvInt("WORKFLOW_MAX_STEPS", defaultMaxStep)),
		compose.WithGraphName(name),
	)
	if err != nil {
		return nil, err
	}

	return &Agent{
		runnable: runnable,
		name:     name,
	}, nil
}

// buildWorkflowGraph assembles the workflow graph:
//
//	Plan (outline) → Execute (generate slides) → QA (review) → Fix (repair)
//	                                                 ↑___________________________|
func buildWorkflowGraph(ctx context.Context, state *workflowState) (*compose.Graph[[]*schema.Message, *schema.Message], error) {
	g := compose.NewGraph[[]*schema.Message, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *workflowState { return state }),
	)

	// ── Plan path ──────────────────────────────────────────────────────────
	planLLM, err := buildPlanModel(ctx, state)
	if err != nil {
		return nil, fmt.Errorf("build plan model: %w", err)
	}

	_ = g.AddChatModelNode(nodeKeyPlannerLLM, planLLM,
		compose.WithStatePreHandler(planPreHandler(state)),
		compose.WithStatePostHandler(planPostHandler(state)),
		compose.WithNodeName(nodeKeyPlannerLLM),
	)

	_ = g.AddLambdaNode(nodeKeyPlannerToOutline, compose.ToList[*schema.Message]())

	_ = g.AddLambdaNode(nodeKeyPlan, compose.InvokableLambda(planNode(state)),
		compose.WithNodeName(nodeKeyPlan),
	)

	// ── Execute path ───────────────────────────────────────────────────────
	execLLM, execTools, err := buildExecutorModelAndTools(ctx, state)
	if err != nil {
		return nil, fmt.Errorf("build executor model and tools: %w", err)
	}

	_ = g.AddChatModelNode(nodeKeyExecutorLLM, execLLM,
		compose.WithStatePreHandler(execPreHandler(state)),
		compose.WithStatePostHandler(execPostHandler(state)),
		compose.WithNodeName(nodeKeyExecutorLLM),
	)

	_ = g.AddToolsNode(nodeKeyExecutorTools, execTools,
		compose.WithStatePreHandler(toolPreHandler(state)),
		compose.WithNodeName(nodeKeyExecutorTools),
	)

	_ = g.AddLambdaNode(nodeKeyExecutorToList, compose.ToList[*schema.Message]())

	_ = g.AddLambdaNode(nodeKeyExecute, compose.InvokableLambda(executeNode(state)),
		compose.WithNodeName(nodeKeyExecute),
	)

	// ── QA path ────────────────────────────────────────────────────────────
	qaLLM, qaTools, err := buildQAModelAndTools(ctx, state)
	if err != nil {
		return nil, fmt.Errorf("build QA model and tools: %w", err)
	}

	_ = g.AddChatModelNode(nodeKeyQALLM, qaLLM,
		compose.WithStatePreHandler(qaPreHandler(state)),
		compose.WithStatePostHandler(qaPostHandler(state)),
		compose.WithNodeName(nodeKeyQALLM),
	)

	_ = g.AddToolsNode(nodeKeyQATools, qaTools,
		compose.WithStatePreHandler(toolPreHandler(state)),
		compose.WithNodeName(nodeKeyQATools),
	)

	_ = g.AddLambdaNode(nodeKeyQAToList, compose.ToList[*schema.Message]())

	_ = g.AddLambdaNode(nodeKeyQA, compose.InvokableLambda(qaNode(state)),
		compose.WithNodeName(nodeKeyQA),
	)

	// ── Fix path ───────────────────────────────────────────────────────────
	fixLLM, fixTools, err := buildFixModelAndTools(ctx, state)
	if err != nil {
		return nil, fmt.Errorf("build fix model and tools: %w", err)
	}

	_ = g.AddChatModelNode(nodeKeyFixLLM, fixLLM,
		compose.WithStatePreHandler(fixPreHandler(state)),
		compose.WithStatePostHandler(fixPostHandler(state)),
		compose.WithNodeName(nodeKeyFixLLM),
	)

	_ = g.AddToolsNode(nodeKeyFixTools, fixTools,
		compose.WithStatePreHandler(toolPreHandler(state)),
		compose.WithNodeName(nodeKeyFixTools),
	)

	_ = g.AddLambdaNode(nodeKeyFixToList, compose.ToList[*schema.Message]())

	_ = g.AddLambdaNode(nodeKeyFix, compose.InvokableLambda(fixNode(state)),
		compose.WithNodeName(nodeKeyFix),
	)

	// ── Final output node ──────────────────────────────────────────────────
	_ = g.AddLambdaNode(nodeKeyToMessage, compose.InvokableLambda(toMessageNode(state)),
		compose.WithNodeName(nodeKeyToMessage),
	)

	// ── Edges ──────────────────────────────────────────────────────────────
	// Start → Plan
	_ = g.AddEdge(compose.START, nodeKeyPlannerLLM)

	// Planner → PlannerToList → PlanNode → Execute
	_ = g.AddEdge(nodeKeyPlannerLLM, nodeKeyPlannerToOutline)
	_ = g.AddEdge(nodeKeyPlannerToOutline, nodeKeyPlan)
	_ = g.AddEdge(nodeKeyPlan, nodeKeyExecutorLLM)

	// Executor: LLM → tools → toList → ExecuteNode
	// Branch on LLM output: if tool calls, go to tools; else go to toList
	_ = g.AddEdge(nodeKeyExecutorLLM, nodeKeyExecutorTools)
	_ = g.AddEdge(nodeKeyExecutorTools, nodeKeyExecutorToList)
	_ = g.AddEdge(nodeKeyExecutorToList, nodeKeyExecute)

	// Execute → QA or ToMessage
	if state.EnableQA {
		// Execute → QA
		_ = g.AddEdge(nodeKeyExecute, nodeKeyQALLM)

		// QA: LLM → tools → toList → QANode
		_ = g.AddEdge(nodeKeyQALLM, nodeKeyQATools)
		_ = g.AddEdge(nodeKeyQATools, nodeKeyQAToList)
		_ = g.AddEdge(nodeKeyQAToList, nodeKeyQA)

		// QA → Fix (issues found) or ToMessage (all pass)
		_ = g.AddBranch(nodeKeyQA, compose.NewGraphBranch(
			qaBranchCondition(state),
			map[string]bool{nodeKeyFixLLM: true, nodeKeyToMessage: true},
		))

		// Fix: LLM → tools → toList → FixNode → QA (loop back)
		_ = g.AddEdge(nodeKeyFixLLM, nodeKeyFixTools)
		_ = g.AddEdge(nodeKeyFixTools, nodeKeyFixToList)
		_ = g.AddEdge(nodeKeyFixToList, nodeKeyFix)
		_ = g.AddEdge(nodeKeyFix, nodeKeyQALLM) // Loop back to QA
	} else {
		_ = g.AddEdge(nodeKeyExecute, nodeKeyToMessage)
	}

	// ToMessage → END
	_ = g.AddEdge(nodeKeyToMessage, compose.END)

	return g, nil
}

// qaBranchCondition determines routing after QA:
//   - If there are QA issues → Fix
//   - If all pages pass → ToMessage (END)
func qaBranchCondition(state *workflowState) func(context.Context, *schema.Message) (string, error) {
	return func(ctx context.Context, msg *schema.Message) (string, error) {
		state.mu.Lock()
		defer state.mu.Unlock()

		if state.Manifest == nil {
			return nodeKeyToMessage, nil
		}

		needsFix := state.Manifest.NeedsFix()
		if len(needsFix) > 0 {
			logger.Info("qa_branch", "action", "fix", "issues", len(needsFix))
			return nodeKeyFixLLM, nil
		}

		logger.Info("qa_branch", "action", "all_pass")
		return nodeKeyToMessage, nil
	}
}

// toMessageNode converts accumulated output messages to a final schema.Message.
func toMessageNode(state *workflowState) func(ctx context.Context, in []*schema.Message) (*schema.Message, error) {
	return func(ctx context.Context, in []*schema.Message) (*schema.Message, error) {
		state.mu.Lock()
		defer state.mu.Unlock()

		var sb strings.Builder
		if state.Manifest != nil {
			sb.WriteString(fmt.Sprintf("PPT 生成完成，共 %d 页，已完成 %d 页。\n",
				len(state.Manifest.Tasks), state.Manifest.CompletedCount()))
			for _, t := range state.Manifest.Tasks {
				sb.WriteString(fmt.Sprintf("- 第%d页: %s (%s) → %s\n",
					t.PageIndex, t.Title, t.ContentType, t.Status))
			}
		}
		for _, m := range state.OutputMessages {
			sb.WriteString(m)
		}

		return schema.UserMessage(sb.String()), nil
	}
}

// isQAEnabled returns whether QA is enabled.
// Online QA is opt-in because it adds model cost and latency.
func isQAEnabled() bool {
	return os.Getenv("ENABLE_QA") == "true"
}
