package task

import (
	"testing"
	"time"
)

func TestTaskStateBroadcastAssignsIncreasingEventIDs(t *testing.T) {
	ts := &TaskState{
		Info:      TaskInfo{Status: TaskStatusRunning},
		listeners: make(map[string]chan SSERichEvent),
	}

	ts.Broadcast(SSERichEvent{Type: "progress"})
	ts.Broadcast(SSERichEvent{Type: "file_ready"})
	ts.Broadcast(SSERichEvent{Type: "complete"})

	if len(ts.Events) != 3 {
		t.Fatalf("len(Events) = %d, want 3", len(ts.Events))
	}
	for i, event := range ts.Events {
		want := uint64(i + 1)
		if event.ID != want {
			t.Fatalf("Events[%d].ID = %d, want %d", i, event.ID, want)
		}
	}
}

func TestTaskStateSubscribeFromReplaysOnlyNewerEvents(t *testing.T) {
	ts := &TaskState{
		Info:      TaskInfo{Status: TaskStatusRunning},
		listeners: make(map[string]chan SSERichEvent),
	}
	for _, eventType := range []string{"system_step", "progress", "file_ready"} {
		ts.Broadcast(SSERichEvent{Type: eventType})
	}

	ch := make(chan SSERichEvent, 1)
	events, done := ts.SubscribeFrom("listener", ch, 1)
	if done {
		t.Fatal("SubscribeFrom reported a running task as done")
	}
	if len(events) != 2 || events[0].ID != 2 || events[1].ID != 3 {
		t.Fatalf("replayed events = %#v, want IDs 2 and 3", events)
	}

	ts.Broadcast(SSERichEvent{Type: "thumbnail_ready"})
	select {
	case event := <-ch:
		if event.ID != 4 {
			t.Fatalf("live event ID = %d, want 4", event.ID)
		}
	default:
		t.Fatal("live event was not delivered after atomic subscription")
	}
}

func TestTaskStateSubscribeFromDoesNotRegisterCompletedTask(t *testing.T) {
	ts := &TaskState{
		Info:      TaskInfo{Status: TaskStatusCompleted},
		listeners: make(map[string]chan SSERichEvent),
	}
	ts.Broadcast(SSERichEvent{Type: "complete"})

	events, done := ts.SubscribeFrom("listener", make(chan SSERichEvent, 1), 0)
	if !done || len(events) != 1 {
		t.Fatalf("done=%v events=%d, want true and one event", done, len(events))
	}
	if len(ts.listeners) != 0 {
		t.Fatalf("completed task registered %d live listeners", len(ts.listeners))
	}
}

func TestReportFileReadyBroadcastsBeforeStartingCallback(t *testing.T) {
	tm := &TaskManager{}
	callbackStarted := make(chan string, 1)
	tm.SetFileReadyCallback(func(taskID, workDir, filename string) {
		callbackStarted <- taskID + "|" + workDir + "|" + filename
	})
	ts := &TaskState{
		Info:      TaskInfo{ID: "task-1", Status: TaskStatusRunning},
		listeners: make(map[string]chan SSERichEvent),
	}

	tm.reportFileReady(ts, "work-dir", "1_intro.pptx")
	if len(ts.Events) != 1 || ts.Events[0].Type != "file_ready" {
		t.Fatalf("events = %#v, want an immediate file_ready event", ts.Events)
	}

	select {
	case got := <-callbackStarted:
		if got != "task-1|work-dir|1_intro.pptx" {
			t.Fatalf("callback payload = %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("file-ready callback was not scheduled")
	}
}
