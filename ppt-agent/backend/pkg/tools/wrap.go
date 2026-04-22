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

// ToolRequestRepairJSON 修复 JSON 格式问题
// 1. 去除模型输出中常见的 markdown 代码块包裹（如 ```json ... ``` 或 ``` ... ```）
// 2. 去除 eino 框架的特殊标记
func ToolRequestRepairJSON(ctx context.Context, request string) (string, error) {
	// Step 1: 去除 markdown 代码块包裹（必须先于 Step 2，因为 fence 可能包裹了 <|FunctionCallBegin|> 等标记）
	request = stripMarkdownFence(request)

	// Step 2: 去除 eino 框架的特殊标记
	request = strings.TrimPrefix(request, "<|FunctionCallBegin|>")
	request = strings.TrimSuffix(request, "<|FunctionCallEnd|>")
	request = strings.TrimPrefix(request, "<think>")
	request = strings.TrimSuffix(request, "</think>")

	return request, nil
}

// stripMarkdownFence 去除 markdown 代码块包裹
// 处理常见的格式：```json\n{...}\n```、```\n{...}\n```、```python\n{...}\n```
// 也处理单行形式：```{...}```、`{...}`
func stripMarkdownFence(s string) string {
	s = strings.TrimSpace(s)

	// 处理 ```lang\n...\n``` 格式（多行代码块，最常见）
	// 先匹配前缀 ```xxx 或 ```（xxx 可以是 json、python、shell 等）
	lines := strings.SplitN(s, "\n", 2)
	if len(lines) >= 2 {
		firstLine := strings.TrimSpace(lines[0])
		// 去除前缀 ```lang，保留 lang 部分以便后续匹配
		if strings.HasPrefix(firstLine, "```") {
			// 去掉开头的 ```lang，保留剩余内容
			remaining := strings.TrimPrefix(firstLine, "```")
			remaining = strings.TrimSpace(remaining)
			// 剩余内容是语言标识（如 "json"、"python"），我们不需要它
			// 从第二行开始到倒数第二行
			content := lines[1]
			// 去掉末尾的 ```（可能在最后一行或倒数第二行）
			// 先去掉尾部换行符相关的部分
			for strings.HasSuffix(content, "\n") {
				content = strings.TrimSuffix(content, "\n")
			}
			if strings.HasSuffix(content, "```") {
				content = strings.TrimSuffix(content, "```")
				content = strings.TrimSpace(content)
			}
			// 现在 content 应该就是纯 JSON 了
			// 检查是否是纯 JSON 对象/数组开头
			content = strings.TrimSpace(content)
			if (strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}")) ||
				(strings.HasPrefix(content, "[") && strings.HasSuffix(content, "]")) {
				return content
			}
			// 不是纯 JSON，继续用原字符串
			s = strings.Join([]string{firstLine, lines[1]}, "\n")
		}
	}

	// 处理 ```{...}``` 单行格式（较少见）
	if strings.HasPrefix(s, "```") && strings.HasSuffix(s, "```") && !strings.Contains(s[3:len(s)-3], "```") {
		s = s[3 : len(s)-3]
		s = strings.TrimSpace(s)
		// 去除语言标识（如 ```json）
		if strings.HasPrefix(s, "json") || strings.HasPrefix(s, "python") || strings.HasPrefix(s, "shell") {
			s = s[4:]
			s = strings.TrimSpace(s)
		}
	}

	// 处理 `{...}` 单行格式（最外层直接是对象或数组）
	s = strings.TrimSpace(s)
	return s
}
