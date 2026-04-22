/*
 * 人机交互管理器
 * 处理工具调用的审批流程
 */

package human

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"

	"github.com/cloudwego/ppt-agent/pkg/human/prints"
)

// Manager 人机交互管理器
type Manager struct {
	interactive bool // 是否启用交互模式
}

// NewManager 创建人机交互管理器
func NewManager(interactive bool) *Manager {
	return &Manager{
		interactive: interactive,
	}
}

// RunWithApproval 运行人机交互循环
// 返回最终消息和错误
func (m *Manager) RunWithApproval(ctx context.Context, runner *adk.Runner, checkpointID string, iter *adk.AsyncIterator[*adk.AgentEvent]) (*adk.AgentEvent, error) {
	eventCount := 0

	for {
		var lastEvent *adk.AgentEvent
		interrupted := false

		// 处理事件流
		for {
			event, ok := iter.Next()
			if !ok {
				prints.Summary("PPT-Agent", eventCount, true)
				return lastEvent, nil
			}

			eventCount++

			if event.Err != nil {
				prints.Event(event)
				return event, event.Err
			}

			lastEvent = event

			// 使用 prints 包打印事件
			prints.Event(event)

			// 检查中断
			if event.Action != nil {
				if event.Action.Interrupted != nil {
					interrupted = true
					break
				}
				if event.Action.Exit {
					prints.Summary(event.AgentName, eventCount, true)
					return lastEvent, nil
				}
			}

			// 防止无限循环
			if eventCount > 100 {
				fmt.Println("[警告] 事件数量超过 100，强制结束")
				prints.Summary("PPT-Agent", eventCount, false)
				return lastEvent, nil
			}
		}

		if !interrupted || lastEvent == nil {
			prints.Summary("PPT-Agent", eventCount, true)
			return lastEvent, nil
		}

		// 处理中断 - 默认自动批准
		interruptCtx := lastEvent.Action.Interrupted.InterruptContexts[0]
		fmt.Println()
		prints.PrintSeparator()
		fmt.Printf("中断类型: %T\n", interruptCtx.Info)
		prints.PrintSeparator()
		fmt.Println("自动批准，继续执行...")
		fmt.Println()

		// 恢复执行 - 自动批准
		var err error
		iter, err = runner.Resume(ctx, checkpointID)
		if err != nil {
			return lastEvent, err
		}
	}
}
