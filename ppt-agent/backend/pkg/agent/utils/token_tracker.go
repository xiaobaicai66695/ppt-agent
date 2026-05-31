package utils

import (
	"context"
	"sync/atomic"
)

// TokenTracker 累积任务生命周期中的 LLM token 使用量。
// 存储在 context 中，由 callback handler 更新。
type TokenTracker struct {
	PromptTokens     atomic.Int64
	CompletionTokens atomic.Int64
	TotalTokens      atomic.Int64
}

// TokenTotals 返回累积 token 的快照。
func (t *TokenTracker) TokenTotals() (prompt, completion, total int64) {
	return t.PromptTokens.Load(), t.CompletionTokens.Load(), t.TotalTokens.Load()
}

// Add 将 token 使用量添加到追踪器。
func (t *TokenTracker) Add(prompt, completion, total int) {
	t.PromptTokens.Add(int64(prompt))
	t.CompletionTokens.Add(int64(completion))
	t.TotalTokens.Add(int64(total))
}

type tokenTrackerKey struct{}

// WithTokenTracker 将 TokenTracker 附加到 context。
func WithTokenTracker(ctx context.Context) (context.Context, *TokenTracker) {
	tt := &TokenTracker{}
	return context.WithValue(ctx, tokenTrackerKey{}, tt), tt
}

// TokenTrackerFromContext 从 context 中获取 TokenTracker（如果有）。
func TokenTrackerFromContext(ctx context.Context) *TokenTracker {
	if tt, ok := ctx.Value(tokenTrackerKey{}).(*TokenTracker); ok {
		return tt
	}
	return nil
}
