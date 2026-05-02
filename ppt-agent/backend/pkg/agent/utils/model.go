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
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	arkmodel "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
)

// ChatModelConfig ChatModel 配置选项
type ChatModelConfig struct {
	MaxTokens       *int
	Temperature     *float32
	TopP            *float32
	DisableThinking *bool
	JsonSchema      *openai.ChatCompletionResponseFormatJSONSchema
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

// NewToolCallingChatModel 创建 ChatModel
func NewToolCallingChatModel(ctx context.Context, opts ...ChatModelOption) (cm model.ToolCallingChatModel, err error) {
	o := &ChatModelConfig{}
	for _, opt := range opts {
		opt(o)
	}

	if modelName := os.Getenv("ARK_MODEL"); modelName != "" {
		conf := &ark.ChatModelConfig{
			APIKey:      os.Getenv("ARK_API_KEY"),
			BaseURL:     os.Getenv("ARK_BASE_URL"),
			Region:      os.Getenv("ARK_REGION"),
			Model:       modelName,
			MaxTokens:   o.MaxTokens,
			Temperature: o.Temperature,
			TopP:        o.TopP,
		}
		if o.DisableThinking != nil && *o.DisableThinking {
			conf.Thinking = &arkmodel.Thinking{
				Type: arkmodel.ThinkingTypeDisabled,
			}
		}
		if o.JsonSchema != nil {
			conf.ResponseFormat = &ark.ResponseFormat{
				Type: arkmodel.ResponseFormatJSONSchema,
				JSONSchema: &arkmodel.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        o.JsonSchema.Name,
					Description: o.JsonSchema.Description,
					Schema:      o.JsonSchema.JSONSchema,
					Strict:      o.JsonSchema.Strict,
				},
			}
		}
		cm, err = ark.NewChatModel(ctx, conf)

	} else if modelName := os.Getenv("OPENAI_MODEL"); modelName != "" {
		conf := &openai.ChatModelConfig{
			APIKey: os.Getenv("OPENAI_API_KEY"),
			ByAzure: func() bool {
				return os.Getenv("OPENAI_BY_AZURE") == "true"
			}(),
			BaseURL:     os.Getenv("OPENAI_BASE_URL"),
			Model:       modelName,
			MaxTokens:   o.MaxTokens,
			Temperature: o.Temperature,
			TopP:        o.TopP,
		}
		if o.JsonSchema != nil {
			conf.ResponseFormat = &openai.ChatCompletionResponseFormat{
				Type:       openai.ChatCompletionResponseFormatTypeJSONSchema,
				JSONSchema: o.JsonSchema,
			}
		}
		cm, err = openai.NewChatModel(ctx, conf)
	}
	if err != nil {
		return nil, err
	}

	return cm, nil
}

func GetCurrentTime() string {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return time.Now().In(loc).Format("2006-01-02 15:04:05")
}

// --- 模型降级管理 ---

const (
	fallbackPauseDuration = 30 * time.Second
)

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
	fmt.Printf("[GlobalRateLimit] 模型 [%s] 全局暂停 %v（含 jitter）\n", modelName, baseDuration+jitter)
}

// FallbackChatModel 包装多个模型，支持 429 后降级和全局暂停。
// 多个 FallbackChatModel 实例通过 globalTracker 共享暂停状态。
type FallbackChatModel struct {
	models        []model.ToolCallingChatModel
	modelNames    []string // 日志显示名（含 backup 后缀）
	rawNames      []string // 原始 modelName，用于全局追踪（如 Qwen/Qwen3.5-122B-A10B）
	mu            sync.Mutex
	pauseDuration time.Duration
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
				fmt.Printf("[Model] 初始化模型 [%s] 失败: %v，跳过\n", name, err)
				continue
			}
			validModels = append(validModels, cm)
			validNames = append(validNames, fmt.Sprintf("%s(backup-%d)", name, i))
			rawNames = append(rawNames, name)
			fmt.Printf("[Model] 模型 [%s] 初始化成功\n", name)
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
				fmt.Printf("[Model] 模型 [%s] 全局暂停中，跳过尝试下一个...\n", f.modelNames[idx])
				continue
			}
			// 所有备选模型都在全局暂停中，等待当前模型恢复
			remaining := time.Until(pauseEnd).Round(time.Second)
			fmt.Printf("[Model] 所有模型均全局暂停中，等待模型 [%s] 恢复 %v...\n", f.modelNames[idx], remaining)
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
				fmt.Printf("[Model] 模型 [%s] 全局暂停中，跳过尝试下一个...\n", f.modelNames[idx])
				continue
			}
			remaining := time.Until(pauseEnd).Round(time.Second)
			fmt.Printf("[Model] 所有模型均全局暂停中，等待模型 [%s] 恢复 %v...\n", f.modelNames[idx], remaining)
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

// WithTools 实现 model.ToolCallingChatModel 接口
// 每个底层模型都绑定相同的工具列表
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

	return &FallbackChatModel{
		models:        modelsWithTools,
		modelNames:    f.modelNames,
		rawNames:      f.rawNames,
		pauseDuration: f.pauseDuration,
	}, nil
}
