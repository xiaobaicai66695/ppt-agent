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
	"bytes"
	"context"
	"encoding/json"
	"regexp"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

var editToolInfo = &schema.ToolInfo{
	Name: "edit_file",
	Desc: `Edit or create a file with the given content.
* If the file exists, it will be overwritten.
* The path should be relative to the working directory.`,
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"path": {
			Type:     "string",
			Desc:     "File path (relative to working directory)",
			Required: true,
		},
		"content": {
			Type:     "string",
			Desc:     "File content to write",
			Required: true,
		},
	}),
}

func NewEditFileToolImpl(op commandline.Operator) tool.InvokableTool {
	return &editFileTool{op: op}
}

type editFileTool struct {
	op commandline.Operator
}

func (e *editFileTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return editToolInfo, nil
}

type editInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// normalizeContentBlock converts Anthropic content block format in the JSON string
// to plain string values. For example: {"content":{"type":"text","text":"..."}}
// becomes {"content":"..."}
func normalizeContentBlock(argsJSON string) (string, error) {
	var buf bytes.Buffer
	err := json.Compact(&buf, []byte(argsJSON))
	if err != nil {
		return argsJSON, nil
	}
	normalized := buf.String()

	contentBlockRegex := regexp.MustCompile(`"content"\s*:\s*\{\s*"type"\s*:\s*"text"\s*,\s*"text"\s*:\s*("([^"\\]|\\.)*")\s*\}`)
	normalized = contentBlockRegex.ReplaceAllStringFunc(normalized, func(match string) string {
		parts := contentBlockRegex.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		return `"content":` + parts[1]
	})

	return normalized, nil
}

func (e *editFileTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	normalized, err := normalizeContentBlock(argumentsInJSON)
	if err != nil {
		return "", err
	}

	input := &editInput{}
	err = json.Unmarshal([]byte(normalized), input)
	if err != nil {
		return "", err
	}

	o := tool.GetImplSpecificOptions(&options{op: e.op}, opts...)
	err = o.op.WriteFile(ctx, input.Path, input.Content)
	if err != nil {
		return err.Error(), nil
	}
	return "file written successfully", nil
}
