package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestRuntimeMetaFreezesPlanAndExposesAlignmentSnapshot(t *testing.T) {
	workDir := t.TempDir()
	meta := NewRuntimeMeta("task-1", workDir)
	meta.RecordIntent(IntentAnchor{
		Summary:        "生成一份季度经营复盘",
		OriginalLength: 1200,
		Intent:         "create",
		Domain:         "business",
		SuggestedPages: 2,
		Template:       "executive",
	})

	plan := []PlanSlide{
		{PageIndex: 1, TaskID: "slide-1", Title: "经营概览", ContentType: "kpi_dashboard", OutputFile: "1_overview.pptx"},
		{PageIndex: 2, TaskID: "slide-2", Title: "行动建议", ContentType: "summary_slide", OutputFile: "2_actions.pptx"},
	}
	meta.FreezePlan(plan)
	meta.FreezePlan([]PlanSlide{{PageIndex: 9, Title: "不得覆盖"}})
	current := PlanSlide{PageIndex: 2, TaskID: "slide-2", Title: "行动建议", Status: "generating"}
	meta.RecordCurrentSlide(&current)
	meta.ComparePlan(plan, nil)

	snapshot := meta.Snapshot()
	if snapshot.IntentAnchor.Summary != "生成一份季度经营复盘" {
		t.Fatalf("unexpected intent anchor: %#v", snapshot.IntentAnchor)
	}
	if len(snapshot.PlanSlides) != 2 || snapshot.PlanSlides[0].Title != "经营概览" {
		t.Fatalf("plan baseline was not frozen: %#v", snapshot.PlanSlides)
	}
	if snapshot.CurrentSlide == nil || snapshot.CurrentSlide.PageIndex != 2 {
		t.Fatalf("current slide missing from snapshot: %#v", snapshot.CurrentSlide)
	}
	if snapshot.AlignmentStatus != "aligned" || len(snapshot.AlignmentWarnings) != 0 {
		t.Fatalf("expected aligned snapshot, got %s %#v", snapshot.AlignmentStatus, snapshot.AlignmentWarnings)
	}

	plan[0].Title = "外部修改"
	snapshot.PlanSlides[0].Title = "快照修改"
	if got := meta.Snapshot().PlanSlides[0].Title; got != "经营概览" {
		t.Fatalf("runtime plan leaked mutable state: %q", got)
	}
}

func TestRuntimeMetaReportsStructuralAndDeliveryDeviations(t *testing.T) {
	meta := NewRuntimeMeta("task-2", t.TempDir())
	expected := []PlanSlide{
		{PageIndex: 1, TaskID: "slide-1", Title: "封面", ContentType: "title_slide", OutputFile: "1_cover.pptx"},
		{PageIndex: 2, TaskID: "slide-2", Title: "数据", ContentType: "chart_slide", OutputFile: "2_chart.pptx"},
	}
	meta.FreezePlan(expected)
	observed := []PlanSlide{
		{PageIndex: 1, TaskID: "slide-1", Title: "新封面", ContentType: "content_slide", OutputFile: "1_cover.pptx"},
		{PageIndex: 1, TaskID: "slide-copy", Title: "重复页", ContentType: "title_slide", OutputFile: "1_copy.pptx"},
	}
	meta.ComparePlan(observed, []string{"2_chart.pptx"})

	snapshot := meta.Snapshot()
	if snapshot.AlignmentStatus != "warning" {
		t.Fatalf("expected warning status, got %q", snapshot.AlignmentStatus)
	}
	wantCodes := map[string]bool{
		"title_changed":        false,
		"content_type_changed": false,
		"duplicate_page_index": false,
		"planned_page_missing": false,
		"output_file_missing":  false,
	}
	for _, warning := range snapshot.AlignmentWarnings {
		if _, ok := wantCodes[warning.Code]; ok {
			wantCodes[warning.Code] = true
		}
	}
	for code, found := range wantCodes {
		if !found {
			t.Errorf("missing alignment warning %q in %#v", code, snapshot.AlignmentWarnings)
		}
	}

	if err := meta.WriteReport("completed"); err != nil {
		t.Fatalf("WriteReport() error = %v", err)
	}
	data, err := os.ReadFile(filepath.Join(meta.WorkDir, "runtime_report.json"))
	if err != nil {
		t.Fatalf("read runtime report: %v", err)
	}
	var report RuntimeReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("decode runtime report: %v", err)
	}
	if report.Snapshot.AlignmentStatus != "warning" || len(report.Snapshot.AlignmentWarnings) == 0 {
		t.Fatalf("alignment fields missing from report: %#v", report.Snapshot)
	}
}

func TestRuntimeMetaEmitsCompactEventDetailsToSink(t *testing.T) {
	workDir := t.TempDir()
	meta := NewRuntimeMeta("task-observe", workDir)
	var events []RuntimeEvent
	meta.SetEventSink(func(event RuntimeEvent) {
		events = append(events, event)
	})
	fullArgs := `{"code":"print(\"hello\")","notes":"keep the complete payload"}`
	fullResult := "stdout:\nhello\nfull result body"
	meta.RecordToolStart("python3", fullArgs)
	meta.RecordToolEnd("python3", fullArgs, fullResult)
	meta.RecordLLMStartDetails("ChatModel", map[string]any{
		"history": []map[string]any{
			{"role": "user", "content": "build a deck"},
			{"role": "meta", "content": "llm_call_metadata", "metadata": map[string]any{"model": "test-model"}},
		},
	})
	meta.RecordLLMEndDetails("ChatModel", 11, 7, 18, map[string]any{
		"assistant_output": "done",
		"output":           map[string]any{"role": "assistant", "content": "done"},
	})
	if err := meta.WriteReport("completed"); err != nil {
		t.Fatalf("WriteReport() error = %v", err)
	}
	if len(events) != 4 {
		t.Fatalf("events = %d, want 4: %#v", len(events), events)
	}
	toolEnd := events[1]
	if toolEnd.Metadata["args"] != fullArgs || toolEnd.Metadata["result"] != fullResult {
		t.Fatalf("tool metadata should persist auditable payloads: %#v", toolEnd.Metadata)
	}
	if toolEnd.Metadata["args_preview"] == "" || toolEnd.Metadata["result_preview"] == "" {
		t.Fatalf("tool previews missing: %#v", toolEnd.Metadata)
	}
	llmStart := events[2]
	history, ok := llmStart.Metadata["history"].([]any)
	if !ok || len(history) != 2 {
		t.Fatalf("llm compact history missing: %#v", llmStart.Metadata)
	}
	llmEnd := events[3]
	if llmEnd.Metadata["output"] != nil {
		t.Fatalf("opaque llm output should still be omitted: %#v", llmEnd.Metadata)
	}
	if llmEnd.Metadata["assistant_output"] != "done" {
		t.Fatalf("assistant output missing from full llm event: %#v", llmEnd.Metadata)
	}
	if llmEnd.Metadata["prompt_tokens"].(int64) != 11 || llmEnd.Metadata["total_tokens"].(int64) != 18 {
		t.Fatalf("token detail missing: %#v", llmEnd.Metadata)
	}
	snapshot := meta.Snapshot()
	if snapshot.RecentEvents[0].Metadata["args_preview"] == nil || snapshot.RecentEvents[1].Metadata["result_preview"] == nil {
		t.Fatalf("snapshot summary should keep bounded tool previews: %#v", snapshot.RecentEvents)
	}
	if snapshot.RecentEvents[0].Metadata["args"] != nil || snapshot.RecentEvents[1].Metadata["result"] != nil || snapshot.RecentEvents[2].Metadata != nil {
		t.Fatalf("snapshot summary should omit raw tool payloads and llm inputs: %#v", snapshot.RecentEvents)
	}
	if got := snapshot.RecentEvents[3].Metadata["assistant_output"]; got != "done" {
		t.Fatalf("snapshot summary should keep assistant output, got %#v", snapshot.RecentEvents[3].Metadata)
	}

	reportSnapshot, err := LoadRuntimeMetaSnapshot(workDir)
	if err != nil {
		t.Fatalf("LoadRuntimeMetaSnapshot() error = %v", err)
	}
	if len(reportSnapshot.RecentEvents) != 0 {
		t.Fatalf("runtime report should not store full events on disk: %#v", reportSnapshot.RecentEvents)
	}
}

func TestRuntimeMetaRedactsSecretsFromRawPayloads(t *testing.T) {
	meta := NewRuntimeMeta("task-redact", t.TempDir())
	var events []RuntimeEvent
	meta.SetEventSink(func(event RuntimeEvent) {
		events = append(events, event)
	})

	meta.RecordToolEnd("call_api", `{"api_key":"sk-secretvalue123456","query":"ok"}`, "Authorization: Bearer abcdefghijklmnop")

	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	args := events[0].Metadata["args"].(string)
	result := events[0].Metadata["result"].(string)
	if strings.Contains(args, "sk-secretvalue") || strings.Contains(result, "abcdefghijklmnop") {
		t.Fatalf("secret was not redacted: %#v", events[0].Metadata)
	}
	if !strings.Contains(args, "[REDACTED]") || !strings.Contains(result, "[REDACTED]") {
		t.Fatalf("redaction marker missing: %#v", events[0].Metadata)
	}
}

func TestRuntimeMetaExtractsObservableSearchAndManifestFields(t *testing.T) {
	meta := NewRuntimeMeta("task-observe-fields", t.TempDir())
	var events []RuntimeEvent
	meta.SetEventSink(func(event RuntimeEvent) {
		events = append(events, event)
	})

	meta.RecordToolStart("search", `{"query":"延安 红色旅游 数据","reason":"核实最新文旅资料"}`)
	meta.RecordToolEnd("search", "", `{"results":[{"title":"延安市人民政府","url":"https://www.yanan.gov.cn/a","description":"发布红色旅游接待数据","source":"延安市人民政府","date":"2026-07-01"},{"title":"百科","url":"https://baike.baidu.com/b"}]}`)
	meta.RecordToolStart("update_tasks_manifest", `{"template":"generic","theme":"government_red","tasks":[{"task_id":"slide-1"},{"task_id":"slide-2"}]}`)

	if events[0].Metadata["search_query"] != "延安 红色旅游 数据" || events[0].Metadata["search_reason"] != "核实最新文旅资料" {
		t.Fatalf("search args were not extracted: %#v", events[0].Metadata)
	}
	urls, ok := events[1].Metadata["source_urls"].([]string)
	if !ok || len(urls) != 2 || urls[0] != "https://www.yanan.gov.cn/a" {
		t.Fatalf("search urls were not extracted: %#v", events[1].Metadata)
	}
	if events[1].Metadata["source_count"] != 2 {
		t.Fatalf("search source count was not extracted: %#v", events[1].Metadata)
	}
	searchResults, ok := events[1].Metadata["search_results"].([]any)
	if !ok || len(searchResults) != 2 {
		t.Fatalf("readable search results were not extracted: %#v", events[1].Metadata)
	}
	firstResult, ok := searchResults[0].(map[string]any)
	if !ok || firstResult["description"] != "发布红色旅游接待数据" || firstResult["source"] != "延安市人民政府" {
		t.Fatalf("search result details invalid: %#v", searchResults[0])
	}
	titles, ok := events[1].Metadata["source_titles"].([]string)
	if !ok || len(titles) != 2 || titles[0] != "延安市人民政府" {
		t.Fatalf("search source titles were not extracted: %#v", events[1].Metadata)
	}
	if events[2].Metadata["slide_count"] != 2 || events[2].Metadata["template"] != "generic" {
		t.Fatalf("manifest fields were not extracted: %#v", events[2].Metadata)
	}
}

func TestRuntimeMetaExtractsImageSearchPreviews(t *testing.T) {
	meta := NewRuntimeMeta("task-image-search", t.TempDir())
	var events []RuntimeEvent
	meta.SetEventSink(func(event RuntimeEvent) {
		events = append(events, event)
	})

	meta.RecordToolStart("search_images", `{"query":"aerial city skyline, wide landscape","asset_purpose":"background","asset_subject":"city skyline","composition":"clean negative space"}`)
	meta.RecordToolEnd("search_images", "", `{
		"provider":"unsplash",
		"asset_purpose":"background",
		"asset_query":"aerial city skyline, wide landscape",
		"photos":[{
			"id":"abc",
			"preview_url":"https://images.unsplash.com/small.jpg",
			"image_url":"https://images.unsplash.com/photo.jpg",
			"source_url":"https://unsplash.com/photos/abc",
			"photographer":"Demo",
			"attribution":"Photo by Demo on Unsplash",
			"local_path":"assets/images/unsplash_abc.jpg"
		}]
	}`)

	if events[0].Metadata["image_query"] != "aerial city skyline, wide landscape" {
		t.Fatalf("image args were not extracted: %#v", events[0].Metadata)
	}
	previews, ok := events[1].Metadata["image_results"].([]any)
	if !ok || len(previews) != 1 {
		t.Fatalf("image previews were not extracted: %#v", events[1].Metadata)
	}
	first, ok := previews[0].(map[string]any)
	if !ok || first["local_path"] != "assets/images/unsplash_abc.jpg" || first["preview_url"] == "" {
		t.Fatalf("image preview metadata invalid: %#v", previews[0])
	}
	if events[1].Metadata["provider"] != "unsplash" || events[1].Metadata["asset_query"] != "aerial city skyline, wide landscape" {
		t.Fatalf("image search summary metadata invalid: %#v", events[1].Metadata)
	}
	urls, ok := events[1].Metadata["source_urls"].([]string)
	if !ok || len(urls) != 1 || urls[0] != "https://unsplash.com/photos/abc" {
		t.Fatalf("image source urls were not extracted: %#v", events[1].Metadata)
	}
}

func TestRuntimeMetaDeduplicatesManifestValidationAndUsesProgressStatus(t *testing.T) {
	meta := NewRuntimeMeta("task-manifest", t.TempDir())
	var events []RuntimeEvent
	meta.SetEventSink(func(event RuntimeEvent) {
		events = append(events, event)
	})

	meta.RecordManifestValidation(0, 2, nil, []string{"slide-1", "slide-2"})
	meta.RecordManifestValidation(0, 2, nil, []string{"slide-1", "slide-2"})
	meta.RecordManifestValidation(1, 2, nil, []string{"slide-2"})
	meta.RecordManifestValidation(2, 2, nil, nil)

	if len(events) != 3 {
		t.Fatalf("manifest events = %d, want 3 state changes: %#v", len(events), events)
	}
	if events[0].Status != "running" || events[1].Status != "running" {
		t.Fatalf("pending manifest should be running, got %#v", events)
	}
	if events[2].Status != "ok" {
		t.Fatalf("complete manifest should be ok, got %#v", events[2])
	}
	if events[1].Detail != "已完成 1/2 页，还有 1 页待生成" {
		t.Fatalf("unexpected progress detail: %q", events[1].Detail)
	}
	if got := meta.Snapshot().EventCounts["deck_spec_validated"]; got != 3 {
		t.Fatalf("manifest event count = %d, want 3", got)
	}
	for _, event := range events {
		if len(event.Metadata) != 2 || event.Metadata["done"] == nil || event.Metadata["total"] == nil {
			t.Fatalf("manifest metadata should contain only done/total: %#v", event.Metadata)
		}
	}
}

func TestRuntimeStatusBarOnlyInjectsGeneratedAndTotalSlides(t *testing.T) {
	meta := NewRuntimeMeta("task-progress", t.TempDir())
	meta.RecordSlideProgress(3, 9, 2)
	meta.RecordToolStart("read_file", `{"path":"tasks.json"}`)
	status := meta.StatusBar()
	if status != "<agent_progress>\ngenerated_slides: 3\ntotal_slides: 9\n</agent_progress>" {
		t.Fatalf("unexpected model metadata: %q", status)
	}
}

func TestSlideProgressEventMetadataOnlyContainsGeneratedAndTotalSlides(t *testing.T) {
	meta := NewRuntimeMeta("task-progress-event", t.TempDir())
	var event RuntimeEvent
	meta.SetEventSink(func(recorded RuntimeEvent) { event = recorded })

	meta.RecordSlideProgress(3, 9, 2)

	if event.Kind != "delivery_progress" {
		t.Fatalf("event kind = %q, want delivery_progress", event.Kind)
	}
	if len(event.Metadata) != 2 || event.Metadata["done"] != 3 || event.Metadata["total"] != 9 {
		t.Fatalf("slide progress metadata should contain only done/total: %#v", event.Metadata)
	}
}

func TestRuntimeMetaPersistsEventsBeyondRecentWindow(t *testing.T) {
	meta := NewRuntimeMeta("task-history", t.TempDir())
	var persisted []RuntimeEvent
	meta.SetEventSink(func(event RuntimeEvent) { persisted = append(persisted, event) })
	for i := 0; i < runtimeRecentEventLimit+25; i++ {
		meta.RecordToolStart("read_file", string(rune('a'+i%20)))
	}
	if len(meta.Snapshot().RecentEvents) != runtimeRecentEventLimit {
		t.Fatalf("recent window size = %d", len(meta.Snapshot().RecentEvents))
	}
	if len(persisted) != runtimeRecentEventLimit+25 || persisted[0].ID != 1 {
		t.Fatalf("persistent sink lost early events: %d %#v", len(persisted), persisted[:1])
	}
}

func TestCompressionEventIncludesDifferenceAndUserAnchor(t *testing.T) {
	meta := NewRuntimeMeta("task-compress", t.TempDir())
	var event RuntimeEvent
	meta.SetEventSink(func(recorded RuntimeEvent) { event = recorded })
	meta.RecordCompressionDetails(120, 28, 42000, 11000, "生成一套生态报告", []string{"使用智能推荐", "突出背景图片"})
	if event.Kind != "planner_context_compressed" || event.Metadata["removed_messages"] != 92 || event.Metadata["saved_tokens"] != 31000 {
		t.Fatalf("compression diff missing: %#v", event)
	}
	requirements, ok := event.Metadata["preserved_requirements"].([]string)
	if !ok || len(requirements) != 2 || event.Metadata["user_intent_summary"] != "生成一套生态报告" {
		t.Fatalf("compression user anchor missing: %#v", event.Metadata)
	}
}

func TestCompressionSummaryAnchorsWholeUserMessages(t *testing.T) {
	summary := ExtractKeyDecisions([]*schema.Message{
		schema.SystemMessage("system"),
		schema.UserMessage("生成一套 12 页生态保护报告，重点展示真实数据和背景图片"),
		schema.AssistantMessage("开始规划", nil),
		schema.UserMessage("<agent_progress>\ngenerated_slides: 3\ntotal_slides: 12\n</agent_progress>"),
		schema.UserMessage("配色改为 report_green，并保留用户提供的大纲标题"),
	})
	if summary.UserIntentSummary != "生成一套 12 页生态保护报告，重点展示真实数据和背景图片" {
		t.Fatalf("user intent anchor = %q", summary.UserIntentSummary)
	}
	if len(summary.PreservedRequirements) != 2 || summary.PreservedRequirements[1] != "配色改为 report_green，并保留用户提供的大纲标题" {
		t.Fatalf("preserved requirements = %#v", summary.PreservedRequirements)
	}
}

func TestRuntimeMetaMarksMissingManifestFilesAsWarning(t *testing.T) {
	meta := NewRuntimeMeta("task-missing", t.TempDir())
	var event RuntimeEvent
	meta.SetEventSink(func(recorded RuntimeEvent) {
		event = recorded
	})

	meta.RecordManifestValidation(2, 2, []string{"2_summary.pptx"}, nil)

	if event.Status != "warning" {
		t.Fatalf("missing files status = %q, want warning", event.Status)
	}
	if event.Detail != "已完成 2/2 页，缺少 1 个文件" {
		t.Fatalf("unexpected warning detail: %q", event.Detail)
	}
}
