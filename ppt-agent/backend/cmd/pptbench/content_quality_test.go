package main

import (
	"testing"

	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
)

func TestAssessContentQualityReportsClaimsAndRepeatedLayout(t *testing.T) {
	manifest := &deck.TasksManifest{Tasks: []*deck.TaskItem{
		{PageIndex: 1, ContentType: "title_slide", ContentPlan: &deck.ContentPlan{Components: []deck.PlanComponent{{Type: "key_point", Body: "服务化拆分是解决发布、性能和稳定性瓶颈的结构性方案。"}}}},
		{PageIndex: 2, ContentType: "content_slide", ContentPlan: &deck.ContentPlan{Components: []deck.PlanComponent{{Type: "insight", Body: "发布窗口、P95 延迟和连接池耗尽共同说明单体架构已到边界。"}}}},
		{PageIndex: 3, ContentType: "content_slide", ContentPlan: &deck.ContentPlan{Components: []deck.PlanComponent{{Type: "insight", Body: "发布窗口、P95 延迟和连接池耗尽共同说明单体架构已到边界。"}}}},
		{PageIndex: 4, ContentType: "summary_slide", ContentPlan: &deck.ContentPlan{}},
	}}

	report := assessContentQuality(manifest)
	if report == nil || report.DeckClaim == "" {
		t.Fatalf("expected deck claim, got %#v", report)
	}
	if report.LongestRepeatedLayoutRun != 2 || len(report.RepeatedLayoutRunContentTypes) != 1 || report.RepeatedLayoutRunContentTypes[0] != "content_slide" {
		t.Fatalf("unexpected repeated-layout report: %#v", report)
	}
	if len(report.DuplicateClaimGroups) != 1 || len(report.DuplicateClaimGroups[0]) != 2 {
		t.Fatalf("expected duplicate claim group, got %#v", report.DuplicateClaimGroups)
	}
	if len(report.MissingClaimPages) != 1 || report.MissingClaimPages[0] != 4 {
		t.Fatalf("expected summary page missing claim, got %#v", report.MissingClaimPages)
	}
}

func TestCompactModelOutputKeepsContentQualityEvidence(t *testing.T) {
	quality := &contentQualityReport{DeckClaim: "核心判断"}
	output := compactModelOutput(agentOutput{CaseID: "quality-case", ContentQuality: quality})
	if output.ContentQuality != quality {
		t.Fatalf("content quality evidence was dropped: %#v", output)
	}
}

func TestAssessContentQualityReportsAgendaSubtitleContract(t *testing.T) {
	manifest := &deck.TasksManifest{Tasks: []*deck.TaskItem{
		{
			PageIndex:   2,
			ContentType: "agenda",
			ContentPlan: &deck.ContentPlan{Components: []deck.PlanComponent{
				{ID: "toc_1", Type: "toc_item", Title: "现状判断", Body: "先用服务指标确认客户体验为何正在恶化。"},
				{ID: "toc_2", Type: "toc_item", Title: "改进路径"},
				{ID: "toc_3", Type: "toc_item", Title: "行动闭环", Body: "行动闭环"},
			}},
		},
	}}

	report := assessContentQuality(manifest)
	if report.AgendaTOCItems != 3 || report.AgendaTOCSubtitles != 1 {
		t.Fatalf("unexpected agenda subtitle counts: %#v", report)
	}
	if len(report.AgendaSubtitleIssues) != 2 || report.AgendaSubtitleIssues[0].Code != "missing_toc_subtitle" || report.AgendaSubtitleIssues[1].Code != "repeated_toc_title" {
		t.Fatalf("expected distinct agenda subtitle issues, got %#v", report.AgendaSubtitleIssues)
	}
}
