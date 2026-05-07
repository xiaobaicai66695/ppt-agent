package task

import (
	"context"
	"sync/atomic"
)

// TokenTracker accumulates LLM token usage across a task's lifecycle.
// It is stored in the context and updated by callback handlers.
type TokenTracker struct {
	PromptTokens     atomic.Int64
	CompletionTokens atomic.Int64
	TotalTokens      atomic.Int64
}

// TokenTotals returns a snapshot of accumulated tokens.
func (t *TokenTracker) TokenTotals() (prompt, completion, total int64) {
	return t.PromptTokens.Load(), t.CompletionTokens.Load(), t.TotalTokens.Load()
}

// Add adds token usage counts to the tracker.
func (t *TokenTracker) Add(prompt, completion, total int) {
	t.PromptTokens.Add(int64(prompt))
	t.CompletionTokens.Add(int64(completion))
	t.TotalTokens.Add(int64(total))
}

type tokenTrackerKey struct{}

// WithTokenTracker attaches a TokenTracker to the context.
func WithTokenTracker(ctx context.Context) (context.Context, *TokenTracker) {
	tt := &TokenTracker{}
	return context.WithValue(ctx, tokenTrackerKey{}, tt), tt
}

// TokenTrackerFromContext retrieves the TokenTracker from the context, if any.
func TokenTrackerFromContext(ctx context.Context) *TokenTracker {
	if tt, ok := ctx.Value(tokenTrackerKey{}).(*TokenTracker); ok {
		return tt
	}
	return nil
}
