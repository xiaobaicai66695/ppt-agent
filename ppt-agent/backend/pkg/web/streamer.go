package web

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/cloudwego/ppt-agent/pkg/session"
	"github.com/cloudwego/ppt-agent/pkg/task"
)

type flushWriter interface {
	Flush()
}

// answerBuffer 缓冲单个 answer 片段，flush 时一次性写入数据库。
type answerBuffer struct {
	mu      sync.Mutex
	content strings.Builder
}

func (b *answerBuffer) Append(chunk string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.content.WriteString(chunk)
}

func (b *answerBuffer) Flush() (content string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.content.Len() == 0 {
		return ""
	}
	content = b.content.String()
	b.content.Reset()
	return content
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

	afterEventID, _ := strconv.ParseUint(c.GetHeader("Last-Event-ID"), 10, 64)
	listenerID := uuid.New().String()
	ch := make(chan task.SSERichEvent, 128)
	events, done := ts.SubscribeFrom(listenerID, ch, afterEventID)
	if !done {
		defer ts.RemoveListener(listenerID)
	}

	// 分离 answer 和 system_step 的缓冲池，各自独立 flush。
	answerBuf := &answerBuffer{}
	systemBuf := &answerBuffer{}

	for _, evt := range events {
		writeSSEToWriter(writer, flusher, evt)
		handleEventBuffers(evt, answerBuf, systemBuf, s.sessionManager, id, ts.Info.WorkDir)
	}

	if done {
		flushBuffer(answerBuf, s.sessionManager, id, ts.Info.WorkDir)
		flushBuffer(systemBuf, s.sessionManager, id, ts.Info.WorkDir)
		if len(events) == 0 || events[len(events)-1].Type != "complete" {
			completeEvt := task.SSERichEvent{
				Type:   "complete",
				Status: ts.Info.Status,
				Done:   ts.Info.DoneCount,
				Total:  ts.Info.TotalCount,
				Files:  ts.Info.Files,
			}
			writeSSEToWriter(writer, flusher, completeEvt)
			flushCompleteToDB(s.sessionManager, id, ts.Info.WorkDir, completeEvt)
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
			flushBuffer(answerBuf, s.sessionManager, id, ts.Info.WorkDir)
			flushBuffer(systemBuf, s.sessionManager, id, ts.Info.WorkDir)
			return false
		case <-heartbeat.C:
			w.Write([]byte(": hb\n\n"))
			if flusher != nil {
				flusher.Flush()
			}
			return true
		case evt, ok := <-ch:
			if !ok {
				flushBuffer(answerBuf, s.sessionManager, id, ts.Info.WorkDir)
				flushBuffer(systemBuf, s.sessionManager, id, ts.Info.WorkDir)
				return false
			}
			writeSSEToWriter(w, flusher, evt)
			handleEventBuffers(evt, answerBuf, systemBuf, s.sessionManager, id, ts.Info.WorkDir)
			if evt.Type == "complete" {
				flushCompleteToDB(s.sessionManager, id, ts.Info.WorkDir, evt)
				c.Writer.Write([]byte(": close\n\n"))
				c.Writer.Flush()
				c.Abort()
				return false
			}
			return true
		}
	})
}

// handleEventBuffers 根据事件类型将内容追加到对应缓冲池，并触发 flush。
//
// answer 片段 → answerBuf，遇到 answer_end 时 flush
// system_step 片段 → systemBuf，遇到 system_step_end 时 flush
// progress / complete / error → 直接 flush 两个缓冲，再处理该事件
func handleEventBuffers(evt task.SSERichEvent, answerBuf, systemBuf *answerBuffer, sm *session.SessionManager, taskID, workDir string) {
	switch evt.Type {
	case "answer":
		answerBuf.Append(evt.Content)
	case "answer_end":
		flushBuffer(answerBuf, sm, taskID, workDir)
	case "system_step":
		systemBuf.Append(evt.Content)
	case "system_step_end":
		flushBuffer(systemBuf, sm, taskID, workDir)
	case "complete", "progress", "error":
		flushBuffer(answerBuf, sm, taskID, workDir)
		flushBuffer(systemBuf, sm, taskID, workDir)
		if evt.Type == "progress" {
			flushProgressToDB(sm, taskID, evt)
		}
	default:
		// tool_call / file_ready 等：只 flush answer
		flushBuffer(answerBuf, sm, taskID, workDir)
	}
}

// flushBuffer 将缓冲池内容一次性写入数据库。
func flushBuffer(buf *answerBuffer, sm *session.SessionManager, taskID, workDir string) {
	if sm == nil || buf == nil {
		return
	}
	content := buf.Flush()
	if content == "" || strings.TrimSpace(content) == "" {
		return
	}
	go func() {
		_ = sm.GetOrCreate(taskID, workDir).AddAssistantMessage(content)
	}()
}

// flushCompleteToDB 将任务完成摘要写入数据库。
func flushCompleteToDB(sm *session.SessionManager, taskID, workDir string, evt task.SSERichEvent) {
	if sm == nil || evt.Message == "" {
		return
	}
	go func() {
		_ = sm.GetOrCreate(taskID, workDir).AddAssistantMessage("【任务完成】" + evt.Message)
	}()
}

// flushProgressToDB 将阶段进度信息写入数据库。
func flushProgressToDB(sm *session.SessionManager, taskID string, evt task.SSERichEvent) {
	if sm == nil || evt.Phase == "" {
		return
	}
	detail := ""
	if evt.PhaseDetail != "" {
		detail = " | " + evt.PhaseDetail
	}
	msg := "【进度】" + phaseLabel(evt.Phase) + detail
	go func() {
		_ = sm.GetOrCreate(taskID, "").AddAssistantMessage(msg)
	}()
}

// phaseLabel 返回阶段的中文标签。
func phaseLabel(phase string) string {
	switch phase {
	case "preparing":
		return "准备阶段：读取模板"
	case "planning":
		return "规划阶段：创建任务清单"
	case "generating":
		return "生成阶段：制作幻灯片"
	case "qa":
		return "质检阶段：审查质量"
	case "fixing":
		return "修复阶段：优化问题"
	case "complete":
		return "任务完成"
	default:
		return "执行中"
	}
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
