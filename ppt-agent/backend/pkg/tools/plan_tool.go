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
	"encoding/json"
	"fmt"
	"sync"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/generic"
)

var planToolInfo = generic.PlanToolInfo

// PlanTool 是一个用于输出 PPT 制作计划的工具。
// 它不执行任何实际操作，只是将 LLM 调用时传入的 plan JSON 捕获并存储，
// 供 Orchestrator 在事件循环中提取。
type PlanTool struct {
	result   string // 捕获的 plan JSON
	resultMu sync.Mutex
}

func NewPlanTool() *PlanTool {
	return &PlanTool{}
}

// GetResult 获取捕获的 plan JSON，并在调用后清空。
func (t *PlanTool) GetResult() string {
	t.resultMu.Lock()
	defer t.resultMu.Unlock()
	r := t.result
	t.result = "" // 清空，避免重复使用
	return r
}

// Info 返回工具信息
func (t *PlanTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return planToolInfo, nil
}

// InvokableRun 捕获 plan JSON 并返回成功信息
func (t *PlanTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 解析验证参数
	var req struct {
		Title  string `json:"title"`
		Theme  string `json:"theme"`
		Slides []struct {
			Index       int     `json:"index"`
			Title       string  `json:"title"`
			ContentType string  `json:"content_type"`
			Description string  `json:"description"`
		} `json:"slides"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &req); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	if req.Title == "" {
		return "", fmt.Errorf("title 不能为空")
	}
	if len(req.Slides) == 0 {
		return "", fmt.Errorf("slides 不能为空")
	}

	// 捕获原始 JSON
	t.resultMu.Lock()
	t.result = argumentsInJSON
	t.resultMu.Unlock()

	return fmt.Sprintf("PPT 计划已创建：标题=%s，包含 %d 页幻灯片", req.Title, len(req.Slides)), nil
}
