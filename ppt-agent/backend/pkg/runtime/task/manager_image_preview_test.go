package task

import (
	"testing"

	agentutils "github.com/cloudwego/ppt-agent/pkg/runtime/model"
)

func TestRuntimeImageSearchResultCarriesPreviewToTimeline(t *testing.T) {
	meta := agentutils.NewRuntimeMeta("task-images", t.TempDir())
	var events []agentutils.RuntimeEvent
	meta.SetEventSink(func(event agentutils.RuntimeEvent) { events = append(events, event) })
	args := `{"query":"flow battery storage"}`
	meta.RecordToolStart("search_images", args)
	meta.RecordToolEnd("search_images", args, `{"provider":"unsplash","photos":[{"preview_url":"https://images.unsplash.com/flow-small.jpg","image_url":"https://images.unsplash.com/flow.jpg","source_url":"https://unsplash.com/photos/flow","attribution":"Photo by Test on Unsplash"}]}`)

	ts := &TaskState{listeners: make(map[string]chan SSERichEvent)}
	for _, event := range events {
		broadcastRuntimeSummary(ts, agentutils.RuntimeEventSummary(event))
	}
	if len(ts.Events) != 2 || ts.Events[1].Type != SSEEventToolResult {
		t.Fatalf("timeline events = %#v", ts.Events)
	}
	images, ok := ts.Events[1].ToolPreview["images"].([]map[string]string)
	if !ok || len(images) != 1 || images[0]["thumbnail_url"] != "https://images.unsplash.com/flow-small.jpg" {
		t.Fatalf("tool preview = %#v", ts.Events[1].ToolPreview)
	}
}
