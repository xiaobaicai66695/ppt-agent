package utils

import "testing"

func TestRuntimeMetaExtractsSafeImageSearchPreview(t *testing.T) {
	meta := NewRuntimeMeta("task-images", t.TempDir())
	var event RuntimeEvent
	meta.SetEventSink(func(recorded RuntimeEvent) { event = recorded })

	meta.RecordToolEnd("search_images", `{"query":"flow battery storage","per_page":2}`, `{
		"provider":"unsplash",
		"photos":[
			{"preview_url":"https://images.unsplash.com/flow-small.jpg","image_url":"https://images.unsplash.com/flow.jpg","source_url":"https://unsplash.com/photos/flow","photographer":"Test Photographer","attribution":"Photo by Test Photographer on Unsplash"},
			{"preview_url":"javascript:alert(1)","image_url":"","source_url":"https://unsplash.com/photos/unsafe"}
		]
	}`)

	if event.Metadata["image_query"] != "flow battery storage" || event.Metadata["provider"] != "unsplash" {
		t.Fatalf("image search metadata = %#v", event.Metadata)
	}
	images, ok := event.Metadata["image_results"].([]any)
	if !ok || len(images) != 1 {
		t.Fatalf("image results = %#v, want one safe preview", event.Metadata["image_results"])
	}
	image, ok := images[0].(map[string]any)
	if !ok || image["thumbnail_url"] != "https://images.unsplash.com/flow-small.jpg" {
		t.Fatalf("image preview = %#v", images[0])
	}

	public := RuntimeEventSummary(event)
	if _, ok := public.Metadata["image_results"].([]any); !ok {
		t.Fatalf("public image results missing: %#v", public.Metadata)
	}
}
