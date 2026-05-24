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

// WithModel overrides the model name used by NewFallbackToolCallingChatModel.
// When set, ARK_MODEL / OPENAI_MODEL env vars are ignored for model selection.
// Intended for lightweight side-tasks like intent classification or style extraction
// that should use a smaller/cheaper model than the main agent.
func WithModel(modelName string) ChatModelOption {
	return func(c *ChatModelConfig) {
		c.Model = &modelName
	}
}

func GetCurrentTime() string {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return time.Now().In(loc).Format("2006-01-02 15:04:05")
}

// --- 模型降级管理 ---

const (
	fallbackPauseDuration = 30 * time.Second
)

// EnvInt reads an integer from an environment variable, returning the default if unset or invalid.
func EnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultVal
}

// isRateLimitError 判断错误是否为 429 限流错误
func isRateLimitError(err error) bool {
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
// NOTE: Two FallbackChatModel instances using the same raw model name share pause state
// even if they have different configurations (MaxTokens, Temperature, etc.). This means
// a rate limit on a lower-priority task can delay a higher-priority one. If strict
// isolation is needed, prefix the model name with a scope tag.
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

// markRateLimited 标记指定模型触发 429，全局暂停 baseDuration+随机jitter。
// jitter 防止多个实例在暂停结束后同时恢复、再次触发限流。
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
	rawNames      []string // 原始 modelName，用于全局追踪（如 Qwen/Qwen3.5-122B-A10B）
	mu            sync.Mutex
	pauseDuration time.Duration
	// compressorCfg is stored when a FallbackChatModel is wrapped by ChatModelCompressor.
	// When WithTools is called, the compressor is re-applied to the new model instances.
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
		APIKey:      os.Getenv("ARK_API_KEY"),
		BaseURL:     os.Getenv("ARK_BASE_URL"),
		Region:      os.Getenv("ARK_REGION"),
		Model:       modelName,
		MaxTokens:   cfg.MaxTokens,
		Temperature: cfg.Temperature,
		TopP:        cfg.TopP,
	}

	// DeepSeek 系列模型默认开启 thinking 模式，流式输出中会夹杂 reasoning_content，
	// 导致 eino ReAct 框架解析 tool call JSON 失败。这里自动禁用。
	// Qwen3.5/3.6, DeepSeek, Kimi-K2 thinking models consume token budget
// with reasoning_content, breaking tool call JSON parsing in ReAct.
// Disable thinking for all such models in tool-calling scenarios.
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
			if isRateLimitError(err) {
				f.markRateLimited(idx)
				continue
			}
			return msg, err
		}
		return msg, nil
	}

	return nil, fmt.Errorf("所有模型均失败")
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
			if isRateLimitError(err) {
				f.markRateLimited(idx)
				continue
			}
			return nil, err
		}
		return stream, nil
	}

	return nil, fmt.Errorf("所有模型均失败")
}

// WithTools implements model.ToolCallingChatModel interface
// Each underlying model is bound to the same tool list.
// If this FallbackChatModel was created via WithTools from a compressor-wrapped model,
// the compressor wrapper is preserved by wrapping the new models with the stored compressor config.
// If there is no compressor config stored, returns a plain FallbackChatModel.
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
		return newChatModelCompressorFromConfig(newModel, f.compressorCfg), nil
	}

	return newModel, nil
}

// compressorCfg is stored in FallbackChatModel so that WithTools can re-apply
// the compressor after the underlying models are rebound to tools.
type compressorConfig struct {
	summarizerFactory func() (model.ToolCallingChatModel, error)
	messageThreshold  int
	tokenThreshold    int
	preserveCount     int
}

// newChatModelCompressorFromConfig creates a ChatModelCompressor from stored config.
// This is used by FallbackChatModel.WithTools to preserve compression across tool binding.
func newChatModelCompressorFromConfig(inner model.ToolCallingChatModel, cfg *compressorConfig) *ChatModelCompressor {
	summarizer, err := cfg.summarizerFactory()
	if err != nil {
		// Cannot recreate compressor — fall back to no compression
		return nil
	}
	return newChatModelCompressor(inner, summarizer, cfg.messageThreshold, cfg.tokenThreshold, cfg.preserveCount)
}

// newChatModelCompressor is the internal constructor that takes raw threshold values.
func newChatModelCompressor(inner model.ToolCallingChatModel, summarizer model.ToolCallingChatModel, msgThresh, tokenThresh, preserve int) *ChatModelCompressor {
	return &ChatModelCompressor{
		inner: inner,
		summarizer: summarizer,
		cfg: &CompressorConfig{
			MessageThreshold: msgThresh,
			TokenThreshold:   tokenThresh,
			PreserveCount:    preserve,
		},
	}
}
