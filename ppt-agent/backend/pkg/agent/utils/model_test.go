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
	"errors"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// mockModel 模拟一个模型，可以配置其行为
type mockModel struct {
	name        string
	shouldFail  bool
	failCount   *atomic.Int32
	maxFailures int32
	fail429     bool
	response    *schema.Message
	tools       []*schema.ToolInfo
}

func (m *mockModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	if m.fail429 && m.failCount.Add(1) <= m.maxFailures {
		return nil, errors.New("429 rate limit exceeded")
	}
	if m.shouldFail && m.failCount.Add(1) <= m.maxFailures {
		return nil, errors.New("model error")
	}
	return m.response, nil
}

func (m *mockModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	if m.fail429 && m.failCount.Add(1) <= m.maxFailures {
		return nil, errors.New("429 rate limit exceeded")
	}
	if m.shouldFail && m.failCount.Add(1) <= m.maxFailures {
		return nil, errors.New("model error")
	}

	reader, writer := schema.Pipe[*schema.Message](1)
	go func() {
		defer writer.Close()
		if m.response != nil {
			writer.Send(m.response, nil)
		}
	}()
	return reader, nil
}

func (m *mockModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	newModel := *m
	newModel.tools = tools
	return &newModel, nil
}

func makeMockModel(name string, response string) *mockModel {
	return &mockModel{
		name:     name,
		response: schema.UserMessage(response),
	}
}

func make429MockModel(name string, maxFailures int32, response string) *mockModel {
	return &mockModel{
		name:        name,
		fail429:     true,
		maxFailures: maxFailures,
		failCount:   new(atomic.Int32),
		response:    schema.UserMessage(response),
	}
}

func makeFailMockModel(name string, maxFailures int32) *mockModel {
	return &mockModel{
		name:        name,
		shouldFail:  true,
		maxFailures: maxFailures,
		failCount:   new(atomic.Int32),
	}
}

// --- 辅助函数测试 ---

func TestIsRateLimitError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"429 in message", errors.New("status code 429"), true},
		{"rate limit phrase", errors.New("rate limit exceeded"), true},
		{"rate_limit phrase", errors.New("rate_limit exceeded"), true},
		{"too many requests", errors.New("too many requests"), true},
		{"normal error", errors.New("connection refused"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRateLimitError(tt.err)
			if result != tt.expected {
				t.Errorf("isRateLimitError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

// --- FallbackChatModel 集成测试 ---

func TestFallbackChatModel_SuccessOnFirstModel(t *testing.T) {
	m1 := makeMockModel("model-1", "hello from model-1")

	fb := &FallbackChatModel{
		models:        []model.ToolCallingChatModel{m1},
		modelNames:    []string{"model-1"},
		pauseDuration: 30 * time.Second,
	}

	ctx := context.Background()
	result, err := fb.Generate(ctx, []*schema.Message{schema.UserMessage("hi")})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if !strings.Contains(result.Content, "model-1") {
		t.Errorf("expected content from model-1, got %s", result.Content)
	}
}

func TestFallbackChatModel_FallbackToSecondOn429(t *testing.T) {
	m1 := make429MockModel("model-1", 999, "hello from model-1")
	m2 := makeMockModel("model-2", "hello from model-2")

	fb := &FallbackChatModel{
		models:        []model.ToolCallingChatModel{m1, m2},
		modelNames:    []string{"model-1", "model-2"},
		pauseDuration: 1 * time.Millisecond,
	}

	ctx := context.Background()
	result, err := fb.Generate(ctx, []*schema.Message{schema.UserMessage("hi")})

	if err != nil {
		t.Fatalf("expected no error after fallback, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if !strings.Contains(result.Content, "model-2") {
		t.Errorf("expected content from model-2 after fallback, got %s", result.Content)
	}
	if m1.failCount.Load() == 0 {
		t.Error("expected model-1 to be called at least once")
	}
}

func TestFallbackChatModel_FallbackThroughAllModels(t *testing.T) {
	m1 := make429MockModel("model-1", 999, "")
	m2 := make429MockModel("model-2", 999, "")
	m3 := makeMockModel("model-3", "hello from model-3")

	fb := &FallbackChatModel{
		models:        []model.ToolCallingChatModel{m1, m2, m3},
		modelNames:    []string{"model-1", "model-2", "model-3"},
		pauseDuration: 1 * time.Millisecond,
	}

	ctx := context.Background()
	result, err := fb.Generate(ctx, []*schema.Message{schema.UserMessage("hi")})

	if err != nil {
		t.Fatalf("expected no error after cascading fallback, got %v", err)
	}
	if !strings.Contains(result.Content, "model-3") {
		t.Errorf("expected content from model-3, got %s", result.Content)
	}
}

func TestFallbackChatModel_AllModelsFail(t *testing.T) {
	m1 := makeFailMockModel("model-1", 999)
	m2 := makeFailMockModel("model-2", 999)

	fb := &FallbackChatModel{
		models:        []model.ToolCallingChatModel{m1, m2},
		modelNames:    []string{"model-1", "model-2"},
		pauseDuration: 1 * time.Millisecond,
	}

	ctx := context.Background()
	_, err := fb.Generate(ctx, []*schema.Message{schema.UserMessage("hi")})

	if err == nil {
		t.Fatal("expected error when all models fail, got nil")
	}
	if !strings.Contains(err.Error(), "model error") {
		t.Errorf("expected 'model error', got %v", err)
	}
}

func TestFallbackChatModel_PauseThenRecover(t *testing.T) {
	callCount := atomic.Int32{}
	m1 := &mockModel{
		name: "model-1",
		response: schema.UserMessage("model-1 success"),
		fail429:  true,
		failCount: &callCount,
		maxFailures: 1,
	}

	m2 := makeMockModel("model-2", "model-2 response")

	fb := &FallbackChatModel{
		models:        []model.ToolCallingChatModel{m1, m2},
		modelNames:    []string{"model-1", "model-2"},
		pauseDuration: 1 * time.Millisecond,
	}

	ctx := context.Background()

	// 第一次: m1 429，切换到 m2
	result1, err1 := fb.Generate(ctx, []*schema.Message{schema.UserMessage("hi")})
	if err1 != nil {
		t.Fatalf("first call failed: %v", err1)
	}
	if !strings.Contains(result1.Content, "model-2") {
		t.Errorf("first call should fallback to model-2, got %s", result1.Content)
	}

	// 第二次: m1 应该从暂停中恢复
	time.Sleep(5 * time.Millisecond)
	result2, err2 := fb.Generate(ctx, []*schema.Message{schema.UserMessage("hi")})
	if err2 != nil {
		t.Fatalf("second call failed: %v", err2)
	}
	if !strings.Contains(result2.Content, "model-1") {
		t.Errorf("second call should use model-1, got %s", result2.Content)
	}
}

func TestFallbackChatModel_StreamFallback(t *testing.T) {
	m1 := make429MockModel("model-1", 999, "")
	m2 := &mockModel{
		name:     "model-2",
		response: schema.UserMessage("stream from model-2"),
	}

	fb := &FallbackChatModel{
		models:        []model.ToolCallingChatModel{m1, m2},
		modelNames:    []string{"model-1", "model-2"},
		pauseDuration: 1 * time.Millisecond,
	}

	ctx := context.Background()
	stream, err := fb.Stream(ctx, []*schema.Message{schema.UserMessage("hi")})

	if err != nil {
		t.Fatalf("Stream fallback failed: %v", err)
	}
	if stream == nil {
		t.Fatal("expected stream, got nil")
	}
	defer stream.Close()

	var content strings.Builder
	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("read stream failed: %v", err)
		}
		content.WriteString(msg.Content)
	}
	if !strings.Contains(content.String(), "model-2") {
		t.Errorf("expected stream content from model-2, got %s", content.String())
	}
}

func TestFallbackChatModel_WithTools(t *testing.T) {
	m1 := makeMockModel("model-1", "response")
	m2 := makeMockModel("model-2", "response")

	fb := &FallbackChatModel{
		models:        []model.ToolCallingChatModel{m1, m2},
		modelNames:    []string{"model-1", "model-2"},
		pauseDuration: 30 * time.Second,
	}

	tools := []*schema.ToolInfo{
		{Name: "tool1", Desc: "tool 1"},
		{Name: "tool2", Desc: "tool 2"},
	}

	fbWithTools, err := fb.WithTools(tools)
	if err != nil {
		t.Fatalf("WithTools failed: %v", err)
	}

	ctx := context.Background()
	result, err := fbWithTools.Generate(ctx, []*schema.Message{schema.UserMessage("hi")})
	if err != nil {
		t.Fatalf("Generate after WithTools failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result after WithTools")
	}
}

func TestFallbackChatModel_ConcurrentGenerate(t *testing.T) {
	m1 := makeMockModel("model-1", "response")
	m2 := makeMockModel("model-2", "response")

	fb := &FallbackChatModel{
		models:        []model.ToolCallingChatModel{m1, m2},
		modelNames:    []string{"model-1", "model-2"},
		pauseDuration: 1 * time.Millisecond,
	}

	ctx := context.Background()
	done := make(chan struct{})
	errCh := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					_, err := fb.Generate(ctx, []*schema.Message{schema.UserMessage("test")})
					if err != nil {
						errCh <- err
						return
					}
				}
			}
		}()
	}

	time.Sleep(100 * time.Millisecond)
	close(done)

	select {
	case err := <-errCh:
		t.Fatalf("concurrent Generate returned error: %v", err)
	default:
	}
}

// --- 基准测试 ---

func BenchmarkFallbackChatModel_Success(b *testing.B) {
	m1 := makeMockModel("model-1", "benchmark response")
	fb := &FallbackChatModel{
		models:        []model.ToolCallingChatModel{m1},
		modelNames:    []string{"model-1"},
		pauseDuration: 30 * time.Second,
	}

	ctx := context.Background()
	msgs := []*schema.Message{schema.UserMessage("benchmark")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = fb.Generate(ctx, msgs)
	}
}

func BenchmarkFallbackChatModel_3Models(b *testing.B) {
	m1 := makeMockModel("model-1", "response-1")
	m2 := makeMockModel("model-2", "response-2")
	m3 := makeMockModel("model-3", "response-3")
	fb := &FallbackChatModel{
		models:        []model.ToolCallingChatModel{m1, m2, m3},
		modelNames:    []string{"model-1", "model-2", "model-3"},
		pauseDuration: 30 * time.Second,
	}

	ctx := context.Background()
	msgs := []*schema.Message{schema.UserMessage("benchmark")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = fb.Generate(ctx, msgs)
	}
}
