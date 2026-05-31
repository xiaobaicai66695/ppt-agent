package web

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/cloudwego/ppt-agent/pkg/task"
)

type flushWriter interface {
	Flush()
}

func (s *Server) handleStreamTask(c *gin.Context) {
	id := c.Param("id")
	ts := s.tasks.GetTaskState(id)
	if ts == nil {
		// 任务不在内存中 — 尝试从 MySQL 恢复
		info := s.tasks.GetTask(id)
		if info == nil {
			c.JSON(404, gin.H{"error": "task not found"})
			return
		}
		ts = s.tasks.NewColdTaskState(*info)
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// 尝试获取 flusher；如果为 nil，SSE 事件会由 transport 缓冲
	flusher, _ := c.Writer.(flushWriter)
	writer := c.Writer

	ts.Mu.Lock()
	done := ts.Info.Status != task.TaskStatusRunning
	events := make([]task.SSERichEvent, len(ts.Events))
	copy(events, ts.Events)
	ts.Mu.Unlock()

	for _, evt := range events {
		writeSSEToWriter(writer, flusher, evt)
	}

	if done {
		if len(events) == 0 || events[len(events)-1].Type != "complete" {
			writeSSEToWriter(writer, flusher, task.SSERichEvent{
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
	ch := make(chan task.SSERichEvent, 64)
	ts.AddListener(listenerID, ch)
	defer ts.RemoveListener(listenerID)

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			return false
		case <-heartbeat.C:
			w.Write([]byte(": hb\n\n"))
			if flusher != nil {
				flusher.Flush()
			}
			return true
		case evt, ok := <-ch:
			if !ok {
				return false
			}
			writeSSEToWriter(w, flusher, evt)
			if evt.Type == "complete" {
				return false
			}
			return true
		}
	})
}

func writeSSEToWriter(writer io.Writer, flusher flushWriter, evt task.SSERichEvent) {
	data, _ := json.Marshal(evt)
	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", evt.Type, data)
	writer.Write([]byte(msg))
	if flusher != nil {
		flusher.Flush()
	}
}
