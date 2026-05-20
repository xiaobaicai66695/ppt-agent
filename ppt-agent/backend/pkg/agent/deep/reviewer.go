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

	"github.com/cloudwego/ppt-agent/pkg/prompts"
	"github.com/cloudwego/ppt-agent/pkg/tools"
)

func newReviewerAgent(ctx context.Context, cfg *PPTTaskConfig) (adk.Agent, error) {
	cm, err := cfg.QAModelFn(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建 QA 模型失败: %w", err)
	}

	qaTool := tools.NewSingleQATool(cfg.Operator, cfg.QAModelFn)
	readTool := tools.NewReadFileTool(cfg.Operator)

	instruction, err := buildReviewerInstruction(cfg.WorkDir)
	if err != nil {
		panic("failed to render reviewer instruction template: " + err.Error())
	}

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "Reviewer",
		Description: "视觉质量批量审查专家，一次性检查所有已完成幻灯片的排版、溢出、重叠、对比度等问题，汇总输出 QA 结果。",
		Instruction: instruction,
		Model:       cm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{qaTool, readTool},
			},
		},
		MaxIterations: 15,
	})
}

func buildReviewerInstruction(workDir string) (string, error) {
	data := &prompts.TemplateData{
		WorkDir: workDir,
	}
	return prompts.RenderDeepAgent("reviewer_instruction", data)
}
