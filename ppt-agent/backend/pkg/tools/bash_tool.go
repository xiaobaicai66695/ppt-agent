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
	"strings"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

var bashToolInfo = &schema.ToolInfo{
	Name: "bash",
	Desc: `在 Bash Shell 中执行命令。
* 调用此工具时，"command" 参数的内容不需要 XML 转义。
* 此工具无法访问互联网。
* 可以通过 apt 和 pip 访问常用的 Linux 和 Python 包镜像。
* 状态在命令调用和与用户的讨论之间保持持久。
* 请使用后台运行长时间命令，例如 'sleep 10 &' 或在后台启动服务器。`,
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"command": {
			Type:     "string",
			Desc:     "要执行的命令",
			Required: true,
		},
	}),
}

func NewBashToolImpl(op commandline.Operator) tool.InvokableTool {
	return &bashTool{op: op}
}

type bashTool struct {
	op commandline.Operator
}

func (b *bashTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return bashToolInfo, nil
}

type shellInput struct {
	Command string `json:"command"`
}

func (b *bashTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	input := &shellInput{}
	err := json.Unmarshal([]byte(argumentsInJSON), input)
	if err != nil {
		return "", err
	}
	if len(input.Command) == 0 {
		return "command cannot be empty", nil
	}
	o := tool.GetImplSpecificOptions(&options{op: b.op}, opts...)
	cmd, err := o.op.RunCommand(ctx, []string{input.Command})
	if err != nil {
		if strings.HasPrefix(err.Error(), "internal error") {
			return err.Error(), nil
		}
		return "", err
	}
	return formatCommandOutput(cmd), nil
}

func formatCommandOutput(output *commandline.CommandOutput) string {
	return fmt.Sprintf("---\nstdout:\n%s\n---\nstderr:\n%s\n---", output.Stdout, output.Stderr)
}
