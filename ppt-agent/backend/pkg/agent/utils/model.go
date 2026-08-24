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

package utils

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	arkmodel "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"

	"github.com/cloudwego/ppt-agent/pkg/logger"
)

// ChatModelConfig ChatModel 配置选项
type ChatModelConfig struct {
	MaxTokens       *int
	Temperature     *float32
	TopP            *float32
	DisableThinking *bool
	JsonSchema      *openai.ChatCompletionResponseFormatJSONSchema
	Model           *string
	APIKey          *string
}

// ChatModelOption ChatModel 配置函数
type ChatModelOption func(*ChatModelConfig)

func WithMaxTokens(tokens int) ChatModelOption {
	return func(c *ChatModelConfig) {
		c.MaxTokens = &tokens
	}
}

func WithTemperature(temp float32) ChatModelOption {
	return func(c *ChatModelConfig) {
		c.Temperature = &temp
	}
}

func WithTopP(topP float32) ChatModelOption {
	return func(c *ChatModelConfig) {
		c.TopP = &topP
	}
}

func WithDisableThinking(disable bool) ChatModelOption {
	return func(c *ChatModelConfig) {
		c.DisableThinking = &disable
	}
}

func WithResponseFormatJsonSchema(schema *openai.ChatCompletionResponseFormatJSONSchema) ChatModelOption {
	return func(c *ChatModelConfig) {
		c.JsonSchema = schema
	}
}

// WithModel 覆盖 NewFallbackToolCallingChatModel 使用的模型名称。
// 设置后，ARK_MODEL / OPENAI_MODEL 环境变量将被忽略。
// 适用于轻量级辅助任务（如意图分类、风格提取），应使用比主 agent 更小/更便宜的模型。
func WithModel(modelName string) ChatModelOption {
	return func(c *ChatModelConfig) {
		c.Model = &modelName
	}
}

func WithAPIKey(apiKey string) ChatModelOption {
	return func(c *ChatModelConfig) {
		apiKey = strings.TrimSpace(apiKey)
		if apiKey != "" {
			c.APIKey = &apiKey
		}
	}
}

// WithTextModel 覆盖 NewFallbackToolCallingChatModel 用于纯文本任务的模型链。
// 使用 ARK_TEXT_MODEL，备用为 ARK_MODEL_BACKUP*。
// 如果 ARK_TEXT_MODEL 未设置，则回退到 ARK_MODEL。
func WithTextModel() ChatModelOption {
	textModel := os.Getenv("ARK_TEXT_MODEL")
	if textModel == "" {
		textModel = os.Getenv("ARK_MODEL")
	}
	return func(c *ChatModelConfig) {
		c.Model = &textModel
	}
}

// NewQAModel 创建一个不带降级的独立 QA 模型。
// 使用 ARK_QA_MODEL；遇到瞬时失败时最多重试 3 次。
// QA 质量不能因模型降级而受损——不使用降级链。
func NewQAModel(ctx context.Context, opts ...ChatModelOption) (model.ToolCallingChatModel, error) {
	qaModel := os.Getenv("ARK_QA_MODEL")
	if qaModel == "" {
		return nil, fmt.Errorf("ARK_QA_MODEL 环境变量未设置")
	}
	baseCfg := &ChatModelConfig{}
	for _, opt := range opts {
		opt(baseCfg)
	}
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		maxTokens := 32768
		temp := float32(0)
		topP := float32(0)
		disableThink := true
		cfg := *baseCfg
		cfg.MaxTokens = &maxTokens
		cfg.Temperature = &temp
		cfg.TopP = &topP
		cfg.DisableThinking = &disableThink
		m, err := newSingleModel(ctx, qaModel, &cfg)
		if err == nil {
			return m, nil
		}
		lastErr = err
		logger.Warn("qa_model_init_retry", "attempt", attempt, "model", qaModel, "error", err.Error())
	}
	return nil, fmt.Errorf("QA 模型 [%s] 创建失败（已重试 3 次）: %w", qaModel, lastErr)
}

func GetCurrentTime() string {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return time.Now().In(loc).Format("2006-01-02 15:04:05")
}

// --- 模型降级管理 ---

const (
	fallbackPauseDuration = 30 * time.Second
)

// EnvInt 从环境变量读取整数，如果未设置或无效则返回默认值。
func EnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultVal
}

// IsRateLimitError 判断错误是否为 429 限流错误
func IsRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "rate_limit") ||
		strings.Contains(errStr, "too many requests")
}

// globalRateLimitTracker 全局 429 限流协调器，所有 FallbackChatModel 实例共享。
// 当任一 Agent 触发 429 后，所有使用相同 modelName 的 Agent 一起等待，避免惊群效应。
// 注意：两个使用相同原始模型名的 FallbackChatModel 实例共享暂停状态，
// 即使它们的配置不同（MaxTokens、Temperature 等）。这意味着
// 低优先级任务的限流可能会延迟高优先级任务。如果需要严格隔离，
// 需要在模型名前加上作用域标签。
type globalRateLimitTracker struct {
	mu            sync.Mutex
	pauseEndTimes map[string]time.Time // key: 原始 modelName（如 Qwen/Qwen3.5-122B-A10B）
}

var globalTracker = &globalRateLimitTracker{
	pauseEndTimes: make(map[string]time.Time),
}

// checkPause 检查指定模型是否在全局暂停中。返回 (是否暂停, 暂停结束时间)。
func (g *globalRateLimitTracker) checkPause(modelName string) (bool, time.Time) {
	g.mu.Lock()
	defer g.mu.Unlock()
	end, ok := g.pauseEndTimes[modelName]
	if !ok {
		return false, time.Time{}
	}
	if time.Now().Before(end) {
		return true, end
	}
	delete(g.pauseEndTimes, modelName)
	return false, time.Time{}
}

// markRateLimited 标记指定模型触发 429，全局暂停 baseDuration+随机抖动。
// 抖动防止多个实例在暂停结束后同时恢复、再次触发限流。
func (g *globalRateLimitTracker) markRateLimited(modelName string, baseDuration time.Duration) {
	g.mu.Lock()
	defer g.mu.Unlock()
	jitter := time.Duration(rand.Int64N(int64(baseDuration / 4))) // 0~25% 随机抖动
	g.pauseEndTimes[modelName] = time.Now().Add(baseDuration + jitter)
	logger.Info("model_global_paused", "model", modelName, "pause", (baseDuration + jitter).String())
}

// FallbackChatModel 包装多个模型，支持 429 后降级和全局暂停。
// 多个 FallbackChatModel 实例通过 globalTracker 共享暂停状态。
type FallbackChatModel struct {
	models        []model.ToolCallingChatModel
	modelNames    []string // 日志显示名（含 backup 后缀）
	rawNames      []string // 原始 modelName，用于全局追踪
	mu            sync.Mutex
	pauseDuration time.Duration
	// compressorCfg 在 FallbackChatModel 被 ChatModelCompressor 包装时存储。
	// 调用 WithTools 时，会重新应用压缩器到新的模型实例。
	compressorCfg *compressorConfig
}

// NewFallbackToolCallingChatModel 创建支持降级的 ChatModel
// 会依次尝试 ARK_MODEL、ARK_MODEL_BACKUP1、ARK_MODEL_BACKUP2
// 遇到 429 时：当前模型暂停 30s 并尝试下一个模型
// 所有模型都失败后才返回错误
func NewFallbackToolCallingChatModel(ctx context.Context, opts ...ChatModelOption) (model.ToolCallingChatModel, error) {
	o := &ChatModelConfig{}
	for _, opt := range opts {
		opt(o)
	}

	modelNames := []string{
		os.Getenv("ARK_MODEL"),
		os.Getenv("ARK_MODEL_BACKUP1"),
		os.Getenv("ARK_MODEL_BACKUP2"),
		os.Getenv("ARK_MODEL_BACKUP3"),
		os.Getenv("ARK_MODEL_BACKUP4"),
	}

	// WithModel overrides the primary model (e.g. for compressor, style, text tasks).
	if o.Model != nil && *o.Model != "" {
		modelNames = []string{*o.Model}
	}

	var validModels []model.ToolCallingChatModel
	var validNames []string
	var rawNames []string

	for i, name := range modelNames {
		if name != "" {
			cm, err := newSingleModel(ctx, name, o)
			if err != nil {
				logger.Warn("model_init_failed", "model", name, "error", err.Error())
				continue
			}
			validModels = append(validModels, cm)
			validNames = append(validNames, fmt.Sprintf("%s(backup-%d)", name, i))
			rawNames = append(rawNames, name)
			logger.Info("model_init_success", "model", name, "backup", i)
		}
	}

	if len(validModels) == 0 {
		return nil, fmt.Errorf("没有任何可用模型")
	}

	return &FallbackChatModel{
		models:        validModels,
		modelNames:    validNames,
		rawNames:      rawNames,
		pauseDuration: fallbackPauseDuration,
	}, nil
}

// newSingleModel 根据模型名称创建单个模型
func newSingleModel(ctx context.Context, modelName string, cfg *ChatModelConfig) (model.ToolCallingChatModel, error) {
	conf := &ark.ChatModelConfig{
		APIKey:      modelAPIKeyFromConfig(cfg),
		BaseURL:     os.Getenv("ARK_BASE_URL"),
		Region:      os.Getenv("ARK_REGION"),
		Model:       modelName,
		MaxTokens:   cfg.MaxTokens,
		Temperature: cfg.Temperature,
		TopP:        cfg.TopP,
	}

	// DeepSeek 系列模型默认开启 thinking 模式，流式输出中会夹杂 reasoning_content，
	// 导致 eino ReAct 框架解析 tool call JSON 失败。这里自动禁用。
	// Qwen3.5/3.6, DeepSeek, Kimi-K2 等 thinking 模型会消耗 token 预算用于 reasoning_content，
	// 在 ReAct 中破坏 tool call JSON 解析。所有此类模型在工具调用场景下禁用 thinking。
	shouldDisable := strings.Contains(strings.ToLower(modelName), "deepseek") ||
		strings.Contains(strings.ToLower(modelName), "qwen3.5") ||
		strings.Contains(strings.ToLower(modelName), "qwen3.6") ||
		strings.Contains(strings.ToLower(modelName), "kimi-k2")
	if shouldDisable {
		conf.Thinking = &arkmodel.Thinking{
			Type: arkmodel.ThinkingTypeDisabled,
		}
	}
	if cfg.DisableThinking != nil && *cfg.DisableThinking {
		conf.Thinking = &arkmodel.Thinking{
			Type: arkmodel.ThinkingTypeDisabled,
		}
	}

	if cfg.JsonSchema != nil {
		conf.ResponseFormat = &ark.ResponseFormat{
			Type: arkmodel.ResponseFormatJSONSchema,
			JSONSchema: &arkmodel.ResponseFormatJSONSchemaJSONSchemaParam{
				Name:        cfg.JsonSchema.Name,
				Description: cfg.JsonSchema.Description,
				Schema:      cfg.JsonSchema.JSONSchema,
				Strict:      cfg.JsonSchema.Strict,
			},
		}
	}

	return ark.NewChatModel(ctx, conf)
}

func modelAPIKeyFromConfig(cfg *ChatModelConfig) string {
	if cfg != nil && cfg.APIKey != nil && strings.TrimSpace(*cfg.APIKey) != "" {
		return strings.TrimSpace(*cfg.APIKey)
	}
	return os.Getenv("ARK_API_KEY")
}

func (f *FallbackChatModel) shouldPause(idx int) (bool, time.Time) {
	name := f.globalTrackerKey(idx)
	if name == "" {
		return false, time.Time{}
	}
	return globalTracker.checkPause(name)
}

func (f *FallbackChatModel) markRateLimited(idx int) {
	name := f.globalTrackerKey(idx)
	if name == "" {
		return
	}
	globalTracker.markRateLimited(name, f.pauseDuration)
}

// globalTrackerKey 返回用于全局追踪的 modelName。
// 优先使用 rawNames（精确匹配），回退到 modelNames（兼容测试直接构造的实例）。
func (f *FallbackChatModel) globalTrackerKey(idx int) string {
	if idx < len(f.rawNames) {
		return f.rawNames[idx]
	}
	if idx < len(f.modelNames) {
		return f.modelNames[idx]
	}
	return ""
}

func (f *FallbackChatModel) callWithFallback(ctx context.Context, callFn func(idx int) (*schema.Message, error)) (*schema.Message, error) {
	for idx := 0; idx < len(f.models); idx++ {
		paused, pauseEnd := f.shouldPause(idx)
		if paused {
			// 检查是否还有其他可用模型（不被全局暂停的）
			hasAlternative := false
			for j := idx + 1; j < len(f.models); j++ {
				if p, _ := f.shouldPause(j); !p {
					hasAlternative = true
					break
				}
			}
			if hasAlternative {
				logger.Debug("model_paused_skipping", "model", f.modelNames[idx])
				continue
			}
			// 所有备选模型都在全局暂停中，等待当前模型恢复
			remaining := time.Until(pauseEnd).Round(time.Second)
			logger.Info("model_all_paused_waiting", "model", f.modelNames[idx], "remaining", remaining.String())
			select {
			case <-time.After(time.Until(pauseEnd)):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		msg, err := callFn(idx)
		if err != nil {
			if IsRateLimitError(err) {
				f.markRateLimited(idx)
				continue
			}
			return msg, err
		}
		return msg, nil
	}

	return nil, fmt.Errorf("所有模型均失败")
}

// PrimaryModelName 返回降级链中主模型（第一个）的名称。
func (f *FallbackChatModel) PrimaryModelName() string {
	if len(f.rawNames) > 0 {
		return f.rawNames[0]
	}
	if len(f.modelNames) > 0 {
		return f.modelNames[0]
	}
	return ""
}

// Generate 实现 model.ToolCallingChatModel 接口
func (f *FallbackChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return f.callWithFallback(ctx, func(idx int) (*schema.Message, error) {
		msgs := make([]*schema.Message, len(messages))
		copy(msgs, messages)
		return f.models[idx].Generate(ctx, msgs, opts...)
	})
}

// Stream 实现 model.ToolCallingChatModel 接口
func (f *FallbackChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	for idx := 0; idx < len(f.models); idx++ {
		paused, pauseEnd := f.shouldPause(idx)
		if paused {
			hasAlternative := false
			for j := idx + 1; j < len(f.models); j++ {
				if p, _ := f.shouldPause(j); !p {
					hasAlternative = true
					break
				}
			}
			if hasAlternative {
				logger.Debug("model_paused_skipping", "model", f.modelNames[idx])
				continue
			}
			remaining := time.Until(pauseEnd).Round(time.Second)
			logger.Info("model_all_paused_waiting", "model", f.modelNames[idx], "remaining", remaining.String())
			select {
			case <-time.After(time.Until(pauseEnd)):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		msgs := make([]*schema.Message, len(messages))
		copy(msgs, messages)
		stream, err := f.models[idx].Stream(ctx, msgs, opts...)
		if err != nil {
			if IsRateLimitError(err) {
				f.markRateLimited(idx)
				continue
			}
			return nil, err
		}
		return stream, nil
	}

	return nil, fmt.Errorf("所有模型均失败")
}

// WithTools 实现 model.ToolCallingChatModel 接口
// 每个底层模型绑定到相同的工具列表。
// 如果此 FallbackChatModel 是通过压缩器包装模型的 WithTools 创建的，
// 压缩器包装器会被保留（通过使用存储的压缩器配置重新包装新模型）。
// 如果没有存储压缩器配置，则返回普通 FallbackChatModel。
func (f *FallbackChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	modelsWithTools := make([]model.ToolCallingChatModel, 0, len(f.models))
	for i, m := range f.models {
		wm, err := m.WithTools(tools)
		if err != nil {
			return nil, fmt.Errorf("模型 [%s] WithTools 失败: %w", f.modelNames[i], err)
		}
		modelsWithTools = append(modelsWithTools, wm)
	}

	newModel := &FallbackChatModel{
		models:        modelsWithTools,
		modelNames:    f.modelNames,
		rawNames:      f.rawNames,
		pauseDuration: f.pauseDuration,
	}

	// Preserve compressor wrapper if this model was created by a ChatModelCompressor.
	// The compressor stores itself in the FallbackChatModel so that WithTools can
	// re-wrap the new models with the same compressor config.
	if f.compressorCfg != nil {
		if compressor := newChatModelCompressorFromConfig(newModel, f.compressorCfg); compressor != nil {
			return compressor, nil
		}
		logger.Warn("compressor_rebuild_failed_fallback_without_compression")
	}

	return newModel, nil
}

// compressorCfg 存储在 FallbackChatModel 中，以便 WithTools 可以在重新绑定工具后重新应用压缩器。
type compressorConfig struct {
	summarizerFactory       func() (model.ToolCallingChatModel, error)
	messageThreshold        int
	tokenThreshold          int
	preserveCount           int
	toolResultPreserveCount int
}

// newChatModelCompressorFromConfig 从存储的配置创建 ChatModelCompressor。
// 用于 FallbackChatModel.WithTools 在工具绑定时保留压缩功能。
func newChatModelCompressorFromConfig(inner model.ToolCallingChatModel, cfg *compressorConfig) *ChatModelCompressor {
	summarizer, err := cfg.summarizerFactory()
	if err != nil {
		// Cannot recreate compressor — fall back to no compression
		return nil
	}
	return newChatModelCompressor(inner, summarizer, cfg.messageThreshold, cfg.tokenThreshold, cfg.preserveCount, cfg.toolResultPreserveCount)
}

// newChatModelCompressor 是内部构造函数，接受原始阈值参数。
func newChatModelCompressor(inner model.ToolCallingChatModel, summarizer model.ToolCallingChatModel, msgThresh, tokenThresh, preserve, toolResultPreserve int) *ChatModelCompressor {
	return &ChatModelCompressor{
		inner:      inner,
		summarizer: summarizer,
		cfg: &CompressorConfig{
			MessageThreshold:        msgThresh,
			TokenThreshold:          tokenThresh,
			PreserveCount:           preserve,
			ToolResultPreserveCount: toolResultPreserve,
		},
	}
}
