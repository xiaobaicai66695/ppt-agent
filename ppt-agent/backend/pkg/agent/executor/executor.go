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

package executor

import (
	"context"
	"fmt"
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
	"github.com/cloudwego/ppt-agent/pkg/params"
	"github.com/cloudwego/ppt-agent/pkg/tools"
)

// qaModelFn 创建用于 QA 视觉审查的多模态 LLM。
// 需要支持图片输入的视觉模型。
var qaModelFn = func(ctx context.Context) (model.ToolCallingChatModel, error) {
	return agentutils.NewFallbackToolCallingChatModel(ctx,
		agentutils.WithMaxTokens(8192),
		agentutils.WithTemperature(0),
		agentutils.WithTopP(0),
	)
}

// executorRunCounter 记录 executor 被调用的次数（用于调试上下文增长）
var executorRunCounter int32

// executorSystemPrompt 是 Executor 的系统提示词（拆分为多段以避免反引号问题）
var executorSystemPromptPart1 = `你是一个PPT执行代理，负责根据 Planner 的计划生成幻灯片。每生成一页后立即进行视觉 QA 检查。

**【执行规则】：**
- Planner 已为某些页面标记了多页建议（multi_page_hint）。你需要根据实际内容量自主决策最终分页数量：
  - 如果实际内容确实需要分多页 → 生成多页，分别调用 update_progress 记录每页
  - 如果单页可以容纳 Planner 的内容描述 → 保持单页，不要强行拆页
  - **最终分页决策权在你**，Planner 的 multi_page_hint 仅作为参考提示
- 当收到批量任务（多页打包）时，统一在 python3 调用中一次性生成所有页

**【内容质量】：**
- 内容空洞或缺少具体数据时，**可调用 search 工具搜索真实信息**
- 几何体内嵌文字必须先估算宽度，确保不超出边界

**【搜索规范】（必须严格遵守）：**
- 每次搜索**只传入一个核心关键词**，不要拼接多个关键词
- 关键词要求：简洁、精准、长约 2-5 个词
- 如果需要搜索多个不同主题，**必须分多次调用** search 工具
- 示例：
  - 正确：{"query": "深度学习发展历程"}
  - 错误：{"query": "深度学习 发展历程 里程碑"}

**【可用工具】：**
- python3: 生成或修复 PPT 文件（主要工具）
- update_progress: 每成功生成一页幻灯片后，必须调用此工具记录页码（如 {"slide_index": 1}），多页则多次调用
- edit_file, read_file, bash, search: 辅助工具，search 必要时使用
- single_qa_review: 每生成或修复一页幻灯片后，必须调用此工具对该页进行视觉 QA 检查，参数为 slide_index（页码）。**每张幻灯片最多进行 2 次 QA 检查，达到次数后该页不再重新审查，直接进入下一页**

**【QA 限制规则】（必须严格遵守）：**
- 每张幻灯片最多调用 single_qa_review 2 次（包括修复后的复检）
- 首次 QA 失败后，可以根据问题修复一次
- 修复后若 QA 仍失败（或者无法定位 JSON 等），**不再继续修复，直接进入下一页**
- 严禁因 QA 失败而在同一页进入无限修复循环

**【文件命名规范】：**
- "页码_标题.pptx" 格式（如 1_标题页.pptx）

**【执行流程】：**

情况A - 生成新页面（当前任务不包含【修复】）：
1. search 搜索内容（如需要）
2. python3 生成 PPT 文件（可一次生成多页）
3. 对每一页调用 update_progress 记录页码
4. 对每一页调用 single_qa_review 进行视觉 QA 检查（传入 slide_index）
5. 在回复中清晰报告 QA 结果

情况B - 修复现有页面（当前任务包含【修复】标记）：
1. 直接执行 python3 修复代码（根据【修复】中的描述和代码）
2. **不需要调用 update_progress**（该页已完成，只是修复）
3. 调用 single_qa_review 重新检查该页，确认修复效果
4. 在回复中报告修复结果

情况C - 该页 QA 次数已达到 2 次：
- 不再修复或重新审查，直接标记该页完成并进入下一页

**【回复格式要求】：**
情况A完成后：
已完成：第 {N} 页 - {标题}
QA结果：has_issues={true/false}, has_high_issue={true/false}
{QA返回的summary内容}
接下来该做第 {M} 页：{标题}

情况B完成后：
已完成：第 {N} 页修复
QA结果：has_issues={true/false}, has_high_issue={true/false}
{QA返回的summary内容}
接下来该做第 {M} 页：{标题}

{{ skills }}`

var executorSystemPrompt = executorSystemPromptPart1

var executorUserPrompt = `## 用户需求
{{ input }}

## 执行状态
{{ executor_context }}

## 当前任务
{{ step }}

**执行流程**：
情况A - 生成新页面（任务不包含【修复】）：
1. search 搜索内容（如需要）
2. python3 生成 PPT 文件（可一次生成多页）
3. 对每一页调用 update_progress 记录页码
4. 对每一页调用 single_qa_review 进行视觉 QA 检查
5. 报告 QA 结果

情况B - 修复页面（任务包含【修复】标记）：
1. 直接执行 python3 修复代码
2. **不调用 update_progress**
3. 调用 single_qa_review 重新检查
4. 报告修复结果

情况C - 该页 QA 次数已达到 2 次：
- 不再修复或重新审查，直接标记该页完成并进入下一页`

var executorPrompt = prompt.FromMessages(schema.Jinja2,
	schema.SystemMessage(executorSystemPrompt),
	schema.UserMessage(executorUserPrompt))

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

			// 优先从 Session 读取原始完整计划（避免 framework plan 被 Replanner 覆盖后信息丢失）
			plan, ok := agentutils.GetSessionValue[*generic.Plan](ctx, "OriginalPlan")
			if !ok {
				plan, ok = in.Plan.(*generic.Plan)
				if !ok {
					plan = &generic.Plan{}
				}
				// 存入 Session，供后续调用复用
				adk.AddSessionValue(ctx, "OriginalPlan", plan)
			}

			executorCtx := agentutils.BuildExecutorContext(ctx, plan, workDir, in.ExecutedSteps)
			executorContextStr := agentutils.FormatExecutorContext(executorCtx)

			var stepStr string
			if executorCtx.IsBatchMode && len(executorCtx.NextBatch) > 0 {
				stepStr = generic.FormatBatchStepsForRequest(executorCtx.NextBatch, workDir)
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

			msgs, err := executorPrompt.Format(ctx, promptValues)
			if err != nil {
				return nil, err
			}

			runCount := atomic.AddInt32(&executorRunCounter, 1)
			var totalLen int
			for _, msg := range msgs {
				totalLen += len(msg.Content)
			}
			fmt.Printf("[Executor #%d] 上下文长度: total=%d chars | userInput=%d | skills=%d | executor_context=%d | step=%d\n",
				runCount, totalLen,
				len(promptValues["input"].(string)),
				len(promptValues["skills"].(string)),
				len(promptValues["executor_context"].(string)),
				len(promptValues["step"].(string)))

			return msgs, nil
		},
	})
	if err != nil {
		return nil, err
	}

	return a, nil
}
