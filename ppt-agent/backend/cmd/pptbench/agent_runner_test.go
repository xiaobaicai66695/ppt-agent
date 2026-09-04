package main

import (
	"testing"

	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
)

func TestMergeReviewerIssuesKeepsTargetedAndDeterministicBlockers(t *testing.T) {
	targeted := []deck.PlanReviewIssue{{Code: "weak_narrative", Severity: "error", PageIndex: 2, Message: "case target"}}
	detected := []deck.PlanReviewIssue{
		{Code: "weak_narrative", Severity: "error", PageIndex: 2, Message: "detected duplicate"},
		{Code: "low_information_density", Severity: "error", PageIndex: 2, Message: "detected density blocker"},
	}

	issues := mergeReviewerIssues(targeted, detected)
	if len(issues) != 2 {
		t.Fatalf("merged issues = %#v, want target plus deterministic blocker", issues)
	}
	if issues[0].Message != "case target" || issues[1].Code != "low_information_density" {
		t.Fatalf("unexpected issue order/content: %#v", issues)
	}
}
