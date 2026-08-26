package modelcompat

import (
	"testing"
	"time"
)

func TestNormalizeProviderAliases(t *testing.T) {
	tests := map[string]Provider{
		"":                  ProviderArk,
		"ARK":               ProviderArk,
		"openai-compatible": ProviderOpenAICompat,
		"silicon_flow":      ProviderSiliconFlow,
		"硅基流动":              ProviderSiliconFlow,
		"DashScope":         ProviderQwen,
	}
	for input, want := range tests {
		if got := NormalizeProvider(input); got != want {
			t.Fatalf("NormalizeProvider(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSiliconFlowDefaults(t *testing.T) {
	spec := NormalizeSpec(ModelSpec{Provider: ProviderSiliconFlow, Model: "deepseek-ai/DeepSeek-V3"})
	if spec.BaseURL != SiliconFlowBaseURL {
		t.Fatalf("BaseURL = %q, want %q", spec.BaseURL, SiliconFlowBaseURL)
	}
	if spec.Timeout != DefaultTimeout {
		t.Fatalf("Timeout = %v, want %v", spec.Timeout, DefaultTimeout)
	}
}

func TestDisplayNameAndTrackerKeyIncludeProvider(t *testing.T) {
	spec := ModelSpec{Provider: ProviderSiliconFlow, Model: "same-model"}
	if got := DisplayName(spec, 1); got != "siliconflow/same-model(backup-1)" {
		t.Fatalf("DisplayName() = %q", got)
	}
	if got := TrackerKey(spec); got != "siliconflow:same-model" {
		t.Fatalf("TrackerKey() = %q", got)
	}
}

func TestBuildOpenAIConfigAppliesSiliconFlowProfileAndThinkingDisable(t *testing.T) {
	maxTokens := 1024
	temp := float32(0)
	spec := ModelSpec{
		Provider:    ProviderSiliconFlow,
		Model:       "Qwen/Qwen3.5-Test",
		APIKey:      "key",
		MaxTokens:   &maxTokens,
		Temperature: &temp,
	}
	conf := BuildOpenAIConfig(spec)
	if conf.BaseURL != SiliconFlowBaseURL {
		t.Fatalf("BaseURL = %q, want %q", conf.BaseURL, SiliconFlowBaseURL)
	}
	if conf.APIKey != "key" || conf.Model != spec.Model {
		t.Fatalf("unexpected config key/model: %#v", conf)
	}
	if conf.Timeout != DefaultTimeout {
		t.Fatalf("Timeout = %v, want %v", conf.Timeout, DefaultTimeout)
	}
	if conf.MaxTokens == nil || *conf.MaxTokens != maxTokens {
		t.Fatalf("MaxTokens = %#v", conf.MaxTokens)
	}
	if got, ok := conf.ExtraFields["enable_thinking"]; !ok || got != false {
		t.Fatalf("enable_thinking extra = %#v, ok=%v", got, ok)
	}
}

func TestBuildArkConfigAppliesTimeoutAndThinkingDisable(t *testing.T) {
	disable := true
	timeout := 2 * time.Minute
	conf := BuildArkConfig(ModelSpec{
		Provider:        ProviderArk,
		Model:           "deepseek-test",
		APIKey:          "key",
		Timeout:         timeout,
		DisableThinking: &disable,
	})
	if conf.Timeout == nil || *conf.Timeout != timeout {
		t.Fatalf("Timeout = %#v, want %v", conf.Timeout, timeout)
	}
	if conf.Thinking == nil {
		t.Fatalf("Thinking is nil, want disabled")
	}
}

func TestResolveProviderAPIKey(t *testing.T) {
	t.Setenv("SILICONFLOW_API_KEY", "sf-key")
	if got := ResolveProviderAPIKey(ProviderSiliconFlow, " explicit "); got != "explicit" {
		t.Fatalf("explicit key = %q", got)
	}
	if got := ResolveProviderAPIKey(ProviderSiliconFlow, " "); got != "sf-key" {
		t.Fatalf("env key = %q", got)
	}
}

func TestDefaultFactoryRejectsUnsupportedProvider(t *testing.T) {
	_, err := (DefaultFactory{}).NewToolCallingChatModel(nil, ModelSpec{Provider: Provider("unknown"), Model: "m"})
	if err == nil {
		t.Fatalf("expected unsupported provider error")
	}
}
