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

func (s *Server) handleStreamTask(c *gin.Context) {
	id := c.Param("id")
	ts := s.tasks.GetTaskState(id)
	if ts == nil {
		c.JSON(404, gin.H{"error": "task not found"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ts.Mu.Lock()
	done := ts.Info.Status != task.TaskStatusRunning
	events := make([]task.SSERichEvent, len(ts.Events))
	copy(events, ts.Events)
	ts.Mu.Unlock()

	flusher, _ := c.Writer.(interface{ Flush() })
	writer := c.Writer

	for _, evt := range events {
		writeSSEGin(writer, flusher, evt)
	}

	if done {
		if len(events) == 0 || events[len(events)-1].Type != "complete" {
			writeSSEGin(writer, flusher, task.SSERichEvent{
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

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			return false
		case evt, ok := <-ch:
			if !ok {
				return false
			}
			writeSSEGin(writer, flusher, evt)
			if evt.Type == "complete" {
				time.Sleep(50 * time.Millisecond)
				return false
			}
			return true
		}
	})
}

func writeSSEGin(writer interface{ Write([]byte) (int, error) }, flusher interface{ Flush() }, evt task.SSERichEvent) {
	data, _ := json.Marshal(evt)
	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", evt.Type, data)
	writer.Write([]byte(msg))
	if flusher != nil {
		flusher.Flush()
	}
}
