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
	"regexp"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/logger"
)

// CompressorConfig 压缩器配置
type CompressorConfig struct {
	// 触发压缩的消息项数量阈值（默认 60）
	MessageThreshold int
	// 触发压缩的估算 token 数阈值（默认 200000，覆盖 system_prompt + 上下文后接近 32K 模型的安全线）
	TokenThreshold int
	// 压缩时保留的留边消息对数量（默认 8）
	PreserveCount int
	// 压缩时每个留边消息对中保留的 tool result 条数上限（默认全部保留）
	// 设为 0 表示不限制；设为正数 N 表示每个留边 pair 只保留最近的 N 条 tool result
	ToolResultPreserveCount int
}

// CompressorOption 压缩器配置选项
type CompressorOption func(*CompressorConfig)

// WithCompressThreshold 设置触发压缩的消息数量阈值（默认 12）
func WithCompressThreshold(n int) CompressorOption {
	return func(c *CompressorConfig) {
		c.MessageThreshold = n
	}
}

// WithTokenThreshold 设置触发压缩的 token 数阈值（默认 8000）
func WithTokenThreshold(n int) CompressorOption {
	return func(c *CompressorConfig) {
		c.TokenThreshold = n
	}
}

// WithPreserveCount 设置压缩时保留的留边消息对数量（默认 4）
func WithPreserveCount(n int) CompressorOption {
	return func(c *CompressorConfig) {
		c.PreserveCount = n
	}
}

// WithToolResultPreserveCount 设置每个留边消息对中保留的 tool result 条数上限（默认全部保留）
func WithToolResultPreserveCount(n int) CompressorOption {
	return func(c *CompressorConfig) {
		c.ToolResultPreserveCount = n
	}
}

// DefaultCompressorConfig 返回默认压缩配置。
// 可以通过环境变量覆盖：
//   - MASTER_COMPRESSOR_MESSAGE_THRESHOLD: 消息数量阈值（默认 60）
//   - MASTER_COMPRESSOR_TOKEN_THRESHOLD: token 估算阈值（默认 200000）
//   - MASTER_COMPRESSOR_PRESERVE_COUNT: 保留的消息对数量（默认 8）
//   - MASTER_COMPRESSOR_TOOL_RESULT_PRESERVE_COUNT: 每个留边 pair 保留的 tool result 条数（默认 0，不限制）
func DefaultCompressorConfig() *CompressorConfig {
	return &CompressorConfig{
		MessageThreshold:        EnvInt("MASTER_COMPRESSOR_MESSAGE_THRESHOLD", 60),
		TokenThreshold:          EnvInt("MASTER_COMPRESSOR_TOKEN_THRESHOLD", 200000),
		PreserveCount:           EnvInt("MASTER_COMPRESSOR_PRESERVE_COUNT", 8),
		ToolResultPreserveCount: EnvInt("MASTER_COMPRESSOR_TOOL_RESULT_PRESERVE_COUNT", 0),
	}
}

// ── 结构化摘要 ────────────────────────────────────────────────────────────────

// CompressionSummary 是压缩后摘要的结构化格式
type CompressionSummary struct {
	// UserIntentSummary is a deterministic anchor copied from the first real user request.
	UserIntentSummary string `json:"user_intent_summary"`

	// PreservedRequirements keeps whole user messages as bounded requirements.
	PreservedRequirements []string `json:"preserved_requirements"`

	// KeyDecisions 关键决策，结构化保留，不会被 LLM 自由文本丢失
	KeyDecisions struct {
		Template       string   `json:"template,omitempty"`        // 模板名称
		ColorScheme    string   `json:"color_scheme,omitempty"`    // 配色方案
		Theme          string   `json:"theme,omitempty"`           // 主题
		TotalPages     int      `json:"total_pages,omitempty"`     // 总页数
		SlideTypes     []string `json:"slide_types,omitempty"`     // 内容类型列表
		OtherDecisions []string `json:"other_decisions,omitempty"` // 其他关键决策
	} `json:"key_decisions"`

	// ProgressSummary 进度摘要：已完成的任务、当前状态
	ProgressSummary string `json:"progress_summary"`

	// ConversationSummary 自由格式对话摘要，描述中间轮次的交互过程
	ConversationSummary string `json:"conversation_summary"`
}

// ExtractKeyDecisions 从对话历史中解析关键决策
func ExtractKeyDecisions(messages []*schema.Message) *CompressionSummary {
	summary := &CompressionSummary{}
	summary.KeyDecisions.OtherDecisions = []string{}
	summary.UserIntentSummary, summary.PreservedRequirements = extractUserIntentAnchor(messages)

	// 正则匹配模板名称
	templateRe := regexp.MustCompile(`(?i)(?:模板|template)[：:\s]*([a-zA-Z0-9_\-]+)`)
	colorRe := regexp.MustCompile(`(?i)(?:配色|color|pall?ete)[：:\s]*([a-zA-Z0-9_\-]+)`)
	themeRe := regexp.MustCompile(`(?i)(?:theme|主题)[：:\s]*([a-zA-Z0-9_\-]+)`)
	pageRe := regexp.MustCompile(`(?i)(?:(\d+)\s*页|total.*?(\d+))`)
	slideTypeRe := regexp.MustCompile(`(?i)(?:content_type|幻灯片类型)[：:\s]*([a-zA-Z_\-]+)`)

	for _, msg := range messages {
		content := msg.Content
		if content == "" {
			continue
		}

		if tmpl := templateRe.FindStringSubmatch(content); len(tmpl) > 1 && tmpl[1] != "" {
			summary.KeyDecisions.Template = tmpl[1]
		}
		if color := colorRe.FindStringSubmatch(content); len(color) > 1 && color[1] != "" {
			summary.KeyDecisions.ColorScheme = color[1]
		}
		if theme := themeRe.FindStringSubmatch(content); len(theme) > 1 && theme[1] != "" {
			summary.KeyDecisions.Theme = theme[1]
		}
		if page := pageRe.FindStringSubmatch(content); len(page) > 1 {
			if page[1] != "" {
				var n int
				fmt.Sscanf(page[1], "%d", &n)
				if n > summary.KeyDecisions.TotalPages {
					summary.KeyDecisions.TotalPages = n
				}
			}
			if page[2] != "" {
				var n int
				fmt.Sscanf(page[2], "%d", &n)
				if n > summary.KeyDecisions.TotalPages {
					summary.KeyDecisions.TotalPages = n
				}
			}
		}
		if st := slideTypeRe.FindStringSubmatch(content); len(st) > 1 && st[1] != "" {
			found := false
			for _, existing := range summary.KeyDecisions.SlideTypes {
				if existing == st[1] {
					found = true
					break
				}
			}
			if !found {
				summary.KeyDecisions.SlideTypes = append(summary.KeyDecisions.SlideTypes, st[1])
			}
		}
	}

	return summary
}

func extractUserIntentAnchor(messages []*schema.Message) (string, []string) {
	var requirements []string
	for _, msg := range messages {
		if msg == nil || msg.Role != schema.User {
			continue
		}
		content := strings.TrimSpace(msg.Content)
		if content == "" || isRuntimeControlMessage(content) {
			continue
		}
		requirements = append(requirements, truncateString(content, 360))
	}
	requirements = boundedStrings(requirements, 6, 360)
	if len(requirements) == 0 {
		return "", nil
	}
	return truncateString(requirements[0], 520), requirements
}

func isRuntimeControlMessage(content string) bool {
	return strings.HasPrefix(content, "<agent_status>") ||
		strings.HasPrefix(content, "<agent_progress>") ||
		strings.HasPrefix(content, "【对话压缩摘要") ||
		strings.HasPrefix(content, "【上下文压缩交接")
}

// conversationToSummary 调用 LLM 将中间对话压缩为自由摘要
func conversationToSummary(ctx context.Context, summarizer model.ToolCallingChatModel,
	base *CompressionSummary, conversationText string) (*CompressionSummary, string, error) {

	baseJSON, _ := json.Marshal(base)

	summaryPrompt := fmt.Sprintf(`你是 PPT Agent 的上下文压缩助手。请以用户原始目标和后续明确要求为最高优先级，将早期执行轨迹压缩为结构化交接。

【用户目标、明确要求与代码提取的关键决策】：
%s

【待压缩的早期执行轨迹】：
%s

输出一个 JSON 对象：
{
	"user_intent_summary": "一句话重述用户最终要完成的目标",
	"preserved_requirements": ["仍然生效的用户要求，保持原意和优先级"],
  "progress_summary": "简要描述已完成的工作和当前进度（50字以内）",
  "conversation_summary": "用50字以内概括中间轮次的交互过程"
}`, baseJSON, conversationText)

	sumCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := summarizer.Generate(sumCtx, []*schema.Message{
		schema.SystemMessage("你负责生成结构化上下文交接。用户原始目标与后续明确要求拥有最高保留优先级，输出为 JSON 对象。"),
		schema.UserMessage(summaryPrompt),
	})
	if err != nil {
		return nil, summaryPrompt, fmt.Errorf("summarizer 调用失败: %w", err)
	}

	raw := resp.Content
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var parsed CompressionSummary
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		logger.Warn("summarizer_json_parse_failed", "raw", truncateString(raw, 200), "error", err.Error())
		return nil, summaryPrompt, err
	}

	result := *base
	result.UserIntentSummary = base.UserIntentSummary
	result.PreservedRequirements = boundedStrings(
		append(append([]string{}, base.PreservedRequirements...), parsed.PreservedRequirements...), 8, 360)
	result.ProgressSummary = truncateString(strings.TrimSpace(parsed.ProgressSummary), 240)
	result.ConversationSummary = truncateString(strings.TrimSpace(parsed.ConversationSummary), 240)
	return &result, summaryPrompt, nil
}

// ChatModelCompressor 包装 ChatModel，在调用前对消息历史进行上下文压缩。
// 当消息对数量超过阈值时，将早期的 user/assistant 对话压缩为摘要，只保留：
//   - system prompt（首条消息）
//   - 最近的 N 条完整消息对（留边区）
//   - 中间段压缩为摘要
//
// 压缩策略：保留 system -> 留边 -> 摘要中间段 -> 留边
type ChatModelCompressor struct {
	inner      model.ToolCallingChatModel
	summarizer model.ToolCallingChatModel
	cfg        *CompressorConfig
	tracker    *TokenTracker // optional; if set, summarizer calls are tracked
	runtime    *RuntimeMeta
}

// NewChatModelCompressor 创建上下文压缩包装器
func NewChatModelCompressor(inner model.ToolCallingChatModel, summarizer model.ToolCallingChatModel, opts ...CompressorOption) *ChatModelCompressor {
	cfg := DefaultCompressorConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	c := &ChatModelCompressor{
		inner:      inner,
		summarizer: summarizer,
		cfg:        cfg,
	}
	// If the inner model is a FallbackChatModel, store the compressor config so that
	// WithTools can re-apply the compressor after rebinding tools.
	if fcm, ok := inner.(*FallbackChatModel); ok {
		fcm.mu.Lock()
		fcm.compressorCfg = &compressorConfig{
			summarizerFactory: func() (model.ToolCallingChatModel, error) {
				return summarizer, nil
			},
			messageThreshold:        cfg.MessageThreshold,
			tokenThreshold:          cfg.TokenThreshold,
			preserveCount:           cfg.PreserveCount,
			toolResultPreserveCount: cfg.ToolResultPreserveCount,
		}
		fcm.mu.Unlock()
	}
	return c
}

// SetTracker 附加 TokenTracker 以便与主模型调用一起追踪压缩器调用，
// 实现准确的任务级 token 统计。
func (c *ChatModelCompressor) SetTracker(tracker *TokenTracker) {
	c.tracker = tracker
}

// SetRuntimeMeta attaches the task runtime metadata sink for compression stats.
func (c *ChatModelCompressor) SetRuntimeMeta(meta *RuntimeMeta) {
	c.runtime = meta
}

// Generate 实现 model.ToolCallingChatModel 接口
func (c *ChatModelCompressor) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	estimatedTokens := totalTokenEstimate(messages)
	// Skip compression only when BOTH message count AND token estimate are below threshold.
	// Compression is triggered when EITHER threshold is exceeded.
	if len(messages) <= c.cfg.MessageThreshold && estimatedTokens <= c.cfg.TokenThreshold {
		return c.inner.Generate(ctx, messages, opts...)
	}

	compressed, handoff, err := c.compress(ctx, messages)
	if err != nil {
		logger.Warn("compress_failed_fallback", "error", err.Error())
		return c.inner.Generate(ctx, messages, opts...)
	}

	beforeLen := estimatedTokens
	afterLen := totalTokenEstimate(compressed)
	saved := float64(0)
	if beforeLen > 0 {
		saved = 100 * (1 - float64(afterLen)/float64(beforeLen))
	}
	logger.Info("context_compressed",
		"before_msgs", len(messages), "before_tokens", beforeLen,
		"after_msgs", len(compressed), "after_tokens", afterLen,
		"saved_pct", fmt.Sprintf("%.1f%%", saved))
	if c.runtime != nil {
		c.runtime.RecordCompressionDetails(len(messages), len(compressed), beforeLen, afterLen,
			handoff.UserIntentSummary, handoff.PreservedRequirements)
	}

	return c.inner.Generate(ctx, compressed, opts...)
}

// Stream 实现 model.ToolCallingChatModel 接口
func (c *ChatModelCompressor) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	estimatedTokens := totalTokenEstimate(messages)
	// Skip compression only when BOTH message count AND token estimate are below threshold.
	// Compression is triggered when EITHER threshold is exceeded.
	if len(messages) <= c.cfg.MessageThreshold && estimatedTokens <= c.cfg.TokenThreshold {
		return c.inner.Stream(ctx, messages, opts...)
	}

	compressed, handoff, err := c.compress(ctx, messages)
	if err != nil {
		logger.Warn("compress_failed_fallback", "error", err.Error())
		return c.inner.Stream(ctx, messages, opts...)
	}

	beforeLen := estimatedTokens
	afterLen := totalTokenEstimate(compressed)
	saved := float64(0)
	if beforeLen > 0 {
		saved = 100 * (1 - float64(afterLen)/float64(beforeLen))
	}
	logger.Info("context_compressed",
		"before_msgs", len(messages), "before_tokens", beforeLen,
		"after_msgs", len(compressed), "after_tokens", afterLen,
		"saved_pct", fmt.Sprintf("%.1f%%", saved))
	if c.runtime != nil {
		c.runtime.RecordCompressionDetails(len(messages), len(compressed), beforeLen, afterLen,
			handoff.UserIntentSummary, handoff.PreservedRequirements)
	}

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
		logger.Warn("summarizer_withtools_failed", "error", err.Error())
		summarizerWithTools = c.summarizer
	}

	return &ChatModelCompressor{
		inner:      innerWithTools,
		summarizer: summarizerWithTools,
		cfg:        c.cfg,
		tracker:    c.tracker,
		runtime:    c.runtime,
	}, nil
}

// compress 执行实际的压缩逻辑
func (c *ChatModelCompressor) compress(ctx context.Context, messages []*schema.Message) ([]*schema.Message, *CompressionSummary, error) {
	if len(messages) < 2 {
		return messages, ExtractKeyDecisions(messages), nil
	}

	// 解析消息结构
	var systemMsg *schema.Message
	var pairs []pairWithToolCalls

	systemMsg = messages[0]
	if systemMsg.Role != schema.System {
		logger.Warn("compress_non_system_first_msg", "role", string(systemMsg.Role))
	}

	// 按 user/tool_call 块分组，而不是严格两两成对
	// 支持 tool result 消息不单独成对，而是附属到对应的 tool_call 轮次
	i := 1
	for i < len(messages) {
		msg := messages[i]
		switch msg.Role {
		case schema.User, schema.Assistant:
			// 新的 user 或 assistant 消息，开启新一轮对话对
			p := pairWithToolCalls{
				user:      nil,
				assistant: nil,
			}
			if msg.Role == schema.User {
				p.user = msg
			} else {
				p.assistant = msg
			}
			// 收集后续的 tool result 消息，附属于这个 assistant
			for i+1 < len(messages) {
				next := messages[i+1]
				if next.Role == schema.Tool {
					p.toolResults = append(p.toolResults, next)
					i++
				} else {
					break
				}
			}
			pairs = append(pairs, p)
		default:
			// tool, system 等角色，尝试追加到最后一对
			if len(pairs) > 0 {
				if msg.Role == schema.Tool {
					pairs[len(pairs)-1].toolResults = append(pairs[len(pairs)-1].toolResults, msg)
				}
			}
		}
		i++
	}

	// 保留最近的 preserveCount 对话
	preservePairs := c.cfg.PreserveCount
	if preservePairs > len(pairs) {
		preservePairs = len(pairs)
	}

	if preservePairs >= len(pairs) {
		return messages, ExtractKeyDecisions(messages), nil
	}

	headPairs := pairs[:len(pairs)-preservePairs]
	tailPairs := pairs[len(pairs)-preservePairs:]

	// 第一步：从全部消息中结构化提取关键决策（不过滤中间段）
	keyDecisions := ExtractKeyDecisions(messages)

	// 第二步：只把 headPairs 转为文本，送给 LLM 生成自由摘要
	var convText strings.Builder
	for _, p := range headPairs {
		role := "user"
		if p.user != nil && p.user.Role != "" {
			role = string(p.user.Role)
		}
		content := ""
		if p.user != nil {
			content = p.user.Content
		}
		convText.WriteString(fmt.Sprintf("[%s]: %s\n", role, content))
		if p.assistant != nil {
			convText.WriteString(fmt.Sprintf("[assistant]: %s\n", p.assistant.Content))
		}
		// tool result 只记录数量，不展开全文（避免太长）
		if len(p.toolResults) > 0 {
			convText.WriteString(fmt.Sprintf("[tool_results]: %d 个工具调用结果（已省略详情）\n", len(p.toolResults)))
		}
	}

	// 调用 summarizer 生成自由摘要
	handoff, summaryPrompt, err := conversationToSummary(ctx, c.summarizer, keyDecisions, convText.String())
	completionText := ""
	if handoff != nil {
		completionBytes, _ := json.Marshal(handoff)
		completionText = string(completionBytes)
	}
	if c.tracker != nil && (err == nil || completionText != "") {
		promptTokens := tokenEstimate(summaryPrompt)
		completionTokens := tokenEstimate(completionText)
		c.tracker.Add(promptTokens, completionTokens, promptTokens+completionTokens)
		if c.runtime != nil {
			c.runtime.RecordLLMTokens(int64(promptTokens), int64(completionTokens), int64(promptTokens+completionTokens))
		}
	}
	if err != nil {
		logger.Warn("conversation_summary_failed", "error", err.Error())
		handoff = keyDecisions
		handoff.ProgressSummary = fmt.Sprintf("已压缩 %d 轮早期轨迹，保留 %d 轮近期上下文", len(headPairs), preservePairs)
		handoff.ConversationSummary = "早期工具轨迹已压缩，用户目标与明确要求由确定性锚点保留"
	}

	// 第三步：构建以用户意图为锚点的结构化摘要消息。
	summaryJSON, _ := json.Marshal(handoff)
	progressPart := handoff.ProgressSummary
	if progressPart == "" {
		progressPart = fmt.Sprintf("已完成 %d/%d 对话轮次，保留 %d 轮近期待查",
			len(headPairs), len(pairs), preservePairs)
	}

	// 构建压缩后的消息列表
	compressed := []*schema.Message{systemMsg}
	compressed = append(compressed, &schema.Message{
		Role: schema.User,
		Content: fmt.Sprintf("【上下文压缩交接 | 已压缩 %d/%d 轮】\n```json\n%s\n```\n当前进度：%s",
			len(headPairs), len(pairs), summaryJSON, progressPart),
	})

	// 添加留边对话对（保留 tool result，不超过配置的条数上限）
	for _, p := range tailPairs {
		if p.user != nil {
			compressed = append(compressed, p.user)
		}
		if p.assistant != nil {
			compressed = append(compressed, p.assistant)
		}
		toolLimit := c.cfg.ToolResultPreserveCount
		if toolLimit <= 0 {
			// 0 表示不限制，全部保留
			for _, tr := range p.toolResults {
				compressed = append(compressed, tr)
			}
		} else {
			// 只保留最近的 N 条 tool result
			start := 0
			if len(p.toolResults) > toolLimit {
				start = len(p.toolResults) - toolLimit
			}
			for j := start; j < len(p.toolResults); j++ {
				compressed = append(compressed, p.toolResults[j])
			}
		}
	}

	return compressed, handoff, nil
}

// pairWithToolCalls 描述一轮 user+assistant 对话及其 tool result
type pairWithToolCalls struct {
	user        *schema.Message
	assistant   *schema.Message
	toolResults []*schema.Message // 紧跟在这轮 assistant 后的 tool result
}

// truncateString 截断字符串用于日志
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
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
	// 估算策略：按字符类别分别统计后求和
	// - CJK 字符（中文/日文/韩文等 Unicode 扩展-B 区）：每个 ≈ 1.5 token
	// - 英文字母/数字：每个 ≈ 0.75 token（接近 GPT 实际分配）
	// - 标点/空格/控制符：每个 ≈ 0.25 token（极低权重）
	var cjk, ascii, other int
	for _, r := range s {
		switch {
		case r >= 0x4E00 && r <= 0x9FFF, // CJK Unified Ideographs
			r >= 0x3000 && r <= 0x303F, // CJK Symbols
			r >= 0xFF00 && r <= 0xFFEF: // Halfwidth/Fullwidth Forms
			cjk++
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'):
			ascii++
		default:
			other++
		}
	}
	return cjk*3/2 + ascii*3/4 + other/4
}
