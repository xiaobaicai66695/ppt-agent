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
	"strings"
	"sync"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"

	agentintent "github.com/cloudwego/ppt-agent/pkg/agent/intent"
	agentlearning "github.com/cloudwego/ppt-agent/pkg/agent/learning"
	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/prompts"
	"github.com/cloudwego/ppt-agent/pkg/style"
	"github.com/cloudwego/ppt-agent/pkg/tools"
)

func NewPPTTaskDeepAgent(ctx context.Context, cfg *PPTTaskConfig) (adk.Agent, error) {
	chatModel, err := agentutils.NewFallbackToolCallingChatModel(ctx,
		agentutils.WithMaxTokens(32768),
		agentutils.WithTemperature(0),
		agentutils.WithTopP(0),
	)
	if err != nil {
		return nil, fmt.Errorf("创建主模型失败: %w", err)
	}

	compressor, err := agentutils.NewFallbackToolCallingChatModel(ctx,
		agentutils.WithTextModel(),
		agentutils.WithMaxTokens(4096),
		agentutils.WithTemperature(0),
	)
	if err != nil {
		return nil, fmt.Errorf("创建压缩器模型失败: %w", err)
	}
	chatModel = agentutils.NewChatModelCompressor(chatModel, compressor,
		agentutils.WithCompressThreshold(60),
		agentutils.WithTokenThreshold(200000),
		agentutils.WithPreserveCount(8),
	)
	if cfg.CompressorTracker != nil {
		if compressor, ok := chatModel.(*agentutils.ChatModelCompressor); ok {
			compressor.SetTracker(cfg.CompressorTracker)
		}
	}
	if cfg.RuntimeMeta != nil {
		if compressor, ok := chatModel.(*agentutils.ChatModelCompressor); ok {
			compressor.SetRuntimeMeta(cfg.RuntimeMeta)
		}
	}
	chatModel = agentutils.NewRuntimeStatusChatModel(chatModel, cfg.RuntimeMeta)

	slideExecutor, err := newSlideExecutorAgent(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("创建 SlideExecutor 子代理失败: %w", err)
	}

	readFileTool := tools.NewReadFileTool(cfg.Operator)
	manifestTool := newManifestTool(cfg.WorkDir)
	searchTool := tools.NewSearchTool()
	subAgents := []adk.Agent{slideExecutor}

	deepAgent, err := deep.New(ctx, &deep.Config{
		Name:        "PPTTaskDeepAgent",
		Description: "PPT 任务调度代理，负责规划并并行生成 PPT 幻灯片",
		ChatModel:   chatModel,
		Instruction: buildDeepAgentInstruction(cfg.WorkDir, cfg.SkillsDir, cfg.StyleContext, cfg.Outline, cfg.Query, false, getConcurrency(cfg.RoutingDecision)),
		SubAgents:   subAgents,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{manifestTool, readFileTool, searchTool},
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

// getConcurrency 从路由决策获取并发数，默认 5
func getConcurrency(route *agentintent.RoutingDecision) int {
	if route != nil && route.Concurrency > 0 {
		// 限制在合理范围
		if route.Concurrency > 10 {
			return 10
		}
		return route.Concurrency
	}
	return 5
}

// buildDeepAgentInstruction 从模板加载深度代理的主指令
// 当提供大纲时，tasks.json 已经预填充了用户的幻灯片计划
// query 提供用户原始主题描述用于内容生成
// concurrency 每批最大并发页数（来自路由决策）
func buildDeepAgentInstruction(workDir string, skillsDir string, styleContext string, outline *TaskOutline, query string, enableQA bool, concurrency int) string {
	tmplDir := filepath.Join(skillsDir, "visual_designer", "templates", "full-decks")
	tasksJSON := filepath.Join(workDir, "tasks.json")

	const templateCatalog = `| 模板文件 | 适用场景 | 页数 |
|---------|---------|------|
| tech-intro.json | 新技术介绍、行业科普、知识分享，从基础概念到应用实践，适合非技术受众 | 18页 |
| tech-sharing.json | 内部技术分享、技术培训、架构讲解，有章节划分，注重内容深度 | 18页 |
| product-launch.json | 新产品发布会、产品宣讲、客户演示，强调价值主张和差异化优势 | 14页 |
| weekly-report.json | 团队周报、项目月报、工作汇报，简洁高效，数据驱动 | 9页 |
| pitch-deck.json | 创业路演，投资人演示、商业计划，逻辑严密，数据驱动，说服力强 | 16页 |
| course-module.json | 教学课件、培训材料、知识分享，内容系统化，便于学习理解 | 17页 |
| current-affairs.json | 时政热点分析、政策解读、国际形势分析，稳重专业，数据支撑强 | 14页 |
| politics-ideology.json | 思政教育、团课培训、爱国主义教育，价值观明确，结构清晰 | 16页 |
| design-defense.json | 课程设计、毕业设计、项目答辩，逻辑清晰，技术扎实 | 12页 |
| innovation-compete.json | 大创/挑战杯/互联网+等科创竞赛汇报，创新性强，数据支撑 | 16页 |
| research-report.json | 市场调研、行业分析、可行性研究，数据详实，结论明确 | 14页 |
| activity-plan.json | 团建活动、校园活动、节日策划，活泼有创意，执行清晰 | 10页 |
| personal-summary.json | 个人总结、述职报告、年终总结，重点突出，成果可见 | 10页 |
| short-class-talk.json | 课堂5-10分钟短时分享、课题介绍，精简高效，快速传达 | 6页 |
| meeting-minutes.json | 会议记录、工作例会、项目评审会，结构清晰，行动明确 | 8页 |
| product-intro.json | 产品介绍、客户演示、功能展示，突出价值，增强信任 | 12页 |
| training-course.json | 内部培训、新人入职培训、技能培训，知识系统，互动引导 | 16页 |
| project-proposal.json | 新项目立项、项目申请、资源申请，理由充分，方案可行 | 12页 |`

	// 当用户提供了大纲时，tasks.json 已经包含 template/theme
	// 读取它以便告知代理使用什么
	outlineTemplate := ""
	outlineTheme := ""
	outlineTitle := ""
	outlineContentMode := ""
	outlineUseBackground := false
	outlineBackground := ""
	outlineQuery := query // user's original topic description
	hasOutline := outline != nil && len(outline.Slides) > 0
	if hasOutline {
		outlineTemplate = outline.Template
		outlineTheme = outline.Theme
		outlineTitle = outline.Title
		outlineContentMode = outline.ContentMode
		outlineUseBackground = outline.UseBackground
		if outline.UseBackground {
			outlineBackground = outline.RecommendedBackground
		}
	}

	data := &prompts.TemplateData{
		TmplDir:              tmplDir,
		TasksJSON:            tasksJSON,
		TemplateCatalog:      templateCatalog,
		StyleContext:         styleContext,
		HasOutline:           hasOutline,
		OutlineQuery:         outlineQuery,
		OutlineTemplate:      outlineTemplate,
		OutlineTheme:         outlineTheme,
		OutlineTitle:         outlineTitle,
		OutlineContentMode:   outlineContentMode,
		OutlineUseBackground: outlineUseBackground,
		OutlineBackground:    outlineBackground,
		SkillsDir:            skillsDir,
		EnableQA:             enableQA,
		Concurrency:          concurrency,
	}

	instruction, err := prompts.RenderDeepAgent("master_instruction", data)
	if err != nil {
		panic("failed to render deep agent master instruction template: " + err.Error())
	}
	return instruction
}

// ── 智能学习引擎 ──────────────────────────────────────────────────────────

var globalLearningEngine *agentlearning.Engine
var learningEngineOnce sync.Once
var learningEngineFactory interface{}
var textLearningEngineFactory interface{}

// InitLearningEngine 初始化全局学习引擎（由 main.go 在 modelFactory 可用后调用）
// factory 是 ServerConfig.AIModelFactory 的工厂函数
// textFactory 是 ServerConfig.TextModelFactory 的工厂函数（轻量级模型，节省意图分类成本）
func InitLearningEngine(factory, textFactory interface{}) {
	learningEngineFactory = factory
	textLearningEngineFactory = textFactory
}

// GetLearningEngine 获取全局学习引擎实例
func GetLearningEngine() *agentlearning.Engine {
	learningEngineOnce.Do(func() {
		cfg := &agentlearning.EngineConfig{
			EnableLLMClassification: true,
			EnableLearning:          true,
			EnableProfileMatch:      true,
		}
		globalLearningEngine = agentlearning.NewEngine(cfg, learningEngineFactory, textLearningEngineFactory)
	})
	return globalLearningEngine
}

// ProcessUserIntent 处理用户意图
// 在任务开始前调用，用于意图识别和路由决策
func ProcessUserIntent(ctx context.Context, query string, userID int) (*PPTTaskConfig, error) {
	engine := GetLearningEngine()
	if engine == nil {
		return nil, nil
	}

	result, err := engine.ProcessTask(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	// 构建配置
	cfg := &PPTTaskConfig{
		UserID:          userID,
		Query:           query,
		IntentResult:    result.Intent,
		RoutingDecision: result.Routing,
		EnhancedProfile: result.Profile,
		// LearningEngine 通过全局单例 GetLearningEngine() 获取
	}

	// 如果有推荐的模板/主题，设置到 StyleContext
	if result.Intent != nil {
		var sb strings.Builder
		sb.WriteString(result.Intent.IntentReasoning)
		if len(result.Intent.SuggestedTemplates) > 0 {
			sb.WriteString("\n推荐模板: ")
			sb.WriteString(strings.Join(result.Intent.SuggestedTemplates, ", "))
		}
		if result.Intent.SuggestedTheme != "" {
			sb.WriteString("\n推荐配色: ")
			sb.WriteString(result.Intent.SuggestedTheme)
		}
		if result.Intent.SuggestedPageCount > 0 {
			sb.WriteString(fmt.Sprintf("\n推荐页数: %d", result.Intent.SuggestedPageCount))
		}
		cfg.StyleContext = sb.String()
	}

	// 如果有用户画像，增强 StyleContext
	if result.Profile != nil {
		cfg.StyleContext = enhanceStyleContextWithProfile(cfg.StyleContext, result.Profile)
	}

	return cfg, nil
}

// RecordUserFeedback 记录用户反馈
func RecordUserFeedback(userID int, taskID string, feedback *agentlearning.Feedback) {
	engine := GetLearningEngine()
	if engine != nil {
		engine.RecordFeedback(userID, taskID, feedback)
	}
}

// UpdateUserProfileFromTask 从任务更新用户画像
func UpdateUserProfileFromTask(userID int, task *agentlearning.TaskContext) {
	engine := GetLearningEngine()
	if engine != nil {
		engine.UpdateProfileFromTask(userID, task)
	}
}

// GetUserRecommendations 获取用户推荐
func GetUserRecommendations(userID int, domain string) *style.RecommendResult {
	engine := GetLearningEngine()
	if engine != nil {
		return engine.GetRecommendations(userID, domain)
	}
	return nil
}

// GetUserInsights 获取用户洞察
func GetUserInsights(userID int) *agentlearning.InsightsReport {
	engine := GetLearningEngine()
	if engine != nil {
		return engine.GetUserInsights(userID)
	}
	return nil
}

// enhanceStyleContextWithProfile 根据用户画像增强 StyleContext
func enhanceStyleContextWithProfile(baseContext string, profile *style.EnhancedProfile) string {
	if profile == nil {
		return baseContext
	}

	var sb strings.Builder
	sb.WriteString(baseContext)
	sb.WriteString("\n\n【用户历史偏好】")

	// 语言风格偏好
	if profile.LanguageTone != "" {
		sb.WriteString(fmt.Sprintf("\n- 语言风格: %s", profile.LanguageTone))
	}

	// 配色偏好
	if len(profile.PreferredColors) > 0 {
		sb.WriteString(fmt.Sprintf("\n- 偏好配色: %s", profile.PreferredColors[0]))
	}

	// 布局偏好
	if len(profile.LayoutPreferences) > 0 {
		sb.WriteString(fmt.Sprintf("\n- 偏好布局: %s", profile.LayoutPreferences[0]))
	}

	// 模板偏好
	if len(profile.PreferredThemes) > 0 {
		sb.WriteString(fmt.Sprintf("\n- 偏好模板风格: %s", profile.PreferredThemes[0]))
	}

	// 动画偏好
	if profile.AnimationLevel != style.AnimationNone {
		sb.WriteString(fmt.Sprintf("\n- 动画偏好: %s", profile.AnimationLevel.String()))
	}

	// 成功模式
	if len(profile.SuccessPatterns) > 0 {
		sb.WriteString("\n- 历史成功经验:")
		for i, sp := range profile.SuccessPatterns {
			if i >= 2 { // 最多显示2个
				break
			}
			sb.WriteString(fmt.Sprintf("\n  • %s领域使用%s模板，评分%.1f",
				sp.Domain, sp.Template, sp.AvgQualityScore))
		}
	}

	// 品牌元素
	if profile.BrandElements.LogoPosition != "" {
		sb.WriteString(fmt.Sprintf("\n- Logo位置: %s", profile.BrandElements.LogoPosition))
	}
	if profile.BrandElements.FooterText != "" {
		sb.WriteString(fmt.Sprintf("\n- 页脚文字: %s", profile.BrandElements.FooterText))
	}

	// 图表偏好
	if len(profile.ChartPreferences.PreferredTypes) > 0 {
		sb.WriteString(fmt.Sprintf("\n- 偏好图表类型: %s", profile.ChartPreferences.PreferredTypes[0]))
	}

	// 典型页数
	if profile.TypicalPageCount > 0 {
		sb.WriteString(fmt.Sprintf("\n- 典型页数: %d页", profile.TypicalPageCount))
	}

	return sb.String()
}
