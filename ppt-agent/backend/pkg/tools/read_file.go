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
)

var readToolInfo = &schema.ToolInfo{
	Name: "read_file",
	Desc: `Read the content of a file.
* Use this tool to read files for debugging or to understand the file content.`,
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"path": {
			Type:     "string",
			Desc:     "File path (relative to working directory)",
			Required: true,
		},
		"start_row": {
			Type:     "integer",
			Desc:     "Start row (0-indexed)",
			Required: false,
		},
		"n_rows": {
			Type:     "integer",
			Desc:     "Number of rows to read",
			Required: false,
		},
	}),
}

func NewReadFileToolImpl(op commandline.Operator) tool.InvokableTool {
	return &readFileTool{op: op}
}

type readFileTool struct {
	op commandline.Operator
}

func (r *readFileTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return readToolInfo, nil
}

type readInput struct {
	Path     string `json:"path"`
	StartRow *int   `json:"start_row"`
	NRows    *int   `json:"n_rows"`
}

func (r *readFileTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	input := &readInput{}
	err := json.Unmarshal([]byte(argumentsInJSON), input)
	if err != nil {
		return "", err
	}

	startRow := 0
	nRows := -1
	if input.StartRow != nil {
		startRow = *input.StartRow
	}
	if input.NRows != nil {
		nRows = *input.NRows
	}

	o := tool.GetImplSpecificOptions(&options{op: r.op}, opts...)
	content, err := o.op.ReadFile(ctx, input.Path)
	if err != nil {
		return "", err
	}

	if nRows > 0 {
		return fmt.Sprintf("File: %s\nContent (lines %d-%d):\n%s", input.Path, startRow, startRow+nRows, content), nil
	}
	return fmt.Sprintf("File: %s\nContent:\n%s", input.Path, content), nil
}
