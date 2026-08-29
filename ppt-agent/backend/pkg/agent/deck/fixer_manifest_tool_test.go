package deck

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

const validToolTestManifest = `{
  "mode":"initialize",
  "title":"组件计划",
  "tasks":[{
    "page_index":1,
    "title":"原始标题",
    "content_type":"content_slide",
    "content_plan":{
      "summary":"用事实和判断解释核心主题。",
      "slide_intent":"建立听众对核心主题的基本认知。",
      "components":[
        {"id":"headline_1","type":"headline","text":"核心主题"},
        {"id":"argument_1","type":"argument_block","body":"这是一段用于验证组件计划的完整说明，涵盖背景、事实、判断和建议，确保页面不是只有空泛短句。"},
        {"id":"bg_1","type":"image","asset_purpose":"background","asset_query":"clean wide landscape background"}
      ]
    }
  }]
}`

func writeValidFinalManifest(t *testing.T, workDir string) {
	t.Helper()
	draftTool := newDraftManifestTool(workDir, nil, "组件计划")
	if _, err := draftTool.InvokableRun(context.Background(), validToolTestManifest); err != nil {
		t.Fatal(err)
	}
	manifest, err := ReadTasksDraftManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteTasksManifest(workDir, manifest); err != nil {
		t.Fatal(err)
	}
}

func TestDraftPatchToolCannotInitializeOrCommit(t *testing.T) {
	workDir := t.TempDir()
	configured := newDraftManifestTool(workDir, nil, "组件计划")
	if _, err := configured.InvokableRun(context.Background(), validToolTestManifest); err != nil {
		t.Fatal(err)
	}

	patcher := newDraftTasksPatchTool(workDir)
	result, err := patcher.InvokableRun(context.Background(), `{"mode":"commit","tasks":[{"page_index":1,"title":"审查后标题"}]}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"ok":false`) || !strings.Contains(result, "mode is not accepted") {
		t.Fatalf("reviewer patch should reject mode: %s", result)
	}
	if _, err := ReadTasksManifest(workDir); err == nil {
		t.Fatal("reviewer tool must not publish tasks.json")
	}
	draft, err := ReadTasksDraftManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if draft.Tasks[0].Title != "原始标题" {
		t.Fatalf("draft title = %q", draft.Tasks[0].Title)
	}
}

func TestDraftPatchToolAcceptsTasksJSONArrayString(t *testing.T) {
	workDir := t.TempDir()
	configured := newDraftManifestTool(workDir, nil, "组件计划")
	if _, err := configured.InvokableRun(context.Background(), validToolTestManifest); err != nil {
		t.Fatal(err)
	}

	patcher := newDraftTasksPatchTool(workDir)
	result, err := patcher.InvokableRun(context.Background(), `{"tasks":"[{\"page_index\":1,\"title\":\"字符串修订标题\"}]"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"ok":true`) {
		t.Fatalf("reviewer string patch failed: %s", result)
	}
	draft, err := ReadTasksDraftManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if draft.Tasks[0].Title != "字符串修订标题" {
		t.Fatalf("draft title = %q", draft.Tasks[0].Title)
	}
}

func TestScopedDraftPatchToolRejectsTopLevelPageIntent(t *testing.T) {
	workDir := t.TempDir()
	configured := newDraftManifestTool(workDir, nil, "组件计划")
	if _, err := configured.InvokableRun(context.Background(), validToolTestManifest); err != nil {
		t.Fatal(err)
	}

	patcher := newScopedDraftTasksPatchTool(workDir, []int{1})
	result, err := patcher.InvokableRun(context.Background(), `{"tasks":[{"page_index":1,"page_intent":"不应由 Reviewer 写入"}]}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"ok":false`) || !strings.Contains(result, "content_plan.slide_intent") {
		t.Fatalf("reviewer top-level page_intent should be rejected: %s", result)
	}
}

func TestPlannerManifestToolOnlyAllowsInitialize(t *testing.T) {
	workDir := t.TempDir()
	planner := newPlannerManifestTool(workDir, nil, "组件计划")
	result, err := planner.InvokableRun(context.Background(), `{"mode":"commit"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"ok":false`) || !strings.Contains(result, "only supports initialize") {
		t.Fatalf("planner boundary was not enforced: %s", result)
	}
	if _, err := ReadTasksDraftManifest(workDir); err == nil {
		t.Fatal("planner commit attempt must not create a draft")
	}
}

func TestSelectedTasksPatchToolRejectsUnauthorizedPage(t *testing.T) {
	workDir := t.TempDir()
	writeValidFinalManifest(t, workDir)

	fixer := newSelectedTasksPatchTool(workDir, []string{"1"})
	result, err := fixer.InvokableRun(context.Background(), `{"tasks":[{"page_index":2,"title":"越权修改"}]}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"ok":false`) || !strings.Contains(result, "not authorized") {
		t.Fatalf("unexpected fixer result: %s", result)
	}
	manifest, err := ReadTasksManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Tasks[0].Title != "原始标题" {
		t.Fatalf("unauthorized patch changed manifest: %q", manifest.Tasks[0].Title)
	}
}

func TestSelectedTasksPatchToolPreservesRuntimeIdentity(t *testing.T) {
	workDir := t.TempDir()
	writeValidFinalManifest(t, workDir)

	manifest, err := ReadTasksManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	fixer := newSelectedTasksPatchTool(workDir, []string{manifest.Tasks[0].TaskID})
	result, err := fixer.InvokableRun(context.Background(), `{"tasks":[{"page_index":1,"task_id":"slide-1","title":"新标题"}]}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"ok":false`) || !strings.Contains(result, "task_id is not part") {
		t.Fatalf("unexpected fixer result: %s", result)
	}
}

func TestSelectedTasksPatchToolAcceptsTasksJSONArrayString(t *testing.T) {
	workDir := t.TempDir()
	writeValidFinalManifest(t, workDir)

	manifest, err := ReadTasksManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	fixer := newSelectedTasksPatchTool(workDir, []string{manifest.Tasks[0].TaskID})
	result, err := fixer.InvokableRun(context.Background(), `{"tasks":"[{\"page_index\":1,\"title\":\"字符串定点修复\"}]"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"ok":true`) {
		t.Fatalf("fixer string patch failed: %s", result)
	}
	manifest, err = ReadTasksManifest(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Tasks[0].Title != "字符串定点修复" {
		t.Fatalf("manifest title = %q", manifest.Tasks[0].Title)
	}
}

func TestBuildFixerTaskSnapshotOnlyIncludesAuthorizedTaskIDs(t *testing.T) {
	workDir := t.TempDir()
	manifest := &TasksManifest{Tasks: []*TaskItem{
		{TaskID: "slide-intro", PageIndex: 1, Title: "引言", ContentType: "title_slide"},
		{TaskID: "slide-metrics", PageIndex: 2, Title: "指标", ContentType: "kpi_dashboard"},
	}}
	if err := WriteTasksManifest(workDir, manifest); err != nil {
		t.Fatal(err)
	}

	snapshot, pages, err := buildFixerTaskSnapshot(workDir, []string{"slide-metrics"})
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 1 || pages[0] != 2 {
		t.Fatalf("snapshot pages = %#v, want [2]", pages)
	}
	var tasks []TaskItem
	if err := json.Unmarshal([]byte(snapshot), &tasks); err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].TaskID != "slide-metrics" {
		t.Fatalf("snapshot tasks = %#v, want only slide-metrics", tasks)
	}
	if strings.Contains(snapshot, "slide-intro") {
		t.Fatalf("snapshot leaks unselected task: %s", snapshot)
	}
}

func TestBuildFixerTaskSnapshotRejectsUnknownTaskID(t *testing.T) {
	workDir := t.TempDir()
	writeValidFinalManifest(t, workDir)
	if _, _, err := buildFixerTaskSnapshot(workDir, []string{"missing-task"}); err == nil {
		t.Fatal("expected unknown task_id to fail")
	}
}

func TestPlanReviewAcceptsBackgroundAssetQueryWithoutDownloadedFile(t *testing.T) {
	workDir := t.TempDir()
	configured := newDraftManifestTool(workDir, nil, "组件计划")
	if _, err := configured.InvokableRun(context.Background(), validToolTestManifest); err != nil {
		t.Fatal(err)
	}
	report, err := ReviewTasksDraftManifest(workDir, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Passed {
		t.Fatalf("asset_query should be executable when image search is unavailable: %#v", report)
	}
	if _, ok, err := CommitReviewedTasksDraftManifestIfPresent(workDir); err != nil || !ok {
		t.Fatalf("reviewed draft was not committed: ok=%v err=%v", ok, err)
	}
}
