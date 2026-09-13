package ppt

import (
	"strings"
	"testing"
)

func TestPlannedAssetSearchToolCallExposesPendingQueries(t *testing.T) {
	workDir := t.TempDir()
	manifest := &TasksManifest{Tasks: []*TaskItem{{
		TaskID: "slide-1", PageIndex: 1, ContentType: "image_text",
		ContentPlan: &ContentPlan{
			VisualIntent: &VisualIntent{AssetPurpose: "background", AssetQuery: "flow battery"},
			Components:   []PlanComponent{{ID: "scene", Type: "image", AssetPurpose: "scene", AssetQuery: "energy storage container"}},
		},
	}}}

	callID, args := plannedAssetSearchToolCall(workDir, manifest, 1)
	if callID != "asset-search-1" {
		t.Fatalf("call ID = %q, want asset-search-1", callID)
	}
	for _, want := range []string{"unsplash", "flow battery", "energy storage container"} {
		if !strings.Contains(args, want) {
			t.Fatalf("tool args %q missing %q", args, want)
		}
	}
}

func TestMaterializedAssetPreviewKeepsOnlyHTTPSImages(t *testing.T) {
	manifest := &TasksManifest{Tasks: []*TaskItem{{
		PageIndex: 2,
		ContentPlan: &ContentPlan{
			VisualIntent: &VisualIntent{
				PreviewURL:  "https://images.unsplash.com/flow-small.jpg",
				ImageURL:    "https://images.unsplash.com/flow.jpg",
				SourceURL:   "https://unsplash.com/photos/flow",
				Attribution: "Photo by Test on Unsplash",
			},
			Components: []PlanComponent{{ID: "unsafe", Type: "image", PreviewURL: "javascript:alert(1)"}},
		},
	}}}

	preview := materializedAssetPreview(manifest)
	images, ok := preview["images"].([]map[string]string)
	if !ok || len(images) != 1 {
		t.Fatalf("preview images = %#v, want one safe image", preview)
	}
	if images[0]["thumbnail_url"] != "https://images.unsplash.com/flow-small.jpg" || images[0]["source_url"] != "https://unsplash.com/photos/flow" {
		t.Fatalf("preview image = %#v", images[0])
	}
}
