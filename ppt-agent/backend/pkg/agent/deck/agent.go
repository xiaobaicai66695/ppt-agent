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
	"github.com/cloudwego/eino/components/model"
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
	chatModel, err := newPlanningChatModel(ctx, cfg, 32768)
	if err != nil {
		return nil, fmt.Errorf("创建主模型失败: %w", err)
	}

	readFileTool := tools.NewReadFileTool(cfg.Operator)
	searchTool := tools.NewSearchTool()
	var imageSearchClient *unsplash.Client
	imageSearchAvailable := false
	if client, err := unsplash.NewClientFromEnv(); err == nil {
		imageSearchClient = client
		imageSearchAvailable = true
	}
	manifestTool := newPlannerManifestTool(cfg.WorkDir, cfg.Outline, cfg.Query)
	plannerTools := []tool.BaseTool{manifestTool, readFileTool, searchTool}
	if imageSearchAvailable {
		plannerTools = append(plannerTools, tools.NewImageSearchTool(imageSearchClient, cfg.WorkDir))
	}

	planner, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "PPTPlanner",
		Description: "PPT 规划代理，无论有无大纲都负责生成完整的 DeckSpec 规划草稿，不负责审查和渲染。",
		Model:       chatModel,
		Instruction: buildPlannerInstruction(cfg.WorkDir, cfg.SkillsDir, cfg.StyleContext, cfg.Outline, cfg.Query, imageSearchAvailable),
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: plannerTools},
		},
		MaxIterations: agentutils.EnvInt("PLANNER_MAX_ITERATIONS", 20),
	})
	if err != nil {
		return nil, fmt.Errorf("创建 Planner Agent 失败: %w", err)
	}

	return planner, nil
}

// NewTaskPlanReviewerAgent 创建独立的 DeckSpec 质量审查与修正 Agent。
// 硬校验、轮次上限和最终提交由 Go workflow 负责。
func NewTaskPlanReviewerAgent(ctx context.Context, cfg *PPTTaskConfig) (adk.Agent, error) {
	chatModel, err := newPlanningChatModel(ctx, cfg, 32768)
	if err != nil {
		return nil, fmt.Errorf("创建 Reviewer 模型失败: %w", err)
	}

	reviewerTools := []tool.BaseTool{
		newDraftTasksPatchTool(cfg.WorkDir),
		tools.NewReadFileTool(cfg.Operator),
	}

	reviewer, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "TaskPlanReviewer",
		Description: "DeckSpec 质量审查代理，只根据审查报告修正 tasks.draft.json，不负责渲染或生成后的定点修复。",
		Model:       chatModel,
		Instruction: buildReviewerInstruction(cfg),
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: reviewerTools},
		},
		MaxIterations: agentutils.EnvInt("PLAN_REVIEWER_MAX_ITERATIONS", 12),
	})
	if err != nil {
		return nil, fmt.Errorf("创建 TaskPlanReviewer Agent 失败: %w", err)
	}
	return reviewer, nil
}

// NewPPTFixerAgent 创建生成后定点修复 Agent。allowedTaskIDs 是唯一允许修改的页面集合。
func NewPPTFixerAgent(ctx context.Context, cfg *PPTTaskConfig, allowedTaskIDs []string) (adk.Agent, error) {
	chatModel, err := newPlanningChatModel(ctx, cfg, 16384)
	if err != nil {
		return nil, fmt.Errorf("创建 Fixer 模型失败: %w", err)
	}

	fixer, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "PPTFixer",
		Description: "PPT 生成后定点修复代理，只修改用户指定页面的 DeckSpec 语义计划。",
		Model:       chatModel,
		Instruction: buildFixerInstruction(cfg),
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: []tool.BaseTool{
				newSelectedTasksPatchTool(cfg.WorkDir, allowedTaskIDs),
				tools.NewReadFileTool(cfg.Operator),
			}},
		},
		MaxIterations: agentutils.EnvInt("PPT_FIXER_MAX_ITERATIONS", 8),
	})
	if err != nil {
		return nil, fmt.Errorf("创建 PPTFixer Agent 失败: %w", err)
	}
	return fixer, nil
}

func newPlanningChatModel(ctx context.Context, cfg *PPTTaskConfig, maxTokens int) (model.ToolCallingChatModel, error) {
	modelOpts := []agentutils.ChatModelOption{
		agentutils.WithMaxTokens(maxTokens),
		agentutils.WithTemperature(0),
		agentutils.WithTopP(0),
	}
	compressorOpts := []agentutils.ChatModelOption{
		agentutils.WithTextModel(),
		agentutils.WithMaxTokens(4096),
		agentutils.WithTemperature(0),
	}
	if strings.TrimSpace(cfg.ModelAPIKey) != "" {
		modelOpts = append(modelOpts, agentutils.WithAPIKeyForProvider(cfg.ModelProvider, cfg.ModelAPIKey))
		compressorOpts = append(compressorOpts, agentutils.WithAPIKeyForProvider(cfg.ModelProvider, cfg.ModelAPIKey))
	}

	chatModel, err := agentutils.NewFallbackToolCallingChatModel(ctx, modelOpts...)
	if err != nil {
		return nil, err
	}

	compressor, err := agentutils.NewFallbackToolCallingChatModel(ctx, compressorOpts...)
	if err != nil {
		return nil, err
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
	return chatModel, nil
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
func buildPlannerInstruction(workDir string, skillsDir string, styleContext string, outline *TaskOutline, query string, imageSearchAvailable bool) string {
	tasksJSON := filepath.Join(workDir, tasksDraftFileName)
	suggestedPageCount := 0
	hasOutline := outline != nil && len(outline.Slides) > 0
	if outline != nil {
		suggestedPageCount = outline.SuggestedPageCount
	}

	data := &prompts.TemplateData{
		TasksJSON:            tasksJSON,
		StyleContext:         styleContext,
		HasOutline:           hasOutline,
		OutlineQuery:         query,
		SuggestedPageCount:   suggestedPageCount,
		SkillsDir:            skillsDir,
		ImageSearchAvailable: imageSearchAvailable,
	}

	instruction, err := prompts.RenderPlanner("master_instruction", data)
	if err != nil {
		panic("failed to render planner instruction template: " + err.Error())
	}
	return instruction
}

func buildReviewerInstruction(cfg *PPTTaskConfig) string {
	data := &prompts.TemplateData{
		TasksJSON:    filepath.Join(cfg.WorkDir, tasksDraftFileName),
		OutlineQuery: cfg.Query,
		SkillsDir:    cfg.SkillsDir,
		StyleContext: cfg.StyleContext,
	}
	instruction, err := prompts.RenderReviewer("master_instruction", data)
	if err != nil {
		panic("failed to render reviewer instruction template: " + err.Error())
	}
	return instruction
}

func buildFixerInstruction(cfg *PPTTaskConfig) string {
	data := &prompts.TemplateData{
		TasksJSON:    filepath.Join(cfg.WorkDir, "tasks.json"),
		OutlineQuery: cfg.Query,
		SkillsDir:    cfg.SkillsDir,
	}
	instruction, err := prompts.RenderFixer("master_instruction", data)
	if err != nil {
		panic("failed to render fixer instruction template: " + err.Error())
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

	// 只把内容意图和规模建议传给 Planner。配色不再由意图/画像推荐；
	// 带背景图的页面由生成器从图片本身提取弱化色系，避免主题色和背景打架。
	if result.Intent != nil {
		var sb strings.Builder
		sb.WriteString(result.Intent.IntentReasoning)
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
	sb.WriteString("- 使用原则: 当前主题、用户显式大纲和显式配色要求优先；历史偏好仅作弱参考，冲突时忽略。\n")
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
			sb.WriteString("  - 同领域历史成功经验:\n")
		}
		for i, sp := range profile.SuccessPatterns {
			if i >= 2 { // 最多显示2个
				break
			}
			if !strings.EqualFold(strings.TrimSpace(sp.Domain), strings.TrimSpace(domain)) {
				continue
			}
			sb.WriteString(fmt.Sprintf("\n  - %s 领域历史质量评分 %.1f", sp.Domain, sp.AvgQualityScore))
		}
	} else if hasSceneSensitivePreferences(profile) {
		sb.WriteString("- 已跳过历史配色、布局和备注等场景敏感偏好：未找到当前领域的历史证据。\n")
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
