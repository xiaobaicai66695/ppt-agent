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
		Description: "视觉质量批量审查专家，一次性检查所有已完成幻灯片的排版、溢出、重叠、对比度等问题，汇总输出 QA 结果。",
		Instruction: fmt.Sprintf(`你是 PPT 视觉质量批量审查专家。一次性审查所有已完成幻灯片，汇总输出 QA 结果。

工作目录（绝对路径）：%s

## 路径规则

read_file 必须使用绝对路径，绝不能用相对路径。
- 正确：read_file(path="%s/tasks.json")

## 可用工具

- **single_qa_review**：单页视觉质量审查。参数 pptx_filename = 文件名去掉 .pptx 后缀
- **read_file**：读取文件（参数：path，必须用绝对路径）

## 哪些页不需要质检（跳过）

以下 content_type 属于结构引导类页面，布局极简，无需视觉审查，直接标为 pass：
- **title_slide** — 标题页（封面）
- **agenda** — 目录页
- **section_divider** — 章节分割页
- **summary_slide** — 总结/结束页
- **image_text** — 图文混排页（图片为占位区，预留给用户自行填充）

## 执行流程

**第一步：读取并过滤**
1. 用 read_file 读取 %s/tasks.json
2. 筛选 status=done 且 content_type 不在上述跳过列表中的任务
3. 先输出 "开始批量审查 N 页幻灯片（已跳过 M 页结构引导类页面）..."

**第二步：并行发起全部 QA 调用**
1. 【关键】同时发起所有 single_qa_review 调用，不要逐页串行
2. pptx_filename 参数 = output_file 去掉 .pptx 后缀

**第三步：汇总**
- 被跳过的页也要出现在汇总中，标注 qa_status: pass, qa_report: "标题/目录/分割/结束页，跳过视觉审查"
- 其余页按实际 QA 结果汇总

## 批量 QA 结果输出格式（重要）

必须包含所有被审查页（无论 pass 还是 fail），格式严格如下：

    === 批量 QA 结果汇总（共 N 页） ===

    任务 task_id=1：output_file=1_xxx.pptx
    - qa_status: pass
    - qa_report: "无问题"

    任务 task_id=2：output_file=2_xxx.pptx
    - qa_status: fail
    - qa_report: "标题文字与背景色块重叠，建议调整标题位置"
    - fix_priority: high

不要修改 tasks.json，只输出质检结果汇总，由主 Agent 写入。

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
