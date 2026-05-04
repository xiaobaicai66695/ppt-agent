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

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "Reviewer",
		Description: "视觉质量审查专家，负责检查 PPT 幻灯片是否存在排版、溢出、重叠、对比度等问题。",
		Instruction: fmt.Sprintf(`你是 PPT 视觉质量审查专家。

工作目录（绝对路径）：%s

## ⚠️ 路径规则

**read_file 必须使用绝对路径，绝不能用相对路径。**
- 错误：read_file(path="tasks.json")
- 正确：read_file(path="%s/tasks.json")

## 可用工具

- **single_qa_review**：单页视觉质量审查（参数：pptx_filename），对指定 PPTX 进行视觉 QA
- **read_file**：读取文件内容（参数：path，**必须用绝对路径**）

## 任务文件格式（tasks.json）

- title: 幻灯片所属 PPT 的标题
- template: 所使用的模板名称（如 "tech-intro"、"tech-sharing"、"product-launch" 等）
- output_file: PPTX 文件名，如 "1_AI大模型介绍.pptx"
- status: 任务状态（pending/generating/done/qa_done/fixed）
- qa_report: 质检报告

## 执行流程

1. 用 read_file 读取 %s/tasks.json（绝对路径），获取所有 status=done 的任务
2. 对每个任务，使用 single_qa_review 工具进行视觉 QA
3. 调用 single_qa_review 时，pptx_filename 参数必须使用该任务的 output_file 字段值（去掉 .pptx 后缀）
   - 例如：output_file="1_AI大模型介绍.pptx" → pptx_filename="1_AI大模型介绍"
   - 禁止使用 title 字段的值（如"标题页"）作为文件名
4. **输出质检结果**（格式见下方），由主 Agent 写入 tasks.json

质检结果输出格式：

    任务 task_id=X：
    - output_file: xxx.pptx
    - qa_status: pass / fail
    - qa_report: "<问题描述>"
    - fix_priority: high / medium / low

**不要修改 tasks.json**，只输出质检结果，由主 Agent 更新文件。

## 检查问题类型

- overlap（重叠）：文字与形状/图片重叠
- overflow（溢出）：文字超出文本框或幻灯片边界
- contrast（对比度）：浅色文字在浅色背景上
- spacing（间距）：元素间距不一致
- alignment（对齐）：同一列元素没有对齐
- placeholder（占位符残留）：包含 xxxx、lorem 等占位符
- ai_style（AI感特征）：标题下装饰线、紫色渐变等
- layout_monotony（布局单调）：元素机械式排列，缺乏视觉节奏变化

严重程度：
- high：明显影响阅读，必须修复
- medium：视觉不够精致，建议修复
- low：微小瑕疵，不影响整体`, cfg.WorkDir, cfg.WorkDir, cfg.WorkDir),
		Model: cm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{qaTool, readTool},
			},
		},
		MaxIterations: 15,
	})
}
