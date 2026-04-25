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
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"

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
	envPath := os.Getenv("ENV_PATH")
	if envPath == "" {
		envPath = "D:\\environment\\codeGo\\llm-examples\\projects\\ppt-agent\\backend\\.env"
	}
	_ = godotenv.Load(envPath)

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

// FallbackChatModel 包装多个模型，支持 429 后降级和暂停
type FallbackChatModel struct {
	models        []model.ToolCallingChatModel
	modelNames    []string
	mu            sync.Mutex
	pauseEndTimes map[int]time.Time
	pauseDuration time.Duration
}

// NewFallbackToolCallingChatModel 创建支持降级的 ChatModel
// 会依次尝试 ARK_MODEL、ARK_MODEL_BACKUP1、ARK_MODEL_BACKUP2
// 遇到 429 时：当前模型暂停 30s 并尝试下一个模型
// 所有模型都失败后才返回错误
func NewFallbackToolCallingChatModel(ctx context.Context, opts ...ChatModelOption) (model.ToolCallingChatModel, error) {
	envPath := os.Getenv("ENV_PATH")
	if envPath == "" {
		envPath = "D:\\environment\\codeGo\\llm-examples\\projects\\ppt-agent\\backend\\.env"
	}
	_ = godotenv.Load(envPath)

	o := &ChatModelConfig{}
	for _, opt := range opts {
		opt(o)
	}

	modelNames := []string{
		os.Getenv("ARK_MODEL"),
		os.Getenv("ARK_MODEL_BACKUP1"),
		os.Getenv("ARK_MODEL_BACKUP2"),
	}

	var validModels []model.ToolCallingChatModel
	var validNames []string

	for i, name := range modelNames {
		if name != "" {
			cm, err := newSingleModel(ctx, name, o)
			if err != nil {
				fmt.Printf("[Model] 初始化模型 [%s] 失败: %v，跳过\n", name, err)
				continue
			}
			validModels = append(validModels, cm)
			validNames = append(validNames, fmt.Sprintf("%s(backup-%d)", name, i))
			fmt.Printf("[Model] 模型 [%s] 初始化成功\n", name)
		}
	}

	if len(validModels) == 0 {
		return nil, fmt.Errorf("没有任何可用模型")
	}

	return &FallbackChatModel{
		models:        validModels,
		modelNames:    validNames,
		pauseEndTimes: make(map[int]time.Time, len(validModels)),
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
	f.mu.Lock()
	defer f.mu.Unlock()
	pauseEnd, ok := f.pauseEndTimes[idx]
	return ok && time.Now().Before(pauseEnd), pauseEnd
}

func (f *FallbackChatModel) markRateLimited(idx int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pauseEndTimes[idx] = time.Now().Add(f.pauseDuration)
	fmt.Printf("[Model] 模型 [%s] 触发 429，暂停 %v\n", f.modelNames[idx], f.pauseDuration)
}

func (f *FallbackChatModel) callWithFallback(ctx context.Context, callFn func(idx int) (*schema.Message, error)) (*schema.Message, error) {
	for idx := 0; idx < len(f.models); idx++ {
		if paused, pauseEnd := f.shouldPause(idx); paused {
			remaining := time.Until(pauseEnd).Round(time.Second)
			fmt.Printf("[Model] 模型 [%s] 仍在暂停中，等待 %v...\n", f.modelNames[idx], remaining)
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
		if paused, pauseEnd := f.shouldPause(idx); paused {
			remaining := time.Until(pauseEnd).Round(time.Second)
			fmt.Printf("[Model] 模型 [%s] 仍在暂停中，等待 %v...\n", f.modelNames[idx], remaining)
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
		pauseEndTimes: f.pauseEndTimes,
		pauseDuration: f.pauseDuration,
	}, nil
}
