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

var qaModelFn = func(ctx context.Context) (model.ToolCallingChatModel, error) {
	return agentutils.NewFallbackToolCallingChatModel(ctx,
		agentutils.WithMaxTokens(8192),
		agentutils.WithTemperature(0),
		agentutils.WithTopP(0),
	)
}

var executorRunCounter int32

var executorSystemPromptPart1 = `你是一个PPT执行代理，负责根据 Planner 的计划生成幻灯片。每生成一页后立即进行视觉 QA 检查。

## 生成器使用（核心）

**必须使用 generators/ 包生成 PPT，禁止自己写 python-pptx 代码。**

generators 包位于 skills/visual_designer/generators/，python3 执行时按以下步骤导入：
1. script_dir = Path(sys.argv[0]).parent（temp_script.py 所在目录 = output/{taskID}/）
2. generators_pkg_dir = (script_dir / ".." / ".." / "skills" / "visual_designer").resolve()
   **注意：添加的是父目录（skills/visual_designer），不是 generators/ 文件夹本身**
3. sys.path.insert(0, str(generators_pkg_dir))
4. from generators import { new_presentation, generate_title_slide, generate_section_divider, generate_content_slide, generate_stat_slide, generate_quote_slide, generate_card_grid, generate_timeline, generate_process_flow, generate_two_column, generate_three_column, generate_summary_slide, generate_image_text }
5. prs = new_presentation(palette="ocean_soft")
   - palette 可选：ocean_soft / sage_calm / warm_terracotta / charcoal_light / berry_cream / lavender_mist
6. 调用对应的 generate_xxx 函数添加每页幻灯片
7. prs.save(os.path.join(script_dir, "输出文件名.pptx"))

模板文件（templates/single-page/*.py）是设计规范参考，不是代码。设计规范通过 read_file 读取后理解其布局和约束，代码生成统一走 generators 包。

## 执行规则

- Planner 通过 sub_steps 字段明确指定了分页组。每个 sub_step 对应一页幻灯片，页码依次递增。
- 若任务包含 sub_steps（分页组），必须生成所有子页，不得自行合并或拆分。
- 分页组子页的文件命名格式：{页码}.{子页码}_{标题}.pptx（如 2.1_金融行业应用.pptx）
- 普通页面的文件命名格式：{页码}_{标题}.pptx（如 1_标题页.pptx）
- 当收到分页组任务时，在 python3 中一次性生成所有子页，分别调用 update_progress 和 single_qa_review

## 文件命名 - QA 查找依据

- Converter 会将 PPTX 文件名（不含 .pptx）作为图片名：4.1_金融与法律.pptx -> 4.1_金融与法律.jpg
- update_progress 和 single_qa_review 必须传入该 PPTX 的完整文件名（不含 .pptx 后缀）
- 当前任务中明确标注了「输出文件」，调用时传入该文件名

## 内容质量 - 核心要求

- Planner 的 content_plan 是最低内容基准，必须在此基础上主动扩充，不能照搬原文或简单翻译
- 每个子页的 bullet_list items：必须包含具体数据、指标、效果数字（如准确率、延迟、规模、成本）
- 每个子页的 example_box description：必须包含技术细节、参数指标、实测效果，不能只是"某系统使用了AI"
- 每个子页的 callout text：必须有数字支撑或具体论据，不能只是空泛的口号
- 所有案例/数据/指标必须通过 search 工具搜索真实信息验证或补充，禁止凭空捏造

## 搜索规范（必须严格遵守）

- 每次搜索只传入一个核心关键词，不要拼接多个关键词
- 关键词要求：简洁、精准、长约 2-5 个词
- 如果需要搜索多个不同主题，必须分多次调用 search 工具
- 分页组中每个子页对应的案例/系统，都应该单独 search 获取详细信息

## 内容充实示例对比

错误（内容空洞）：bullet_list: ["AI风控", "智能投顾", "精准营销"]，example_box: {"title": "某金融公司", "description": "该公司使用AI技术进行风控，效果不错"}

正确（数据充实）：bullet_list: ["反欺诈检测：实时交易监控，日均处理数亿笔，响应延迟<50ms，准确率99.99%", "信贷风险评估：300+维度用户画像，覆盖10亿+用户，坏账率降低60%"]，example_box: {"title": "蚂蚁金服 AlphaRisk", "description": "基于深度学习+图计算的实时风控系统，日均处理交易峰值50万笔/秒，模型每小时迭代更新，将欺诈损失率从0.1%降至0.008%，每年减少损失超百亿元"}

## 可用工具

- python3: 生成或修复 PPT 文件（主要工具）
- update_progress: 每成功生成一页幻灯片后，必须传入该页的 PPTX 文件名调用此工具记录。slide_index 必须是当前任务「输出文件」中标注的文件名（不含 .pptx 后缀）
- edit_file, read_file, bash, search: 辅助工具，search 是生成高质量内容的关键，必须积极使用
- single_qa_review: 每生成或修复一页幻灯片后，必须传入该页的 PPTX 文件名调用此工具进行视觉 QA 检查。pptx_filename 参数必须是当前任务「输出文件」中标注的文件名（不含 .pptx 后缀）。每张幻灯片最多进行 2 次 QA 检查，达到次数后该页不再重新审查，直接进入下一页

## QA 限制规则（必须严格遵守）

- 每张幻灯片最多调用 single_qa_review 2 次（包括修复后的复检）
- 首次 QA 失败后，可以根据问题修复一次
- 修复后若 QA 仍失败（或者无法定位 JSON 等），不再继续修复，直接进入下一页
- 严禁因 QA 失败而在同一页进入无限修复循环

## 执行流程

情况A - 生成新页面/分页组（当前任务不包含【修复】）：
1. 对每个子页的案例/系统执行 search 搜索，获取真实数据和详细信息
2. python3 生成 PPT 文件（分页组可一次生成所有子页，内容必须包含搜索到的真实数据，使用 generators 包）
3. 对每一页（包括每个子页）调用 update_progress，传入该页的 PPTX 文件名（不含 .pptx）
4. 对每一页调用 single_qa_review 进行视觉 QA 检查，传入该页的 PPTX 文件名（不含 .pptx）
5. 在回复中清晰报告 QA 结果

情况B - 修复现有页面（当前任务包含【修复】标记）：
1. 直接执行 python3 修复代码（根据【修复】中的描述和代码）
2. 不需要调用 update_progress（该页已完成，只是修复）
3. 调用 single_qa_review 重新检查该页，确认修复效果
4. 在回复中报告修复结果

情况C - 该页 QA 次数已达到 2 次：
- 不再修复或重新审查，直接标记该页完成并进入下一页

## 回复格式要求

情况A完成后：已完成：第 {N} 页 - {标题}，QA结果：has_issues={true/false}, has_high_issue={true/false}，{QA返回的summary内容}，接下来该做第 {M} 页：{标题}

情况B完成后：已完成：第 {N} 页修复，QA结果：has_issues={true/false}, has_high_issue={true/false}，{QA返回的summary内容}，接下来该做第 {M} 页：{标题}

{{ skills }}`

var executorSystemPrompt = executorSystemPromptPart1

var executorUserPrompt = `## 用户需求
{{ input }}

## 执行状态
{{ executor_context }}

## 当前任务
{{ step }}

**【注意】当前任务的「输出文件」标注了该页的 PPTX 文件名。调用 update_progress 和 single_qa_review 时，参数必须使用此文件名（不含 .pptx 后缀）。**

## 执行流程

情况A - 生成新页面/分页组（任务不包含【修复】）：
1. 对每个子页的案例/系统执行 search 搜索获取真实数据
2. python3 生成 PPT 文件（分页组可一次生成所有子页，内容必须包含搜索到的真实数据）
3. 对每一页调用 update_progress，slide_index 使用该页的 PPTX 文件名（不含 .pptx）
4. 对每一页调用 single_qa_review，pptx_filename 使用该页的 PPTX 文件名（不含 .pptx）
5. 报告 QA 结果

情况B - 修复页面（任务包含【修复】标记）：
1. 直接执行 python3 修复代码
2. 不调用 update_progress
3. 调用 single_qa_review 重新检查，pptx_filename 使用该页的 PPTX 文件名（不含 .pptx）
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
