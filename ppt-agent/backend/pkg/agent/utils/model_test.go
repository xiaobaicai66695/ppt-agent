package utils

import "testing"

func TestModelAPIKeyFromConfigPrefersExplicitKey(t *testing.T) {
	t.Setenv("ARK_API_KEY", "env-key")

	cfg := &ChatModelConfig{}
	WithAPIKey(" user-key ")(cfg)

	if got := modelAPIKeyFromConfig(cfg); got != "user-key" {
		t.Fatalf("modelAPIKeyFromConfig() = %q, want user-key", got)
	}
}

func TestModelAPIKeyFromConfigFallsBackToEnv(t *testing.T) {
	t.Setenv("ARK_API_KEY", "env-key")

	if got := modelAPIKeyFromConfig(&ChatModelConfig{}); got != "env-key" {
		t.Fatalf("modelAPIKeyFromConfig() = %q, want env-key", got)
	}
}

func TestResolveModelAPIKeyPrefersAccountKey(t *testing.T) {
	if got := ResolveModelAPIKey(" account-key ", "env-key"); got != "account-key" {
		t.Fatalf("ResolveModelAPIKey() = %q, want account-key", got)
	}
}

func TestResolveModelAPIKeyFallsBackToEnvironment(t *testing.T) {
	if got := ResolveModelAPIKey("  ", " env-key "); got != "env-key" {
		t.Fatalf("ResolveModelAPIKey() = %q, want env-key", got)
	}
}
