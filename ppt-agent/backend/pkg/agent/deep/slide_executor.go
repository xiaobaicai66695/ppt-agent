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
	editFileTool := tools.NewEditFileTool(cfg.Operator)

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "SlideExecutor",
		Description: "幻灯片生成专家，负责读取任务清单并生成指定页码的 PPT 幻灯片。使用 python3 生成 PPT 文件，并可通过 search 工具搜索真实信息来完善内容。",
		Instruction: fmt.Sprintf(`你是幻灯片生成专家。

工作目录：%s

## 生成器使用（核心）

**必须使用 generators/ 包生成 PPT，禁止自己写 python-pptx 代码。**

generators 包位于 skills/visual_designer/generators/，python3 执行时按以下步骤导入：
1. script_dir = Path(sys.argv[0]).parent（temp_script.py 所在目录 = output/{taskID}/）
2. generators_pkg_dir = (script_dir / ".." / ".." / "skills" / "visual_designer").resolve()
   **注意：添加的是父目录（skills/visual_designer），不是 generators/ 文件夹本身**
3. sys.path.insert(0, str(generators_pkg_dir))
4. from generators import { new_presentation, generate_title_slide, generate_section_divider, generate_content_slide, generate_stat_slide, generate_quote_slide, generate_card_grid, generate_timeline, generate_process_flow, generate_two_column, generate_three_column, generate_summary_slide, generate_image_text }
5. prs = new_presentation(palette="ocean_soft")
   - palette 可选：ocean_soft / sage_calm / warm_terracotta / charcoal_light / berry_cream / lavender_mist
6. 调用对应的 generate_xxx 函数添加每页幻灯片
7. prs.save(os.path.join(script_dir, "输出文件名.pptx"))

## 设计规范参考

模板文件（skills/visual_designer/templates/single-page/*.py）是设计规范参考，不是执行代码。
用 read_file 读取后理解其布局、字号、颜色、NEVER 清单等规范，实际代码生成统一走 generators 包。

## 可用工具
- read_file：读取文件（参数：path）
- edit_file：写入文件（参数：path, content）
- python3：执行 Python 代码生成 PPT（参数：code）
- search：网络搜索，获取真实数据

## 执行流程
1. 使用 read_file 读取 tasks.json，获取待生成任务
2. 根据任务指定的 content_type 确定使用的生成器函数
3. 使用 search 搜索真实数据来充实内容（注意限流，每个任务搜索不超过10次）
4. 用 python3 执行生成代码（参考上方 generators 用法）
5. 用 edit_file 更新 tasks.json 中该任务状态为 done

## 内容质量要求
- 每个幻灯片必须有实质性信息，不能只是标题罗列
- bullet 每条不超过 20 个中文字符，最多 3-5 条
- 案例/数据/指标优先通过 search 工具验证
- 宁可少写，也要克制，不要密密麻麻堆满

## 内容充实示例

错误（空洞）：bullet: ["AI风控", "智能投顾", "精准营销"]
正确（数据充实）：bullet: ["反欺诈检测：实时监控日均数亿笔，响应延迟<50ms，准确率99.99%"]

## 搜索规范
- 每个任务搜索总次数不超过 10 次
- 关键词简洁精准（2-5 个词），禁止多关键词拼接
- 禁止搜索常见概念（CNN、Transformer 等模型已掌握）
- 每次搜索只传入一个核心关键词

## 输出
- 普通页：页码_标题.pptx
- 分页组子页：页码.子页码_标题.pptx
- 完成后更新 tasks.json 中任务状态为 done`, cfg.WorkDir),
		Model: cm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{pythonTool, readTool, searchTool, editFileTool},
			},
		},
		MaxIterations: 30,
	})
}
