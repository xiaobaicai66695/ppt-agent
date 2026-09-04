// Package retry owns bounded runtime retry decisions.  It deliberately keeps
// dependency-specific recovery work out of callers: a Factory selects a
// Strategy, then invokes the supplied executor for that strategy.
package retry

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"time"
)

type Operation string

const (
	OperationUnsplashSearch  Operation = "unsplash_search"
	OperationModelFallback   Operation = "model_fallback"
	OperationModelStreamRead Operation = "model_stream_read"
	OperationQAModelInit     Operation = "qa_model_init"
	OperationDeckSpecReview  Operation = "deck_spec_review"
)

type Policy struct {
	// MaxAttempts is the number of recovery executions after the original
	// failed operation, not the number of network requests.
	MaxAttempts int
	Cooldown    time.Duration
	// MaxCooldown caps exponential backoff. A zero value uses Cooldown as-is.
	MaxCooldown time.Duration
}

// ErrorClass is a stable, provider-agnostic failure category. Callers should
// branch on it instead of matching a provider's error string directly.
type ErrorClass string

const (
	ErrorUnknown       ErrorClass = "unknown"
	ErrorRateLimited   ErrorClass = "rate_limited"
	ErrorUnexpectedEOF ErrorClass = "unexpected_eof"
	ErrorTimeout       ErrorClass = "timeout"
	ErrorUnavailable   ErrorClass = "service_unavailable"
	ErrorConnection    ErrorClass = "connection"
)

// RetryableError keeps the original error available to errors.Is/As while
// exposing a normalized class for task status, metrics and retry selection.
type RetryableError struct {
	Class ErrorClass
	Cause error
}

func (e *RetryableError) Error() string {
	if e == nil || e.Cause == nil {
		return string(ErrorUnknown)
	}
	return e.Cause.Error()
}

func (e *RetryableError) Unwrap() error { return e.Cause }

// ClassifyError recognizes provider/network failures that are safe to retry
// before a workflow has reached a durable checkpoint. It intentionally does
// not classify auth, schema or validation failures as retryable.
func ClassifyError(err error) ErrorClass {
	if err == nil {
		return ErrorUnknown
	}
	var classified *RetryableError
	if errors.As(err, &classified) && classified.Class != "" {
		return classified.Class
	}
	if IsRateLimited(err) {
		return ErrorRateLimited
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrorTimeout
	}
	var networkErr net.Error
	if errors.As(err, &networkErr) && networkErr.Timeout() {
		return ErrorTimeout
	}
	message := strings.ToLower(err.Error())
	switch {
	case errors.Is(err, io.ErrUnexpectedEOF), strings.Contains(message, "unexpected eof"):
		return ErrorUnexpectedEOF
	case strings.Contains(message, "context deadline exceeded"), strings.Contains(message, "timeout"), strings.Contains(message, "timed out"), strings.Contains(message, "超时"):
		return ErrorTimeout
	case strings.Contains(message, "http 502"), strings.Contains(message, "http 503"), strings.Contains(message, "http 504"),
		strings.Contains(message, "bad gateway"), strings.Contains(message, "service unavailable"), strings.Contains(message, "gateway timeout"):
		return ErrorUnavailable
	case strings.Contains(message, "connection reset"), strings.Contains(message, "connection refused"),
		strings.Contains(message, "broken pipe"), strings.Contains(message, "transport is closing"), strings.Contains(message, "stream reset"):
		return ErrorConnection
	default:
		return ErrorUnknown
	}
}

func IsRetryable(err error) bool {
	switch ClassifyError(err) {
	case ErrorRateLimited, ErrorUnexpectedEOF, ErrorTimeout, ErrorUnavailable, ErrorConnection:
		return true
	default:
		return false
	}
}

// WrapRetryable preserves already classified errors and otherwise adds the
// normalized class only when the cause is safe to retry.
func WrapRetryable(err error) error {
	if err == nil {
		return nil
	}
	var classified *RetryableError
	if errors.As(err, &classified) {
		return err
	}
	class := ClassifyError(err)
	if class == ErrorUnknown {
		return err
	}
	return &RetryableError{Class: class, Cause: err}
}

// FixedAttemptStrategy centralizes a workflow's bounded retry/revision budget.
// It intentionally matches without an error because its caller already owns
// the phase-specific validation that decides whether another attempt is needed.
type FixedAttemptStrategy struct {
	Operation    Operation
	StrategyName string
	MaxAttempts  int
}

func (s FixedAttemptStrategy) Name() string { return s.StrategyName }
func (s FixedAttemptStrategy) Match(operation Operation, _ error) bool {
	return operation == s.Operation
}
func (s FixedAttemptStrategy) Policy() Policy { return Policy{MaxAttempts: s.MaxAttempts} }

type Strategy interface {
	Name() string
	Match(Operation, error) bool
	Policy() Policy
}

type Decision struct {
	StrategyName string
	Policy       Policy
}

// Factory selects the first matching strategy in registration order.
type Factory struct{ strategies []Strategy }

func NewFactory(strategies ...Strategy) *Factory {
	filtered := make([]Strategy, 0, len(strategies))
	for _, strategy := range strategies {
		if strategy != nil {
			filtered = append(filtered, strategy)
		}
	}
	return &Factory{strategies: filtered}
}

// Execute selects a policy for cause and runs its recovery executor when the
// one-based recoveryAttempt stays inside the policy bound.  It returns false
// when no strategy applies or all allowed recovery attempts are exhausted.
func (f *Factory) Execute(ctx context.Context, operation Operation, cause error, recoveryAttempt int, executor func(context.Context, Decision) error) (bool, error) {
	if executor == nil || recoveryAttempt < 1 {
		return false, nil
	}
	decision, ok := f.Select(operation, cause)
	if !ok || recoveryAttempt > decision.Policy.MaxAttempts {
		return false, nil
	}
	return true, executor(ctx, decision)
}

func (f *Factory) Select(operation Operation, cause error) (Decision, bool) {
	if f == nil {
		return Decision{}, false
	}
	for _, strategy := range f.strategies {
		if strategy == nil || !strategy.Match(operation, cause) {
			continue
		}
		policy := strategy.Policy()
		if policy.MaxAttempts < 1 {
			return Decision{}, false
		}
		return Decision{StrategyName: strategy.Name(), Policy: policy}, true
	}
	return Decision{}, false
}

// MaxAttempts returns the registered bound for operations whose strategy is
// phase-based rather than error-based. A missing policy is deliberately zero so
// callers fail closed instead of accidentally looping forever.
func (f *Factory) MaxAttempts(operation Operation) int {
	decision, ok := f.Select(operation, nil)
	if !ok {
		return 0
	}
	return decision.Policy.MaxAttempts
}

// RetryDelay returns the delay for the one-based recovery attempt. It is
// deliberately owned here so model, stream and task workflows share one
// bounded backoff policy.
func (f *Factory) RetryDelay(operation Operation, cause error, recoveryAttempt int) (time.Duration, bool) {
	decision, ok := f.Select(operation, cause)
	if !ok || recoveryAttempt < 1 || recoveryAttempt > decision.Policy.MaxAttempts {
		return 0, false
	}
	delay := decision.Policy.Cooldown
	if delay <= 0 {
		return 0, true
	}
	for attempt := 1; attempt < recoveryAttempt; attempt++ {
		delay *= 2
		if decision.Policy.MaxCooldown > 0 && delay >= decision.Policy.MaxCooldown {
			return decision.Policy.MaxCooldown, true
		}
	}
	return delay, true
}

func Wait(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// HTTP410SearchTermRevisionStrategy delegates the retry to an LLM that can
// understand the slide visual intent and write a different asset_query. It is
// intentionally not a network retry and never retries a download operation.
type HTTP410SearchTermRevisionStrategy struct{ MaxAttempts int }

func (HTTP410SearchTermRevisionStrategy) Name() string { return "http_410_llm_search_term_revision" }

func (HTTP410SearchTermRevisionStrategy) Match(operation Operation, err error) bool {
	return operation == OperationUnsplashSearch && IsHTTP410(err)
}

func (s HTTP410SearchTermRevisionStrategy) Policy() Policy {
	maxAttempts := s.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	return Policy{MaxAttempts: maxAttempts}
}

func IsHTTP410(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "http 410")
}

// RateLimitFallbackStrategy controls the shared cooldown before a model is
// considered again after a 429. The model package still performs the actual
// fallback-chain switch because it owns the model instances.
type RateLimitFallbackStrategy struct{}

func (RateLimitFallbackStrategy) Name() string { return "model_rate_limit_fallback" }
func (RateLimitFallbackStrategy) Match(operation Operation, err error) bool {
	return operation == OperationModelFallback && IsRateLimited(err)
}
func (RateLimitFallbackStrategy) Policy() Policy {
	return Policy{MaxAttempts: 1, Cooldown: 30 * time.Second}
}

func IsRateLimited(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "429") || strings.Contains(message, "rate limit") ||
		strings.Contains(message, "rate_limit") || strings.Contains(message, "too many requests")
}

// TransientModelRetryStrategy covers provider disconnects and temporary 5xx
// responses. After its same-model retry budget is exhausted, callers may use
// the same decision to advance to an independently configured fallback model.
type TransientModelRetryStrategy struct{}

func (TransientModelRetryStrategy) Name() string { return "model_transient_retry" }
func (TransientModelRetryStrategy) Match(operation Operation, err error) bool {
	return (operation == OperationModelFallback || operation == OperationModelStreamRead) &&
		ClassifyError(err) != ErrorRateLimited && IsRetryable(err)
}
func (TransientModelRetryStrategy) Policy() Policy {
	return Policy{MaxAttempts: 2, Cooldown: 500 * time.Millisecond, MaxCooldown: 2 * time.Second}
}

// ModelStreamReadFallbackStrategy permits a fallback-model retry only before
// the current stream emitted content; callers retain the stream state check to
// avoid duplicating an answer after partial output.
type ModelStreamReadFallbackStrategy struct{}

func (ModelStreamReadFallbackStrategy) Name() string { return "model_stream_read_fallback" }
func (ModelStreamReadFallbackStrategy) Match(operation Operation, err error) bool {
	return operation == OperationModelStreamRead && IsRateLimited(err)
}
func (ModelStreamReadFallbackStrategy) Policy() Policy { return Policy{MaxAttempts: 1} }

var defaultFactory = NewFactory(
	HTTP410SearchTermRevisionStrategy{},
	RateLimitFallbackStrategy{},
	TransientModelRetryStrategy{},
	ModelStreamReadFallbackStrategy{},
	FixedAttemptStrategy{Operation: OperationQAModelInit, StrategyName: "qa_model_initialization", MaxAttempts: 3},
	FixedAttemptStrategy{Operation: OperationDeckSpecReview, StrategyName: "deck_spec_reviewer", MaxAttempts: 3},
)

// Default is the one registry used by all current retrying runtime phases.
func Default() *Factory { return defaultFactory }
