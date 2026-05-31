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
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

var editToolInfo = &schema.ToolInfo{
	Name: "edit_file",
	Desc: `编辑或创建文件，使用给定内容写入指定路径。
* 如果文件已存在，将被覆盖。
* 路径应相对于工作目录。`,
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"path": {
			Type:     "string",
			Desc:     "文件路径（相对于工作目录）",
			Required: true,
		},
		"content": {
			Type:     "string",
			Desc:     "要写入的文件内容",
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

// fixContentObject 检测 LLM 是否将 content 字段传递为原始 JSON 对象而非字符串，
// 并将其转换为字符串。例如：
//
//	{"path":"tasks.json","content":{"title":"x","tasks":[...]}}
//
// 转换为：
//
//	{"path":"tasks.json","content":"{\"title\":\"x\",\"tasks\":[...]}"}
func fixContentObject(argsJSON string) string {
	prefix := `"content":`
	idx := strings.Index(argsJSON, prefix)
	if idx < 0 {
		return argsJSON
	}
	start := idx + len(prefix)
	for start < len(argsJSON) && (argsJSON[start] == ' ' || argsJSON[start] == '\t') {
		start++
	}
	if start >= len(argsJSON) || (argsJSON[start] != '{' && argsJSON[start] != '[') {
		return argsJSON
	}

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
				inner := argsJSON[start : end+1]
				escaped, err := json.Marshal(inner)
				if err != nil {
					return argsJSON
				}
				prefixPart := argsJSON[:idx+len(prefix)]
				suffixPart := argsJSON[end+1:]
				return prefixPart + string(escaped) + suffixPart
			}
		}
	}
	return argsJSON
}

// normalizeContentBlock 将 JSON 字符串中的 Anthropic 内容块格式转换为普通字符串值。
// 例如：{"content":{"type":"text","text":"..."}} 转换为 {"content":"..."}
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

// tryFixUnescapedQuotes 尝试修复内容中包含未转义双引号的 JSON 字符串值。
// 它尝试修复最常见的模式：LLM 生成的 JSON 字符串中包含字面量 " 字符但没有转义它们。
//
// 策略：如果内容看起来像 JSON 片段，尝试解析每个顶层字符串值并重新序列化以正确转义。
func tryFixUnescapedQuotes(content string) string {
	// 尝试 json.Marshal 处理原始字符串——这会转义任何特殊字符
	fixed, err := json.Marshal(content)
	if err != nil {
		return content
	}
	// json.Marshal 将字符串包装在引号中，但我们只想要内容
	decoded := &struct{ S string }{}
	if err := json.Unmarshal(fixed, decoded); err != nil {
		return content
	}
	// 仅在实际更改时使用（引号已被修复）
	if decoded.S != content {
		return decoded.S
	}
	return content
}

// validateAndRepairJSON 检查内容是否为有效 JSON。
// 如果是 JSON 文件（如 tasks.json）且内容无效，
// 它会在写入前尝试修复。
func validateAndRepairJSON(path, content string) (string, bool) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".json" {
		return content, true
	}

	// 快速路径：已经是有效 JSON
	if json.Valid([]byte(content)) {
		return content, true
	}

	// 专门针对 tasks.json 的修复尝试
	base := filepath.Base(path)
	if base == "tasks.json" || base == "outline.json" {
		// 尝试未转义引号的修复
		repaired := tryFixUnescapedQuotes(content)
		if json.Valid([]byte(repaired)) {
			return repaired, true
		}
	}

	return content, false
}

func (e *editFileTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 清理常见的 LLM JSON 错误（尾随逗号等）
	argumentsInJSON = SanitizeJSON(argumentsInJSON)

	normalized, err := normalizeContentBlock(argumentsInJSON)
	if err != nil {
		return "", err
	}

	// LLM 有时将 content 作为原始 JSON 对象而非字符串传递。
	// 检测并字符串化：{"content": {...}} → {"content": "{...}"}
	normalized = fixContentObject(normalized)

	input := &editInput{}
	err = json.Unmarshal([]byte(normalized), input)
	if err != nil {
		return "", err
	}

	// 写入前验证并自动修复 JSON 文件。
	// 这可以防止 tasks.json 因未转义引号而损坏。
	if repaired, ok := validateAndRepairJSON(input.Path, input.Content); !ok {
		return "content is not valid JSON and could not be auto-repaired: " + input.Path, nil
	} else if repaired != input.Content {
		input.Content = repaired
	}

	o := tool.GetImplSpecificOptions(&options{op: e.op}, opts...)
	err = o.op.WriteFile(ctx, input.Path, input.Content)
	if err != nil {
		return err.Error(), nil
	}

	// 写入后验证：重新读取 tasks.json 确认它是有效 JSON。
	// 这可以捕获 validateAndRepairJSON 中漏掉的任何损坏。
	if ext := strings.ToLower(filepath.Ext(input.Path)); ext == ".json" {
		if raw, readErr := os.ReadFile(input.Path); readErr == nil && !json.Valid(raw) {
			return "WARNING: file written but contains invalid JSON: " + input.Path, nil
		}
	}

	return "file written successfully", nil
}
