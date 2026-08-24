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

package deck

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"

	agentintent "github.com/cloudwego/ppt-agent/pkg/agent/intent"
	agentlearning "github.com/cloudwego/ppt-agent/pkg/agent/learning"
	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/assets/unsplash"
	"github.com/cloudwego/ppt-agent/pkg/prompts"
	"github.com/cloudwego/ppt-agent/pkg/style"
	"github.com/cloudwego/ppt-agent/pkg/tools"
)

func NewPPTPlannerAgent(ctx context.Context, cfg *PPTTaskConfig) (adk.Agent, error) {
	modelOpts := []agentutils.ChatModelOption{
		agentutils.WithMaxTokens(32768),
		agentutils.WithTemperature(0),
		agentutils.WithTopP(0),
	}
	compressorOpts := []agentutils.ChatModelOption{
		agentutils.WithTextModel(),
		agentutils.WithMaxTokens(4096),
		agentutils.WithTemperature(0),
	}
	if strings.TrimSpace(cfg.ModelAPIKey) != "" {
		modelOpts = append(modelOpts, agentutils.WithAPIKey(cfg.ModelAPIKey))
		compressorOpts = append(compressorOpts, agentutils.WithAPIKey(cfg.ModelAPIKey))
	}

	chatModel, err := agentutils.NewFallbackToolCallingChatModel(ctx, modelOpts...)
	if err != nil {
		return nil, fmt.Errorf("创建主模型失败: %w", err)
	}

	compressor, err := agentutils.NewFallbackToolCallingChatModel(ctx, compressorOpts...)
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

	readFileTool := tools.NewReadFileTool(cfg.Operator)
	searchTool := tools.NewSearchTool()
	var imageSearchClient *unsplash.Client
	imageSearchAvailable := false
	if client, err := unsplash.NewClientFromEnv(); err == nil {
		imageSearchClient = client
		imageSearchAvailable = true
	}
	manifestTool := newConfiguredManifestTool(cfg.WorkDir, cfg.SkillsDir, cfg.Outline, cfg.Query, imageSearchAvailable)
	planReviewTool := newPlanReviewTool(cfg.WorkDir)
	plannerTools := []tool.BaseTool{manifestTool, planReviewTool, readFileTool, searchTool}
	if imageSearchAvailable {
		plannerTools = append(plannerTools, tools.NewImageSearchTool(imageSearchClient, cfg.WorkDir))
	}

	planner, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "PPTPlanner",
		Description: "PPT 规划代理，负责在意图分类后生成可执行的 DeckSpec/tasks.json，不负责渲染页面。",
		Model:       chatModel,
		Instruction: buildPlannerInstructionWithImageSearch(cfg.WorkDir, cfg.SkillsDir, cfg.StyleContext, cfg.Outline, cfg.Query, false, getConcurrency(cfg.RoutingDecision), imageSearchAvailable),
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: plannerTools,
			},
		},
		MaxIterations: agentutils.EnvInt("PLANNER_MAX_ITERATIONS", 40),
	})
	if err != nil {
		return nil, fmt.Errorf("创建 Planner Agent 失败: %w", err)
	}

	return planner, nil
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

// buildPlannerInstruction 从模板加载 Planner 指令。
// 当提供大纲时，tasks.json 已经预填充了用户的幻灯片计划。
// query 提供用户原始主题描述用于内容生成。
// concurrency 是后续渲染 worker pool 的并发提示。
func buildPlannerInstruction(workDir string, skillsDir string, styleContext string, outline *TaskOutline, query string, enableQA bool, concurrency int) string {
	return buildPlannerInstructionWithImageSearch(
		workDir,
		skillsDir,
		styleContext,
		outline,
		query,
		enableQA,
		concurrency,
		unsplash.IsConfigured(),
	)
}

func buildPlannerInstructionWithImageSearch(workDir string, skillsDir string, styleContext string, outline *TaskOutline, query string, enableQA bool, concurrency int, imageSearchAvailable bool) string {
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
	suggestedPageCount := 0
	outlineQuery := query // user's original topic description
	hasOutline := outline != nil && len(outline.Slides) > 0
	hasStyleRecommendation := outline != nil && outline.ContentMode == OutlineContentModeRecommendedStyle
	if outline != nil {
		outlineTemplate = outline.Template
		outlineTheme = outline.Theme
		outlineTitle = outline.Title
		outlineContentMode = outline.ContentMode
		outlineUseBackground = outline.UseBackground
		if outline.UseBackground {
			outlineBackground = outline.RecommendedBackground
		}
		suggestedPageCount = outline.SuggestedPageCount
	}

	data := &prompts.TemplateData{
		TasksJSON:              tasksJSON,
		TemplateCatalog:        templateCatalog,
		StyleContext:           styleContext,
		HasOutline:             hasOutline,
		HasStyleRecommendation: hasStyleRecommendation,
		OutlineQuery:           outlineQuery,
		OutlineTemplate:        outlineTemplate,
		OutlineTheme:           outlineTheme,
		OutlineTitle:           outlineTitle,
		OutlineContentMode:     outlineContentMode,
		OutlineUseBackground:   outlineUseBackground,
		OutlineBackground:      outlineBackground,
		SuggestedPageCount:     suggestedPageCount,
		AvailableBackgrounds:   "",
		SkillsDir:              skillsDir,
		EnableQA:               enableQA,
		Concurrency:            concurrency,
		ImageSearchAvailable:   imageSearchAvailable,
	}

	instruction, err := prompts.RenderPlanner("master_instruction", data)
	if err != nil {
		panic("failed to render planner instruction template: " + err.Error())
	}
	return instruction
}

// ── 意图路由与只读画像引擎 ─────────────────────────────────────────────────

var globalLearningEngine *agentlearning.Engine
var learningEngineOnce sync.Once
var learningEngineFactory interface{}
var textLearningEngineFactory interface{}

// InitLearningEngine 初始化全局意图路由与只读画像引擎（由 main.go 在 modelFactory 可用后调用）
// factory 是 ServerConfig.AIModelFactory 的工厂函数
// textFactory 是 ServerConfig.TextModelFactory 的工厂函数（轻量级模型，节省意图分类成本）
func InitLearningEngine(factory, textFactory interface{}) {
	learningEngineFactory = factory
	textLearningEngineFactory = textFactory
}

// GetLearningEngine 获取全局意图路由与只读画像引擎实例
func GetLearningEngine() *agentlearning.Engine {
	learningEngineOnce.Do(func() {
		cfg := &agentlearning.EngineConfig{
			EnableLLMClassification: true,
			EnableLearning:          false,
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
		domain := ""
		if result.Intent != nil {
			domain = result.Intent.Domain.String()
		}
		cfg.StyleContext = enhanceStyleContextWithProfile(cfg.StyleContext, result.Profile, domain)
	}

	return cfg, nil
}

// enhanceStyleContextWithProfile 根据当前领域筛选用户画像后增强 StyleContext。
// 历史偏好只是参考信号；与当前主题、outline 或显式模板选择冲突时必须让位。
func enhanceStyleContextWithProfile(baseContext string, profile *style.EnhancedProfile, domain string) string {
	if profile == nil {
		return baseContext
	}

	var sb strings.Builder
	if strings.TrimSpace(baseContext) != "" {
		sb.WriteString(strings.TrimSpace(baseContext))
		sb.WriteString("\n")
	}
	if lines := profile.UserFacts.PromptLines(); len(lines) > 0 {
		sb.WriteString("\n【用户确定性资料】\n")
		sb.WriteString("- 使用原则: 这些资料可直接作为称谓、署名、组织背景和工作场景上下文；若当前任务另有明确说明，以当前任务为准。\n")
		sb.WriteString(strings.Join(lines, "\n"))
		sb.WriteString("\n")
	}
	sb.WriteString("\n【用户偏好参考】\n")
	sb.WriteString("- 使用原则: 当前主题、用户显式大纲、显式模板/配色选择优先；历史偏好仅作弱参考，冲突时忽略。\n")
	if strings.TrimSpace(domain) != "" && !strings.EqualFold(domain, "unknown") {
		sb.WriteString(fmt.Sprintf("- 当前识别领域: %s\n", domain))
	}

	// 低敏偏好：可跨领域参考。
	if profile.LanguageTone != "" {
		sb.WriteString(fmt.Sprintf("- 语言风格参考: %s\n", profile.LanguageTone))
	}
	if profile.TypicalPageCount > 0 {
		sb.WriteString(fmt.Sprintf("- 常用页数参考: %d 页；如果当前需求指定页数，以当前需求为准。\n", profile.TypicalPageCount))
	}
	if profile.AnimationLevel != style.AnimationNone {
		sb.WriteString(fmt.Sprintf("- 动画强度参考: %s\n", profile.AnimationLevel.String()))
	}

	// 高敏偏好：只有同领域历史才注入，避免跨场景迁移。
	if hasExactDomainHistory(profile, domain) {
		sb.WriteString("- 同领域历史可参考项:\n")
		if theme := profile.GetPreferredThemeForDomain(domain); theme != "" {
			sb.WriteString(fmt.Sprintf("  - 历史常用主题/模板风格: %s\n", theme))
		}
		if len(profile.PreferredColors) > 0 {
			sb.WriteString(fmt.Sprintf("  - 历史配色参考: %s\n", profile.PreferredColors[0]))
		}
		if len(profile.LayoutPreferences) > 0 {
			sb.WriteString(fmt.Sprintf("  - 历史布局参考: %s\n", profile.LayoutPreferences[0]))
		}
		if len(profile.SpecialNotes) > 0 {
			notes := profile.SpecialNotes
			if len(notes) > 3 {
				notes = notes[len(notes)-3:]
			}
			sb.WriteString(fmt.Sprintf("  - 历史备注参考: %s\n", strings.Join(notes, "；")))
		}
		if len(profile.SuccessPatterns) > 0 {
			sb.WriteString("  - 历史成功经验:\n")
		}
		for i, sp := range profile.SuccessPatterns {
			if i >= 2 { // 最多显示2个
				break
			}
			if !strings.EqualFold(strings.TrimSpace(sp.Domain), strings.TrimSpace(domain)) {
				continue
			}
			sb.WriteString(fmt.Sprintf("\n  • %s领域使用%s模板，评分%.1f",
				sp.Domain, sp.Template, sp.AvgQualityScore))
		}
	} else if hasSceneSensitivePreferences(profile) {
		sb.WriteString("- 已跳过历史模板/配色/布局/备注等场景敏感偏好：未找到当前领域的历史证据。\n")
	}

	// 显式品牌资产可跨领域复用，但仍低于当前任务中用户显式选择。
	if profile.BrandElements.LogoPosition != "" {
		sb.WriteString(fmt.Sprintf("- 品牌 Logo 位置参考: %s\n", profile.BrandElements.LogoPosition))
	}
	if profile.BrandElements.FooterText != "" {
		sb.WriteString(fmt.Sprintf("- 品牌页脚参考: %s\n", profile.BrandElements.FooterText))
	}

	// 图表偏好偏向表达习惯，可作为弱参考。
	if len(profile.ChartPreferences.PreferredTypes) > 0 {
		sb.WriteString(fmt.Sprintf("- 图表类型参考: %s\n", profile.ChartPreferences.PreferredTypes[0]))
	}

	return strings.TrimSpace(sb.String())
}

func hasExactDomainHistory(profile *style.EnhancedProfile, domain string) bool {
	if profile == nil || strings.TrimSpace(domain) == "" || strings.EqualFold(domain, "unknown") {
		return false
	}
	return profile.HasExactDomainHistory(domain)
}

func hasSceneSensitivePreferences(profile *style.EnhancedProfile) bool {
	if profile == nil {
		return false
	}
	return len(profile.PreferredThemes) > 0 ||
		len(profile.PreferredColors) > 0 ||
		len(profile.LayoutPreferences) > 0 ||
		len(profile.SuccessPatterns) > 0 ||
		len(profile.SpecialNotes) > 0
}
