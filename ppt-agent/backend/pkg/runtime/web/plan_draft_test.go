package web

import "testing"

func TestNewPlanDraftRecordDefaults(t *testing.T) {
	server := &Server{}
	record := server.newPlanDraftRecord(3, "先规划一下 AI 分享", "", "", "conv-1", "msg-1")
	if record.ID == "" {
		t.Fatal("draft id is empty")
	}
	if record.UserID != 3 || record.ConversationID != "conv-1" || record.SourceMessageID != "msg-1" {
		t.Fatalf("draft metadata not carried: %#v", record)
	}
	if record.NormalizedRequest != "先规划一下 AI 分享" || record.Status != "draft" {
		t.Fatalf("draft defaults invalid: %#v", record)
	}
	if record.DraftContent == "" {
		t.Fatal("draft content should be generated when omitted")
	}
}
