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
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// CompressorConfig 压缩器配置
type CompressorConfig struct {
	Threshold     int // 触发压缩的消息对数量阈值（默认 20）
	PreserveCount int // 压缩时保留的留边消息对数量（默认 4）
}

// CompressorOption 压缩器配置选项
type CompressorOption func(*CompressorConfig)

// WithCompressThreshold 设置触发压缩的消息数量阈值（默认 20）
func WithCompressThreshold(n int) CompressorOption {
	return func(c *CompressorConfig) {
		c.Threshold = n
	}
}

// WithPreserveCount 设置压缩时保留的留边消息数量（默认 4）
func WithPreserveCount(n int) CompressorOption {
	return func(c *CompressorConfig) {
		c.PreserveCount = n
	}
}

// DefaultCompressorConfig 返回默认压缩配置
func DefaultCompressorConfig() *CompressorConfig {
	return &CompressorConfig{
		Threshold:     20,
		PreserveCount: 4,
	}
}

// ChatModelCompressor 包装 ChatModel，在调用前对消息历史进行上下文压缩。
// 当消息对数量超过阈值时，将早期的 user/assistant 对话压缩为摘要，只保留：
//   - system prompt（首条消息）
//   - 最近的 N 条完整消息对（留边区）
//   - 中间段压缩为摘要
//
// 压缩策略：保留 system → 留边 → 摘要中间段 → 留边
type ChatModelCompressor struct {
	inner      model.ToolCallingChatModel
	summarizer model.ToolCallingChatModel
	cfg        *CompressorConfig
}

// NewChatModelCompressor 创建上下文压缩包装器
func NewChatModelCompressor(inner model.ToolCallingChatModel, summarizer model.ToolCallingChatModel, opts ...CompressorOption) *ChatModelCompressor {
	cfg := DefaultCompressorConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return &ChatModelCompressor{
		inner:      inner,
		summarizer: summarizer,
		cfg:        cfg,
	}
}

// Generate 实现 model.ToolCallingChatModel 接口
func (c *ChatModelCompressor) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	if len(messages) <= c.cfg.Threshold {
		return c.inner.Generate(ctx, messages, opts...)
	}

	compressed, err := c.compress(ctx, messages)
	if err != nil {
		fmt.Printf("[Compressor] 上下文压缩失败，降级到原始消息: %v\n", err)
		return c.inner.Generate(ctx, messages, opts...)
	}

	beforeLen := totalTokenEstimate(messages)
	afterLen := totalTokenEstimate(compressed)
	saved := float64(0)
	if beforeLen > 0 {
		saved = 100 * (1 - float64(afterLen)/float64(beforeLen))
	}
	fmt.Printf("[Compressor] 压缩: %d → %d 条消息 (估算 token: %d → %d, 节省约 %.1f%%)\n",
		len(messages), len(compressed), beforeLen, afterLen, saved)

	return c.inner.Generate(ctx, compressed, opts...)
}

// Stream 实现 model.ToolCallingChatModel 接口
func (c *ChatModelCompressor) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	if len(messages) <= c.cfg.Threshold {
		return c.inner.Stream(ctx, messages, opts...)
	}

	compressed, err := c.compress(ctx, messages)
	if err != nil {
		fmt.Printf("[Compressor] 上下文压缩失败，降级到原始消息: %v\n", err)
		return c.inner.Stream(ctx, messages, opts...)
	}

	beforeLen := totalTokenEstimate(messages)
	afterLen := totalTokenEstimate(compressed)
	saved := float64(0)
	if beforeLen > 0 {
		saved = 100 * (1 - float64(afterLen)/float64(beforeLen))
	}
	fmt.Printf("[Compressor] 压缩: %d → %d 条消息 (估算 token: %d → %d, 节省约 %.1f%%)\n",
		len(messages), len(compressed), beforeLen, afterLen, saved)

	return c.inner.Stream(ctx, compressed, opts...)
}

// WithTools 实现 model.ToolCallingChatModel 接口
func (c *ChatModelCompressor) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	innerWithTools, err := c.inner.WithTools(tools)
	if err != nil {
		return nil, err
	}

	summarizerWithTools, err := c.summarizer.WithTools(tools)
	if err != nil {
		summarizerWithTools = c.summarizer
	}

	return &ChatModelCompressor{
		inner:      innerWithTools,
		summarizer: summarizerWithTools,
		cfg:        c.cfg,
	}, nil
}

// compress 执行实际的压缩逻辑
func (c *ChatModelCompressor) compress(ctx context.Context, messages []*schema.Message) ([]*schema.Message, error) {
	// 策略：保留 system → 保留最近 preserveCount 对 → 摘要中间段
	// messages 结构: [system, user, assistant, user, assistant, ...]
	// 我们按 user-assistant 对来处理

	var systemMsg *schema.Message
	var pairs []struct {
		user      *schema.Message
		assistant *schema.Message
	}

	// 解析出 system + pairs
	systemMsg = messages[0]
	rest := messages[1:]

	for i := 0; i+1 <= len(rest); i += 2 {
		pairs = append(pairs, struct {
			user      *schema.Message
			assistant *schema.Message
		}{
			user:      rest[i],
			assistant: rest[i+1],
		})
	}

	// 保留最近的 preserveCount 对话
	preservePairs := c.cfg.PreserveCount
	if preservePairs > len(pairs) {
		preservePairs = len(pairs)
	}

	if preservePairs >= len(pairs) {
		return messages, nil
	}

	headPairs := pairs[:len(pairs)-preservePairs]
	tailPairs := pairs[len(pairs)-preservePairs:]

	// 构建摘要提示
	var summaryBuilder strings.Builder
	for _, p := range headPairs {
		role := "user"
		if p.user.Role != "" {
			role = string(p.user.Role)
		}
		summaryBuilder.WriteString(fmt.Sprintf("[%s]: %s\n", role, p.user.Content))
		if p.assistant != nil {
			summaryBuilder.WriteString(fmt.Sprintf("[assistant]: %s\n", p.assistant.Content))
		}
	}

	summaryPrompt := fmt.Sprintf(`请将以下对话历史压缩为简洁的摘要。摘要需要保留：
1. 所有关键决策和结论
2. 已完成的任务和结果
3. 当前进度状态
4. 重要的上下文信息（如已选择的模板、配色方案、PPT结构等）

原始对话：
%s

请只输出压缩后的摘要，不要输出其他内容。`, summaryBuilder.String())

	// 调用 summarizer 生成摘要
	summaryMsg, err := c.summarizer.Generate(ctx, []*schema.Message{
		schema.SystemMessage("你是一个对话摘要助手，专门将长对话压缩为简洁的摘要。"),
		schema.UserMessage(summaryPrompt),
	})
	if err != nil {
		return nil, fmt.Errorf("summarizer 调用失败: %w", err)
	}

	// 构建压缩后的消息列表
	compressed := []*schema.Message{systemMsg}

	// 添加摘要消息（作为 user 角色）
	summaryContent := summaryMsg.Content
	if summaryContent == "" {
		summaryContent = "[早期对话已压缩省略]"
	}
	compressed = append(compressed, &schema.Message{
		Role:    schema.User,
		Content: fmt.Sprintf("【早期对话摘要】\n%s", summaryContent),
	})

	// 添加留边对话对
	for _, p := range tailPairs {
		compressed = append(compressed, p.user)
		if p.assistant != nil {
			compressed = append(compressed, p.assistant)
		}
	}

	return compressed, nil
}

// totalTokenEstimate 估算消息列表的总 token 数量（粗略估算）
func totalTokenEstimate(messages []*schema.Message) int {
	total := 0
	for _, msg := range messages {
		total += tokenEstimate(msg.Content)
	}
	return total
}

// tokenEstimate 估算单个字符串的 token 数量
func tokenEstimate(s string) int {
	// 粗略估算：中文每个字约 1.5 token，英文每字符约 0.25 token（含空格和标点约 0.75）
	chars := 0
	for _, r := range s {
		if r > 127 {
			chars += 2 // 中文字符权重高
		} else {
			chars++ // ASCII
		}
	}
	return chars * 3 / 4 // 近似 token（中文 ≈1.5，英文 ≈0.75）
}
