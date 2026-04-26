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

package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"

	"github.com/cloudwego/ppt-agent/pkg/generic"
)

func FormatInput(input []adk.Message) string {
	var sb strings.Builder
	for _, msg := range input {
		sb.WriteString(msg.Content)
		sb.WriteString("\n")
	}
	return sb.String()
}

func ToJSONString(v interface{}) string {
	str, _ := json.Marshal(v)
	return string(str)
}

func PtrOf[T any](v T) *T {
	return &v
}

func GetSessionValue[T any](ctx context.Context, key string) (T, bool) {
	v, ok := adk.GetSessionValue(ctx, key)
	if !ok {
		var zero T
		return zero, false
	}
	t, ok := v.(T)
	if !ok {
		var zero T
		return zero, false
	}
	return t, true
}

func FormatExecutedSteps(in []planexecute.ExecutedStep) string {
	var sb strings.Builder
	for idx, m := range in {
		_, _ = fmt.Fprintf(&sb, "## %d. Step: %v\n  Result: %v\n\n", idx+1, m.Step, m.Result)
	}
	return sb.String()
}

func FormatExecutedStepsStr(executedSteps string) string {
	if executedSteps == "" {
		return "暂无"
	}
	return executedSteps
}

// ExecutorContext 用于构建 Executor 的精简上下文
type ExecutorContext struct {
	CompletedSlides  []int           // 已完成的页码列表
	TotalCount       int             // 总页数
	NextSlide       *generic.Step   // 下一个待执行的幻灯片
	LastStepSuccess  bool            // 上一轮是否成功
	LastStepError   string          // 上一轮的错误信息（如果有）
	RemainingTitles []string        // 剩余幻灯片的标题列表（简洁）
}

// BuildExecutorContext 构建精简的 Executor 上下文
// 替代原有的完整历史记录，只传递简洁的完成状态和上一轮结果
func BuildExecutorContext(ctx context.Context, plan *generic.Plan, workDir string, executedSteps []planexecute.ExecutedStep) *ExecutorContext {
	ec := &ExecutorContext{
		LastStepSuccess: true,
	}

	// 获取总幻灯片数
	allSlides := plan.GetSlides()
	ec.TotalCount = len(allSlides)

	// 获取已完成幻灯片（优先级：checkpoint > filesystem > framework）
	completedSet := make(map[int]bool)

	// 1. 从 checkpoint 读取（最可靠）
	if checkpoint, err := generic.LoadCheckpoint(workDir); err == nil && checkpoint != nil {
		for _, idx := range checkpoint.CompletedSlides {
			completedSet[idx] = true
		}
	}

	// 2. 从 filesystem 补充
	for idx := range generic.GetExistingStepFiles(workDir) {
		completedSet[idx] = true
	}

	// 3. 从 framework executedSteps 补充（兜底）
	for _, es := range executedSteps {
		var step generic.Step
		if err := json.Unmarshal([]byte(es.Step), &step); err == nil {
			completedSet[step.Index] = true
		}
	}

	// 转换为有序列表
	for idx := range completedSet {
		ec.CompletedSlides = append(ec.CompletedSlides, idx)
	}

	// 构建剩余幻灯片列表和标题摘要
	var remainingSlides []generic.Step
	for _, slide := range allSlides {
		if !completedSet[slide.Index] {
			remainingSlides = append(remainingSlides, slide)
			ec.RemainingTitles = append(ec.RemainingTitles, fmt.Sprintf("%d. %s", slide.Index, slide.Title))
		}
	}

	// 设置下一个待执行的幻灯片
	if len(remainingSlides) > 0 {
		ec.NextSlide = &remainingSlides[0]
	}

	// 分析上一轮执行结果（通过 Result 是否为空来判断成功/失败）
	if len(executedSteps) > 0 {
		lastStep := executedSteps[len(executedSteps)-1]
		// Result 为空表示失败，有内容表示成功
		ec.LastStepSuccess = lastStep.Result != ""
		if !ec.LastStepSuccess {
			ec.LastStepError = "执行结果为空"
		}
	}

	return ec
}

// FormatExecutorContext 将 ExecutorContext 格式化为字符串
// 生成精简的上下文信息，用于传递给 Executor 的 prompt
func FormatExecutorContext(ec *ExecutorContext) string {
	var sb strings.Builder

	// 1. 简洁的完成状态
	if len(ec.CompletedSlides) == 0 {
		sb.WriteString("已完成幻灯片：无（共0页）\n")
	} else {
		// 排序
		for i := 0; i < len(ec.CompletedSlides)-1; i++ {
			for j := i + 1; j < len(ec.CompletedSlides); j++ {
				if ec.CompletedSlides[i] > ec.CompletedSlides[j] {
					ec.CompletedSlides[i], ec.CompletedSlides[j] = ec.CompletedSlides[j], ec.CompletedSlides[i]
				}
			}
		}
		completedStr := fmt.Sprintf("%v", ec.CompletedSlides)
		sb.WriteString(fmt.Sprintf("已完成幻灯片：%s（共%d页）\n", completedStr, len(ec.CompletedSlides)))
	}

	// 2. 上一轮执行状态
	if len(ec.CompletedSlides) > 0 {
		if ec.LastStepSuccess {
			sb.WriteString("上一轮：执行成功\n")
		} else {
			sb.WriteString(fmt.Sprintf("上一轮：执行失败 - %s\n", ec.LastStepError))
		}
	}

	// 3. 当前任务
	sb.WriteString("\n当前任务：")
	if ec.NextSlide != nil {
		sb.WriteString(fmt.Sprintf("生成第%d页 - %s\n", ec.NextSlide.Index, ec.NextSlide.Title))
	} else {
		sb.WriteString("所有幻灯片都已生成完毕\n")
	}

	// 4. 剩余幻灯片标题摘要（只需要标题，不需要完整描述）
	if len(ec.RemainingTitles) > 0 {
		sb.WriteString("\n待生成（共" + fmt.Sprintf("%d", len(ec.RemainingTitles)) + "页）：\n")
		for _, title := range ec.RemainingTitles {
			sb.WriteString("- " + title + "\n")
		}
	}

	return sb.String()
}

// FormatRemainingPlanSummary 将剩余计划格式化为简洁的摘要
// 只包含标题列表，不包含完整的描述内容
func FormatRemainingPlanSummary(plan *generic.Plan, completedIndexes map[int]bool) string {
	allSlides := plan.GetSlides()
	var sb strings.Builder

	for _, slide := range allSlides {
		if !completedIndexes[slide.Index] {
			sb.WriteString(fmt.Sprintf("- [ ] 第%d页: %s\n", slide.Index, slide.Title))
		} else {
			sb.WriteString(fmt.Sprintf("- [x] 第%d页: %s\n", slide.Index, slide.Title))
		}
	}

	return sb.String()
}
