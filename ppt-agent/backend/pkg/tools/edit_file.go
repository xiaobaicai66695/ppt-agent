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
	"strings"

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

// fixContentObject detects when the LLM passes a raw JSON object for the content field
// instead of a string, and converts it. For example:
//
//	{"path":"tasks.json","content":{"title":"x","tasks":[...]}}
//
// becomes:
//
//	{"path":"tasks.json","content":"{\"title\":\"x\",\"tasks\":[...]}"}
func fixContentObject(argsJSON string) string {
	prefix := `"content":`
	idx := strings.Index(argsJSON, prefix)
	if idx < 0 {
		return argsJSON
	}
	start := idx + len(prefix)
	// skip whitespace
	for start < len(argsJSON) && (argsJSON[start] == ' ' || argsJSON[start] == '\t') {
		start++
	}
	if start >= len(argsJSON) || argsJSON[start] != '{' && argsJSON[start] != '[' {
		return argsJSON // already a string, number, or other primitive
	}

	// Find the matching end by counting braces
	depth := 0
	end := start
	inString := false
	var escapeNext bool
	for ; end < len(argsJSON); end++ {
		c := argsJSON[end]
		if escapeNext {
			escapeNext = false
			continue
		}
		if inString {
			if c == '\\' {
				escapeNext = true
			} else if c == '"' {
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '{', '[':
			depth++
		case '}', ']':
			depth--
			if depth == 0 {
				// end is the closing brace/bracket — need to include it
				inner := argsJSON[start : end+1]
				escaped, err := json.Marshal(inner)
				if err != nil {
					return argsJSON
				}
				// json.Marshal double-escapes the string; we need the JSON string form
				prefixPart := argsJSON[:idx+len(prefix)]
				suffixPart := argsJSON[end+1:]
				return prefixPart + string(escaped) + suffixPart
			}
		}
	}
	return argsJSON
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

	// LLM sometimes passes content as a raw JSON object instead of a string.
	// Detect and stringify: {"content": {...}} → {"content": "{...}"}
	normalized = fixContentObject(normalized)

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
