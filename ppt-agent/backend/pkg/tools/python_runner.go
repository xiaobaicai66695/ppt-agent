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
	"os"
	"path/filepath"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/params"
)

var pythonRunnerToolInfo = &schema.ToolInfo{
	Name: "python3",
	Desc: `Execute Python code. The code will be saved to a temporary .py file and executed.
* Use this tool to run Python scripts for PPT generation.`,
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"code": {
			Type:     "string",
			Desc:     "Python code to execute",
			Required: true,
		},
	}),
}

func NewPythonRunnerToolImpl(op commandline.Operator) tool.InvokableTool {
	return &pythonRunnerTool{op: op}
}

type pythonRunnerTool struct {
	op commandline.Operator
}

func (p *pythonRunnerTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return pythonRunnerToolInfo, nil
}

type pythonInput struct {
	Code string `json:"code"`
}

func (p *pythonRunnerTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	input := &pythonInput{}
	if err := json.Unmarshal([]byte(argumentsInJSON), input); err != nil {
		return "", err
	}

	if len(input.Code) == 0 {
		return "code cannot be empty", nil
	}

	wd, _ := params.GetTypedContextParams[string](ctx, params.WorkDirSessionKey)
	if wd == "" {
		wd, _ = os.Getwd()
	}

	tmpFile := filepath.Join(wd, "temp_script.py")
	err := os.WriteFile(tmpFile, []byte(input.Code), 0o644)
	if err != nil {
		return fmt.Sprintf("Failed to write temp file: %v", err), nil
	}

	cmd := []string{"/root/pptx_env/bin/python", tmpFile}
	o := tool.GetImplSpecificOptions(&options{op: p.op}, opts...)
	output, err := o.op.RunCommand(ctx, cmd)
	if err != nil {
		return fmt.Sprintf("Python execution failed: %v", err), nil
	}

	os.Remove(tmpFile)

	return formatPythonOutput(output), nil
}

func formatPythonOutput(output *commandline.CommandOutput) string {
	result := ""
	if output.Stdout != "" {
		result += fmt.Sprintf("stdout:\n%s\n", output.Stdout)
	}
	if output.Stderr != "" {
		result += fmt.Sprintf("stderr:\n%s\n", output.Stderr)
	}
	if result == "" {
		result = "(no output)"
	}
	return result
}
