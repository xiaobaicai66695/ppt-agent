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
