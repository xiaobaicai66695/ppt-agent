package service

import (
	"testing"

	"github.com/cloudwego/ppt-agent/pkg/agent/modelcompat"
)

func TestMaskAPIKeyNeverReturnsFullSecret(t *testing.T) {
	secret := "sk-1234567890abcdef"
	masked := MaskAPIKey(secret)
	if masked == secret || masked == "" {
		t.Fatalf("masked key = %q", masked)
	}
	if masked[:3] != "sk-" || masked[len(masked)-4:] != "cdef" {
		t.Fatalf("masked key lost stable edges: %q", masked)
	}
}

func TestSupportedProviderBoundary(t *testing.T) {
	if !IsSupportedAccountProvider(modelcompat.ProviderDeepSeek) {
		t.Fatal("deepseek should be supported")
	}
	if IsSupportedAccountProvider(modelcompat.Provider("unknown")) {
		t.Fatal("unknown provider should be rejected")
	}
}
