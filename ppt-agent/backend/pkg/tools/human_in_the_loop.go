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

package tools

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

func init() {
	gob.Register(&SearchApprovalInfo{})
	gob.Register(&SearchApprovalResult{})
	schema.Register[*SearchApprovalInfo]()
	schema.Register[*SearchApprovalResult]()
}

type SearchApprovalInfo struct {
	ToolName string                `json:"tool_name"`
	Query    string                `json:"query"`
	Reason   string                `json:"reason,omitempty"`
	Result   *SearchApprovalResult `json:"result,omitempty"`
}

type SearchApprovalResult struct {
	Option      int     `json:"option"`
	EditedQuery *string `json:"edited_query,omitempty"`
}

func (s *SearchApprovalInfo) String() string {
	reasonStr := ""
	if s.Reason != "" {
		reasonStr = fmt.Sprintf(" | 原因: %s", s.Reason)
	}
	return fmt.Sprintf("工具 '%s' 即将搜索: %s%s", s.ToolName, s.Query, reasonStr)
}

type InvokableSearchApprovalTool struct {
	tool.InvokableTool
}

// extractSearchParams 从可能带噪声的输入中提取搜索参数。
// 支持三种策略：直接解析、JSON块提取、正则字段提取。
func extractSearchParams(rawInput string) (query string, reason string, _ error) {
	// 1. 尝试直接解析
	var searchInput map[string]any
	if err := json.Unmarshal([]byte(rawInput), &searchInput); err == nil {
		query, _ = searchInput["query"].(string)
		reason, _ = searchInput["reason"].(string)
		return query, reason, nil
	}

	// 2. 提取 JSON 块（第一个 { 到最后一个 }）
	start := strings.Index(rawInput, "{")
	end := strings.LastIndex(rawInput, "}")
	if start != -1 && end != -1 && end > start {
		jsonStr := rawInput[start : end+1]
		if err := json.Unmarshal([]byte(jsonStr), &searchInput); err == nil {
			query, _ = searchInput["query"].(string)
			reason, _ = searchInput["reason"].(string)
			return query, reason, nil
		}
	}

	// 3. 从原始文本中用正则提取 "query": "xxx"
	queryMatch := extractJSONField(rawInput, "query")
	reasonMatch := extractJSONField(rawInput, "reason")
	if queryMatch != "" {
		return queryMatch, reasonMatch, nil
	}

	return "", "", fmt.Errorf("无法从输入中提取搜索参数")
}

// extractJSONField 从包含 JSON 的文本中提取指定字段的值。
func extractJSONField(text string, field string) string {
	pattern := fmt.Sprintf(`"%s"\s*:\s*"([^"]*)"`, field)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

func (i InvokableSearchApprovalTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return i.InvokableTool.Info(ctx)
}

func (i InvokableSearchApprovalTool) InvokableRun(ctx context.Context, argumentsInJSON string,
	opts ...tool.Option,
) (string, error) {
	toolInfo, err := i.Info(ctx)
	if err != nil {
		return "", err
	}

	// 使用健壮的参数提取，支持 JSON 块提取和正则匹配
	query, reason, err := extractSearchParams(argumentsInJSON)
	if err != nil {
		// 参数完全无法解析时，使用原始文本作为 query（不报错，避免流程中断）
		query = strings.TrimSpace(argumentsInJSON)
		if query == "" {
			return "", fmt.Errorf("搜索关键词不能为空，且无法从输入中提取")
		}
		reason = ""
	}
	if query == "" {
		return "", fmt.Errorf("搜索关键词不能为空")
	}

	wasInterrupted, _, storedArguments := tool.GetInterruptState[string](ctx)
	if !wasInterrupted {
		return "", tool.StatefulInterrupt(ctx, &SearchApprovalInfo{
			ToolName: toolInfo.Name,
			Query:    query,
			Reason:   reason,
		}, argumentsInJSON)
	}

	// 被恢复执行，检查是否是本次中断的恢复目标
	isResumeTarget, hasData, data := tool.GetResumeContext[*SearchApprovalInfo](ctx)
	if !isResumeTarget {
		// 不是本次中断的直接恢复目标，但曾被中断过（wasInterrupted=true）
		// 说明是其他节点的恢复目标，让当前节点继续执行原始工具
		return i.InvokableTool.InvokableRun(ctx, storedArguments, opts...)
	}

	if hasData && data != nil {
		// 是本次中断的直接恢复目标，处理用户的审批结果
		result := data.Result
		if result == nil {
			return "", fmt.Errorf("工具 '%s' 恢复时缺少审批结果", toolInfo.Name)
		}

		switch result.Option {
		case 1:
			return fmt.Sprintf("用户选择跳过搜索。关键词 '%s' 未执行。", query), nil
		case 3:
			if result.EditedQuery != nil && *result.EditedQuery != "" {
				// 使用健壮的参数提取，避免 storedArguments 格式损坏导致解析失败
				editedQuery, _, _ := extractSearchParams(storedArguments)
				if editedQuery == "" {
					// 参数无法解析时，直接使用新的 query
					editedQuery = *result.EditedQuery
				}
				stored := map[string]any{
					"query": editedQuery,
				}
				editedJSON, err := json.Marshal(stored)
				if err != nil {
					return "", fmt.Errorf("序列化编辑后参数失败: %w", err)
				}
				res, err := i.InvokableTool.InvokableRun(ctx, string(editedJSON), opts...)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("用户编辑了关键词：从 '%s' 修改为 '%s'。搜索结果：%s",
					query, *result.EditedQuery, res), nil
			}
			return "", fmt.Errorf("编辑后的关键词为空")
		default:
			// Option == 2 或其他：使用原始参数执行搜索
			return i.InvokableTool.InvokableRun(ctx, storedArguments, opts...)
		}
	}

	// hasData == false 或 data == nil：应该重新中断
	return "", tool.StatefulInterrupt(ctx, &SearchApprovalInfo{
		ToolName: toolInfo.Name,
		Query:    query,
		Reason:   reason,
	}, storedArguments)
}

func ParseSearchApprovalResult(input string) (*SearchApprovalResult, error) {
	input = strings.TrimSpace(input)

	if strings.HasPrefix(input, "{") {
		var result SearchApprovalResult
		if err := json.Unmarshal([]byte(input), &result); err != nil {
			return nil, fmt.Errorf("JSON 解析失败: %w", err)
		}
		return &result, nil
	}

	lower := strings.ToLower(input)
	switch {
	case lower == "1" || lower == "skip" || lower == "s" || lower == "跳过":
		return &SearchApprovalResult{Option: 1}, nil
	case lower == "2" || lower == "confirm" || lower == "c" || lower == "y" || lower == "确认":
		return &SearchApprovalResult{Option: 2}, nil
	case strings.HasPrefix(lower, "3"):
		parts := strings.SplitN(input, " ", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("选项 3 需要提供编辑后的关键词")
		}
		editedQuery := strings.TrimSpace(parts[1])
		if editedQuery == "" {
			return nil, fmt.Errorf("编辑后的关键词不能为空")
		}
		return &SearchApprovalResult{Option: 3, EditedQuery: &editedQuery}, nil
	default:
		words := strings.Fields(input)
		if len(words) >= 2 {
			editedQuery := strings.TrimSpace(strings.Join(words, " "))
			if editedQuery != "" {
				return &SearchApprovalResult{Option: 3, EditedQuery: &editedQuery}, nil
			}
		}
		return nil, fmt.Errorf("无效输入，请输入：1（跳过）、2（确认）或 3 <编辑后的关键词>")
	}
}
