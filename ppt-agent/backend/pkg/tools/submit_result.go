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

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

var SubmitResultToolInfo = &schema.ToolInfo{
	Name: "submit_result",
	Desc: "提交最终结果，标记任务完成。当所有幻灯片都已生成并确认满意后，调用此工具提交最终结果。",
	ParamsOneOf: schema.NewParamsOneOfByParams(
		map[string]*schema.ParameterInfo{
			"result": {
				Type:     schema.String,
				Desc:     "最终结果描述",
				Required: true,
			},
			"files": {
				Type: schema.Array,
				ElemInfo: &schema.ParameterInfo{
					Type: schema.Object,
					SubParams: map[string]*schema.ParameterInfo{
						"path": {
							Type:     schema.String,
							Desc:     "文件路径",
							Required: true,
						},
						"desc": {
							Type:     schema.String,
							Desc:     "文件描述",
							Required: false,
						},
					},
				},
				Desc: "生成的文件列表",
			},
		},
	),
}

type ToolSubmitResult struct{}

func (sr *ToolSubmitResult) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return SubmitResultToolInfo, nil
}

func (sr *ToolSubmitResult) Invoke(ctx context.Context, params string, opts ...tool.Option) (string, error) {
	return "任务已完成", nil
}

func (sr *ToolSubmitResult) Stream(ctx context.Context, params string, opts ...tool.Option) (*schema.StreamReader[string], error) {
	return nil, nil
}
