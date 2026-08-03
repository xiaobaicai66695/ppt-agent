package web

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
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

	flusher, _ := c.Writer.(flushWriter)
	writer := c.Writer

	afterValue := c.Query("after_id")
	if afterValue == "" {
		afterValue = c.GetHeader("Last-Event-ID")
	}
	afterEventID, _ := strconv.ParseUint(afterValue, 10, 64)
	listenerID := uuid.New().String()
	ch := make(chan task.SSERichEvent, 128)
	events, done := ts.SubscribeFrom(listenerID, ch, afterEventID)
	if !done {
		defer ts.RemoveListener(listenerID)
	}

	for _, evt := range events {
		writeSSEToWriter(writer, flusher, evt)
	}

	if done {
		if len(events) == 0 || events[len(events)-1].Type != "complete" {
			completeEvt := task.SSERichEvent{
				Type:   "complete",
				Status: ts.Info.Status,
				Done:   ts.Info.DoneCount,
				Total:  ts.Info.TotalCount,
				Files:  ts.Info.Files,
			}
			writeSSEToWriter(writer, flusher, completeEvt)
		}
		c.Writer.Write([]byte(": close\n\n"))
		c.Writer.Flush()
		c.Abort()
		return
	}

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
				c.Writer.Write([]byte(": close\n\n"))
				c.Writer.Flush()
				c.Abort()
				return false
			}
			return true
		}
	})
}

func writeSSEToWriter(writer io.Writer, flusher flushWriter, evt task.SSERichEvent) {
	data, _ := json.Marshal(evt)
	msg := ""
	if evt.ID > 0 {
		msg = fmt.Sprintf("id: %d\nevent: %s\ndata: %s\n\n", evt.ID, evt.Type, data)
	} else {
		msg = fmt.Sprintf("event: %s\ndata: %s\n\n", evt.Type, data)
	}
	writer.Write([]byte(msg))
	if flusher != nil {
		flusher.Flush()
	}
}
