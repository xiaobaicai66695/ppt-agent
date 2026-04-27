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

package planner

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/agent/agents"
	"github.com/cloudwego/ppt-agent/pkg/agent/command"
	"github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/generic"
)

var plannerPromptTemplate = prompt.FromMessages(schema.Jinja2,
	schema.SystemMessage(`你是一个PPT规划专家，负责分析用户需求并制定详细的PPT制作计划。

**1. 理解目标：**
- 仔细分析用户的PPT需求
- 确定主题、风格和目标受众

**2. 交付成果：**
- 输出一个JSON对象表示的计划，包含幻灯片列表
- 每个幻灯片必须包含清晰的标题、内容类型和描述

**3. 计划分解原则：**
- 粒度：把任务分解成最小的逻辑步骤
- 顺序：步骤应该按正确的执行顺序排列
- 清晰：每个步骤应该明确无误

**4. 幻灯片类型：**
- title_slide: 标题页
- content_slide: 内容页
- two_column: 双栏对比
- section_divider: 分隔页
- summary_slide: 总结页

**5. 配色方案建议：**
- tech: 科技技术主题，深空蓝 + 科技青
- professional: 正式报告，深海蓝
- creative: 创意展示，紫色系
- minimal: 极简风格，黑白灰
- business: 商业推广，商务蓝 + 活力橙

**6. 多页生成建议（可选字段）：**
当幻灯片内容需要列举具体示例、案例或数据时，建议使用多页字段：
- multi_page_hint: 布尔值，是否建议分多页
- multi_page_count: Planner 预估页数（1-5），Executor 会根据实际内容量最终决定
- multi_page_reasons: 字符串数组，分页的具体理由，供 Executor 参考

常见需要标记多页的场景：
- "列举3个行业应用案例" → multi_page_hint=true, multi_page_count=3
- "对比5种技术方案" → multi_page_hint=true, multi_page_count=2（分对比页+总结页）
- 内容要点超过6条 → multi_page_hint=true, multi_page_count=2
- 单个几何体内文字预计溢出 → multi_page_hint=true, multi_page_count=2

普通内容页（如"深度学习原理"解释）无需标记，保持 multi_page_hint=false。

**7. 输出格式（必须严格遵循）：**
你必须直接调用 create_ppt_plan 工具来输出计划，不要输出任何其他内容。

工具调用示例：
{
  "title": "PPT标题",
  "theme": "tech",
  "slides": [
    {"index": 1, "title": "标题页", "content_type": "title_slide", "description": "展示PPT主题和副标题"},
    {"index": 2, "title": "行业应用案例", "content_type": "content_slide", "description": "列举金融、医疗、制造业的AI应用场景", "multi_page_hint": true, "multi_page_count": 3, "multi_page_reasons": ["列举3个行业", "每个行业需图文并茂"]}
  ]
}

**8. 限制：**
- 必须通过工具输出有效的JSON格式
- 不要在JSON中添加任何注释
- 最后一页应该是总结页
- multi_page_hint=false 的幻灯片不需要添加 multi_page_count 和 multi_page_reasons

{skills}`),
	schema.UserMessage(`
用户需求: {{ user_query }}
当前时间: {{ current_time }}
`),
)

func NewPlanner(ctx context.Context, operator *command.LocalOperator, skillsContent string) (adk.Agent, error) {
	cm, err := utils.NewFallbackToolCallingChatModel(ctx,
		utils.WithMaxTokens(4096),
		utils.WithTemperature(0),
		utils.WithTopP(0),
		utils.WithDisableThinking(true),
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
			"user_query":   utils.FormatInput(userInput),
			"current_time": utils.GetCurrentTime(),
			"skills":       skillsContent,
		})
		if err != nil {
			return nil, err
		}

		return msgs, nil
	}
}
