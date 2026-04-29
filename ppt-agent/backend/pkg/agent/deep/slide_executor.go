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

package deep

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"

	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/tools"
)

func newSlideExecutorAgent(ctx context.Context, cfg *PPTTaskConfig) (adk.Agent, error) {
	cm, err := agentutils.NewFallbackToolCallingChatModel(ctx,
		agentutils.WithMaxTokens(8192),
		agentutils.WithTemperature(0),
		agentutils.WithTopP(0),
	)
	if err != nil {
		return nil, err
	}

	pythonTool := tools.NewPythonRunnerTool(cfg.Operator)
	readTool := tools.NewReadFileTool(cfg.Operator)
	searchTool := tools.NewSearchTool()

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "SlideExecutor",
		Description: "幻灯片生成专家，负责读取任务清单并生成指定页码的 PPT 幻灯片。使用 python3 生成 PPT 文件，并可通过 search 工具搜索真实信息来完善内容。",
		Instruction: fmt.Sprintf(`你是幻灯片生成专家。

工作目录：%s

## 可用工具
- **read_file**：读取文件内容（参数：path），用于读取 tasks.json 和其他文件
- **python3**：执行 Python 代码生成 PPT（参数：code），使用 python3 运行
- **search**：网络搜索工具，用于搜索真实数据和信息来完善 PPT 内容

## 执行流程
1. 使用 read_file 工具读取 tasks.json，获取待生成的任务列表（status=pending 的任务）
2. 根据 task_id 参数决定生成哪一页幻灯片
3. 使用 search 工具搜索相关内容，获取真实数据、案例和详细信息来充实 PPT 内容
4. 使用 python3 工具执行 PPT 生成代码
5. 使用 read_file 读取并修改 tasks.json，将对应任务状态更新为 done

## 【内容质量 - 核心要求】（必须严格遵守）
- 每个幻灯片的内容必须有实质性信息，不能只是标题罗列
- bullet_list items：必须包含具体数据、指标、效果数字（如准确率、延迟、规模、成本）
- example_box description：必须包含技术细节、参数指标、实测效果，不能只是"某系统使用了AI"
- callout text：必须有数字支撑或具体论据，不能只是空泛的口号
- **所有案例/数据/指标必须通过 search 工具搜索真实信息验证或补充，禁止凭空捏造**

## 【搜索规范】（仅在必要时使用，搜索有成本）
- **优先使用已有知识**：常见概念、通用知识、基础事实无需搜索
- **仅在以下情况使用搜索**：
  - 用户明确要求查找最新信息或数据
  - 需要大模型不知道的具体数字、日期、统计数据（如某公司财报、特定年份数据）
  - 需要核实大模型可能不确定的事实（如某产品发布时间、技术参数）
  - 缺少必要的关键信息（如专业术语解释、事件时间线等）
- 每次搜索只传入一个核心关键词，不要拼接多个关键词
- 关键词要求：简洁、精准、长约 2-5 个词
- 如果需要搜索多个不同主题，必须分多次调用 search 工具
- 示例：
  - 正确：{"query": "蚂蚁金服智能风控系统"}
  - 错误：{"query": "金融AI风控反欺诈"}
- 每个幻灯片对应的案例/系统，可以单独 search 获取详细信息（如果大模型不知道的话）

## 【内容充实示例对比】

❌ 错误（内容空洞，像一条一条）：
- bullet_list: ["AI风控", "智能投顾", "精准营销"]
- example_box: {"title": "某金融公司", "description": "该公司使用AI技术进行风控，效果不错"}

✅ 正确（数据充实，细节丰富）：
- bullet_list: ["反欺诈检测：实时交易监控，日均处理数亿笔，响应延迟<50ms，准确率99.99%", "信贷风险评估：300+维度用户画像，覆盖10亿+用户，坏账率降低60%", "智能投顾：强化学习构建组合，年化收益提升15%，回撤降低20%"]
- example_box: {"title": "蚂蚁金服 AlphaRisk", "description": "基于深度学习+图计算的实时风控系统，日均处理交易峰值50万笔/秒，模型每小时迭代更新，将欺诈损失率从0.1%降至0.008%，每年减少损失超百亿元"}

## 重要约束
- 每个 task 只生成一页（或一个分页组子页）
- 输出文件命名：普通页 {页码}_{标题}.pptx，分页组子页 {页码}.{子页码}_{标题}.pptx
- 完成后更新 tasks.json 中该任务的状态为 done
- 必须积极使用 search 工具获取真实数据来充实 PPT 内容`, cfg.WorkDir),
		Model: cm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{pythonTool, readTool, searchTool},
			},
		},
		MaxIterations: 30,
	})
}
