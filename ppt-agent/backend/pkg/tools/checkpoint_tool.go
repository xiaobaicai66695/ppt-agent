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

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/generic"
	"github.com/cloudwego/ppt-agent/pkg/params"
)

var checkpointToolInfo = &schema.ToolInfo{
	Name: "update_progress",
	Desc: `记录幻灯片生成进度。每当成功生成一页幻灯片后，调用此工具记录完成状态。
此工具用于确保即使框架状态丢失，也能通过文件系统追踪真实的生成进度。`,
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"slide_index": {
			Type:        schema.String,
			Desc:        "已完成幻灯片的页码标识，如 4_标题页 或 4.1_金融行业",
			Required:    true,
		},
	}),
}

func NewCheckpointTool(op commandline.Operator) tool.InvokableTool {
	return &checkpointTool{op: op}
}

type checkpointTool struct {
	op commandline.Operator
}

func (c *checkpointTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return checkpointToolInfo, nil
}

type checkpointInput struct {
	SlideIndex string `json:"slide_index"`
}

func (c *checkpointTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	input := &checkpointInput{}
	if err := json.Unmarshal([]byte(argumentsInJSON), input); err != nil {
		return "", err
	}

	if input.SlideIndex == "" {
		return "", fmt.Errorf("slide_index 不能为空")
	}

	wd, ok := params.GetTypedContextParams[string](ctx, params.WorkDirSessionKey)
	if !ok || wd == "" {
		return "", fmt.Errorf("无法获取工作目录")
	}

	if err := generic.AddCompletedSlide(wd, input.SlideIndex); err != nil {
		return "", fmt.Errorf("更新进度失败: %v", err)
	}

	completedCount, _ := generic.GetCompletedCountFromCheckpoint(wd)
	return fmt.Sprintf("已记录第 %s 页完成。当前共完成 %d 页。", input.SlideIndex, completedCount), nil
}
