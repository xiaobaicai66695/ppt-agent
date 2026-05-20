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

package planexecute

import (
	"context"
	"sync/atomic"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/generic"
	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/params"
	"github.com/cloudwego/ppt-agent/pkg/prompts"
	"github.com/cloudwego/ppt-agent/pkg/tools"
)

var qaModelFn = func(ctx context.Context) (model.ToolCallingChatModel, error) {
	return agentutils.NewFallbackToolCallingChatModel(ctx,
		agentutils.WithMaxTokens(8192),
		agentutils.WithTemperature(0),
		agentutils.WithTopP(0),
	)
}

var executorRunCounter int32

var executorPromptTemplate prompt.ChatTemplate

func init() {
	systemTmpl, err := prompts.RenderPlanExecute("executor_system", &prompts.TemplateData{})
	if err != nil {
		panic("failed to load executor system template: " + err.Error())
	}

	userTmpl, err := prompts.RenderPlanExecute("executor_user", &prompts.TemplateData{})
	if err != nil {
		panic("failed to load executor user template: " + err.Error())
	}

	executorPromptTemplate = prompt.FromMessages(schema.Jinja2,
		schema.SystemMessage(systemTmpl),
		schema.UserMessage(userTmpl),
	)
}

func NewExecutor(ctx context.Context, operator commandline.Operator, skillsContent string) (adk.Agent, error) {
	cm, err := agentutils.NewFallbackToolCallingChatModel(ctx,
		agentutils.WithMaxTokens(4096),
		agentutils.WithTemperature(0),
		agentutils.WithTopP(0),
	)
	if err != nil {
		return nil, err
	}

	searchTool := &tools.InvokableSearchApprovalTool{InvokableTool: tools.NewSearchTool()}
	pythonTool := tools.NewPythonRunnerTool(operator)
	editFileTool := tools.NewEditFileTool(operator)
	readFileTool := tools.NewReadFileTool(operator)
	bashTool := tools.NewBashTool(operator)
	checkpointTool := tools.NewCheckpointTool(operator)
	singleQATool := tools.NewSingleQATool(operator, qaModelFn)

	a, err := planexecute.NewExecutor(ctx, &planexecute.ExecutorConfig{
		Model: cm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{
					pythonTool,
					editFileTool,
					readFileTool,
					bashTool,
					searchTool,
					checkpointTool,
					singleQATool,
				},
			},
		},
		MaxIterations: 20,
		GenInputFn: func(ctx context.Context, in *planexecute.ExecutionContext) ([]adk.Message, error) {
			workDir, _ := params.GetTypedContextParams[string](ctx, params.WorkDirSessionKey)

			plan, ok := agentutils.GetSessionValue[*generic.Plan](ctx, "OriginalPlan")
			if !ok {
				plan, ok = in.Plan.(*generic.Plan)
				if !ok {
					plan = &generic.Plan{}
				}
				adk.AddSessionValue(ctx, "OriginalPlan", plan)
			}

			executorCtx := agentutils.BuildExecutorContext(ctx, plan, workDir, in.ExecutedSteps)
			executorContextStr := agentutils.FormatExecutorContext(executorCtx)

			var stepStr string
			if executorCtx.IsBatchMode && executorCtx.NextSlide != nil {
				stepStr = generic.FormatBatchStepsForRequest([]generic.Step{*executorCtx.NextSlide}, workDir)
			} else if executorCtx.NextSlide != nil {
				stepStr = generic.FormatStepForRequest(executorCtx.NextSlide, workDir)
			} else {
				stepStr = "[完成] 所有幻灯片都已生成完毕。"
			}

			promptValues := map[string]any{
				"input":            agentutils.FormatInput(in.UserInput),
				"executor_context": executorContextStr,
				"step":             stepStr,
				"skills":           skillsContent,
			}

			msgs, err := executorPromptTemplate.Format(ctx, promptValues)
			if err != nil {
				return nil, err
			}

			runCount := atomic.AddInt32(&executorRunCounter, 1)
			var totalLen int
			for _, msg := range msgs {
				totalLen += len(msg.Content)
			}
			logger.Info("executor_ctx_stats",
				"run", runCount,
				"total_chars", totalLen,
				"user_input_chars", len(promptValues["input"].(string)),
				"skills_chars", len(promptValues["skills"].(string)),
				"context_chars", len(promptValues["executor_context"].(string)),
				"step_chars", len(promptValues["step"].(string)))

			return msgs, nil
		},
	})
	if err != nil {
		return nil, err
	}

	return a, nil
}
