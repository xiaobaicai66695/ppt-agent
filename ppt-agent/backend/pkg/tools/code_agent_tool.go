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

package tools

import (
	"context"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/params"
)

type CodeAgentConfig struct {
	Name        string
	Description string
	Instruction string
	MaxTokens   int
	Temperature float32
	TopP        float32
}

func DefaultCodeAgentConfig() *CodeAgentConfig {
	return &CodeAgentConfig{
		Name:        "CodeAgent",
		Description: "这是一个代码代理，专门用于处理 PPT 相关任务。它接收包含多个幻灯片的计划，一次性生成所有 PPT 文件。",
		Instruction: `你是一个 PPT 代码生成代理，负责批量生成所有幻灯片。

【输入格式】
你会收到一个包含多个幻灯片的批量任务，每个幻灯片包含：
- index: 页码
- title: 标题
- content_type: 内容类型 (title_slide, content_slide, two_column, section_divider, summary_slide)
- description: 描述

【工作流程】：
1. 解析所有幻灯片信息
2. 为每个幻灯片生成独立的 PPT 文件
3. 文件名格式：{页号}_{标题}.pptx（例如：1_标题页.pptx, 2_目录页.pptx）
4. 所有文件保存到工作目录
5. 完成后列出所有生成的文件

【生成规则】
- 每个 PPT 文件必须包含该页的所有内容
- 标题页包含主标题和副标题
- 内容页包含标题和正文内容
- 使用 python-pptx 库生成
- 必须调用 Python Runner 工具执行代码

【代码执行】
必须通过 python_runner 工具执行 Python 代码来生成 PPT。
在一个 Python 脚本中生成所有幻灯片。
不要重复调用工具，一次脚本执行完成所有生成。

【示例输入】
批量任务包含3个幻灯片：
1. index=1, title="标题页", content_type="title_slide"
2. index=2, title="目录", content_type="content_slide"
3. index=3, title="概述", content_type="section_divider"

【重要约束】
1. 一个 Python 脚本生成所有 PPT 文件
2. 文件名严格遵循 {页号}_{标题}.pptx 格式
3. 完成后直接返回结果，不要重复生成`,
		MaxTokens:   14125,
		Temperature: 1,
		TopP:        1,
	}
}

func newCodeAgent(ctx context.Context, operator commandline.Operator, opts ...func(*CodeAgentConfig)) (adk.Agent, error) {
	config := DefaultCodeAgentConfig()
	for _, opt := range opts {
		opt(config)
	}

	cm, err := agentutils.NewFallbackToolCallingChatModel(ctx,
		agentutils.WithMaxTokens(config.MaxTokens),
		agentutils.WithTemperature(config.Temperature),
		agentutils.WithTopP(config.TopP),
	)
	if err != nil {
		return nil, err
	}

	preprocess := []ToolRequestPreprocess{ToolRequestRepairJSON}

	ca, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        config.Name,
		Description: config.Description,
		Instruction: config.Instruction,
		Model:       cm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{
					NewWrapTool(NewBashTool(operator), preprocess, nil),
					NewWrapTool(NewEditFileTool(operator), preprocess, nil),
					NewWrapTool(NewReadFileTool(operator), preprocess, nil),
					NewWrapTool(NewPythonRunnerTool(operator), preprocess, nil),
				},
			},
		},
		GenModelInput: func(ctx context.Context, instruction string, input *adk.AgentInput) ([]adk.Message, error) {
			wd, _ := params.GetTypedContextParams[string](ctx, params.WorkDirSessionKey)

			tpl := prompt.FromMessages(schema.Jinja2,
				schema.SystemMessage(instruction),
				schema.UserMessage(`工作目录: {{ working_dir }}
用户查询: {{ user_query }}
当前时间: {{ current_time }}
`))

			msgs, err := tpl.Format(ctx, map[string]any{
				"working_dir":  wd,
				"user_query":   agentutils.FormatInput(input.Messages),
				"current_time": agentutils.GetCurrentTime(),
			})
			if err != nil {
				return nil, err
			}

			return msgs, nil
		},
		OutputKey:     "",
		MaxIterations: 1000,
	})
	if err != nil {
		return nil, err
	}

	return ca, nil
}

func newCodeAgentTool(ctx context.Context, operator commandline.Operator) (tool.BaseTool, error) {
	ca, err := newCodeAgent(ctx, operator)
	if err != nil {
		return nil, err
	}

	return adk.NewAgentTool(ctx, ca), nil
}

func NewCodeAgentTool(ctx context.Context, operator commandline.Operator) (tool.BaseTool, error) {
	return newCodeAgentTool(ctx, operator)
}
