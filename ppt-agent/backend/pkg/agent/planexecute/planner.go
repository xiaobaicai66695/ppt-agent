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

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/agent/agents"
	"github.com/cloudwego/ppt-agent/pkg/agent/command"
	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/generic"
	"github.com/cloudwego/ppt-agent/pkg/prompts"
)

var plannerPromptTemplate prompt.ChatTemplate

func init() {
	systemTmpl, err := prompts.RenderPlanExecute("planner_system", &prompts.TemplateData{})
	if err != nil {
		panic("failed to load planner system template: " + err.Error())
	}

	userTmpl, err := prompts.RenderPlanExecute("planner_user", &prompts.TemplateData{})
	if err != nil {
		panic("failed to load planner user template: " + err.Error())
	}

	plannerPromptTemplate = prompt.FromMessages(schema.Jinja2,
		schema.SystemMessage(systemTmpl),
		schema.UserMessage(userTmpl),
	)
}

func NewPlanner(ctx context.Context, operator *command.LocalOperator, skillsContent string) (adk.Agent, error) {
	cm, err := agentutils.NewFallbackToolCallingChatModel(ctx,
		agentutils.WithMaxTokens(4096),
		agentutils.WithTemperature(0),
		agentutils.WithTopP(0),
		agentutils.WithDisableThinking(true),
	)
	if err != nil {
		return nil, err
	}

	a, err := planexecute.NewPlanner(ctx, &planexecute.PlannerConfig{
		ChatModelWithFormattedOutput: cm,
		GenInputFn:                  newPlannerInputGen(plannerPromptTemplate, skillsContent),
		NewPlan: func(ctx context.Context) planexecute.Plan {
			return &generic.Plan{}
		},
	})
	if err != nil {
		return nil, err
	}

	return agents.NewWrite2PlanMDWrapper(a, operator), nil
}

func newPlannerInputGen(plannerPrompt prompt.ChatTemplate, skillsContent string) planexecute.GenPlannerModelInputFn {
	return func(ctx context.Context, userInput []adk.Message) ([]adk.Message, error) {
		msgs, err := plannerPrompt.Format(ctx, map[string]any{
			"user_query":   agentutils.FormatInput(userInput),
			"current_time": agentutils.GetCurrentTime(),
			"skills":       skillsContent,
		})
		if err != nil {
			return nil, err
		}

		return msgs, nil
	}
}
