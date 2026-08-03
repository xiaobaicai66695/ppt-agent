package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
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
