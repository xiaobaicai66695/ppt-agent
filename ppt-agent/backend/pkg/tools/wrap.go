package tools

import (
	"context"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type ToolRequestPreprocess func(ctx context.Context, request string) (string, error)

type ToolResponsePostprocess func(ctx context.Context, response string) (string, error)

func NewWrapTool(t tool.InvokableTool, preprocess []ToolRequestPreprocess, postprocess []ToolResponsePostprocess) tool.InvokableTool {
	return &wrapTool{
		baseTool:    t,
		preprocess:  preprocess,
		postprocess: postprocess,
	}
}

type wrapTool struct {
	baseTool    tool.InvokableTool
	preprocess  []ToolRequestPreprocess
	postprocess []ToolResponsePostprocess
}

func (w *wrapTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return w.baseTool.Info(ctx)
}

func (w *wrapTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	req := argumentsInJSON
	for _, pre := range w.preprocess {
		var err error
		req, err = pre(ctx, req)
		if err != nil {
			return "", err
		}
	}

	resp, err := w.baseTool.InvokableRun(ctx, req, opts...)
	if err != nil {
		return "", err
	}

	for _, post := range w.postprocess {
		resp, err = post(ctx, resp)
		if err != nil {
			return "", err
		}
	}

	return resp, nil
}

// ToolRequestRepairJSON 修复 LLM 输出中常见的 JSON 格式问题。
// 1. 去除 markdown 代码块包装（```json ... ``` 或 ``` ... ```）
// 2. 移除 eino 框架的特殊标记
// 3. 修剪尾随的 {} 或 } 字符（LLM 有时会生成 {"query": "..."}{}）
func ToolRequestRepairJSON(ctx context.Context, request string) (string, error) {
	request = stripMarkdownFence(request)

	request = strings.TrimPrefix(request, "<|FunctionCallBegin|>")
	request = strings.TrimSuffix(request, "<|FunctionCallEnd|>")
	request = strings.TrimPrefix(request, "<think>")
	request = strings.TrimSuffix(request, "</think>")
	request = strings.TrimSpace(request)

	// 修剪尾随的 } 或 {}（LLM 有时附加额外的闭合花括号）
	for {
		trimmed := strings.TrimRight(request, "}")
		if len(trimmed) == len(request) {
			break
		}
		if len(trimmed) == 0 {
			break
		}
		request = trimmed
	}

	// 修剪尾随的标点符号/空白
	request = strings.TrimRight(request, ",，;； \t\n")

	return request, nil
}

// stripMarkdownFence 去除字符串中的 markdown 代码块包装。
// 处理：```json\n{...}\n```, ```\n{...}\n```, ```python\n{...}\n```
// 也处理单行形式：```{...}``` 和 `{...}`
func stripMarkdownFence(s string) string {
	s = strings.TrimSpace(s)

	lines := strings.SplitN(s, "\n", 2)
	if len(lines) >= 2 {
		firstLine := strings.TrimSpace(lines[0])
		if strings.HasPrefix(firstLine, "```") {
			content := lines[1]
			for strings.HasSuffix(content, "\n") {
				content = strings.TrimSuffix(content, "\n")
			}
			if strings.HasSuffix(content, "```") {
				content = strings.TrimSuffix(content, "```")
				content = strings.TrimSpace(content)
			}
			content = strings.TrimSpace(content)
			if (strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}")) ||
				(strings.HasPrefix(content, "[") && strings.HasSuffix(content, "]")) {
				return content
			}
			s = strings.Join([]string{firstLine, lines[1]}, "\n")
		}
	}

	if strings.HasPrefix(s, "```") && strings.HasSuffix(s, "```") && !strings.Contains(s[3:len(s)-3], "```") {
		s = s[3 : len(s)-3]
		s = strings.TrimSpace(s)
		if strings.HasPrefix(s, "json") || strings.HasPrefix(s, "python") || strings.HasPrefix(s, "shell") {
			s = s[4:]
			s = strings.TrimSpace(s)
		}
	}

	return strings.TrimSpace(s)
}
