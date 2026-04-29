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
	"path/filepath"
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
	CompletedSlides  []string      // 已完成的页码标识列表（如 "4"、"4.1"）
	TotalCount      int           // 总页数
	NextSlide       *generic.Step // 下一个待执行的幻灯片（单页或分页组）
	NextSlideKey    string        // 当前任务的 slideKey（用于 update_progress 和 QA）
	IsBatchMode     bool          // 是否批量模式（Planner 指定了 SubSteps）
	LastStepSuccess bool          // 上一轮是否成功
	LastStepError   string        // 上一轮的错误信息（如果有）
	RemainingTitles []string      // 剩余幻灯片的标题列表（简洁）
}

// BuildExecutorContext 构建精简的 Executor 上下文
// 替代原有的完整历史记录，只传递简洁的完成状态和上一轮结果
func BuildExecutorContext(ctx context.Context, plan *generic.Plan, workDir string, executedSteps []planexecute.ExecutedStep) *ExecutorContext {
	ec := &ExecutorContext{
		LastStepSuccess: true,
	}

	// 1. 展开所有幻灯片为扁平列表，同时生成唯一的 slideKey（用于完成状态追踪）
	// slideKey 格式：普通页 "N"，分页子页 "N.M"
	allSlides := plan.GetSlides()
	var flatSlides []SlideWithKey // SlideWithKey 包含 Step 和它对应的 slideKey
	for _, slide := range allSlides {
		if len(slide.SubSteps) > 0 {
			for _, sub := range slide.SubSteps {
				flatSlides = append(flatSlides, SlideWithKey{
					Step: generic.Step{
						Index:         slide.Index,
						Title:         sub.Title,
						ContentType:   sub.ContentType,
						Description:   sub.Description,
						ContentPlan:   sub.ContentPlan,
						LayoutHint:    sub.LayoutHint,
					},
					SlideKey: fmt.Sprintf("%d.%d", slide.Index, sub.Index),
				})
			}
		} else {
			flatSlides = append(flatSlides, SlideWithKey{
				Step:     slide,
				SlideKey: fmt.Sprintf("%d", slide.Index),
			})
		}
	}
	ec.TotalCount = len(flatSlides)

	// 2. 获取已完成幻灯片的 slideKey（优先级：checkpoint > filesystem）
	completedSet := make(map[string]bool)

	// checkpoint 中存储的是 slideKey 列表
	if checkpoint, err := generic.LoadCheckpoint(workDir); err == nil && checkpoint != nil {
		for _, key := range checkpoint.CompletedSlides {
			completedSet[key] = true
		}
	}

	// filesystem 中的文件名：普通页 4_标题页.pptx，子页 4.1_金融.pptx
	for _, path := range generic.GetExistingStepFiles(workDir) {
		// 从文件名提取 slideKey: "4_标题页" -> "4"，"4.1_金融" -> "4.1"
		stem := filepath.Base(path)
		stem = strings.TrimSuffix(stem, ".pptx")
		// 检查是否为子页（包含 ".N" 模式）
		parts := strings.Split(stem, "_")
		if len(parts) >= 1 {
			prefix := parts[0]
			// 如果前缀包含 "."，说明是子页，保留完整前缀
			if strings.Contains(prefix, ".") {
				completedSet[prefix] = true
			} else {
				// 普通页，取前缀作为 key
				completedSet[prefix] = true
			}
		}
	}

	// framework executedSteps 兜底（从 Step JSON 中解析 slideKey）
	for _, es := range executedSteps {
		var step generic.Step
		if err := json.Unmarshal([]byte(es.Step), &step); err == nil {
			// 只能通过 Index 判断，如果 Index 已存在则视为完成
			key := fmt.Sprintf("%d", step.Index)
			completedSet[key] = true
		}
	}

	// 3. 转换为有序列表
	for key := range completedSet {
		ec.CompletedSlides = append(ec.CompletedSlides, key)
	}

	// 4. 构建剩余幻灯片列表（基于 slideKey）
	var remainingSlides []SlideWithKey
	for _, sw := range flatSlides {
		if !completedSet[sw.SlideKey] {
			remainingSlides = append(remainingSlides, sw)
			ec.RemainingTitles = append(ec.RemainingTitles,
				fmt.Sprintf("第%s页: %s", sw.SlideKey, sw.Title))
		}
	}

	// 5. 设置下一个待执行的幻灯片
	if len(remainingSlides) > 0 {
		first := &remainingSlides[0].Step
		first.SlideKey = remainingSlides[0].SlideKey
		ec.NextSlideKey = remainingSlides[0].SlideKey
		if len(first.SubSteps) > 0 {
			ec.IsBatchMode = true
		}
		ec.NextSlide = first
	}

	// 6. 分析上一轮执行结果
	if len(executedSteps) > 0 {
		lastStep := executedSteps[len(executedSteps)-1]
		ec.LastStepSuccess = lastStep.Result != ""
		if !ec.LastStepSuccess {
			ec.LastStepError = "执行结果为空"
		}
	}

	return ec
}

// SlideWithKey 包含展开后的幻灯片及其 slideKey
type SlideWithKey struct {
	generic.Step
	SlideKey string
}

// FormatExecutorContext 将 ExecutorContext 格式化为字符串
// 生成精简的上下文信息，用于传递给 Executor 的 prompt
func FormatExecutorContext(ec *ExecutorContext) string {
	var sb strings.Builder

	// 1. 简洁的完成状态
	if len(ec.CompletedSlides) == 0 {
		sb.WriteString("已完成幻灯片：无（共0页）\n")
	} else {
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
		var contentTypeHint string
		if ec.NextSlide.ContentType != "" {
			contentTypeHint = ec.NextSlide.ContentType
		} else {
			contentTypeHint = "待 Executor 决定"
		}
		if ec.IsBatchMode {
			sb.WriteString(fmt.Sprintf("批量模式（Planner 指定分页组，共 %d 页）：\n", ec.TotalCount))
			sb.WriteString(fmt.Sprintf("  - 第%s页: %s (%s)\n",
				ec.NextSlideKey, ec.NextSlide.Title, contentTypeHint))
		} else {
			sb.WriteString(fmt.Sprintf("生成第%s页 - %s (%s)\n",
				ec.NextSlideKey, ec.NextSlide.Title, contentTypeHint))
		}
	} else {
		sb.WriteString(fmt.Sprintf("所有幻灯片都已生成完毕（共 %d 页）\n", ec.TotalCount))
	}

	// 4. 剩余幻灯片标题摘要
	if len(ec.RemainingTitles) > 0 {
		sb.WriteString(fmt.Sprintf("\n待生成（共%d页）：\n", len(ec.RemainingTitles)))
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
