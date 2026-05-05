package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (s *Server) handleStreamTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ts := s.tasks.GetTaskState(id)
	if ts == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "streaming not supported"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// 对于已结束的任务：直接回放历史事件并发送完成信号，不添加监听器
	ts.mu.Lock()
	done := ts.Info.Status != TaskStatusRunning
	events := make([]SSERichEvent, len(ts.events))
	copy(events, ts.events)
	ts.mu.Unlock()

	for _, evt := range events {
		writeSSE(w, flusher, evt)
	}

	if done {
		// 如果历史事件中没有 complete，补发一个
		if len(events) == 0 || events[len(events)-1].Type != "complete" {
			writeSSE(w, flusher, SSERichEvent{
				Type:   "complete",
				Status: ts.Info.Status,
				Done:   ts.Info.DoneCount,
				Total:  ts.Info.TotalCount,
				Files:  ts.Info.Files,
			})
		}
		return
	}

	listenerID := uuid.New().String()
	ch := make(chan SSERichEvent, 64)
	ts.addListener(listenerID, ch)
	defer ts.removeListener(listenerID)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-ch:
			if !ok {
				return
			}
			writeSSE(w, flusher, evt)
			if evt.Type == "complete" {
				time.Sleep(50 * time.Millisecond)
				return
			}
		}
	}
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, evt SSERichEvent) {
	data, _ := json.Marshal(evt)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Type, data)
	flusher.Flush()
}
