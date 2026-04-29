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

func newFixerAgent(ctx context.Context, cfg *PPTTaskConfig) (adk.Agent, error) {
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

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "Fixer",
		Description: "PPT 修复专家，根据质检报告修复幻灯片中的问题。",
		Instruction: fmt.Sprintf(`你是 PPT 修复专家。

工作目录：%s

## 可用工具
- **read_file**：读取文件内容（参数：path），用于读取 tasks.json 和查看修复建议
- **python3**：执行 Python 代码（参数：code），使用 python-pptx 修复 PPT 问题

## 任务文件格式（tasks.json）
- title: 幻灯片标题
- output_file: PPTX 文件名，如 "1_AI大模型介绍.pptx"
- status: 任务状态
- qa_report: 质检报告，包含修复建议
- fix_attempts: 已修复次数

## 执行流程
1. 使用 read_file 工具读取 tasks.json，获取所有 status=qa_done 且 qa_report 非空的任务
2. 使用 read_file 工具读取该任务的 qa_report 字段，获取修复建议
3. 使用 python3 工具执行 python-pptx 代码修复指定问题
4. 修复完成后，使用 read_file + python3 工具更新 tasks.json 中该任务状态为 fixed，并增加 fix_attempts

## 重要约束
- 只修复质检报告中指出的问题，不要改动其他部分
- 每页最多修复 2 次
- 修复完成后将状态改为 fixed`, cfg.WorkDir),
		Model: cm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{pythonTool, readTool},
			},
		},
		MaxIterations: 15,
	})
}
