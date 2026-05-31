package web

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/cloudwego/ppt-agent/pkg/session"
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
		// 实时写助手消息到数据库（answer 类型）。
		persistSSEEventToDB(s.sessionManager, id, ts.Info.WorkDir, evt)
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
			persistSSEEventToDB(s.sessionManager, id, ts.Info.WorkDir, completeEvt)
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
			// 实时写助手消息到数据库。
			persistSSEEventToDB(s.sessionManager, id, ts.Info.WorkDir, evt)
			if evt.Type == "complete" {
				return false
			}
			return true
		}
	})
}

// persistSSEEventToDB 将 SSE 事件中的有意义内容写入 conversation_messages 表。
// 目前只处理 answer 类型（助手回复）和 complete 类型的 message（完成摘要）。
// 为避免阻塞 SSE 流，写操作在 goroutine 中异步执行。
func persistSSEEventToDB(sm *session.SessionManager, taskID, workDir string, evt task.SSERichEvent) {
	switch evt.Type {
	case "answer":
		content := strings.TrimSpace(evt.Content)
		if content != "" && content != "..." && content != "……" {
			go func() {
				if sm != nil {
					_ = sm.GetOrCreate(taskID, workDir).AddAssistantMessage(content)
				}
			}()
		}
	case "complete":
		// 将完成摘要也写入数据库（用户后续对话可能需要）。
		if evt.Message != "" {
			go func() {
				if sm != nil {
					_ = sm.GetOrCreate(taskID, workDir).AddAssistantMessage("【任务完成】" + evt.Message)
				}
			}()
		}
	}
}

func writeSSEToWriter(writer io.Writer, flusher flushWriter, evt task.SSERichEvent) {
	data, _ := json.Marshal(evt)
	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", evt.Type, data)
	writer.Write([]byte(msg))
	if flusher != nil {
		flusher.Flush()
	}
}
