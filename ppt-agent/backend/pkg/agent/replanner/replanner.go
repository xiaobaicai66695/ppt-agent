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

package replanner

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/agent/agents"
	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/generic"
	"github.com/cloudwego/ppt-agent/pkg/params"
	"github.com/cloudwego/ppt-agent/pkg/tools"
)

const (
	OriginalPlanSessionKey = "OriginalPlan"
)

// replannerSystemPrompt 是系统提示词，拆分为字符串常量避免反引号问题
var replannerSystemPrompt = `你是一个PPT执行进度评估专家，负责判断当前任务状态，并决定下一步操作。

**决策规则（按优先级执行）：**

1. has_high_issue=true：必须生成修复指令，修复所有问题，包括high,medium,low，调用 create_ppt_plan。
   修复指令格式（必须严格遵守）：
   slides = [{"index": <页码>, "title": "<原始标题>", "content_type": "content_slide", "description": "【修复】<一句话描述问题>。大模型给出的具体建议\n}]
   - description 必须以【修复】开头
   - description 中写清楚要修复什么，后面紧跟 python-pptx 代码
   - 不要改动已完成且无问题的页面
   - 调用 create_ppt_plan 后，Executor 会执行修复
   - **每张幻灯片最多修复 2 次，超过后不再生成修复指令，直接进入下一页**

2. has_issues=true（无 high 但有 medium/low 问题）：可以忽略/

3. has_issues=false 且 has_high_issue=false：
   - 已完成幻灯片数 < 总幻灯片数：调用 create_ppt_plan（生成下一页）
   - 已完成幻灯片数 == 总幻灯片数：调用 submit_result

**禁止行为**：
- has_high_issue=true 时禁止调用 submit_result
- has_high_issue=true 时禁止继续生成新页面，必须先修复
- 不要直接输出文本，必须通过工具输出`

var replannerPromptTemplate = prompt.FromMessages(schema.Jinja2,
	schema.SystemMessage(replannerSystemPrompt),
	schema.UserMessage(`## 用户需求
{{ user_query }}

## 进度数字
- 总幻灯片数: {{ total_count }}
- 已完成幻灯片数: {{ executed_count }}

## 剩余幻灯片计划
{{ remaining_plan }}

## QA 检查结果
{{ qa_summary }}

## 判断`))

func NewReplanner(ctx context.Context, operator commandline.Operator) (adk.Agent, error) {
	cm, err := agentutils.NewFallbackToolCallingChatModel(ctx,
		agentutils.WithMaxTokens(4096),
		agentutils.WithTemperature(1.0),
		agentutils.WithTopP(0),
		agentutils.WithDisableThinking(true),
	)
	if err != nil {
		return nil, err
	}

	submitTool := tools.NewToolSubmitResult()

	submitToolInfo, err := submitTool.Info(ctx)
	if err != nil {
		return nil, err
	}

	genInputFn := func(ctx context.Context, in *planexecute.ExecutionContext) ([]adk.Message, error) {
		if _, found := adk.GetSessionValue(ctx, OriginalPlanSessionKey); !found {
			if p, ok := in.Plan.(*generic.Plan); ok {
				adk.AddSessionValue(ctx, OriginalPlanSessionKey, p)
			}
		}
		return replannerInputGen(ctx, in)
	}

	a, err := planexecute.NewReplanner(ctx, &planexecute.ReplannerConfig{
		ChatModel:   cm,
		PlanTool:    generic.PlanToolInfo,
		GenInputFn:  genInputFn,
		RespondTool: submitToolInfo,
		NewPlan: func(ctx context.Context) planexecute.Plan {
			return &generic.Plan{}
		},
	})
	if err != nil {
		return nil, err
	}

	return agents.NewWrite2PlanMDWrapper(a, operator), nil
}

func replannerInputGen(ctx context.Context, in *planexecute.ExecutionContext) ([]adk.Message, error) {
	var plan *generic.Plan
	if p, found := adk.GetSessionValue(ctx, OriginalPlanSessionKey); found {
		plan = p.(*generic.Plan)
	} else if p, ok := in.Plan.(*generic.Plan); ok {
		plan = p
	} else {
		plan = &generic.Plan{}
	}

	workDir, _ := params.GetTypedContextParams[string](ctx, params.WorkDirSessionKey)
	allSlides := plan.GetSlides()
	totalCount := len(allSlides)

	// 展开所有幻灯片为扁平列表，生成 slideKey
	var flatSlides []struct {
		Step     generic.Step
		SlideKey string
	}
	for _, slide := range allSlides {
		if len(slide.SubSteps) > 0 {
			for _, sub := range slide.SubSteps {
				flatSlides = append(flatSlides, struct {
					Step     generic.Step
					SlideKey string
				}{
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
			flatSlides = append(flatSlides, struct {
				Step     generic.Step
				SlideKey string
			}{
				Step:     slide,
				SlideKey: fmt.Sprintf("%d", slide.Index),
			})
		}
	}

	// 从 checkpoint 获取已完成 keys（string 类型）
	completedSet := make(map[string]bool)
	checkpoint, _ := generic.LoadCheckpoint(workDir)
	if checkpoint != nil {
		for _, key := range checkpoint.CompletedSlides {
			completedSet[key] = true
		}
	}

	// 从文件系统补充（解析文件名提取 slideKey）
	for _, path := range generic.GetExistingStepFiles(workDir) {
		stem := filepath.Base(path)
		stem = strings.TrimSuffix(stem, ".pptx")
		parts := strings.Split(stem, "_")
		if len(parts) >= 1 {
			prefix := parts[0]
			if strings.Contains(prefix, ".") {
				completedSet[prefix] = true
			} else {
				completedSet[prefix] = true
			}
		}
	}

	completedCount := len(completedSet)

	// 将 QA 尝试次数已达 2 次的页面标记为完成（不再重试）
	if qaAttempts, err := generic.LoadQAAttempts(workDir); err == nil && qaAttempts != nil {
		for slideKey, count := range qaAttempts.Attempts {
			if count >= 2 {
				completedSet[slideKey] = true
			}
		}
	}

	// 构建剩余幻灯片列表
	var remainingSlides []generic.Step
	for _, sw := range flatSlides {
		if !completedSet[sw.SlideKey] {
			sw.Step.SlideKey = sw.SlideKey
			remainingSlides = append(remainingSlides, sw.Step)
		}
	}

	// 更新扁平列表的 totalCount（包含子页的总数）
	totalCount = len(flatSlides)

	remainingPlan := &generic.Plan{
		Title:  plan.Title,
		Theme:  plan.Theme,
		Slides: remainingSlides,
		Steps:  remainingSlides,
	}
	remainingPlanStr, _ := remainingPlan.MarshalJSON()

	var nextSlideStr string
	if len(remainingSlides) > 0 {
		nextSlideStr = generic.FormatStepForRequest(&remainingSlides[0], workDir)
	} else {
		nextSlideStr = "[完成] 所有幻灯片都已生成完毕。"
	}

	qaResultStr := buildQAResultSummary(ctx, workDir, in)

	return replannerPromptTemplate.Format(ctx, map[string]any{
		"current_time":   agentutils.GetCurrentTime(),
		"user_query":     agentutils.FormatInput(in.UserInput),
		"next_slide":     nextSlideStr,
		"executed_count": completedCount,
		"total_count":    totalCount,
		"remaining_plan": string(remainingPlanStr),
		"qa_summary":     qaResultStr,
	})
}

// buildQAResultSummary 从 QA 结果文件中提取摘要。
func buildQAResultSummary(ctx context.Context, workDir string, in *planexecute.ExecutionContext) string {
	var sb strings.Builder

	qaAttempts, _ := generic.LoadQAAttempts(workDir)

	if qaResult, err := generic.LoadQAResult(workDir); err == nil && qaResult != nil {
		if qaResult.HasHighIssue {
			for _, report := range qaResult.Reports {
				// report 格式: "slideKey|reportContent"
				parts := strings.SplitN(report, "|", 2)
				slideKey := parts[0]
				reportContent := ""
				if len(parts) > 1 {
					reportContent = parts[1]
				}
				attemptCount := 0
				if qaAttempts != nil {
					attemptCount = qaAttempts.Attempts[slideKey]
				}
				attemptInfo := ""
				if attemptCount > 0 {
					attemptInfo = fmt.Sprintf("（已尝试 %d/2 次）", attemptCount)
				}
				if attemptCount >= 2 {
					sb.WriteString(fmt.Sprintf("【严重问题】%s has_high_issue=true%s，但已达到修复上限，不再生成修复指令。\n", slideKey, attemptInfo))
					continue
				}
				sb.WriteString(fmt.Sprintf("【严重问题】has_high_issue=true，必须修复%s。\n", attemptInfo))
				sb.WriteString(fmt.Sprintf("- 页面：%s\n", slideKey))
				sb.WriteString("  审查报告：\n" + indentText(reportContent, 4) + "\n")
			}
		} else if qaResult.HasIssues {
			sb.WriteString("【QA 问题】has_issues=true，存在 medium/low 级别问题：\n")
			for _, report := range qaResult.Reports {
				parts := strings.SplitN(report, "|", 2)
				slideKey := parts[0]
				reportContent := ""
				if len(parts) > 1 {
					reportContent = parts[1]
				}
				sb.WriteString(fmt.Sprintf("- 页面：%s\n", slideKey))
				sb.WriteString("  " + indentText(reportContent, 2) + "\n")
			}
		} else if qaResult.TotalSlides > 0 {
			sb.WriteString("【QA 结果】所有已审查页面检查通过，无严重问题。\n")
		}
	}

	if len(in.ExecutedSteps) > 0 {
		lastStep := in.ExecutedSteps[len(in.ExecutedSteps)-1]
		if strings.Contains(lastStep.Result, "has_high_issue=true") {
			sb.WriteString("【Executor 报告】has_high_issue=true。\n")
		} else if strings.Contains(lastStep.Result, "has_issues=true") {
			sb.WriteString("【Executor 报告】has_issues=true。\n")
		}
	}

	if sb.Len() == 0 {
		return "【QA 结果】暂无 QA 数据（尚未完成任何页面的视觉检查）。"
	}

	// 附加：明确告知已达到 2 次 QA 上限的页面，供 Replanner 感知
	if qaAttempts != nil {
		var maxedSlides []string
		for slideIdx, count := range qaAttempts.Attempts {
			if count >= 2 {
				maxedSlides = append(maxedSlides, slideIdx)
			}
		}
		if len(maxedSlides) > 0 {
			sb.WriteString(fmt.Sprintf("【QA 上限】以下页面已完成 2 次 QA 审查，标记为已处理：第 %v\n", maxedSlides))
		}
	}

	return sb.String()
}

// indentText 将文本的每一行前面加上指定数量的空格缩进。
func indentText(text string, spaces int) string {
	prefix := strings.Repeat(" ", spaces)
	lines := strings.Split(text, "\n")
	var sb strings.Builder
	for i, line := range lines {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(prefix)
		sb.WriteString(line)
	}
	return sb.String()
}
