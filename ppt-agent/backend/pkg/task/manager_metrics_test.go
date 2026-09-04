package task

import (
	"testing"
	"time"

	"github.com/cloudwego/ppt-agent/pkg/db"
)

func TestTaskGenerationMetricsRoundTrip(t *testing.T) {
	started := time.Date(2026, 9, 3, 10, 0, 0, 0, time.UTC)
	finished := started.Add(42*time.Second + 17*time.Millisecond)
	info := TaskInfo{
		ID:                   "deck-metrics",
		UserID:               9,
		Intent:               "create",
		GenerationStartedAt:  &started,
		GenerationFinishedAt: &finished,
		GenerationDurationMS: 42017,
		FixerRunCount:        3,
	}

	record := taskInfoToRecord(&info)
	if record.GenerationStartedAt == nil || !record.GenerationStartedAt.Equal(started) {
		t.Fatalf("generation_started_at = %v, want %v", record.GenerationStartedAt, started)
	}
	if record.GenerationFinishedAt == nil || !record.GenerationFinishedAt.Equal(finished) {
		t.Fatalf("generation_finished_at = %v, want %v", record.GenerationFinishedAt, finished)
	}
	if record.GenerationDurationMS != 42017 || record.FixerRunCount != 3 {
		t.Fatalf("record metrics = %#v", record)
	}

	got := recordToTaskInfo(&db.TaskRecord{
		ID:                   record.ID,
		UserID:               record.UserID,
		Intent:               record.Intent,
		GenerationStartedAt:  record.GenerationStartedAt,
		GenerationFinishedAt: record.GenerationFinishedAt,
		GenerationDurationMS: record.GenerationDurationMS,
		FixerRunCount:        record.FixerRunCount,
	})
	if got.GenerationDurationMS != 42017 || got.FixerRunCount != 3 {
		t.Fatalf("round-trip metrics = %#v", got)
	}
}

func TestRecordFixerRunCountsAttempts(t *testing.T) {
	state := &TaskState{Info: TaskInfo{ID: "deck-fixer"}}
	state.RecordFixerRun()
	state.RecordFixerRun()
	if state.Info.FixerRunCount != 2 {
		t.Fatalf("fixer_run_count = %d, want 2", state.Info.FixerRunCount)
	}
}

func TestNormalizeAnswerChunkKeepsURLsContiguous(t *testing.T) {
	left := "来源：https://baijiahao.baidu.com/s?id=173896838743"
	if got := normalizeAnswerChunk(left, "452741&wfr=spider"); got != "452741&wfr=spider" {
		t.Fatalf("URL chunk = %q, want contiguous suffix", got)
	}
	if got := normalizeAnswerChunk("[来源](h", "ttps://example.com/guide)"); got != "ttps://example.com/guide)" {
		t.Fatalf("Markdown URL protocol suffix = %q, want raw suffix", got)
	}
	if got := normalizeAnswerChunk("The source", "continues"); got != " continues" {
		t.Fatalf("ordinary English chunk = %q, want separating space", got)
	}
}
