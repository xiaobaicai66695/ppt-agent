package task

import (
	"strings"
	"testing"
)

func TestValidateDeliveryFeedback(t *testing.T) {
	if _, err := ValidateDeliveryFeedback(0, ""); err == nil {
		t.Fatal("zero rating must be rejected")
	}
	if _, err := ValidateDeliveryFeedback(6, ""); err == nil {
		t.Fatal("rating above five must be rejected")
	}
	if _, err := ValidateDeliveryFeedback(5, strings.Repeat("a", MaxDeliveryFeedbackSuggestionRunes+1)); err == nil {
		t.Fatal("overlong suggestion must be rejected")
	}
	got, err := ValidateDeliveryFeedback(4, "  内容清晰，建议补充数据来源。  ")
	if err != nil || got != "内容清晰，建议补充数据来源。" {
		t.Fatalf("got %q, %v", got, err)
	}
}
