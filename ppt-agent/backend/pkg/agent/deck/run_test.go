package deck

import (
	"testing"
	"time"
)

func TestStreamTimeoutDefaultsToEightMinutes(t *testing.T) {
	t.Setenv("STREAM_TIMEOUT", "")

	if got := streamTimeout(); got != 8*time.Minute {
		t.Fatalf("streamTimeout() = %s, want 8m", got)
	}
}

func TestStreamTimeoutUsesEnvValue(t *testing.T) {
	t.Setenv("STREAM_TIMEOUT", "30s")

	if got := streamTimeout(); got != 30*time.Second {
		t.Fatalf("streamTimeout() = %s, want 30s", got)
	}
}

func TestStreamTimeoutInvalidEnvFallsBack(t *testing.T) {
	t.Setenv("STREAM_TIMEOUT", "soon")

	if got := streamTimeout(); got != 8*time.Minute {
		t.Fatalf("streamTimeout() = %s, want 8m", got)
	}
}

func TestStreamTimeoutCanBeDisabled(t *testing.T) {
	for _, value := range []string{"0", "-1s"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("STREAM_TIMEOUT", value)

			if got := streamTimeout(); got > 0 {
				t.Fatalf("streamTimeout() = %s, want disabled timeout", got)
			}
		})
	}
}

func TestADKStreamingEnabledDefaultsOff(t *testing.T) {
	t.Setenv("ADK_ENABLE_STREAMING", "")

	if adkStreamingEnabled() {
		t.Fatal("expected ADK streaming to be disabled by default")
	}
}

func TestADKStreamingEnabledOptIn(t *testing.T) {
	t.Setenv("ADK_ENABLE_STREAMING", "true")

	if !adkStreamingEnabled() {
		t.Fatal("expected ADK streaming to be enabled when ADK_ENABLE_STREAMING=true")
	}
}

func TestADKStreamingEnabledIgnoresOtherValues(t *testing.T) {
	for _, value := range []string{"false", "1", "yes"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("ADK_ENABLE_STREAMING", value)

			if adkStreamingEnabled() {
				t.Fatalf("expected ADK streaming to be disabled for %q", value)
			}
		})
	}
}
