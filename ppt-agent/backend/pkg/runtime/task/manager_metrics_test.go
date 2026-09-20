package task

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/ppt-agent/pkg/db"
	agentutils "github.com/cloudwego/ppt-agent/pkg/runtime/model"
)

func TestBroadcastStreamsToolCallThenToolResultInOrder(t *testing.T) {
	listener := make(chan SSERichEvent, 4)
	ts := &TaskState{listeners: map[string]chan SSERichEvent{"test": listener}}
	call := ts.Broadcast(SSERichEvent{Type: SSEEventToolCall, ToolCallID: "call-1", ToolName: "search", ToolArgs: `{"query":"PPT"}`})
	if call.Type != SSEEventToolCall || call.ID == 0 {
		t.Fatalf("tool call = %#v, want immediate tool_call with an event ID", call)
	}
	if len(ts.Events) != 1 || ts.Events[0].Type != SSEEventToolCall || ts.Events[0].ID != call.ID {
		t.Fatalf("tool call was not immediately replayable: %#v", ts.Events)
	}
	select {
	case event := <-listener:
		if event.Type != SSEEventToolCall || event.ID != call.ID || event.ToolResult != "" {
			t.Fatalf("listener tool call = %#v", event)
		}
	default:
		t.Fatal("listener did not receive the immediate tool call")
	}

	result := ts.Broadcast(SSERichEvent{Type: SSEEventToolResult, ToolCallID: "call-1", ToolResult: "找到 3 条资料"})
	if result.Type != SSEEventToolResult || result.ID <= call.ID || result.ToolName != "search" || result.ToolStatus != "success" || result.ToolResult != "找到 3 条资料" {
		t.Fatalf("tool result = %#v", result)
	}
	if len(ts.Events) != 2 || ts.Events[1].Type != SSEEventToolResult || ts.Events[1].ID != result.ID || ts.Events[1].ToolCallID != "call-1" {
		t.Fatalf("replay events = %#v, want call followed by result", ts.Events)
	}
	select {
	case event := <-listener:
		if event.Type != SSEEventToolResult || event.ID != result.ID || event.ToolResult != "找到 3 条资料" {
			t.Fatalf("listener tool result = %#v", event)
		}
	default:
		t.Fatal("listener did not receive the independent tool result")
	}
	if len(ts.pendingTools) != 0 {
		t.Fatalf("pending tools = %#v, want result to close the call", ts.pendingTools)
	}
}

func TestBroadcastStreamsToolFailureAsToolResult(t *testing.T) {
	ts := &TaskState{listeners: make(map[string]chan SSERichEvent)}
	call := ts.Broadcast(SSERichEvent{Type: SSEEventToolCall, ToolCallID: "call-images", ToolName: "search_images"})
	result := ts.Broadcast(SSERichEvent{Type: SSEEventToolResult, ToolCallID: "call-images", Error: "图片搜索不可用"})
	if result.Type != SSEEventToolResult || result.ToolStatus != "error" || !strings.Contains(result.ToolResult, "不可用") {
		t.Fatalf("tool failure = %#v", result)
	}
	if len(ts.Events) != 2 || ts.Events[0].ID != call.ID || ts.Events[1].ID != result.ID || ts.Events[1].Type != SSEEventToolResult {
		t.Fatalf("events = %#v, want distinct call and failed result", ts.Events)
	}
}

func TestBroadcastCorrelatesRepeatedToolNamesByArgumentPreview(t *testing.T) {
	ts := &TaskState{listeners: make(map[string]chan SSERichEvent)}
	first := ts.Broadcast(SSERichEvent{Type: SSEEventToolCall, ToolCallID: "call-first", ToolName: "search", ToolArgs: `{"query":"first"}`})
	second := ts.Broadcast(SSERichEvent{Type: SSEEventToolCall, ToolCallID: "call-second", ToolName: "search", ToolArgs: `{"query":"second"}`})

	firstResult := ts.Broadcast(SSERichEvent{Type: SSEEventToolResult, ToolName: "search", ToolArgs: `{"query":"first"}`, ToolResult: "first result"})
	secondResult := ts.Broadcast(SSERichEvent{Type: SSEEventToolResult, ToolName: "search", ToolArgs: `{"query":"second"}`, ToolResult: "second result"})

	if firstResult.ToolCallID != first.ToolCallID || secondResult.ToolCallID != second.ToolCallID {
		t.Fatalf("result correlation = (%q, %q), want (%q, %q)", firstResult.ToolCallID, secondResult.ToolCallID, first.ToolCallID, second.ToolCallID)
	}
	if len(ts.pendingTools) != 0 {
		t.Fatalf("pending tools = %#v, want both matching calls closed", ts.pendingTools)
	}
	if got := []string{ts.Events[0].Type, ts.Events[1].Type, ts.Events[2].Type, ts.Events[3].Type}; strings.Join(got, ",") != "tool_call,tool_call,tool_result,tool_result" {
		t.Fatalf("source order = %#v", ts.Events)
	}
}

func TestBroadcastSuppressesDuplicateToolCompletion(t *testing.T) {
	listener := make(chan SSERichEvent, 4)
	ts := &TaskState{listeners: map[string]chan SSERichEvent{"test": listener}}

	ts.Broadcast(SSERichEvent{Type: SSEEventToolCall, ToolCallID: "call-duplicate", ToolName: "search"})
	first := ts.Broadcast(SSERichEvent{Type: SSEEventToolResult, ToolCallID: "call-duplicate", ToolResult: "首个结果"})
	duplicate := ts.Broadcast(SSERichEvent{Type: SSEEventToolResult, ToolCallID: "call-duplicate", ToolResult: "重复结果"})

	if duplicate.Type != "" || duplicate.ID != 0 || duplicate.ToolCallID != "" {
		t.Fatalf("duplicate completion = %#v, want suppressed zero event", duplicate)
	}
	if len(ts.Events) != 2 || ts.Events[0].Type != SSEEventToolCall || ts.Events[1].Type != SSEEventToolResult || ts.Events[1].ID != first.ID || ts.Events[1].ToolResult != "首个结果" {
		t.Fatalf("events = %#v, want one call and one terminal result", ts.Events)
	}
	if _, completed := ts.completedTools["call-duplicate"]; !completed {
		t.Fatal("completed tool call was not recorded")
	}
	if len(ts.pendingTools) != 0 {
		t.Fatalf("pending tools = %#v, want none after terminal result", ts.pendingTools)
	}
	for _, wantType := range []string{SSEEventToolCall, SSEEventToolResult} {
		select {
		case event := <-listener:
			if event.Type != wantType {
				t.Fatalf("listener event type = %q, want %q", event.Type, wantType)
			}
		default:
			t.Fatalf("listener missing %s event", wantType)
		}
	}
	select {
	case event := <-listener:
		t.Fatalf("listener received duplicate completion: %#v", event)
	default:
	}
}

func TestBroadcastRuntimeSummaryForwardsLLMBoundariesAndKeepsSlideRenderTools(t *testing.T) {
	ts := &TaskState{listeners: make(map[string]chan SSERichEvent)}
	fullArgs := `{"task_id":"slide-1","notes":"` + strings.Repeat("x", 24000) + `"}`
	fullResult := "render complete\n" + strings.TrimSuffix(strings.Repeat("output line\n", 1800), "\n")

	broadcastRuntimeSummary(ts, agentutils.RuntimeEventSummary(agentutils.RuntimeEvent{
		Kind: "llm_start",
		Name: "ChatModel",
	}))
	broadcastRuntimeSummary(ts, agentutils.RuntimeEventSummary(agentutils.RuntimeEvent{
		Kind: "llm_end",
		Name: "ChatModel",
	}))
	if len(ts.Events) != 2 || ts.Events[0].Type != SSEEventLLMStart || ts.Events[1].Type != SSEEventLLMEnd {
		t.Fatalf("llm runtime summaries = %#v, want start/end SSE boundaries", ts.Events)
	}

	broadcastRuntimeSummary(ts, agentutils.RuntimeEventSummary(agentutils.RuntimeEvent{
		Kind:  "slide_render_start",
		Name:  "generate_slide",
		Phase: "rendering",
		Metadata: map[string]any{
			"args": fullArgs,
		},
	}))
	if len(ts.Events) != 3 || ts.Events[2].Type != SSEEventToolCall || ts.Events[2].ToolName != "generate_slide" || ts.Events[2].ToolArgs != fullArgs {
		t.Fatal("slide render start did not retain the complete arguments")
	}

	broadcastRuntimeSummary(ts, agentutils.RuntimeEventSummary(agentutils.RuntimeEvent{
		Kind:   "slide_render_end",
		Name:   "generate_slide",
		Phase:  "rendering",
		Status: "ok",
		Metadata: map[string]any{
			"args":   fullArgs,
			"result": fullResult,
		},
	}))
	if len(ts.Events) != 4 || ts.Events[3].Type != SSEEventToolResult || ts.Events[3].ToolCallID == "" || ts.Events[3].ToolResult != fullResult {
		t.Fatal("slide render end did not retain the complete result")
	}
}

func TestBroadcastRuntimeSummarySeparatesExecutedWithoutDisplayableOutput(t *testing.T) {
	ts := &TaskState{listeners: make(map[string]chan SSERichEvent)}
	args := `{"path":"component_contracts.json"}`

	broadcastRuntimeSummary(ts, agentutils.RuntimeEventSummary(agentutils.RuntimeEvent{
		Kind: "tool_start", Name: "read_file", Phase: "planning", Metadata: map[string]any{"args": args},
	}))
	broadcastRuntimeSummary(ts, agentutils.RuntimeEventSummary(agentutils.RuntimeEvent{
		Kind: "tool_end", Name: "read_file", Phase: "planning", Status: "ok", Metadata: map[string]any{"args": args},
	}))

	if len(ts.Events) != 2 {
		t.Fatalf("timeline events = %#v", ts.Events)
	}
	result := ts.Events[1]
	if result.Type != SSEEventToolResult || result.ToolStatus != "success" || result.ToolResult != "工具已执行，未返回可展示内容" {
		t.Fatalf("terminal result = %#v", result)
	}
}

func TestCompletePendingToolsMarksUnverifiedCallsWhenNoResultCallbackArrives(t *testing.T) {
	ts := &TaskState{listeners: make(map[string]chan SSERichEvent)}
	firstCall := ts.Broadcast(SSERichEvent{Type: SSEEventToolCall, ToolCallID: "call-read", ToolName: "read_file"})
	secondCall := ts.Broadcast(SSERichEvent{Type: SSEEventToolCall, ToolCallID: "call-search", ToolName: "search"})

	ts.CompletePendingTools(nil)

	if len(ts.pendingTools) != 0 {
		t.Fatalf("pending tools = %#v, want all calls closed", ts.pendingTools)
	}
	if len(ts.Events) != 4 {
		t.Fatalf("events = %#v, want calls followed by two terminal results", ts.Events)
	}
	for index, wantID := range []string{"call-read", "call-search"} {
		result := ts.Events[index+2]
		if result.Type != SSEEventToolResult || result.ToolCallID != wantID || result.ToolStatus != "unverified" || result.ToolResult != "未收到工具执行结果" || result.Error != "" {
			t.Fatalf("terminal result[%d] = %#v", index, result)
		}
	}
	if ts.Events[0].ID != firstCall.ID || ts.Events[1].ID != secondCall.ID || ts.Events[2].ID <= secondCall.ID || ts.Events[3].ID <= ts.Events[2].ID {
		t.Fatalf("event IDs are not ordered: %#v", ts.Events)
	}

	ts.CompletePendingTools(nil)
	if len(ts.Events) != 4 {
		t.Fatalf("second completion pass emitted duplicate results: %#v", ts.Events)
	}
}

func TestCompletePendingToolsMarksUnfinishedCallsAsErrors(t *testing.T) {
	ts := &TaskState{listeners: make(map[string]chan SSERichEvent)}
	ts.Broadcast(SSERichEvent{Type: SSEEventToolCall, ToolCallID: "call-read", ToolName: "read_file"})

	ts.CompletePendingTools(errors.New("planner stopped"))

	if len(ts.pendingTools) != 0 {
		t.Fatalf("pending tools = %#v, want all calls closed", ts.pendingTools)
	}
	if len(ts.Events) != 2 {
		t.Fatalf("events = %#v, want one call followed by one terminal result", ts.Events)
	}
	result := ts.Events[1]
	if result.Type != SSEEventToolResult || result.ToolCallID != "call-read" || result.ToolStatus != "error" || result.ToolResult != "工具调用未完成" || result.Error != "工具调用未完成" {
		t.Fatalf("terminal error result = %#v", result)
	}
}

func TestTaskGenerationMetricsRoundTrip(t *testing.T) {
	started := time.Date(2026, 9, 3, 10, 0, 0, 0, time.UTC)
	finished := started.Add(42*time.Second + 17*time.Millisecond)
	info := TaskInfo{
		ID:                   "ppt-metrics",
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
	state := &TaskState{Info: TaskInfo{ID: "ppt-fixer"}}
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
