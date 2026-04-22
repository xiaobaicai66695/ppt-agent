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
	"strings"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/generic"
	"github.com/cloudwego/ppt-agent/pkg/params"
	"github.com/cloudwego/ppt-agent/pkg/tools"
)

var executorPrompt = prompt.FromMessages(schema.Jinja2,
	schema.SystemMessage(`你是一个PPT执行代理，每次只生成一页幻灯片。

**【执行规则】：**
- 只处理 "当前需要执行的任务" 中指定的那一页
- 不要处理计划中的其他页面
- 不要回头执行已完成的页面
- 完成当前页后，必须调用 update_progress 工具记录进度，然后回复"接下来该做第 X 页：{标题}"

**【可用工具】：**
- python3: 生成 PPT 文件（主要工具）
- update_progress: 每成功生成一页幻灯片后，必须调用此工具记录页码（如 {"slide_index": 1}）
- edit_file, read_file, bash, search, search_image: 辅助工具

**【visual_designer 使用方式】：**
- 它是设计规范参考文档（已注入本 prompt）
- 参考其配色、字体、布局规范
- 不要调用 visual_designer 工具

**【文件命名规范】：**
- "页码_标题.pptx" 格式（如 1_标题页.pptx）

**【重要】完成流程**：
1. 调用 python3 生成 PPT 文件
2. 调用 update_progress 记录完成的页码
3. 回复"接下来该做第 X 页：{标题}"

{{ skills }}`), schema.UserMessage(`## 用户需求
{{ input }}

## 完整计划
{{ plan }}

## 已执行步骤
{{ executed_steps }}

## 当前任务（只处理这一页！）
{{ step }}

**完成流程**：
1. 调用 python3 生成 PPT 文件
2. 调用 update_progress 记录页码
3. 回复"接下来该做第 X 页：{标题}"`))

func getNextSlideFromDisk(plan *generic.Plan, workDir string) *generic.Step {
	if plan == nil {
		return nil
	}
	existingFiles := generic.GetExistingStepFiles(workDir)
	allSlides := plan.GetSlides()
	for i := range allSlides {
		slide := &allSlides[i]
		if _, exists := existingFiles[slide.Index]; !exists {
			return slide
		}
	}
	return nil
}

func NewExecutor(ctx context.Context, operator commandline.Operator, skillsContent string) (adk.Agent, error) {
	cm, err := agentutils.NewToolCallingChatModel(ctx,
		agentutils.WithMaxTokens(4096),
		agentutils.WithTemperature(0),
		agentutils.WithTopP(0),
	)
	if err != nil {
		return nil, err
	}

	// 直接配置所有工具，不再使用嵌套子代理
	searchTool := tools.NewSearchTool()
	imageSearchTool := tools.NewImageSearchTool()
	pythonTool := tools.NewPythonRunnerTool(operator)
	editFileTool := tools.NewEditFileTool(operator)
	readFileTool := tools.NewReadFileTool(operator)
	bashTool := tools.NewBashTool(operator)
	checkpointTool := tools.NewCheckpointTool(operator)

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
					imageSearchTool,
					checkpointTool,
				},
			},
		},
		MaxIterations: 20,
		GenInputFn: func(ctx context.Context, in *planexecute.ExecutionContext) ([]adk.Message, error) {
			planContent, err := in.Plan.MarshalJSON()
			if err != nil {
				return nil, err
			}

			// 获取工作目录
			workDir, _ := params.GetTypedContextParams[string](ctx, params.WorkDirSessionKey)

			// 获取 Plan 对象
			plan, ok := in.Plan.(*generic.Plan)
			if !ok {
				plan = &generic.Plan{}
			}

			// 核心修复：优先使用 filesystem 检查真正的进度
			// 因为框架的 ExecutedSteps 可能没有正确累积
			nextSlide := getNextSlideFromDisk(plan, workDir)

			// 如果 filesystem 也找不到剩余幻灯片，尝试用框架的 ExecutedSteps
			var stepStr string
			if nextSlide != nil {
				stepStr = generic.FormatStepForRequest(nextSlide, workDir)
			} else {
				// 回退到框架的 ExecutedSteps
				var executedStepJSONs []string
				for _, es := range in.ExecutedSteps {
					executedStepJSONs = append(executedStepJSONs, es.Step)
				}
				remainingSlides := plan.GetRemainingSlides(executedStepJSONs)
				if len(remainingSlides) > 0 {
					nextSlide = &remainingSlides[0]
					stepStr = generic.FormatStepForRequest(nextSlide, workDir)
				} else {
					stepStr = "[完成] 所有幻灯片都已生成完毕。"
				}
			}

			// 格式化已执行步骤（用于 prompt 显示）
			executedSummary := agentutils.FormatExecutedSteps(in.ExecutedSteps)

			// 如果 filesystem 显示有下一个幻灯片但框架没有，将其追加到 executedSteps 摘要中
			// 以便 prompt 能看到正确的历史进度
			if nextSlide != nil {
				existingFiles := generic.GetExistingStepFiles(workDir)
				allSlides := plan.GetSlides()
				for i := range allSlides {
					slide := &allSlides[i]
					if _, exists := existingFiles[slide.Index]; exists {
						// 检查是否已经在 ExecutedSteps 中
						found := false
						for _, es := range in.ExecutedSteps {
							if strings.Contains(es.Step, fmt.Sprintf(`"index":%d`, slide.Index)) ||
							   strings.Contains(es.Step, fmt.Sprintf(`"index": %d`, slide.Index)) {
								found = true
								break
							}
						}
						if !found {
							// 追加到摘要中
							executedSummary += fmt.Sprintf("## %d. Step: {\"index\":%d,\"title\":\"%s\"}\n  Result: 已生成文件\n\n",
								len(in.ExecutedSteps)+1, slide.Index, slide.Title)
						}
					}
				}
			}

			return executorPrompt.Format(ctx, map[string]any{
				"input":          agentutils.FormatInput(in.UserInput),
				"plan":           string(planContent),
				"executed_steps": executedSummary,
				"step":           stepStr,
				"skills":         skillsContent,
			})
		},
	})
	if err != nil {
		return nil, err
	}

	return a, nil
}
