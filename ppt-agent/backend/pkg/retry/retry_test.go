package retry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"
)

func TestFactoryExecutesHTTP410LLMRevisionStrategy(t *testing.T) {
	factory := NewFactory(HTTP410SearchTermRevisionStrategy{})
	called := 0
	handled, err := factory.Execute(context.Background(), OperationUnsplashSearch, fmt.Errorf("Unsplash API HTTP 410"), 1, func(_ context.Context, decision Decision) error {
		called++
		if decision.StrategyName != "http_410_llm_search_term_revision" {
			t.Fatalf("strategy = %q", decision.StrategyName)
		}
		return nil
	})
	if err != nil || !handled || called != 1 {
		t.Fatalf("handled=%t called=%d err=%v", handled, called, err)
	}
}

func TestFactoryRejectsSecondHTTP410RevisionByDefault(t *testing.T) {
	factory := NewFactory(HTTP410SearchTermRevisionStrategy{})
	handled, err := factory.Execute(context.Background(), OperationUnsplashSearch, errors.New("HTTP 410"), 2, func(context.Context, Decision) error {
		t.Fatal("executor must not run after the recovery budget is exhausted")
		return nil
	})
	if err != nil || handled {
		t.Fatalf("handled=%t err=%v", handled, err)
	}
}

func TestFactoryDoesNotMatchOtherFailures(t *testing.T) {
	factory := NewFactory(HTTP410SearchTermRevisionStrategy{})
	if _, ok := factory.Select(OperationUnsplashSearch, errors.New("HTTP 401")); ok {
		t.Fatal("HTTP 401 must not select the 410 strategy")
	}
}

func TestDefaultFactoryContainsSharedAttemptBounds(t *testing.T) {
	if got := Default().MaxAttempts(OperationQAModelInit); got != 3 {
		t.Fatalf("QA retry budget = %d, want 3", got)
	}
	if got := Default().MaxAttempts(OperationDeckSpecReview); got != 3 {
		t.Fatalf("review retry budget = %d, want 3", got)
	}
	decision, ok := Default().Select(OperationModelFallback, errors.New("HTTP 429 rate limit"))
	if !ok || decision.Policy.Cooldown <= 0 {
		t.Fatalf("missing model fallback policy: %#v ok=%t", decision, ok)
	}
	if _, ok := Default().Select(OperationModelStreamRead, errors.New("stream reset")); !ok {
		t.Fatal("missing stream read fallback policy")
	}
}

func TestClassifyRetryableTransportFailures(t *testing.T) {
	cases := []struct {
		err  error
		want ErrorClass
	}{
		{io.ErrUnexpectedEOF, ErrorUnexpectedEOF},
		{context.DeadlineExceeded, ErrorTimeout},
		{errors.New("Post request: HTTP 503 service unavailable"), ErrorUnavailable},
		{errors.New("read tcp: connection reset by peer"), ErrorConnection},
	}
	for _, tc := range cases {
		if got := ClassifyError(tc.err); got != tc.want {
			t.Fatalf("ClassifyError(%v) = %q, want %q", tc.err, got, tc.want)
		}
		if !IsRetryable(tc.err) {
			t.Fatalf("IsRetryable(%v) = false", tc.err)
		}
		if !errors.Is(WrapRetryable(tc.err), tc.err) {
			t.Fatalf("wrapped error does not preserve cause: %v", tc.err)
		}
	}
}

func TestTransientModelRetryUsesBoundedBackoff(t *testing.T) {
	first, ok := Default().RetryDelay(OperationModelFallback, io.ErrUnexpectedEOF, 1)
	if !ok || first != 500*time.Millisecond {
		t.Fatalf("first retry = %s ok=%t", first, ok)
	}
	second, ok := Default().RetryDelay(OperationModelFallback, io.ErrUnexpectedEOF, 2)
	if !ok || second != time.Second {
		t.Fatalf("second retry = %s ok=%t", second, ok)
	}
	if _, ok := Default().RetryDelay(OperationModelFallback, io.ErrUnexpectedEOF, 3); ok {
		t.Fatal("third transient retry must exceed the shared retry budget")
	}
}
