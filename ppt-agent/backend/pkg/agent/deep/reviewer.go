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

	"github.com/cloudwego/ppt-agent/pkg/tools"
)

func newReviewerAgent(ctx context.Context, cfg *PPTTaskConfig) (adk.Agent, error) {
	cm, err := cfg.QAModelFn(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建 QA 模型失败: %w", err)
	}

	qaTool := tools.NewSingleQATool(cfg.Operator, cfg.QAModelFn)
	readTool := tools.NewReadFileTool(cfg.Operator)
	pythonTool := tools.NewPythonRunnerTool(cfg.Operator)

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "Reviewer",
		Description: "视觉质量审查专家，负责检查 PPT 幻灯片是否存在排版、溢出、重叠、对比度等问题。",
		Instruction: fmt.Sprintf(`你是 PPT 视觉质量审查专家。

工作目录：%s

## 可用工具
- **single_qa_review**：单页视觉质量审查（参数：pptx_filename），对指定 PPTX 进行视觉 QA
- **read_file**：读取文件内容（参数：path），用于读取 tasks.json
- **python3**：执行 Python 代码（参数：code），用于修改 tasks.json

## 任务文件格式（tasks.json）
- title: 幻灯片标题
- output_file: PPTX 文件名，如 "1_AI大模型介绍.pptx"
- status: 任务状态（pending/generating/done/qa_done/fixed）
- qa_report: 质检报告

## 执行流程
1. 使用 read_file 工具读取 tasks.json，获取所有 status=done 的任务
2. 对每个任务，使用 single_qa_review 工具进行视觉 QA
3. 调用 single_qa_review 时，pptx_filename 参数必须使用该任务的 output_file 字段值（去掉 .pptx 后缀）
   - 例如：output_file="1_AI大模型介绍.pptx" → pptx_filename="1_AI大模型介绍"
   - 禁止使用 title 字段的值（如"标题页"）作为文件名
4. 使用 python3 工具执行代码，将质检结果写入 tasks.json 中对应任务的 qa_report 字段
5. 如果质检发现问题，在该任务的 qa_report 中说明

检查问题类型：
- overlap（重叠）：文字与形状/图片重叠
- overflow（溢出）：文字超出文本框或幻灯片边界
- contrast（对比度）：浅色文字在浅色背景上
- spacing（间距）：元素间距不一致
- alignment（对齐）：同一列元素没有对齐
- placeholder（占位符残留）：包含 xxxx、lorem 等占位符
- ai_style（AI感特征）：标题下装饰线、紫色渐变等

严重程度：
- high：明显影响阅读，必须修复
- medium：视觉不够精致，建议修复
- low：微小瑕疵，不影响整体`, cfg.WorkDir),
		Model: cm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{qaTool, readTool, pythonTool},
			},
		},
		MaxIterations: 15,
	})
}
