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
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/agent/modelcompat"
	"github.com/cloudwego/ppt-agent/pkg/logger"
)

const (
	modelRoleText = "text"
	modelRoleQA   = "qa"
)

// ChatModelConfig ChatModel 配置选项
type ChatModelConfig struct {
	MaxTokens           *int
	MaxCompletionTokens *int
	Temperature         *float32
	TopP                *float32
	DisableThinking     *bool
	JsonSchema          *openai.ChatCompletionResponseFormatJSONSchema
	Model               *string
	APIKey              *string
	APIKeyProvider      *modelcompat.Provider
	ModelRole           string
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

func WithAPIKeyForProvider(provider string, apiKey string) ChatModelOption {
	return func(c *ChatModelConfig) {
		apiKey = strings.TrimSpace(apiKey)
		if apiKey == "" {
			return
		}
		normalized := modelcompat.NormalizeProvider(provider)
		c.APIKey = &apiKey
		c.APIKeyProvider = &normalized
	}
}

// WithTextModel 覆盖 NewFallbackToolCallingChatModel 用于纯文本任务的模型链。
// 优先使用 MODEL_TEXT_* provider-aware 配置，旧环境下继续使用 ARK_TEXT_MODEL，
// 如果 ARK_TEXT_MODEL 未设置，则回退到 ARK_MODEL。
func WithTextModel() ChatModelOption {
	return func(c *ChatModelConfig) {
		c.ModelRole = modelRoleText
	}
}

// NewQAModel 创建一个不带降级的独立 QA 模型。
// 使用 ARK_QA_MODEL；遇到瞬时失败时最多重试 3 次。
// QA 质量不能因模型降级而受损——不使用降级链。
func NewQAModel(ctx context.Context, opts ...ChatModelOption) (model.ToolCallingChatModel, error) {
	baseCfg := &ChatModelConfig{}
	for _, opt := range opts {
		opt(baseCfg)
	}
	baseCfg.ModelRole = modelRoleQA
	qaSpec, err := resolveSingleRoleModelSpec(modelRoleQA, baseCfg)
	if err != nil {
		return nil, err
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
		spec := applyConfigToModelSpec(qaSpec, &cfg)
		m, err := newSingleModel(ctx, spec)
		if err == nil {
			return m, nil
		}
		lastErr = err
		logger.Warn("qa_model_init_retry", "attempt", attempt, "model", modelcompat.DisplayName(spec, -1), "error", err.Error())
	}
	return nil, fmt.Errorf("QA 模型 [%s] 创建失败（已重试 3 次）: %w", modelcompat.DisplayName(qaSpec, -1), lastErr)
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

type modelCallLimiter struct {
	mu    sync.Mutex
	slots map[string]chan struct{}
}

var globalModelCallLimiter = &modelCallLimiter{
	slots: make(map[string]chan struct{}),
}

func acquireModelCallSlot(ctx context.Context, resourceKey string) (func(), error) {
	if strings.TrimSpace(resourceKey) == "" {
		return func() {}, nil
	}
	return globalModelCallLimiter.acquire(ctx, resourceKey, modelAPIConcurrency())
}

func modelAPIConcurrency() int {
	if v := EnvInt("MODEL_API_CONCURRENCY", 0); v > 0 {
		return v
	}
	return EnvInt("MODEL_RESOURCE_CONCURRENCY", 1)
}

func (l *modelCallLimiter) acquire(ctx context.Context, key string, limit int) (func(), error) {
	if limit <= 0 {
		return func() {}, nil
	}
	l.mu.Lock()
	slot := l.slots[key]
	if slot == nil {
		slot = make(chan struct{}, limit)
		l.slots[key] = slot
	}
	l.mu.Unlock()

	select {
	case slot <- struct{}{}:
		return func() { <-slot }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
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
	trackerNames  []string // provider:model，用于跨供应商限流隔离
	resourceKeys  []string // provider:model:key-hash，用于跨供应商/API key 并发隔离
	profiles      []modelCallProfile
	mu            sync.Mutex
	pauseDuration time.Duration
	// compressorCfg 在 FallbackChatModel 被 ChatModelCompressor 包装时存储。
	// 调用 WithTools 时，会重新应用压缩器到新的模型实例。
	compressorCfg *compressorConfig
}

type modelCallProfile struct {
	Provider string
	Model    string
	Timeout  time.Duration
}

// NewFallbackToolCallingChatModel 创建支持降级的 ChatModel。
// 优先使用 MODEL_CHAIN/MODEL_<ENTRY>_* provider-aware 配置；
// 未配置时会依次尝试 ARK_MODEL、ARK_MODEL_BACKUP1、ARK_MODEL_BACKUP2
// 遇到 429 时：当前模型暂停 30s 并尝试下一个模型
// 所有模型都失败后才返回错误
func NewFallbackToolCallingChatModel(ctx context.Context, opts ...ChatModelOption) (model.ToolCallingChatModel, error) {
	o := &ChatModelConfig{}
	for _, opt := range opts {
		opt(o)
	}

	modelSpecs, err := resolveFallbackModelSpecs(o)
	if err != nil {
		return nil, err
	}

	var validModels []model.ToolCallingChatModel
	var validNames []string
	var rawNames []string
	var trackerNames []string
	var resourceKeys []string
	var profiles []modelCallProfile

	for i, spec := range modelSpecs {
		if strings.TrimSpace(spec.Model) == "" {
			continue
		}
		spec = modelcompat.NormalizeSpec(spec)
		cm, err := newSingleModel(ctx, spec)
		displayName := modelcompat.DisplayName(spec, i)
		if err != nil {
			logger.Warn("model_init_failed", "model", displayName, "error", err.Error())
			continue
		}
		validModels = append(validModels, cm)
		validNames = append(validNames, displayName)
		rawNames = append(rawNames, spec.Model)
		trackerNames = append(trackerNames, modelcompat.TrackerKey(spec))
		resourceKeys = append(resourceKeys, modelcompat.ConcurrencyKey(spec))
		profiles = append(profiles, modelCallProfile{
			Provider: string(spec.Provider),
			Model:    spec.Model,
			Timeout:  spec.Timeout,
		})
		logger.Info("model_init_success", "model", displayName, "backup", i)
	}

	if len(validModels) == 0 {
		return nil, fmt.Errorf("没有任何可用模型")
	}

	return &FallbackChatModel{
		models:        validModels,
		modelNames:    validNames,
		rawNames:      rawNames,
		trackerNames:  trackerNames,
		resourceKeys:  resourceKeys,
		profiles:      profiles,
		pauseDuration: fallbackPauseDuration,
	}, nil
}

// newSingleModel 根据 provider-aware spec 创建单个模型。
func newSingleModel(ctx context.Context, spec modelcompat.ModelSpec) (model.ToolCallingChatModel, error) {
	return (modelcompat.DefaultFactory{}).NewToolCallingChatModel(ctx, spec)
}

func resolveFallbackModelSpecs(cfg *ChatModelConfig) ([]modelcompat.ModelSpec, error) {
	if cfg == nil {
		cfg = &ChatModelConfig{}
	}
	if cfg.ModelRole == modelRoleText {
		spec, err := resolveSingleRoleModelSpec(modelRoleText, cfg)
		if err != nil {
			return nil, err
		}
		return []modelcompat.ModelSpec{spec}, nil
	}
	if cfg.Model != nil && strings.TrimSpace(*cfg.Model) != "" {
		spec := modelSpecFromEntry("primary", cfg)
		spec.Model = strings.TrimSpace(*cfg.Model)
		return []modelcompat.ModelSpec{spec}, nil
	}

	entries := configuredModelChain()
	if len(entries) > 0 {
		specs := make([]modelcompat.ModelSpec, 0, len(entries))
		for _, entry := range entries {
			spec := modelSpecFromEntry(entry, cfg)
			if spec.Model != "" {
				specs = append(specs, spec)
			}
		}
		if len(specs) == 0 {
			return nil, fmt.Errorf("MODEL_CHAIN 已配置但没有任何有效 MODEL_<ENTRY>_NAME")
		}
		return specs, nil
	}

	return legacyArkFallbackSpecs(cfg), nil
}

func resolveSingleRoleModelSpec(role string, cfg *ChatModelConfig) (modelcompat.ModelSpec, error) {
	entry := strings.ToLower(strings.TrimSpace(role))
	spec := modelSpecFromEntry(entry, cfg)
	if spec.Model != "" {
		return spec, nil
	}

	switch entry {
	case modelRoleText:
		modelName := strings.TrimSpace(os.Getenv("ARK_TEXT_MODEL"))
		if modelName == "" {
			modelName = strings.TrimSpace(os.Getenv("ARK_MODEL"))
		}
		if modelName == "" {
			return modelcompat.ModelSpec{}, fmt.Errorf("ARK_TEXT_MODEL/ARK_MODEL 环境变量未设置")
		}
		spec := legacyArkSpec(modelName, cfg)
		return spec, nil
	case modelRoleQA:
		modelName := strings.TrimSpace(os.Getenv("ARK_QA_MODEL"))
		if modelName == "" {
			return modelcompat.ModelSpec{}, fmt.Errorf("MODEL_QA_NAME/ARK_QA_MODEL 环境变量未设置")
		}
		return legacyArkSpec(modelName, cfg), nil
	default:
		return modelcompat.ModelSpec{}, fmt.Errorf("unsupported model role %q", role)
	}
}

func configuredModelChain() []string {
	chain := splitEnvList(os.Getenv("MODEL_CHAIN"))
	if len(chain) > 0 {
		return chain
	}
	if strings.TrimSpace(os.Getenv("MODEL_PRIMARY_NAME")) != "" || strings.TrimSpace(os.Getenv("MODEL_PRIMARY_MODEL")) != "" {
		return []string{"primary"}
	}
	return nil
}

func splitEnvList(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func modelSpecFromEntry(entry string, cfg *ChatModelConfig) modelcompat.ModelSpec {
	entryKey := modelEntryKey(entry)
	provider := modelcompat.NormalizeProvider(firstNonEmpty(
		os.Getenv("MODEL_"+entryKey+"_PROVIDER"),
		os.Getenv("MODEL_PROVIDER"),
	))
	spec := modelcompat.ModelSpec{
		Provider:            provider,
		Model:               strings.TrimSpace(firstNonEmpty(os.Getenv("MODEL_"+entryKey+"_NAME"), os.Getenv("MODEL_"+entryKey+"_MODEL"))),
		APIKey:              modelAPIKeyForProvider(provider, entryKey, cfg),
		BaseURL:             modelBaseURLForProvider(provider, entryKey),
		Region:              modelRegionForProvider(provider, entryKey),
		Timeout:             modelTimeoutForProviderEntry(provider, entryKey),
		MaxTokens:           cfg.MaxTokens,
		MaxCompletionTokens: cfg.MaxCompletionTokens,
		Temperature:         cfg.Temperature,
		TopP:                cfg.TopP,
		DisableThinking:     cfg.DisableThinking,
		JSONSchema:          cfg.JsonSchema,
	}
	return modelcompat.NormalizeSpec(spec)
}

func legacyArkFallbackSpecs(cfg *ChatModelConfig) []modelcompat.ModelSpec {
	modelNames := []string{
		os.Getenv("ARK_MODEL"),
		os.Getenv("ARK_MODEL_BACKUP1"),
		os.Getenv("ARK_MODEL_BACKUP2"),
		os.Getenv("ARK_MODEL_BACKUP3"),
		os.Getenv("ARK_MODEL_BACKUP4"),
	}
	specs := make([]modelcompat.ModelSpec, 0, len(modelNames))
	for _, name := range modelNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		specs = append(specs, legacyArkSpec(name, cfg))
	}
	return specs
}

func legacyArkSpec(modelName string, cfg *ChatModelConfig) modelcompat.ModelSpec {
	spec := modelcompat.ModelSpec{
		Provider:            modelcompat.ProviderArk,
		Model:               strings.TrimSpace(modelName),
		APIKey:              modelAPIKeyFromConfig(cfg),
		BaseURL:             strings.TrimSpace(os.Getenv("ARK_BASE_URL")),
		Region:              strings.TrimSpace(os.Getenv("ARK_REGION")),
		Timeout:             modelTimeoutFromEnv("MODEL_TIMEOUT_SECONDS", modelcompat.DefaultTimeout),
		MaxTokens:           cfg.MaxTokens,
		MaxCompletionTokens: cfg.MaxCompletionTokens,
		Temperature:         cfg.Temperature,
		TopP:                cfg.TopP,
		DisableThinking:     cfg.DisableThinking,
		JSONSchema:          cfg.JsonSchema,
	}
	return modelcompat.NormalizeSpec(spec)
}

func applyConfigToModelSpec(spec modelcompat.ModelSpec, cfg *ChatModelConfig) modelcompat.ModelSpec {
	spec.MaxTokens = cfg.MaxTokens
	spec.MaxCompletionTokens = cfg.MaxCompletionTokens
	spec.Temperature = cfg.Temperature
	spec.TopP = cfg.TopP
	spec.DisableThinking = cfg.DisableThinking
	spec.JSONSchema = cfg.JsonSchema
	if cfg.APIKey != nil && strings.TrimSpace(*cfg.APIKey) != "" &&
		(cfg.APIKeyProvider == nil || modelcompat.NormalizeProvider(string(spec.Provider)) == *cfg.APIKeyProvider) {
		spec.APIKey = strings.TrimSpace(*cfg.APIKey)
	}
	return modelcompat.NormalizeSpec(spec)
}

func modelEntryKey(entry string) string {
	key := strings.ToUpper(strings.TrimSpace(entry))
	key = strings.ReplaceAll(key, "-", "_")
	return strings.ReplaceAll(key, " ", "_")
}

func modelAPIKeyForProvider(provider modelcompat.Provider, entryKey string, cfg *ChatModelConfig) string {
	if cfg != nil && cfg.APIKey != nil && strings.TrimSpace(*cfg.APIKey) != "" {
		if cfg.APIKeyProvider == nil || modelcompat.NormalizeProvider(string(provider)) == *cfg.APIKeyProvider {
			return strings.TrimSpace(*cfg.APIKey)
		}
	}
	if key := strings.TrimSpace(os.Getenv("MODEL_" + entryKey + "_API_KEY")); key != "" {
		return key
	}
	if keyEnv := strings.TrimSpace(os.Getenv("MODEL_" + entryKey + "_API_KEY_ENV")); keyEnv != "" {
		return strings.TrimSpace(os.Getenv(keyEnv))
	}
	return modelcompat.ResolveProviderAPIKey(provider, "")
}

func modelBaseURLForProvider(provider modelcompat.Provider, entryKey string) string {
	if value := strings.TrimSpace(os.Getenv("MODEL_" + entryKey + "_BASE_URL")); value != "" {
		return value
	}
	switch modelcompat.NormalizeProvider(string(provider)) {
	case modelcompat.ProviderArk:
		return strings.TrimSpace(os.Getenv("ARK_BASE_URL"))
	case modelcompat.ProviderOpenAI:
		return strings.TrimSpace(os.Getenv("OPENAI_BASE_URL"))
	case modelcompat.ProviderSiliconFlow:
		return strings.TrimSpace(os.Getenv("SILICONFLOW_BASE_URL"))
	case modelcompat.ProviderDeepSeek:
		return strings.TrimSpace(os.Getenv("DEEPSEEK_BASE_URL"))
	case modelcompat.ProviderQwen:
		return strings.TrimSpace(firstNonEmpty(os.Getenv("DASHSCOPE_BASE_URL"), os.Getenv("QWEN_BASE_URL")))
	default:
		return strings.TrimSpace(os.Getenv("OPENAI_BASE_URL"))
	}
}

func modelRegionForProvider(provider modelcompat.Provider, entryKey string) string {
	if value := strings.TrimSpace(os.Getenv("MODEL_" + entryKey + "_REGION")); value != "" {
		return value
	}
	if modelcompat.NormalizeProvider(string(provider)) == modelcompat.ProviderArk {
		return strings.TrimSpace(os.Getenv("ARK_REGION"))
	}
	return ""
}

func modelTimeoutForProviderEntry(provider modelcompat.Provider, entryKey string) time.Duration {
	if timeout := modelTimeoutFromEnv("MODEL_"+entryKey+"_TIMEOUT_SECONDS", 0); timeout > 0 {
		return timeout
	}
	providerKey := modelEntryKey(string(modelcompat.NormalizeProvider(string(provider))))
	if providerKey != "" {
		if timeout := modelTimeoutFromEnv("MODEL_"+providerKey+"_TIMEOUT_SECONDS", 0); timeout > 0 {
			return timeout
		}
	}
	return modelTimeoutFromEnv("MODEL_TIMEOUT_SECONDS", modelcompat.DefaultTimeout)
}

func modelTimeoutFromEnv(key string, defaultValue time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return defaultValue
	}
	return time.Duration(seconds) * time.Second
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func modelAPIKeyFromConfig(cfg *ChatModelConfig) string {
	accountKey := ""
	if cfg != nil && cfg.APIKey != nil && strings.TrimSpace(*cfg.APIKey) != "" {
		accountKey = *cfg.APIKey
	}
	return ResolveModelAPIKey(accountKey, os.Getenv("ARK_API_KEY"))
}

// ResolveModelAPIKey applies the service-wide key precedence rule:
// account-specific configuration wins, then the process environment fallback.
func ResolveModelAPIKey(accountKey, environmentKey string) string {
	if key := strings.TrimSpace(accountKey); key != "" {
		return key
	}
	return strings.TrimSpace(environmentKey)
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

// globalTrackerKey 返回用于全局追踪的 provider:model。
// 优先使用 trackerNames，回退到 rawNames/modelNames 以兼容测试直接构造的实例。
func (f *FallbackChatModel) globalTrackerKey(idx int) string {
	if idx < len(f.trackerNames) {
		return f.trackerNames[idx]
	}
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
		release, err := acquireModelCallSlot(ctx, f.modelResourceKey(idx))
		if err != nil {
			return nil, err
		}
		defer release()
		msgs := make([]*schema.Message, len(messages))
		copy(msgs, messages)
		f.recordModelRequest(ctx, idx, "generate", msgs)
		msg, err := f.models[idx].Generate(ctx, msgs, opts...)
		if err != nil {
			f.recordModelError(ctx, idx, "generate", err)
		}
		return msg, err
	})
}

// Stream 实现 model.ToolCallingChatModel 接口
func (f *FallbackChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	if len(f.models) == 0 {
		return nil, fmt.Errorf("所有模型均失败")
	}
	return f.streamWithFallback(ctx, messages, opts...), nil
}

func (f *FallbackChatModel) streamWithFallback(ctx context.Context, messages []*schema.Message, opts ...model.Option) *schema.StreamReader[*schema.Message] {
	reader, writer := schema.Pipe[*schema.Message](8)
	go func() {
		defer writer.Close()

		nextIdx := 0
		for {
			idx, stream, err := f.openStreamFrom(ctx, nextIdx, messages, opts...)
			if err != nil {
				writer.Send(nil, err)
				return
			}

			emitted := false
			for {
				chunk, recvErr := stream.Recv()
				if recvErr != nil {
					stream.Close()
					if recvErr == io.EOF {
						return
					}
					f.recordModelError(ctx, idx, "stream_read", recvErr)
					if ctx.Err() != nil {
						writer.Send(nil, recvErr)
						return
					}
					if !emitted && idx+1 < len(f.models) {
						logger.Warn("model_stream_read_failed_retrying", "model", f.modelDisplayName(idx), "error", recvErr.Error())
						nextIdx = idx + 1
						break
					}
					writer.Send(nil, recvErr)
					return
				}
				emitted = true
				writer.Send(chunk, nil)
			}
		}
	}()
	return reader
}

func (f *FallbackChatModel) openStreamFrom(ctx context.Context, startIdx int, messages []*schema.Message, opts ...model.Option) (int, *schema.StreamReader[*schema.Message], error) {
	for idx := startIdx; idx < len(f.models); idx++ {
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
				logger.Debug("model_paused_skipping", "model", f.modelDisplayName(idx))
				continue
			}
			remaining := time.Until(pauseEnd).Round(time.Second)
			logger.Info("model_all_paused_waiting", "model", f.modelDisplayName(idx), "remaining", remaining.String())
			select {
			case <-time.After(time.Until(pauseEnd)):
			case <-ctx.Done():
				return idx, nil, ctx.Err()
			}
		}

		msgs := make([]*schema.Message, len(messages))
		copy(msgs, messages)
		release, err := acquireModelCallSlot(ctx, f.modelResourceKey(idx))
		if err != nil {
			return idx, nil, err
		}
		stream, err := f.models[idx].Stream(ctx, msgs, opts...)
		if err != nil {
			release()
			if IsRateLimitError(err) {
				f.markRateLimited(idx)
				continue
			}
			f.recordModelError(ctx, idx, "stream_open", err)
			return idx, nil, err
		}
		f.recordModelRequest(ctx, idx, "stream", msgs)
		return idx, releaseOnStreamDone(sanitizeStreamingToolCallDeltas(stream), release), nil
	}

	return -1, nil, fmt.Errorf("所有模型均失败")
}

func (f *FallbackChatModel) modelResourceKey(idx int) string {
	if idx >= 0 && idx < len(f.resourceKeys) {
		return f.resourceKeys[idx]
	}
	return f.globalTrackerKey(idx)
}

func (f *FallbackChatModel) modelDisplayName(idx int) string {
	if idx >= 0 && idx < len(f.modelNames) {
		return f.modelNames[idx]
	}
	if idx >= 0 && idx < len(f.rawNames) {
		return f.rawNames[idx]
	}
	return fmt.Sprintf("model-%d", idx)
}

func (f *FallbackChatModel) modelProfile(idx int) modelCallProfile {
	if idx >= 0 && idx < len(f.profiles) {
		return f.profiles[idx]
	}
	return modelCallProfile{Model: f.modelDisplayName(idx)}
}

func (f *FallbackChatModel) recordModelRequest(ctx context.Context, idx int, mode string, messages []*schema.Message) {
	meta := RuntimeMetaFromContext(ctx)
	if meta == nil {
		return
	}
	metadata := modelRequestMetadata(f.modelProfile(idx), mode, messages)
	meta.RecordEvent("model_request", f.modelDisplayName(idx), "running", "", metadata)
}

func (f *FallbackChatModel) recordModelError(ctx context.Context, idx int, mode string, err error) {
	meta := RuntimeMetaFromContext(ctx)
	if meta == nil || err == nil {
		return
	}
	metadata := modelRequestMetadata(f.modelProfile(idx), mode, nil)
	meta.RecordLLMErrorDetails(f.modelDisplayName(idx), err.Error(), metadata)
}

func modelRequestMetadata(profile modelCallProfile, mode string, messages []*schema.Message) map[string]any {
	metadata := map[string]any{
		"mode":          strings.TrimSpace(mode),
		"provider":      strings.TrimSpace(profile.Provider),
		"model":         strings.TrimSpace(profile.Model),
		"timeout":       profile.Timeout.String(),
		"timeout_ms":    profile.Timeout.Milliseconds(),
		"message_count": len(messages),
	}
	if profile.Timeout <= 0 {
		delete(metadata, "timeout")
		delete(metadata, "timeout_ms")
	}
	if history := compactModelRequestHistory(messages); len(history) > 0 {
		metadata["history"] = history
	}
	if roles := modelRequestRoleCounts(messages); len(roles) > 0 {
		metadata["role_counts"] = roles
	}
	if preview := firstModelMessagePreview(messages, string(schema.System), 220); preview != "" {
		metadata["system_preview"] = preview
	}
	if preview := lastUserModelMessagePreview(messages, 220); preview != "" {
		metadata["last_user_preview"] = preview
	}
	if names := modelRequestToolNames(messages); len(names) > 0 {
		metadata["tool_names"] = names
	}
	return metadata
}

func compactModelRequestHistory(messages []*schema.Message) []any {
	if len(messages) == 0 {
		return nil
	}
	const maxMessages = 18
	start := 0
	if len(messages) > maxMessages {
		start = len(messages) - maxMessages
	}
	history := make([]any, 0, len(messages)-start)
	for i := start; i < len(messages); i++ {
		message := messages[i]
		if message == nil {
			continue
		}
		role := strings.TrimSpace(string(message.Role))
		if role == "" {
			role = "message"
		}
		item := map[string]any{
			"index": i,
			"role":  role,
		}
		if strings.TrimSpace(message.Name) != "" {
			item["name"] = truncateString(strings.TrimSpace(message.Name), 120)
		}
		if role == string(schema.System) {
			item["content_preview"] = "[系统指令已隐藏，仅保留下方 system_preview 摘要]"
		} else if content := strings.TrimSpace(message.Content); content != "" {
			item["content"] = content
			item["content_preview"] = truncateString(content, 600)
		}
		if strings.TrimSpace(message.ToolCallID) != "" {
			item["tool_call_id"] = truncateString(strings.TrimSpace(message.ToolCallID), 120)
		}
		if strings.TrimSpace(message.ToolName) != "" {
			item["tool_name"] = truncateString(strings.TrimSpace(message.ToolName), 120)
		}
		if len(message.ToolCalls) > 0 {
			names := make([]string, 0, len(message.ToolCalls))
			calls := make([]any, 0, len(message.ToolCalls))
			for _, call := range message.ToolCalls {
				names = append(names, strings.TrimSpace(call.Function.Name))
				calls = append(calls, map[string]any{
					"id":                truncateString(strings.TrimSpace(call.ID), 120),
					"name":              truncateString(strings.TrimSpace(call.Function.Name), 120),
					"arguments_preview": truncateString(strings.TrimSpace(call.Function.Arguments), 300),
				})
			}
			item["tool_calls"] = strings.Join(names, ", ")
			item["tool_call_details"] = calls
		}
		history = append(history, item)
	}
	return history
}

func modelRequestRoleCounts(messages []*schema.Message) map[string]any {
	counts := map[string]any{}
	for _, message := range messages {
		if message == nil {
			continue
		}
		role := strings.TrimSpace(string(message.Role))
		if role == "" {
			role = "message"
		}
		current, _ := counts[role].(int)
		counts[role] = current + 1
	}
	return counts
}

func firstModelMessagePreview(messages []*schema.Message, role string, maxLen int) string {
	for _, message := range messages {
		if message != nil && string(message.Role) == role {
			return truncateString(strings.TrimSpace(message.Content), maxLen)
		}
	}
	return ""
}

func lastUserModelMessagePreview(messages []*schema.Message, maxLen int) string {
	for i := len(messages) - 1; i >= 0; i-- {
		message := messages[i]
		if message == nil || message.Role != schema.User {
			continue
		}
		content := strings.TrimSpace(message.Content)
		if strings.HasPrefix(content, "<agent_progress>") && strings.Contains(content, "</agent_progress>") {
			continue
		}
		return truncateString(content, maxLen)
	}
	return ""
}

func modelRequestToolNames(messages []*schema.Message) []string {
	seen := map[string]struct{}{}
	names := []string{}
	for _, message := range messages {
		if message == nil {
			continue
		}
		if name := strings.TrimSpace(message.ToolName); name != "" {
			if _, ok := seen[name]; !ok {
				seen[name] = struct{}{}
				names = append(names, name)
			}
		}
		for _, call := range message.ToolCalls {
			name := strings.TrimSpace(call.Function.Name)
			if name == "" {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			names = append(names, name)
		}
	}
	return names
}

func releaseOnStreamDone(stream *schema.StreamReader[*schema.Message], release func()) *schema.StreamReader[*schema.Message] {
	if stream == nil {
		release()
		return nil
	}
	reader, writer := schema.Pipe[*schema.Message](8)
	go func() {
		defer writer.Close()
		defer release()
		defer stream.Close()
		for {
			chunk, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					writer.Send(nil, err)
				}
				return
			}
			writer.Send(chunk, nil)
		}
	}()
	return reader
}

func sanitizeStreamingToolCallDeltas(stream *schema.StreamReader[*schema.Message]) *schema.StreamReader[*schema.Message] {
	if stream == nil {
		return nil
	}

	reader, writer := schema.Pipe[*schema.Message](8)
	go func() {
		defer writer.Close()
		defer stream.Close()

		toolCallChunks := map[string][]*schema.Message{}
		var toolCallOrder []string
		for {
			chunk, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					writer.Send(nil, err)
				}
				break
			}
			if chunk == nil {
				continue
			}

			if len(chunk.ToolCalls) == 0 {
				writer.Send(chunk, nil)
				continue
			}

			for _, toolCall := range chunk.ToolCalls {
				key := streamingToolCallKey(toolCall, len(toolCallOrder))
				if _, ok := toolCallChunks[key]; !ok {
					toolCallOrder = append(toolCallOrder, key)
				}
				toolCallChunks[key] = append(toolCallChunks[key], singleStreamingToolCallChunk(chunk, toolCall))
			}
			if chunk.Content != "" || chunk.ReasoningContent != "" {
				visible := *chunk
				visible.ToolCalls = nil
				writer.Send(&visible, nil)
			}
			if err := flushCompletedStreamingToolCalls(writer, toolCallChunks, &toolCallOrder, false); err != nil {
				writer.Send(nil, err)
				return
			}
		}

		if len(toolCallChunks) == 0 {
			return
		}
		if err := flushCompletedStreamingToolCalls(writer, toolCallChunks, &toolCallOrder, true); err != nil {
			writer.Send(nil, err)
			return
		}
	}()
	return reader
}

func streamingToolCallKey(toolCall schema.ToolCall, fallback int) string {
	if toolCall.Index != nil {
		return fmt.Sprintf("index:%d", *toolCall.Index)
	}
	if strings.TrimSpace(toolCall.ID) != "" {
		return "id:" + strings.TrimSpace(toolCall.ID)
	}
	if strings.TrimSpace(toolCall.Function.Name) != "" {
		return fmt.Sprintf("name:%s:%d", strings.TrimSpace(toolCall.Function.Name), fallback)
	}
	return fmt.Sprintf("fallback:%d", fallback)
}

func singleStreamingToolCallChunk(chunk *schema.Message, toolCall schema.ToolCall) *schema.Message {
	msg := *chunk
	msg.Content = ""
	msg.ReasoningContent = ""
	msg.MultiContent = nil
	msg.AssistantGenMultiContent = nil
	msg.ToolCalls = []schema.ToolCall{toolCall}
	if msg.Role == "" {
		msg.Role = schema.Assistant
	}
	return &msg
}

func flushCompletedStreamingToolCalls(
	writer *schema.StreamWriter[*schema.Message],
	toolCallChunks map[string][]*schema.Message,
	toolCallOrder *[]string,
	final bool,
) error {
	if len(toolCallChunks) == 0 || len(*toolCallOrder) == 0 {
		return nil
	}
	var completed []*schema.Message
	remainingOrder := (*toolCallOrder)[:0]
	for _, key := range *toolCallOrder {
		chunks := toolCallChunks[key]
		merged, err := schema.ConcatMessages(chunks)
		if err != nil {
			return err
		}
		if len(merged.ToolCalls) == 0 {
			delete(toolCallChunks, key)
			continue
		}
		if !final && !streamingToolCallComplete(merged.ToolCalls[0]) {
			remainingOrder = append(remainingOrder, key)
			continue
		}
		merged.Content = ""
		merged.ReasoningContent = ""
		merged.MultiContent = nil
		merged.AssistantGenMultiContent = nil
		if merged.Role == "" {
			merged.Role = schema.Assistant
		}
		completed = append(completed, merged)
		delete(toolCallChunks, key)
	}
	*toolCallOrder = remainingOrder
	if len(completed) == 0 {
		return nil
	}
	merged, err := schema.ConcatMessages(completed)
	if err != nil {
		return err
	}
	if len(merged.ToolCalls) == 0 {
		return nil
	}
	merged.Content = ""
	merged.ReasoningContent = ""
	merged.MultiContent = nil
	merged.AssistantGenMultiContent = nil
	if merged.Role == "" {
		merged.Role = schema.Assistant
	}
	writer.Send(merged, nil)
	return nil
}

func streamingToolCallComplete(toolCall schema.ToolCall) bool {
	if strings.TrimSpace(toolCall.Function.Name) == "" {
		return false
	}
	args := strings.TrimSpace(toolCall.Function.Arguments)
	if args == "" {
		return false
	}
	return json.Valid([]byte(args))
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
		trackerNames:  f.trackerNames,
		resourceKeys:  f.resourceKeys,
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
