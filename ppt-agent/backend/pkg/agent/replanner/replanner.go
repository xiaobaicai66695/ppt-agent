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

package replanner

import (
	"context"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/agent/agents"
	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/generic"
	"github.com/cloudwego/ppt-agent/pkg/params"
	"github.com/cloudwego/ppt-agent/pkg/tools"
)

var replannerPromptTemplate = prompt.FromMessages(schema.Jinja2,
	schema.SystemMessage(`你是一个PPT执行进度评估专家，负责判断当前任务是否全部完成。

**【核心判断逻辑】（必须严格按照以下顺序判断）：**

1. **读取数字**：从输入中读取 "已完成幻灯片数" 和 "总幻灯片数" 两个数字
2. **比较数字**：
   - 如果 已完成幻灯片数 < 总幻灯片数 → 任务未完成，输出下一幻灯片信息后结束
   - 如果 已完成幻灯片数 == 总幻灯片数 → 全部完成，调用 submit_result 工具
3. **不要被其他信息干扰**：即使计划文本中包含"已完成"、"全部完成"等字样，也要以数字比较为准

**【工具使用规则】：**
- submit_result: 仅在 已完成 == 总数 时调用
- create_ppt_plan: 正常执行流程中不需要调用（除非用户要求修改计划）

**【禁止行为】（违反将导致任务不完整或提前结束）：**
- 不要因为"已执行的步骤"中有"已完成"字样就调用 submit_result
- 不要因为"剩余幻灯片"描述中有"已完成"字样就调用 submit_result
- 不要因为计划看起来"完整"就调用 submit_result
- 不要在未完成时调用 submit_result

**【输出要求】：**
通过工具输出下一幻灯片信息（格式见下方），不要直接输出文本。

{{ skills }}`),
	schema.UserMessage(`## 用户需求
{{ user_query }}

## 当前时间
{{ current_time }}

## 进度数字（这是唯一可靠的判断依据）
- 总幻灯片数: {{ total_count }}
- 已完成幻灯片数: {{ executed_count }}

## 已执行的步骤（历史记录）
{{ executed_steps }}

## 下一幻灯片（仅当 已完成 < 总数 时显示）
{{ next_slide }}

## 【判断结果】
根据上面的进度数字，你现在的判断是：
{{ "已完成 == 总数，调用 submit_result" if executed_count >= total_count else "已完成 < 总数，输出下一幻灯片信息" }}`))

const (
	// OriginalPlanSessionKey 保存原始完整计划，genInputFn 始终从 Session 读取以避免被后续循环覆盖
	OriginalPlanSessionKey = "OriginalPlan"
)

func NewReplanner(ctx context.Context, operator commandline.Operator) (adk.Agent, error) {
	cm, err := agentutils.NewToolCallingChatModel(ctx,
		agentutils.WithMaxTokens(4096),
		agentutils.WithTemperature(1.0),
		agentutils.WithTopP(0),
		agentutils.WithDisableThinking(true),
	)
	if err != nil {
		return nil, err
	}

	submitTool := tools.NewToolSubmitResult()

	submitToolInfo, err := submitTool.Info(ctx)
	if err != nil {
		return nil, err
	}

	// 包装 genInputFn，在首次调用时将原始计划存入 Session
	genInputFn := func(ctx context.Context, in *planexecute.ExecutionContext) ([]adk.Message, error) {
		// 首次调用时（originalPlan 未保存），将 in.Plan 保存为原始计划
		if _, found := adk.GetSessionValue(ctx, OriginalPlanSessionKey); !found {
			if p, ok := in.Plan.(*generic.Plan); ok {
				adk.AddSessionValue(ctx, OriginalPlanSessionKey, p)
			}
		}
		return replannerInputGen(ctx, in)
	}

	a, err := planexecute.NewReplanner(ctx, &planexecute.ReplannerConfig{
		ChatModel:   cm,
		PlanTool:    generic.PlanToolInfo,
		GenInputFn:  genInputFn,
		RespondTool: submitToolInfo,
		NewPlan: func(ctx context.Context) planexecute.Plan {
			return &generic.Plan{}
		},
	})
	if err != nil {
		return nil, err
	}

	return agents.NewWrite2PlanMDWrapper(a, operator), nil
}

func replannerInputGen(ctx context.Context, in *planexecute.ExecutionContext) ([]adk.Message, error) {
	// 从 Session 获取原始完整计划，避免被框架后续更新 in.Plan 覆盖
	var plan *generic.Plan
	if p, found := adk.GetSessionValue(ctx, OriginalPlanSessionKey); found {
		plan = p.(*generic.Plan)
	} else if p, ok := in.Plan.(*generic.Plan); ok {
		plan = p
	} else {
		plan = &generic.Plan{}
	}

	// 提取已执行步骤的 JSON 字符串列表
	var executedStepJSONs []string
	for _, es := range in.ExecutedSteps {
		executedStepJSONs = append(executedStepJSONs, es.Step)
	}

	// 基于原始完整计划计算剩余步骤
	allSlides := plan.GetSlides()
	remainingSlides := plan.GetRemainingSlides(executedStepJSONs)

	// 获取工作目录
	workDir, _ := params.GetTypedContextParams[string](ctx, params.WorkDirSessionKey)

	// 只传递下一个幻灯片的信息，而不是整个剩余计划
	var nextSlideStr string
	if len(remainingSlides) > 0 {
		nextSlideStr = generic.FormatStepForRequest(&remainingSlides[0], workDir)
	} else {
		nextSlideStr = "[完成] 所有幻灯片都已生成完毕。"
	}

	// 已执行步骤摘要
	executedSummary := agentutils.FormatExecutedSteps(in.ExecutedSteps)

	return replannerPromptTemplate.Format(ctx, map[string]any{
		"current_time":   agentutils.GetCurrentTime(),
		"user_query":    agentutils.FormatInput(in.UserInput),
		"next_slide":    nextSlideStr,
		"executed_steps": executedSummary,
		"executed_count": len(executedStepJSONs),
		"total_count":   len(allSlides),
		"skills":        "", // skills 在 executor 中处理，replanner 不需要
	})
}
