package router

import (
	"testing"

	"github.com/cloudwego/ppt-agent/pkg/agent/intent"
	"github.com/cloudwego/ppt-agent/pkg/style"
)

func TestEnhanceWithProfileSkipsSceneSensitivePreferencesAcrossDomains(t *testing.T) {
	matcher := NewProfileMatcher()
	classification := &intent.ClassificationResult{
		Domain:             intent.DomainAcademic,
		SuggestedTheme:     "ocean_soft",
		SuggestedPageCount: 6,
		Complexity: intent.Complexity{
			PageCountEstimate: 6,
		},
	}
	profile := &style.EnhancedProfile{
		UserProfile: style.UserProfile{
			PreferredThemes:  []string{"charcoal_light"},
			TypicalPageCount: 18,
			ContentTypes:     style.ContentTypeCount{"chart_slide": 5},
		},
		DomainPreferences: map[string]int{"business": 5},
		SuccessPatterns: []style.SuccessPattern{{
			Domain:       "business",
			Template:     "pitch-deck",
			Theme:        "charcoal_light",
			PageCount:    18,
			SuccessCount: 5,
		}},
	}

	matcher.EnhanceWithProfile(classification, profile)

	if classification.SuggestedTheme != "ocean_soft" {
		t.Fatalf("cross-domain theme should not override current suggestion: %s", classification.SuggestedTheme)
	}
	if classification.SuggestedPageCount != 6 {
		t.Fatalf("explicit page estimate should not be averaged with history, got %d", classification.SuggestedPageCount)
	}
}

func TestEnhanceWithProfileUsesSameDomainHistoryAsFallback(t *testing.T) {
	matcher := NewProfileMatcher()
	classification := &intent.ClassificationResult{
		Domain: intent.DomainBusiness,
	}
	profile := &style.EnhancedProfile{
		UserProfile: style.UserProfile{
			PreferredThemes:  []string{"charcoal_light"},
			TypicalPageCount: 16,
		},
		DomainPreferences: map[string]int{"business": 3},
		SuccessPatterns: []style.SuccessPattern{{
			Domain:       "business",
			Template:     "pitch-deck",
			Theme:        "charcoal_light",
			PageCount:    16,
			SuccessCount: 3,
		}},
	}

	matcher.EnhanceWithProfile(classification, profile)

	if classification.SuggestedTheme != "charcoal_light" {
		t.Fatalf("same-domain theme should fill empty suggestion, got %q", classification.SuggestedTheme)
	}
	if classification.SuggestedPageCount != 16 {
		t.Fatalf("typical page count should fill missing estimate, got %d", classification.SuggestedPageCount)
	}
}
