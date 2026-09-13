package task

import "testing"

func TestTaskStatePersistsOnlyPublicTimelineEvents(t *testing.T) {
	var persisted []SSERichEvent
	ts := &TaskState{
		Info:            TaskInfo{ID: "timeline-task"},
		listeners:       make(map[string]chan SSERichEvent),
		timelineEventFn: func(_ string, event SSERichEvent) { persisted = append(persisted, event) },
	}
	ts.Broadcast(SSERichEvent{Type: SSEEventThought, Content: "正在审查大纲"})
	ts.Broadcast(SSERichEvent{Type: SSEEventToolCall, ToolName: "read_file"})
	ts.Broadcast(SSERichEvent{Type: SSEEventToolResult, ToolName: "read_file", ToolResult: "已读取"})
	ts.Broadcast(SSERichEvent{Type: SSEEventFinalAnswer, Content: "PPT 已完成交付"})

	if len(persisted) != 3 {
		t.Fatalf("persisted events = %#v, want thought and tool call/result only", persisted)
	}
	for _, event := range persisted {
		if event.Type == SSEEventFinalAnswer {
			t.Fatal("final assistant turn must stay in conversation_messages, not duplicate the durable trace")
		}
	}
}
