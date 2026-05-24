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
	"path/filepath"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"

	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/prompts"
	"github.com/cloudwego/ppt-agent/pkg/tools"
)

func NewPPTTaskDeepAgent(ctx context.Context, cfg *PPTTaskConfig) (adk.Agent, error) {
	chatModel, err := agentutils.NewFallbackToolCallingChatModel(ctx,
		agentutils.WithMaxTokens(16384),
		agentutils.WithTemperature(0),
		agentutils.WithTopP(0),
	)
	if err != nil {
		return nil, fmt.Errorf("创建主模型失败: %w", err)
	}

	compressor, err := agentutils.NewFallbackToolCallingChatModel(ctx,
		agentutils.WithMaxTokens(4096),
		agentutils.WithTemperature(0),
	)
	if err != nil {
		return nil, fmt.Errorf("创建压缩器模型失败: %w", err)
	}
	chatModel = agentutils.NewChatModelCompressor(chatModel, compressor,
		agentutils.WithCompressThreshold(8),
		agentutils.WithTokenThreshold(20000),
		agentutils.WithPreserveCount(4),
	)
	if cfg.CompressorTracker != nil {
		if compressor, ok := chatModel.(*agentutils.ChatModelCompressor); ok {
			compressor.SetTracker(cfg.CompressorTracker)
		}
	}

	slideExecutor, err := newSlideExecutorAgent(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("创建 SlideExecutor 子代理失败: %w", err)
	}

	reviewer, err := newReviewerAgent(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("创建 Reviewer 子代理失败: %w", err)
	}

	fixer, err := newFixerAgent(ctx, cfg, "")
	if err != nil {
		return nil, fmt.Errorf("创建 Fixer 子代理失败: %w", err)
	}

	editFileTool := tools.NewEditFileTool(cfg.Operator)
	readFileTool := tools.NewReadFileTool(cfg.Operator)
	searchTool := tools.NewSearchTool()
	bashTool := tools.NewBashTool(cfg.Operator)
	batchConvertTool := tools.NewBatchConvertTool(cfg.Operator)

	deepAgent, err := deep.New(ctx, &deep.Config{
		Name:        "PPTTaskDeepAgent",
		Description: "PPT 任务调度代理，负责规划、并行生成、质检和修复 PPT 幻灯片",
		ChatModel:   chatModel,
		Instruction: buildDeepAgentInstruction(cfg.WorkDir, cfg.SkillsDir, cfg.StyleContext, cfg.Outline, cfg.Query),
		SubAgents:   []adk.Agent{slideExecutor, reviewer, fixer},
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{editFileTool, readFileTool, searchTool, bashTool, batchConvertTool},
			},
		},
		WithoutWriteTodos:      true,
		WithoutGeneralSubAgent: true,
		MaxIteration:           agentutils.EnvInt("MASTER_MAX_ITERATIONS", 120),
	})
	if err != nil {
		return nil, fmt.Errorf("创建 Deep Agent 失败: %w", err)
	}

	return deepAgent, nil
}

// buildDeepAgentInstruction loads the deep agent master instruction from template.
// When outline is provided, tasks.json is already pre-populated with user's slide plan.
// query provides the user's original topic description for content generation.
func buildDeepAgentInstruction(workDir string, skillsDir string, styleContext string, outline *TaskOutline, query string) string {
	tmplDir := filepath.Join(skillsDir, "visual_designer", "templates", "full-decks")
	tasksJSON := filepath.Join(workDir, "tasks.json")

	const templateCatalog = `| 模板文件 | 适用场景 | 页数 |
|---------|---------|------|
| tech-intro.py | 新技术介绍、行业科普、知识分享，从基础概念到应用实践，适合非技术受众 | 18页 |
| tech-sharing.py | 内部技术分享、技术培训、架构讲解，有章节划分，注重内容深度 | 18页 |
| product-launch.py | 新产品发布会、产品宣讲、客户演示，强调价值主张和差异化优势 | 14页 |
| weekly-report.py | 团队周报、项目月报、工作汇报，简洁高效，数据驱动 | 9页 |
| pitch-deck.py | 创业路演，投资人演示、商业计划，逻辑严密，数据驱动，说服力强 | 16页 |
| course-module.py | 教学课件、培训材料、知识分享，内容系统化，便于学习理解 | 17页 |
| current-affairs.py | 时政热点分析、政策解读、国际形势分析，稳重专业，数据支撑强 | 14页 |
| politics-ideology.py | 思政教育、团课培训、爱国主义教育，价值观明确，结构清晰 | 16页 |
| design-defense.py | 课程设计、毕业设计、项目答辩，逻辑清晰，技术扎实 | 12页 |
| innovation-compete.py | 大创/挑战杯/互联网+等科创竞赛汇报，创新性强，数据支撑 | 16页 |
| research-report.py | 市场调研、行业分析、可行性研究，数据详实，结论明确 | 14页 |
| activity-plan.py | 团建活动、校园活动、节日策划，活泼有创意，执行清晰 | 10页 |
| personal-summary.py | 个人总结、述职报告、年终总结，重点突出，成果可见 | 10页 |
| short-class-talk.py | 课堂5-10分钟短时分享、课题介绍，精简高效，快速传达 | 6页 |
| meeting-minutes.py | 会议记录、工作例会、项目评审会，结构清晰，行动明确 | 8页 |
| product-intro.py | 产品介绍、客户演示、功能展示，突出价值，增强信任 | 12页 |
| training-course.py | 内部培训、新人入职培训、技能培训，知识系统，互动引导 | 16页 |
| project-proposal.py | 新项目立项、项目申请、资源申请，理由充分，方案可行 | 12页 |`

	// When user provided an outline, tasks.json already contains template/theme.
	// Read it so we can tell the agent what to use.
	outlineTemplate := ""
	outlineTheme := ""
	outlineTitle := ""
	outlineQuery := query // user's original topic description
	hasOutline := outline != nil && len(outline.Slides) > 0
	if hasOutline {
		outlineTemplate = outline.Template
		outlineTheme = outline.Theme
		outlineTitle = outline.Title
	}

	data := &prompts.TemplateData{
		TmplDir:         tmplDir,
		TasksJSON:       tasksJSON,
		TemplateCatalog: templateCatalog,
		StyleContext:    styleContext,
		HasOutline:      hasOutline,
		OutlineQuery:    outlineQuery,
		OutlineTemplate: outlineTemplate,
		OutlineTheme:    outlineTheme,
		OutlineTitle:    outlineTitle,
		SkillsDir:       skillsDir,
	}

	instruction, err := prompts.RenderDeepAgent("master_instruction", data)
	if err != nil {
		panic("failed to render deep agent master instruction template: " + err.Error())
	}
	return instruction
}
