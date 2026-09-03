package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"
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
