package task

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
	"github.com/cloudwego/ppt-agent/pkg/db"
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

func TestTaskInfoToRecordDropsFourByteRunesForLegacyMySQL(t *testing.T) {
	record := taskInfoToRecord(&TaskInfo{
		ID:                  "task-emoji",
		Query:               "介绍桂林📋",
		ConversationContent: "中文保留，emoji移除✨",
		FullAnswer:          "完成✅",
		Error:               "错误🚫",
		Files:               []string{"1_桂林.pptx"},
	})

	if strings.ContainsAny(record.Query+record.ConversationContent+record.FullAnswer+record.Error, "📋✨✅🚫") {
		t.Fatalf("record still contains four-byte runes: %#v", record)
	}
	if !strings.Contains(record.ConversationContent, "中文保留") || !strings.Contains(record.Query, "介绍桂林") {
		t.Fatalf("BMP text was not preserved: %#v", record)
	}
}

func TestPersistConversationContentDoesNotHoldTaskLockDuringDatabaseRead(t *testing.T) {
	oldList := listConversationMessagesForSummary
	started := make(chan struct{})
	release := make(chan struct{})
	listConversationMessagesForSummary = func(string) ([]db.ConversationMessage, error) {
		close(started)
		<-release
		return nil, nil
	}
	defer func() { listConversationMessagesForSummary = oldList }()

	ts := &TaskState{Info: TaskInfo{ID: "task-slow-db", WorkDir: t.TempDir()}}
	ts.fullAnswer.WriteString("内存中的完整回答")
	persisted := make(chan struct{})
	go func() {
		ts.persistConversationContent()
		close(persisted)
	}()
	<-started

	answer := make(chan string, 1)
	go func() { answer <- ts.FullAnswer() }()
	select {
	case got := <-answer:
		if got != "内存中的完整回答" {
			t.Fatalf("full answer = %q", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("FullAnswer blocked behind database persistence")
	}

	close(release)
	select {
	case <-persisted:
	case <-time.After(time.Second):
		t.Fatal("conversation persistence did not finish after database release")
	}
}

func TestTaskStatePersistsOneMarkdownTurnAtExplicitBoundary(t *testing.T) {
	var turns []string
	ts := &TaskState{
		Info:      TaskInfo{ID: "task-1", Status: TaskStatusRunning},
		listeners: make(map[string]chan SSERichEvent),
		assistantTurnFn: func(_, _ string, content string) {
			turns = append(turns, content)
		},
	}

	ts.Broadcast(SSERichEvent{Type: "answer", Content: "## 结果\n\n"})
	ts.Broadcast(SSERichEvent{Type: "tool_call", ToolName: "python"})
	ts.Broadcast(SSERichEvent{Type: "progress", Done: 1, Total: 2})
	ts.Broadcast(SSERichEvent{Type: "answer", Content: "- 第一页完成\n- 第二页完成"})
	if len(turns) != 0 {
		t.Fatalf("turn persisted before answer_end: %#v", turns)
	}
	ts.Broadcast(SSERichEvent{Type: "answer_end"})
	ts.Broadcast(SSERichEvent{Type: "complete"})

	want := []string{"## 结果\n\n- 第一页完成\n- 第二页完成"}
	if !reflect.DeepEqual(turns, want) {
		t.Fatalf("turns = %#v, want %#v", turns, want)
	}
}

func TestTaskStateTelemetryDoesNotEnterAssistantTurn(t *testing.T) {
	var turns []string
	ts := &TaskState{
		Info:      TaskInfo{ID: "task-telemetry", Status: TaskStatusRunning},
		listeners: make(map[string]chan SSERichEvent),
		assistantTurnFn: func(_, _ string, content string) {
			turns = append(turns, content)
		},
	}

	ts.Broadcast(SSERichEvent{Type: "system_step", Content: "【步骤1/3】模型意图分析"})
	ts.Broadcast(SSERichEvent{Type: "progress", Phase: "planning", PhaseDetail: "正在写入 DeckSpec"})
	ts.Broadcast(SSERichEvent{Type: "tool_call", ToolName: "read_file", ToolArgs: `{"path":"template.json"}`})
	ts.Broadcast(SSERichEvent{Type: "answer_end"})
	ts.Broadcast(SSERichEvent{Type: "complete"})

	if got := strings.TrimSpace(ts.FullAnswer()); got != "" {
		t.Fatalf("telemetry leaked into full answer: %q", got)
	}
	if len(turns) != 0 {
		t.Fatalf("telemetry persisted as assistant turns: %#v", turns)
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

func TestApplyDeliverySnapshotOutcomeRejectsIncompleteMetadata(t *testing.T) {
	ts := &TaskState{Info: TaskInfo{Status: TaskStatusCompleted}}
	delivery := DeliverySnapshot{
		Total: 23, Done: 21, PendingTasks: []string{"22", "23"},
	}

	if !applyDeliverySnapshotOutcome(ts, delivery) {
		t.Fatal("incomplete delivery was not rejected")
	}
	if ts.Info.Status != TaskStatusFailed {
		t.Fatalf("status = %q, want failed", ts.Info.Status)
	}
	if !strings.Contains(ts.Info.Error, "已交付 21/23 页") || !strings.Contains(ts.Info.Error, "仍有 2 页待完成") {
		t.Fatalf("unexpected error: %q", ts.Info.Error)
	}
}

func TestApplyDeliverySnapshotOutcomeRejectsEmptyMetadata(t *testing.T) {
	ts := &TaskState{Info: TaskInfo{Status: TaskStatusCompleted}}
	if !applyDeliverySnapshotOutcome(ts, DeliverySnapshot{}) {
		t.Fatal("empty delivery metadata was not rejected")
	}
	if ts.Info.Status != TaskStatusFailed || !strings.Contains(ts.Info.Error, "没有有效页面") {
		t.Fatalf("status=%q error=%q", ts.Info.Status, ts.Info.Error)
	}
}

func TestApplyDeliverySnapshotOutcomeAllowsCompleteMetadata(t *testing.T) {
	ts := &TaskState{Info: TaskInfo{Status: TaskStatusCompleted}}
	delivery := DeliverySnapshot{Total: 23, Done: 23, Files: []string{"1.pptx"}}
	if applyDeliverySnapshotOutcome(ts, delivery) {
		t.Fatal("verified delivery was rejected")
	}
	if ts.Info.Status != TaskStatusCompleted || ts.Info.Error != "" {
		t.Fatalf("status=%q error=%q", ts.Info.Status, ts.Info.Error)
	}
}

func TestApplyDeliverySnapshotOutcomePreservesUserCancellation(t *testing.T) {
	ts := &TaskState{Info: TaskInfo{Status: TaskStatusCancelled, Error: "任务已被用户中断"}}
	if applyDeliverySnapshotOutcome(ts, DeliverySnapshot{}) {
		t.Fatal("cancelled task was revalidated")
	}
	if ts.Info.Status != TaskStatusCancelled {
		t.Fatalf("status=%q, want cancelled", ts.Info.Status)
	}
}

func TestShouldSyncDeliveryAfterRunSkipsPlannerFailureWithoutManifest(t *testing.T) {
	workDir := t.TempDir()

	if shouldSyncDeliveryAfterRun(workDir, true) {
		t.Fatal("planner failure without tasks.json should not sync delivery")
	}
	if !shouldSyncDeliveryAfterRun(workDir, false) {
		t.Fatal("successful planner path should sync delivery")
	}
	if err := os.WriteFile(filepath.Join(workDir, "tasks.json"), []byte(`{"tasks":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if !shouldSyncDeliveryAfterRun(workDir, true) {
		t.Fatal("render failure with tasks.json should sync delivery")
	}
}

func TestPollProgressSignalsMetadataCompletion(t *testing.T) {
	workDir := t.TempDir()
	manifest := &deck.TasksManifest{Tasks: []*deck.TaskItem{{
		TaskID: "1", PageIndex: 1, Title: "封面", ContentType: "title_slide",
		OutputFile: "1_封面.pptx", Status: deck.StatusPending,
	}}}
	if err := deck.WriteTasksManifest(workDir, manifest); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workDir, "1_封面.pptx"), []byte("pptx"), 0600); err != nil {
		t.Fatal(err)
	}
	ts := &TaskState{
		Info:      TaskInfo{ID: "task-1", Status: TaskStatusRunning},
		listeners: make(map[string]chan SSERichEvent), reportedFiles: make(map[string]bool),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	completed := make(chan DeliverySnapshot, 1)
	go (&TaskManager{}).pollProgress(ctx, ts, workDir, func(snapshot DeliverySnapshot) {
		completed <- snapshot
	})

	select {
	case snapshot := <-completed:
		if !snapshot.Complete() || snapshot.Done != 1 || snapshot.Total != 1 {
			t.Fatalf("snapshot=%#v", snapshot)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("metadata completion was not signaled")
	}
}
