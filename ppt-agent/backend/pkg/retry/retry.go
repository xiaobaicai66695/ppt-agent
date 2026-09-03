// Package retry owns bounded runtime retry decisions.  It deliberately keeps
// dependency-specific recovery work out of callers: a Factory selects a
// Strategy, then invokes the supplied executor for that strategy.
package retry

import (
	"context"
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

// ModelStreamReadFallbackStrategy permits a fallback-model retry only before
// the current stream emitted content; callers retain the stream state check to
// avoid duplicating an answer after partial output.
type ModelStreamReadFallbackStrategy struct{}

func (ModelStreamReadFallbackStrategy) Name() string { return "model_stream_read_fallback" }
func (ModelStreamReadFallbackStrategy) Match(operation Operation, err error) bool {
	return operation == OperationModelStreamRead && err != nil
}
func (ModelStreamReadFallbackStrategy) Policy() Policy { return Policy{MaxAttempts: 1} }

var defaultFactory = NewFactory(
	HTTP410SearchTermRevisionStrategy{},
	RateLimitFallbackStrategy{},
	ModelStreamReadFallbackStrategy{},
	FixedAttemptStrategy{Operation: OperationQAModelInit, StrategyName: "qa_model_initialization", MaxAttempts: 3},
	FixedAttemptStrategy{Operation: OperationDeckSpecReview, StrategyName: "deck_spec_reviewer", MaxAttempts: 3},
)

// Default is the one registry used by all current retrying runtime phases.
func Default() *Factory { return defaultFactory }
