/*
 * 人机交互管理器
 * 处理工具调用的审批流程
 */

package human

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cloudwego/eino/adk"

	"github.com/cloudwego/ppt-agent/pkg/human/prints"
	"github.com/cloudwego/ppt-agent/pkg/tools"
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

			lastEvent = event

			// 优先处理 interrupt（即使 event.Err != nil）
			if event.Action != nil && event.Action.Interrupted != nil {
				interrupted = true
				break
			}

			// fatal error：没有 interrupt 才视为真正的错误
			if event.Err != nil {
				prints.Event(event)
				return event, event.Err
			}

			// 使用 prints 包打印事件
			prints.Event(event)

			// 检查退出
			if event.Action != nil && event.Action.Exit {
				prints.Summary(event.AgentName, eventCount, true)
				return lastEvent, nil
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

		// 处理所有中断（可能同时有多个）
		targets := make(map[string]any)
		interruptContexts := lastEvent.Action.Interrupted.InterruptContexts

		for _, ic := range interruptContexts {
			if info, ok := ic.Info.(*tools.SearchApprovalInfo); ok {
				if m.interactive {
					// 逐个询问用户，不自动跳过任何搜索
					info.Result = m.promptSearchApprovalLoop(info)
				} else {
					fmt.Println("[人机交互] 非交互模式，默认跳过搜索")
					info.Result = &tools.SearchApprovalResult{Option: 1}
				}
				targets[ic.ID] = info
			} else {
				// 其他类型的中断：打印信息并默认自动继续
				fmt.Println()
				prints.PrintSeparator()
				fmt.Printf("中断类型: %T\n", ic.Info)
				prints.PrintSeparator()
				fmt.Println("自动批准，继续执行...")
				fmt.Println()
			}
		}

		// 恢复执行
		var err error
		if len(targets) > 0 {
			iter, err = runner.ResumeWithParams(ctx, checkpointID, &adk.ResumeParams{
				Targets: targets,
			})
		} else {
			iter, err = runner.Resume(ctx, checkpointID)
		}
		if err != nil {
			return lastEvent, err
		}
	}
}

// promptSearchApprovalLoop 循环询问用户，直到输入有效选项
func (m *Manager) promptSearchApprovalLoop(info *tools.SearchApprovalInfo) *tools.SearchApprovalResult {
	for {
		fmt.Println()
		fmt.Println(strings.Repeat("=", 60))
		fmt.Printf("人机交互：网络搜索审批\n")
		fmt.Printf("工具: %s\n", info.ToolName)
		fmt.Printf("关键词: %s\n", info.Query)
		if info.Reason != "" {
			fmt.Printf("搜索原因: %s\n", info.Reason)
		}
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("\n请选择操作：")
		fmt.Println("  1 - 跳过此次搜索（不调用工具）")
		fmt.Println("  2 - 确认该关键词进行搜索")
		fmt.Println("  3 - 编辑关键词后再搜索")
		fmt.Println()
		fmt.Print("请输入选项：")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		userInput := strings.TrimSpace(scanner.Text())
		fmt.Println()

		// 单独处理选项 3：先确认选项，再询问新关键词
		lower := strings.ToLower(userInput)
		if lower == "3" {
			editedQuery := m.promptNewQuery(info)
			return &tools.SearchApprovalResult{Option: 3, EditedQuery: &editedQuery}
		}

		// 其余选项直接解析
		approvalResult, err := tools.ParseSearchApprovalResult(userInput)
		if err != nil {
			fmt.Printf("输入解析错误: %v\n", err)
			fmt.Println("请重新输入。")
			continue
		}
		return approvalResult
	}
}

// promptNewQuery 提示用户输入新的搜索关键词
func (m *Manager) promptNewQuery(info *tools.SearchApprovalInfo) string {
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("当前关键词: %s\n", info.Query)
	fmt.Println("请输入编辑后的新关键词：")
	fmt.Println()
	fmt.Print("新关键词：")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	newQuery := strings.TrimSpace(scanner.Text())
	fmt.Println()

	for newQuery == "" {
		fmt.Println("关键词不能为空，请重新输入。")
		fmt.Print("新关键词：")
		scanner.Scan()
		newQuery = strings.TrimSpace(scanner.Text())
		fmt.Println()
	}
	return newQuery
}
