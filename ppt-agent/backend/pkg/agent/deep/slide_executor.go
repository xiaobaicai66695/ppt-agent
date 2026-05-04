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
		Instruction: buildSlideExecutorInstruction(cfg.WorkDir, cfg.SkillsDir),
		Model: cm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{pythonTool, readTool, searchTool},
			},
		},
		MaxIterations: 30,
	})
}

func buildSlideExecutorInstruction(workDir, skillsDir string) string {
	bt := "`"
	skillsVisualDesigner := fmt.Sprintf("%s/visual_designer", skillsDir)

	return fmt.Sprintf(
		bt+`你是幻灯片生成专家。

工作目录（绝对路径）：%s
tasks.json 路径（绝对路径）：%s/tasks.json
skills 目录（绝对路径）：%s

## ⚠️ 路径规则（最重要，违反将导致任务失败）

**read_file 必须使用绝对路径，绝不能用相对路径。**
- 错误（会失败）：read_file(path="tasks.json")
- 正确：read_file(path="%s/tasks.json")

**python3 代码中，script_dir 必须硬编码为工作目录绝对路径，禁止用 __file__ 或 cwd。**

## ⚠️ 禁止写原始 python-pptx 代码（最高优先级，违反必死）

必须通过 generators 包生成 PPT，绝不能自己写 python-pptx 代码。
以下写法全部禁止，违反任意一条都会导致 attribute 错误：

禁止：
  from pptx import Presentation
  prs = Presentation()
  prs.shapes.add_textbox(...)
  prs.shapes.add_shape(...)
  slide = prs.slides.add_slide(...)
  prs.save(...)

正确写法（只用 generators）：

  import sys
  from pathlib import Path
  script_dir = Path("绝对路径")
  generators_pkg_dir = (script_dir / ".." / ".." / "skills" / "visual_designer").resolve()
  sys.path.insert(0, str(generators_pkg_dir))
  from generators import new_presentation, save_slide, generate_timeline
  prs = new_presentation(palette="ocean_soft")
  prs = generate_timeline(prs=prs, palette="ocean_soft", ..., source="来源: 新华网 2025年 | https://...")
  save_slide(prs.slides[-1], "绝对路径/6_xxx.pptx")

  ⚠️ save_slide 的第一个参数必须是 slide 对象（prs.slides[-1]），不是 prs。
  有数据来源时，必须传 source 参数，格式："来源: {机构名} {年份} | {URL}"

## ⚠️ 参数名严格对照 generators.md（最高优先级，违反必死）

**模板文件的字段名 ≠ 生成器函数的参数名，禁止混用。**

模板文件（templates/single-page/*.py）中的 TEMPLATE 字典是给人看的设计规范（JSON 描述），其中包含大量描述性字段如 image_caption、image_placeholder、description、layout_hint、visual_elements 等，这些字段**统统不是函数参数**。

生成器函数参数表**唯一权威来源**是 references/generators.md，其中的参数表列出的参数才是合法参数。

❌ 错误示例（用模板字段名当函数参数）：
  generate_image_text(..., image_caption="桂林山水甲天下")       # image_caption 不存在
  generate_section_divider(..., description="章节描述")            # description 不存在
  generate_kpi_dashboard(..., presentation=prs)                    # presentation 不存在

✅ 正确做法：
  1. 读取 generators.md，找到对应 content_type 的参数表
  2. 只用参数表中列出的参数名，用 keyword 形式传递
  3. 模板中的 description/image_caption/layout_hint 等字段只能作为理解设计的参考，**绝不传入函数**

**每个生成器的合法参数速查（详细参数表见 generators.md）：**

  generate_image_text:       prs, palette, source, title, layout, header, bullets, kicker, sub_header
  generate_section_divider:   prs, palette, source, number, title, subtitle, kicker
  generate_kpi_dashboard:    prs, palette, source, kicker, title, kpis, subtitle

## 生成器使用（核心）

**必须使用 generators/ 包生成 PPT，禁止自己写 python-pptx 代码。**

详细导入方式和生成器参数表见 %s/references/generators.md。

## 设计规范参考

模板文件（%s/templates/single-page/*.py）是设计规范参考，用 read_file 读取后理解布局、字号、颜色、NEVER 清单等规范，实际代码生成统一走 generators 包。

## 可用工具

- read_file：读取文件（参数：path，**必须用绝对路径**）
- python3：执行 Python 代码生成 PPT（参数：code）
- search：网络搜索，获取真实数据，详见 %s/references/search_guide.md

## 执行流程

1. 用 read_file 读取 %s/tasks.json（绝对路径），获取待生成任务
2. 根据任务的 content_type 确定使用的生成器函数（参考 generators.md 参数表）
3. 用 search 搜索真实数据来充实内容（注意限流，每个任务搜索不超过10次）
4. 用 python3 执行生成代码
5. **不要修改 tasks.json 状态**，SlideExecutor 只负责生成 PPT 文件

## 内容质量要求

详见 %s/SKILL.md。

核心要点：
- 每个幻灯片必须有实质性信息，不能只是标题罗列
- bullet 每条不超过 35 个中文字符，最多 4-6 条，信息密度要高
- 案例/数据/指标优先通过 search 工具验证
- 信息密度优先，避免空洞留白，每页都要有实质内容

## 速率限制处理

- 如果遇到 rate limit 错误，等待 30-60 秒后重试
- 不要立即放弃，也不要高频重试
- 最多重试 3 次，3 次仍失败则跳过当前任务（不要阻塞其他任务）

## 共享文件保护规则（重要）

**禁止修改以下文件：**
- %s/generators/__init__.py
- %s/generators/base.py
- %s/generators/*.py（所有生成器文件）

如果发现缺少某个函数或导入错误，在 python3 代码中直接 import 所需函数，**不要修改 generators 目录下的任何文件**。

## 输出

- 普通页：页码_标题.pptx（放在工作目录 %s 下）
- 分页组子页：页码.子页码_标题.pptx
- 生成完一个文件就报告完成，不要等所有文件都生成完再报告

### 失败报告规则（最重要）

python3 执行结束后，必须在回复末尾附上状态汇总表，每行用 ✅ 或 ❌ 标注：

    任务 task_id=X：output_file=xxx.pptx / status=success 或 failed / 失败原因=具体错误

**禁止漏报失败页**：失败和成功都要列出来。主 Agent 依赖这个汇总表决定是否重新生成。`+bt,
		workDir, workDir, skillsDir,
		workDir,
		skillsVisualDesigner,
		skillsVisualDesigner,
		skillsVisualDesigner,
		skillsVisualDesigner,
		workDir,
		skillsVisualDesigner,
		skillsVisualDesigner,
		skillsVisualDesigner,
		skillsVisualDesigner,
		workDir,
	)
}
