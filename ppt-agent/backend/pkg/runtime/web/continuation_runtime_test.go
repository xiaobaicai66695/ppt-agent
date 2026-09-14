package web

import (
	"testing"

	"github.com/cloudwego/ppt-agent/pkg/agent/ppt"
)

func TestEnsurePendingContinuationVisualsRepairsRequiredPages(t *testing.T) {
	manifest := &ppt.TasksManifest{
		VisualPolicy: &ppt.VisualPolicy{Mode: "required", MinImagePages: 1},
		Tasks: []*ppt.TaskItem{
			{
				TaskID: "slide-1", PageIndex: 1, Status: ppt.StatusDone,
				ContentPlan: &ppt.ContentPlan{VisualIntent: &ppt.VisualIntent{
					Role: "supporting_photo", AssetPurpose: "background", AssetQuery: "energy storage landscape",
				}},
			},
			{TaskID: "slide-2", PageIndex: 2, Status: ppt.StatusPending, ContentPlan: &ppt.ContentPlan{}},
		},
	}

	if got := ensurePendingContinuationVisuals(manifest); got != 1 {
		t.Fatalf("repaired = %d, want 1", got)
	}
	visual := manifest.Tasks[1].ContentPlan.VisualIntent
	if visual == nil || visual.AssetPurpose != "background" || visual.AssetQuery != "energy storage landscape" {
		t.Fatalf("pending page background = %#v", visual)
	}
	if manifest.VisualPolicy.MinImagePages != 2 {
		t.Fatalf("min_image_pages = %d, want 2", manifest.VisualPolicy.MinImagePages)
	}
}

func TestPPTRenderSSEMapsAssetSearchToToolTimeline(t *testing.T) {
	event := pptRenderSSE(ppt.PPTRenderEvent{Type: "asset_search_done", ToolCallID: "asset-1", ToolName: "search_images", ToolArgs: `{"query":"battery"}`, ToolResult: "素材已下载", ToolPreview: map[string]any{"images": []string{"https://example.com/image.jpg"}}})
	if event.Type != "tool_result" || event.ToolStatus != "success" || event.ToolCallID != "asset-1" || event.ToolPreview == nil {
		t.Fatalf("asset search event = %#v", event)
	}
}
