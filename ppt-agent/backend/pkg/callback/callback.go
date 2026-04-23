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

package callback

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// startTime 程序启动时间，用于计算相对耗时
var startTime = time.Now()

// NewLogHandler 创建一个日志追踪 Handler
// 基于 eino callbacks.Handler 接口，追踪 ChatModel 和 Tool 调用
func NewLogHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(onStart).
		OnEndFn(onEnd).
		OnErrorFn(onError).
		OnStartWithStreamInputFn(onStartWithStreamInput).
		OnEndWithStreamOutputFn(onEndWithStreamOutput).
		Build()
}

func elapsed() int64 {
	return time.Since(startTime).Milliseconds()
}

func onStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	if info == nil {
		return ctx
	}

	if info.Component == components.ComponentOfTool {
		tci := tool.ConvCallbackInput(input)
		args := tci.ArgumentsInJSON
		log.Printf("[%dms] [TRACE] Tool/%s 开始 | 类型=%s | 输入长度=%d", elapsed(), info.Name, info.Type, len(args))
		log.Printf("    输入: %s", args)
		return ctx
	}

	if info.Component == components.ComponentOfChatModel {
		log.Printf("[%dms] [TRACE] ChatModel/%s 开始 | 类型=%s",
			elapsed(), info.Name, info.Type)
		return ctx
	}

	log.Printf("[%dms] [TRACE] %s/%s 开始 | 类型=%s",
		elapsed(), info.Component, info.Name, info.Type)

	return ctx
}

func onEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	if info == nil {
		return ctx
	}

	if info.Component == components.ComponentOfTool {
		tco := tool.ConvCallbackOutput(output)
		response := tco.Response
		log.Printf("[%dms] [TRACE] Tool/%s 完成 | 类型=%s | 输出长度=%d", elapsed(), info.Name, info.Type, len(response))
		log.Printf("    输出: %s", response)
		return ctx
	}

	if info.Component == components.ComponentOfChatModel {
		cco := model.ConvCallbackOutput(output)
		if cco.Message != nil {
			contentLen := len(cco.Message.Content)
			log.Printf("[%dms] [TRACE] ChatModel/%s 完成 | 类型=%s | 内容长度=%d",
				elapsed(), info.Name, info.Type, contentLen)
		} else {
			log.Printf("[%dms] [TRACE] ChatModel/%s 完成 | 类型=%s",
				elapsed(), info.Name, info.Type)
		}
		return ctx
	}

	log.Printf("[%dms] [TRACE] %s/%s 完成 | 类型=%s",
		elapsed(), info.Component, info.Name, info.Type)

	return ctx
}

func onError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	if info == nil {
		log.Printf("[%dms] [ERROR] 组件出错: %v", elapsed(), err)
		return ctx
	}

	log.Printf("[%dms] [ERROR] %s/%s 错误: %v",
		elapsed(), info.Component, info.Name, err)

	return ctx
}

func onStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo, input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	if info == nil || input == nil {
		return ctx
	}

	defer input.Close()
	log.Printf("[%dms] [TRACE] %s/%s 开始流式输入 | 类型=%s",
		elapsed(), info.Component, info.Name, info.Type)

	return ctx
}

func onEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
	if info == nil || output == nil {
		return ctx
	}

	defer output.Close()
	log.Printf("[%dms] [TRACE] %s/%s 开始流式输出 | 类型=%s",
		elapsed(), info.Component, info.Name, info.Type)

	return ctx
}

// TraceSummary 执行摘要收集器
type TraceSummary struct {
	SessionID  string
	StartTime  time.Time
	EndTime    time.Time
	TotalCalls int
	LLMCalls   int
	ToolCalls  int
	ErrorCount int
}

// NewTraceSummary 创建一个新的追踪摘要
func NewTraceSummary(sessionID string) *TraceSummary {
	return &TraceSummary{
		SessionID: sessionID,
		StartTime: time.Now(),
	}
}

// RecordLLMCall 记录一次 LLM 调用
func (s *TraceSummary) RecordLLMCall() {
	s.TotalCalls++
	s.LLMCalls++
}

// RecordToolCall 记录一次工具调用
func (s *TraceSummary) RecordToolCall() {
	s.TotalCalls++
	s.ToolCalls++
}

// RecordError 记录一次错误
func (s *TraceSummary) RecordError() {
	s.ErrorCount++
}

// Print 打印追踪摘要
func (s *TraceSummary) Print() {
	if s.SessionID == "" {
		return
	}

	s.EndTime = time.Now()
	duration := s.EndTime.Sub(s.StartTime)

	log.Printf("")
	log.Printf("========== [CALLBACK] 执行摘要 ==========")
	log.Printf("会话ID: %s", s.SessionID)
	log.Printf("总耗时: %v", duration)
	log.Printf("总调用次数: %d", s.TotalCalls)
	log.Printf("  - LLM 调用: %d", s.LLMCalls)
	log.Printf("  - Tool 调用: %d", s.ToolCalls)
	log.Printf("错误次数: %d", s.ErrorCount)
	log.Printf("========================================")
	log.Printf("")
}

// TraceInfo 追踪信息结构
type TraceInfo struct {
	SessionID string
	Component string
	Name      string
	Type      string
	StartTime time.Time
	EndTime   time.Time
	InputLen  int
	OutputLen int
	Error     error
}

// LogTrace 打印追踪信息
func LogTrace(info *TraceInfo) {
	if info == nil {
		return
	}

	duration := time.Duration(0)
	if !info.EndTime.IsZero() {
		duration = info.EndTime.Sub(info.StartTime)
	}

	if info.Error != nil {
		log.Printf("[TRACE] session=%s component=%s name=%s type=%s duration=%v error=%v",
			info.SessionID, info.Component, info.Name, info.Type, duration, info.Error)
	} else {
		log.Printf("[TRACE] session=%s component=%s name=%s type=%s duration=%v input=%d output=%d",
			info.SessionID, info.Component, info.Name, info.Type, duration, info.InputLen, info.OutputLen)
	}
}

// SensitiveFilter 敏感信息过滤器
func SensitiveFilter(s string) string {
	sensitivePatterns := []string{
		"api_key", "APIToken", "token", "password", "secret",
	}
	result := s
	for _, pattern := range sensitivePatterns {
		if strings.Contains(strings.ToLower(result), pattern) {
			idx := strings.Index(strings.ToLower(result), pattern)
			if idx >= 0 {
				end := idx + len(pattern) + 10
				if end > len(result) {
					end = len(result)
				}
				result = result[:idx] + "***" + result[end:]
			}
		}
	}
	return result
}
