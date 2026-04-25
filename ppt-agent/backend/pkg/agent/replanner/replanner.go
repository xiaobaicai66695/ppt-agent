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
	"encoding/json"

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
	schema.SystemMessage(`你是一个PPT执行进度评估专家，负责判断当前任务是否全部完成，并决定下一步操作。

**【核心判断规则】必须严格按照以下规则判断：**
- 查看 "已完成幻灯片数" 和 "总幻灯片数"
- 如果 已完成幻灯片数 < 总幻灯片数 → 调用 create_ppt_plan 工具
- 如果 已完成幻灯片数 == 总幻灯片数 → 调用 submit_result 工具
- 绝对不能在 未完成时调用 submit_result

**【示例】：**
- 总数=13，已完成=4 → 4 < 13，未完成，调用 create_ppt_plan
- 总数=13，已完成=13 → 13 == 13，全部完成，调用 submit_result

**【禁止行为】：**
- 不要因为看到计划或描述中有"已完成"字样就调用 submit_result
- 不要因为 Executor 说"已完成"就误判
- 不要在 已完成 < 总数 时调用 submit_result
- 不要直接输出文本，必须通过工具输出`), schema.UserMessage(`## 用户需求
{{ user_query }}

## 进度数字
- 总幻灯片数: {{ total_count }}
- 已完成幻灯片数: {{ executed_count }}

## 剩余幻灯片计划
{{ remaining_plan }}

## 下一幻灯片（当前需要执行的）
{{ next_slide }}

## 判断
已完成 {{ executed_count }} 页，总共 {{ total_count }} 页。
{{ "所有幻灯片都已完成，调用 submit_result" if executed_count >= total_count else "还有幻灯片未完成，调用 create_ppt_plan" }}`))

const (
	// OriginalPlanSessionKey 保存原始完整计划，genInputFn 始终从 Session 读取以避免被后续循环覆盖
	OriginalPlanSessionKey = "OriginalPlan"
)

func NewReplanner(ctx context.Context, operator commandline.Operator) (adk.Agent, error) {
	cm, err := agentutils.NewFallbackToolCallingChatModel(ctx,
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
	// 从 Session 获取原始完整计划
	var plan *generic.Plan
	if p, found := adk.GetSessionValue(ctx, OriginalPlanSessionKey); found {
		plan = p.(*generic.Plan)
	} else if p, ok := in.Plan.(*generic.Plan); ok {
		plan = p
	} else {
		plan = &generic.Plan{}
	}

	// 获取工作目录
	workDir, _ := params.GetTypedContextParams[string](ctx, params.WorkDirSessionKey)

	// 获取总幻灯片数
	allSlides := plan.GetSlides()
	totalCount := len(allSlides)

	// 优先级：checkpoint > filesystem > framework ExecutedSteps
	// checkpoint 是 executor 显式记录的状态，最可靠
	checkpointCount, _ := generic.GetCompletedCountFromCheckpoint(workDir)

	// 如果 checkpoint 有记录，直接使用
	if checkpointCount > 0 {
		completedCount := checkpointCount
		// 构建已完成索引集合（从 checkpoint 文件）
		checkpoint, _ := generic.LoadCheckpoint(workDir)
		trueDoneIndexes := make(map[int]bool)
		if checkpoint != nil {
			for _, idx := range checkpoint.CompletedSlides {
				trueDoneIndexes[idx] = true
			}
		}

		// 计算剩余幻灯片
		var remainingSlides []generic.Step
		for _, slide := range allSlides {
			if !trueDoneIndexes[slide.Index] {
				remainingSlides = append(remainingSlides, slide)
			}
		}

		// 构建剩余计划 JSON
		remainingPlan := &generic.Plan{
			Title:  plan.Title,
			Theme:  plan.Theme,
			Slides: remainingSlides,
			Steps:  remainingSlides,
		}
		remainingPlanStr, _ := remainingPlan.MarshalJSON()

		// 下一幻灯片信息
		var nextSlideStr string
		if len(remainingSlides) > 0 {
			nextSlideStr = generic.FormatStepForRequest(&remainingSlides[0], workDir)
		} else {
			nextSlideStr = "[完成] 所有幻灯片都已生成完毕。"
		}

		return replannerPromptTemplate.Format(ctx, map[string]any{
			"current_time":    agentutils.GetCurrentTime(),
			"user_query":     agentutils.FormatInput(in.UserInput),
			"next_slide":     nextSlideStr,
			"executed_steps": "[从 checkpoint 读取进度]",
			"executed_count": completedCount,
			"total_count":    totalCount,
			"remaining_plan": string(remainingPlanStr),
		})
	}

	// 如果 checkpoint 没有记录，回退到原有的 filesystem + framework 逻辑
	var executedIndexes []int
	for _, es := range in.ExecutedSteps {
		var step generic.Step
		if err := json.Unmarshal([]byte(es.Step), &step); err == nil {
			executedIndexes = append(executedIndexes, step.Index)
		}
	}

	existingFiles := generic.GetExistingStepFiles(workDir)

	trueDoneIndexes := make(map[int]bool)
	for _, idx := range executedIndexes {
		trueDoneIndexes[idx] = true
	}
	for idx := range existingFiles {
		trueDoneIndexes[idx] = true
	}

	var remainingSlides []generic.Step
	for _, slide := range allSlides {
		if !trueDoneIndexes[slide.Index] {
			remainingSlides = append(remainingSlides, slide)
		}
	}

	executedSummary := agentutils.FormatExecutedSteps(in.ExecutedSteps)

	var nextSlideStr string
	if len(remainingSlides) > 0 {
		nextSlideStr = generic.FormatStepForRequest(&remainingSlides[0], workDir)
	} else {
		nextSlideStr = "[完成] 所有幻灯片都已生成完毕。"
	}

	remainingPlan := &generic.Plan{
		Title:  plan.Title,
		Theme:  plan.Theme,
		Slides: remainingSlides,
		Steps:  remainingSlides,
	}
	remainingPlanStr, err := remainingPlan.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return replannerPromptTemplate.Format(ctx, map[string]any{
		"current_time":    agentutils.GetCurrentTime(),
		"user_query":     agentutils.FormatInput(in.UserInput),
		"next_slide":     nextSlideStr,
		"executed_steps": executedSummary,
		"executed_count": len(trueDoneIndexes),
		"total_count":    totalCount,
		"remaining_plan": string(remainingPlanStr),
	})
}
