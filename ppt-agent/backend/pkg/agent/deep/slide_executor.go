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

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"

	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/prompts"
	"github.com/cloudwego/ppt-agent/pkg/style"
	"github.com/cloudwego/ppt-agent/pkg/tools"
)

func NewSlideExecutorAgent(ctx context.Context, cfg *PPTTaskConfig) (adk.Agent, error) {
	return newSlideExecutorAgent(ctx, cfg)
}

func newSlideExecutorAgent(ctx context.Context, cfg *PPTTaskConfig) (adk.Agent, error) {
	cm, err := agentutils.NewFallbackToolCallingChatModel(ctx,
		agentutils.WithMaxTokens(32768),
		agentutils.WithTemperature(0),
		agentutils.WithTopP(0),
	)
	if err != nil {
		return nil, err
	}

	cm = wrapSlideExecutorCompressor(ctx, cm)

	pythonTool := tools.NewPythonRunnerTool(cfg.Operator)
	readTool := tools.NewReadFileTool(cfg.Operator)
	searchTool := tools.NewSearchTool()

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "SlideExecutor",
		Description: "幻灯片生成专家，负责读取任务清单并生成指定页码的 PPT 幻灯片。使用 python3 生成 PPT 文件，并可通过 search 工具搜索真实信息来完善内容。",
		Instruction: buildSlideExecutorInstruction(cfg.WorkDir, cfg.SkillsDir, cfg.EnhancedProfile),
		Model:       cm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{pythonTool, readTool, searchTool},
			},
		},
		MaxIterations: agentutils.EnvInt("SLIDE_EXECUTOR_MAX_ITERATIONS", 50),
	})
}

// wrapSlideExecutorCompressor 用轻量级压缩器包装聊天模型
// 摘要器使用单独的 fallback 模型，token 阈值较低（15000）
// 因为 SlideExecutor 的对话更简单（读任务 → 搜索 → 生成 → 报告）
// 阈值可通过 SLIDE_EXECUTOR_COMPRESSOR_MESSAGE_THRESHOLD、
// SLIDE_EXECUTOR_COMPRESSOR_TOKEN_THRESHOLD、SLIDE_EXECUTOR_COMPRESSOR_PRESERVE_COUNT 覆盖
func wrapSlideExecutorCompressor(ctx context.Context, inner model.ToolCallingChatModel) model.ToolCallingChatModel {
	summarizer, err := agentutils.NewFallbackToolCallingChatModel(ctx,
		agentutils.WithTextModel(),
		agentutils.WithMaxTokens(4096),
		agentutils.WithTemperature(0),
	)
	if err != nil {
		return inner
	}
	return agentutils.NewChatModelCompressor(inner, summarizer,
		agentutils.WithCompressThreshold(agentutils.EnvInt("SLIDE_EXECUTOR_COMPRESSOR_MESSAGE_THRESHOLD", 50)),
		agentutils.WithTokenThreshold(agentutils.EnvInt("SLIDE_EXECUTOR_COMPRESSOR_TOKEN_THRESHOLD", 150000)),
		agentutils.WithPreserveCount(agentutils.EnvInt("SLIDE_EXECUTOR_COMPRESSOR_PRESERVE_COUNT", 8)),
	)
}

func BuildSlideExecutorInstruction(workDir, skillsDir string, profile *style.EnhancedProfile) string {
	return buildSlideExecutorInstruction(workDir, skillsDir, profile)
}

func buildSlideExecutorInstruction(workDir, skillsDir string, profile *style.EnhancedProfile) string {
	data := &prompts.TemplateData{
		WorkDir:   workDir,
		SkillsDir: skillsDir,
	}

	// Convert EnhancedProfile to UserPreferences for template
	if profile != nil {
		data.UserPreferences = &prompts.UserPreferences{
			LanguageTone:    profile.LanguageTone,
			PreferredColors:  profile.PreferredColors,
			PreferredLayouts: profile.LayoutPreferences,
			PreferredFonts:  nil, // EnhancedProfile doesn't have font field yet
			PreferredThemes: profile.PreferredThemes,
			AnimationLevel:  profile.AnimationLevel.String(),
			TypicalPageCount: profile.TypicalPageCount,
		}

		// Convert success patterns
		if len(profile.SuccessPatterns) > 0 {
			patterns := make([]prompts.SuccessPatternInfo, 0, len(profile.SuccessPatterns))
			for _, sp := range profile.SuccessPatterns {
				patterns = append(patterns, prompts.SuccessPatternInfo{
					Domain:   sp.Domain,
					Template: sp.Template,
					Theme:    sp.Theme,
					AvgScore: sp.AvgQualityScore,
				})
			}
			data.UserPreferences.SuccessPatterns = patterns
		}

		// Convert brand elements
		if profile.BrandElements.LogoPosition != "" || profile.BrandElements.FooterText != "" {
			data.UserPreferences.BrandElements = &prompts.BrandPreferenceInfo{
				LogoPosition:  profile.BrandElements.LogoPosition,
				FooterText:   profile.BrandElements.FooterText,
				WatermarkText: profile.BrandElements.WatermarkText,
			}
		}

		// Convert chart preferences
		if len(profile.ChartPreferences.PreferredTypes) > 0 {
			data.UserPreferences.ChartPreferences = &prompts.ChartPreferenceInfo{
				PreferredTypes: profile.ChartPreferences.PreferredTypes,
				Use3D:         profile.ChartPreferences.Use3D,
				ColorScheme:   profile.ChartPreferences.ColorScheme,
			}
		}
	}

	instruction, err := prompts.RenderDeepAgent("slide_executor_instruction", data)
	if err != nil {
		panic("failed to render slide executor instruction template: " + err.Error())
	}
	return instruction
}
