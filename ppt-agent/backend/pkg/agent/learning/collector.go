/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package learning

import (
	"context"
	"sync"
	"time"

	"github.com/cloudwego/ppt-agent/pkg/logger"
)

// SignalType 学习信号类型
type SignalType int

const (
	SignalExplicitFeedback SignalType = iota // 显式反馈（用户评分）
	SignalImplicitFeedback                 // 隐式反馈（用户行为）
	SignalEditAction                       // 编辑行为
	SignalAbandonTask                      // 放弃任务
	SignalCompletion                       // 任务完成
	SignalQAResult                        // QA结果
	SignalTimeSpent                       // 时间消耗
)

// LearningSignal 学习信号
type LearningSignal struct {
	Type      SignalType              // 信号类型
	UserID    int                     // 用户ID
	TaskID    string                  // 任务ID
	Data      map[string]interface{}   // 信号数据
	Timestamp time.Time               // 时间戳
	Context   *SignalContext          // 上下文信息
}

// SignalContext 信号上下文
type SignalContext struct {
	TaskPhase     string  // 任务阶段: plan/generate/qa/fix/complete
	PageIndex     int     // 页面索引（如果是页面级别）
	ActionType    string  // 操作类型
	Duration      float64 // 持续时间（秒）
	QualityScore  float64 // 质量评分
	InteractionCount int // 交互次数
}

// Collector 反馈采集器
type Collector struct {
	signals chan *LearningSignal
	buffer  []*LearningSignal
	mu      sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	// 配置
	bufferSize   int
	flushInterval time.Duration
}

// CollectorConfig 采集器配置
type CollectorConfig struct {
	BufferSize    int           // 缓冲区大小
	FlushInterval time.Duration // 刷新间隔
}

// NewCollector 创建反馈采集器
func NewCollector(cfg *CollectorConfig) *Collector {
	if cfg == nil {
		cfg = &CollectorConfig{
			BufferSize:    100,
			FlushInterval: 30 * time.Second,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := &Collector{
		signals:       make(chan *LearningSignal, cfg.BufferSize),
		buffer:        make([]*LearningSignal, 0, cfg.BufferSize),
		ctx:           ctx,
		cancel:        cancel,
		bufferSize:    cfg.BufferSize,
		flushInterval: cfg.FlushInterval,
	}

	// 启动后台处理
	c.wg.Add(1)
	go c.processLoop()

	return c
}

// processLoop 后台信号处理循环
func (c *Collector) processLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			// 关闭前刷新缓冲区
			c.flush()
			return
		case signal := <-c.signals:
			c.mu.Lock()
			c.buffer = append(c.buffer, signal)
			shouldFlush := len(c.buffer) >= c.bufferSize
			c.mu.Unlock()

			if shouldFlush {
				c.flush()
			}
		case <-ticker.C:
			c.flush()
		}
	}
}

// flush 刷新缓冲区
func (c *Collector) flush() {
	c.mu.Lock()
	if len(c.buffer) == 0 {
		c.mu.Unlock()
		return
	}

	signals := c.buffer
	c.buffer = make([]*LearningSignal, 0, c.bufferSize)
	c.mu.Unlock()

	// 处理信号
	for _, signal := range signals {
		c.processSignal(signal)
	}
}

// processSignal 处理单个信号
func (c *Collector) processSignal(signal *LearningSignal) {
	// 记录日志
	logger.Debug("learning_signal",
		"type", signal.Type.String(),
		"user_id", signal.UserID,
		"task_id", signal.TaskID,
		"phase", signal.Context.TaskPhase)

	// TODO: 发送到外部学习系统或存储
}

// Record 记录学习信号
func (c *Collector) Record(signal *LearningSignal) {
	if signal.Timestamp.IsZero() {
		signal.Timestamp = time.Now()
	}

	select {
	case c.signals <- signal:
	default:
		logger.Warn("learning_signal_buffer_full", "dropping_signal", true)
	}
}

// RecordExplicitFeedback 记录显式反馈
func (c *Collector) RecordExplicitFeedback(userID int, taskID string, score float64, data map[string]interface{}) {
	signal := &LearningSignal{
		Type:   SignalExplicitFeedback,
		UserID: userID,
		TaskID: taskID,
		Data:   data,
		Context: &SignalContext{
			QualityScore: score,
		},
	}
	c.Record(signal)
}

// RecordImplicitFeedback 记录隐式反馈
func (c *Collector) RecordImplicitFeedback(userID int, taskID string, action string, duration float64) {
	signal := &LearningSignal{
		Type:   SignalImplicitFeedback,
		UserID: userID,
		TaskID: taskID,
		Data: map[string]interface{}{
			"action":  action,
			"implicit": true,
		},
		Context: &SignalContext{
			ActionType: action,
			Duration:   duration,
		},
	}
	c.Record(signal)
}

// RecordEditAction 记录编辑行为
func (c *Collector) RecordEditAction(userID int, taskID string, pageIndex int, before, after string) {
	signal := &LearningSignal{
		Type:   SignalEditAction,
		UserID: userID,
		TaskID: taskID,
		Data: map[string]interface{}{
			"page_index": pageIndex,
			"edit_type":  "content_modification",
		},
		Context: &SignalContext{
			PageIndex: pageIndex,
			ActionType: "edit",
		},
	}
	c.Record(signal)
}

// RecordTaskCompletion 记录任务完成
func (c *Collector) RecordTaskCompletion(userID int, taskID string, duration time.Duration, qualityScore float64) {
	signal := &LearningSignal{
		Type:   SignalCompletion,
		UserID: userID,
		TaskID: taskID,
		Data: map[string]interface{}{
			"duration": duration.Seconds(),
		},
		Context: &SignalContext{
			TaskPhase:    "complete",
			Duration:    duration.Seconds(),
			QualityScore: qualityScore,
		},
	}
	c.Record(signal)
}

// RecordAbandonTask 记录放弃任务
func (c *Collector) RecordAbandonTask(userID int, taskID string, reason string, progress float64) {
	signal := &LearningSignal{
		Type:   SignalAbandonTask,
		UserID: userID,
		TaskID: taskID,
		Data: map[string]interface{}{
			"reason":   reason,
			"progress": progress,
		},
	}
	c.Record(signal)
}

// RecordQAResult 记录QA结果
func (c *Collector) RecordQAResult(userID int, taskID string, pageIndex int, hasIssue bool, severity string) {
	signal := &LearningSignal{
		Type:   SignalQAResult,
		UserID: userID,
		TaskID: taskID,
		Data: map[string]interface{}{
			"page_index": pageIndex,
			"has_issue":  hasIssue,
			"severity":   severity,
		},
		Context: &SignalContext{
			PageIndex: pageIndex,
			TaskPhase: "qa",
		},
	}
	c.Record(signal)
}

// Close 关闭采集器
func (c *Collector) Close() {
	c.cancel()
	c.wg.Wait()
}

// String 实现 Stringer 接口
func (t SignalType) String() string {
	switch t {
	case SignalExplicitFeedback:
		return "explicit_feedback"
	case SignalImplicitFeedback:
		return "implicit_feedback"
	case SignalEditAction:
		return "edit_action"
	case SignalAbandonTask:
		return "abandon_task"
	case SignalCompletion:
		return "completion"
	case SignalQAResult:
		return "qa_result"
	case SignalTimeSpent:
		return "time_spent"
	default:
		return "unknown"
	}
}
