package web

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cloudwego/ppt-agent/pkg/runtime/task"
)

func TestWriteSSEIncludesEventID(t *testing.T) {
	var output bytes.Buffer
	writeSSEToWriter(&output, nil, task.SSERichEvent{ID: 17, Type: "progress", Done: 2, Total: 4})

	message := output.String()
	if !strings.HasPrefix(message, "id: 17\nevent: progress\n") {
		t.Fatalf("unexpected SSE envelope: %q", message)
	}
	if !strings.Contains(message, `"id":17`) {
		t.Fatalf("SSE JSON payload does not include event ID: %q", message)
	}
}
