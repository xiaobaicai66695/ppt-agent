package schema

import (
	"reflect"
	"testing"
)

func TestLegacyAssistantMessagesPrefersTurnRows(t *testing.T) {
	got := legacyAssistantMessages(`["第一段", "", "第二段"]`, "完整旧回答", "旧摘要")
	want := []string{"第一段", "第二段"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("legacy assistant messages = %#v, want %#v", got, want)
	}
}

func TestLegacyAssistantMessagesFallsBackToFullAnswerThenSummary(t *testing.T) {
	if got := legacyAssistantMessages("not-json", "完整旧回答", "旧摘要"); !reflect.DeepEqual(got, []string{"完整旧回答"}) {
		t.Fatalf("full answer fallback = %#v", got)
	}
	if got := legacyAssistantMessages("", "", "旧摘要"); !reflect.DeepEqual(got, []string{"旧摘要"}) {
		t.Fatalf("summary fallback = %#v", got)
	}
}
