package deck

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestPlanReviewRevisionPayloadIncludesFailedSectionOnly(t *testing.T) {
	manifest := &TasksManifest{
		Title: "测试", Theme: "ocean_soft", Template: "generic",
		Tasks: []*TaskItem{
			{TaskID: "slide-1", PageIndex: 1, SectionID: "intro", SectionTitle: "引入", Title: "封面", ContentType: "title_slide", Description: "封面", OutputFile: "1.pptx", Status: StatusPending},
			{TaskID: "slide-2", PageIndex: 2, SectionID: "body", SectionTitle: "正文", Title: "问题", ContentType: "content_slide", Description: "短", OutputFile: "2.pptx", Status: StatusPending},
			{TaskID: "slide-3", PageIndex: 3, SectionID: "body", SectionTitle: "正文", Title: "分析", ContentType: "content_slide", Description: "分析", OutputFile: "3.pptx", Status: StatusPending},
			{TaskID: "slide-4", PageIndex: 4, SectionID: "end", SectionTitle: "结尾", Title: "总结", ContentType: "summary_slide", Description: "总结", OutputFile: "4.pptx", Status: StatusPending},
		},
	}
	report := &PlanReviewReport{
		Summary: "第 2 页信息密度不足",
		Issues: []PlanReviewIssue{
			{Code: "low_information_density", Severity: "error", PageIndex: 2, Message: "页面 description 过短"},
			{Code: "weak_narrative", Severity: "warning", PageIndex: 4, Message: "非阻塞提示"},
		},
	}

	payload := buildPlanReviewRevisionPayload(manifest, 1, report)

	if got := strings.Join(payload.Scope.AllowedTaskIDs, ","); got != "slide-2,slide-3" {
		t.Fatalf("allowed task ids = %q, want failed section only", got)
	}
	if got := strings.Join(payload.Scope.SectionIDs, ","); got != "body" {
		t.Fatalf("section ids = %q, want body", got)
	}
	if len(payload.IncludedTasks) != 2 || payload.IncludedTasks[0].TaskID != "slide-2" || payload.IncludedTasks[1].TaskID != "slide-3" {
		t.Fatalf("included tasks = %#v, want slide-2 and slide-3", payload.IncludedTasks)
	}
	if len(payload.Issues) != 1 || payload.Issues[0].PageIndex != 2 {
		t.Fatalf("issues = %#v, want only issues for included pages/deck", payload.Issues)
	}
}

func TestBuildPlanReviewRevisionInputSerializesOnlyIncludedTasks(t *testing.T) {
	workDir := t.TempDir()
	manifest := &TasksManifest{
		Title: "测试", Theme: "ocean_soft", Template: "generic",
		Tasks: []*TaskItem{
			{TaskID: "slide-1", PageIndex: 1, Title: "正常页", ContentType: "content_slide", Description: "正常页内容", OutputFile: "1.pptx", Status: StatusPending},
			{TaskID: "slide-2", PageIndex: 2, Title: "问题页", ContentType: "content_slide", Description: "短", OutputFile: "2.pptx", Status: StatusPending},
		},
	}
	if err := WriteTasksDraftManifest(workDir, manifest); err != nil {
		t.Fatal(err)
	}
	report := &PlanReviewReport{
		Summary: "第 2 页信息密度不足",
		Issues:  []PlanReviewIssue{{Code: "low_information_density", Severity: "error", PageIndex: 2}},
	}

	input, allowed, err := buildPlanReviewRevisionInput(workDir, 1, report)
	if err != nil {
		t.Fatal(err)
	}
	if len(allowed) != 1 || allowed[0] != "slide-2" {
		t.Fatalf("allowed = %#v, want slide-2", allowed)
	}
	if strings.Contains(input, "正常页") || !strings.Contains(input, "问题页") {
		t.Fatalf("revision input should contain only failed task slice, got: %s", input)
	}
	var payload planReviewRevisionPayload
	start := strings.Index(input, "{")
	if start < 0 {
		t.Fatalf("revision input does not contain JSON payload: %s", input)
	}
	if err := json.Unmarshal([]byte(input[start:]), &payload); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}
	if len(payload.IncludedTasks) != 1 || payload.IncludedTasks[0].TaskID != "slide-2" {
		t.Fatalf("payload included tasks = %#v", payload.IncludedTasks)
	}
}

func TestScopedDraftPatchToolRejectsUnauthorizedTask(t *testing.T) {
	workDir := t.TempDir()
	manifest := &TasksManifest{
		Title: "测试", Theme: "ocean_soft", Template: "generic",
		Tasks: []*TaskItem{
			{TaskID: "slide-1", PageIndex: 1, Title: "授权页", ContentType: "content_slide", Description: "授权页内容", OutputFile: "1.pptx", Status: StatusPending},
			{TaskID: "slide-2", PageIndex: 2, Title: "未授权页", ContentType: "content_slide", Description: "未授权页内容", OutputFile: "2.pptx", Status: StatusPending},
		},
	}
	if err := WriteTasksDraftManifest(workDir, manifest); err != nil {
		t.Fatal(err)
	}

	patcher := newScopedDraftTasksPatchTool(workDir, []string{"slide-1"})
	result, err := patcher.InvokableRun(context.Background(), `{"tasks":[{"task_id":"slide-2","title":"越界修改"}]}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"ok":false`) || !strings.Contains(result, "not authorized") {
		t.Fatalf("unexpected patch result: %s", result)
	}
	updated, err := ReadTasksDraftManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Tasks[1].Title != "未授权页" {
		t.Fatalf("unauthorized patch changed draft: %#v", updated.Tasks[1])
	}
}
