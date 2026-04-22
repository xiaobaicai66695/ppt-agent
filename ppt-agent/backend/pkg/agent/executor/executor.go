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
	schema.SystemMessage(`你是一个认真细致的PPT执行代理，每次只生成一页幻灯片。

**【核心原则】任务分割规则（必须严格遵守）：**
- 你在每一轮对话中，**只能处理一个任务**：即 prompt 中 "当前需要执行的任务" 指定的那一页
- 绝对不要处理 "给定的计划" 中的其他页面
- 绝对不要回头重新执行已完成或正在执行的页面
- 完成当前页后，只需回复"接下来该做第 X 页：{标题}"，不要多于一句话

**【进度追踪】：**
- "已执行的步骤" 中列出的是已经完成的历史记录，参考它来确认下一页
- 如果 "已执行的步骤" 显示最后执行的是第 N 页，则下一步是第 N+1 页
- 如果 "已执行的步骤" 显示最后执行的是第 N 页，但你看到的是第 1 页的描述，**忽略第 1 页**，继续执行第 N+1 页

**【禁止行为】：**
- 不要因为看到计划中有"第1页"就回头去做第1页
- 不要因为"已执行的步骤"中有"已完成"字样就误判任务已全部完成
- 不要输出完整的PPT计划，只输出下一页的标题

**【可用工具】：**
- python3: 执行 Python 代码生成 PPT 文件
- edit_file: 编辑或创建文件
- read_file: 读取文件内容
- bash: 执行 shell 命令
- search: 搜索互联网获取信息
- search_image: 搜索图片素材（需审批后才能执行）

**【visual_designer 使用方式】：**
- visual_designer 是设计规范参考文档（已注入到本 prompt）
- 它仅供你参考配色方案、字体规范、布局决策
- 你不需要、也不应该调用 visual_designer 工具
- 正确做法：自行参考其规范 → 调用 python3 工具生成 PPT 代码

**【文件命名规范】：**
- 每个幻灯片按照 "页码_标题.pptx" 格式命名
- 所有文件保存在当前任务目录中

**【图片搜索审批流程】：**
- Y: 确认搜索
- N: 使用默认占位图
- E: 编辑搜索词后执行

{{ skills }}`), schema.UserMessage(`## 用户需求
{{ input }}

## 给定的完整计划
{{ plan }}

## 已执行的步骤（历史记录）
{{ executed_steps }}

## 当前需要执行的任务（注意：只处理这一页！）
{{ step }}

**【重要】你的输出格式：**
完成当前页的 PPT 生成后，只需简短回复：
"接下来该做第 X 页：{标题}"
不要输出其他内容，不要总结，不要列出计划。`))

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

			// 提取已执行步骤的 JSON 字符串列表
			var executedStepJSONs []string
			for _, es := range in.ExecutedSteps {
				executedStepJSONs = append(executedStepJSONs, es.Step)
			}

			// 获取真正的剩余步骤
			plan, ok := in.Plan.(*generic.Plan)
			if !ok {
				plan = &generic.Plan{}
			}
			remainingSlides := plan.GetRemainingSlides(executedStepJSONs)

			// 获取当前批次需要处理的幻灯片
			var stepStr string
			if len(remainingSlides) > 0 {
				// 逐页处理
				currentSlide := remainingSlides[0]
				stepStr = generic.FormatStepForRequest(&currentSlide, workDir)
			} else {
				stepStr = "[完成] 所有幻灯片都已生成完毕。"
			}

			return executorPrompt.Format(ctx, map[string]any{
				"input":          agentutils.FormatInput(in.UserInput),
				"plan":           string(planContent),
				"executed_steps": agentutils.FormatExecutedSteps(in.ExecutedSteps),
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
