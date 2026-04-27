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

package qa

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/generic"
	"github.com/cloudwego/ppt-agent/pkg/params"
)

var singleQAToolInfo = &schema.ToolInfo{
	Name: "single_qa_review",
	Desc: `执行单页视觉质量审查。每生成一页幻灯片后，立即使用此工具进行该页的 QA 检查。

该工具会：
1. 将指定的单个 PPTX 文件转换为图片
2. 使用视觉 AI 模型审查该页幻灯片
3. 返回该页的 QA 报告，包含问题描述和具体修复建议（必须包含 python-pptx 代码片段）

传入参数 slide_index（页码），工具会自动查找对应的 pptx 文件。`,
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"slide_index": {
			Type:     schema.Integer,
			Desc:     "要审查的幻灯片页码（序号）",
			Required: true,
		},
	}),
}

// qaSystemPrompt 是视觉 QA 的系统提示词。
const qaSystemPrompt = `你是 PPT 视觉质量审查专家，负责对幻灯片进行严格检查，并给出可执行的修复指令。

## 你的职责

你的审查结果将直接交给另一个 AI Agent 执行修复。因此：
- 如果发现问题，必须给出**具体、可执行**的 python-pptx 代码片段
- 如果没有发现问题，明确返回空 issues 数组
- 不要泛泛而谈，必须精确到形状对象、颜色值、坐标、字号等

## 必须检查的问题类型

1. **overlap（重叠）** — 文字与形状/图片重叠、线条穿过文字
2. **overflow（溢出）** — 文字超出文本框边界被截断、超出幻灯片边界
3. **contrast（对比度）** — 浅色文字在浅色背景上、深色文字在深色背景上
4. **spacing（间距）** — 元素间距不一致、元素过于靠近（<0.3英寸）、边距不足（<0.5英寸）
5. **alignment（对齐）** — 同一列的元素没有对齐、视觉重心不稳
6. **placeholder（占位符残留）** — 包含 "xxxx"、"lorem"、"ipsum"、"placeholder" 等占位符文本
7. **ai_style（AI感特征）** — 标题下有装饰线、紫色渐变科技风、过于均匀的配色

## 严重程度定义

- **high** — 明显影响阅读或观感，必须修复（如文字被截断、重叠）
- **medium** — 视觉上不够精致，建议修复（如间距不均、对比度略低）
- **low** — 微小瑕疵，不影响整体

## 审查规则

1. 假设有错，不要因为"看起来还行"就跳过
2. 发现了问题一定要报告，不要遗漏
3，一定给出具体建议，如1，颜色：rgb()，2，排版：某部分应该在某一块，如，左栏的内容超出了页面，以左上为坐标原点，向下向右为正方向，应该位于(10%,20%)处

## 输出格式

请严格以纯 JSON 格式输出，**不要使用任何 markdown 代码块包裹**，直接输出 JSON 文本。

**字符约束（严格遵守）：**
- 所有字符必须是：中文字符、英文字母（a-zA-Z）、数字（0-9）、ASCII 可见符号（空格、标点、运算符号等）
- **禁止使用任何非 ASCII 字符**，包括但不限于：德语/法语/俄语等西方字母变体（ä、é、ö、ß、â、ñ 等）、全角符号、中文引号（已在规范化步骤处理）
- JSON 字符串中的中文内容必须使用标准 UTF-8 编码（Go 的 json 包原生支持）
- RGB 颜色值使用纯 ASCII 格式，如 rgb(255, 0, 0) 或 #FF0000

{
  "total_slides": <总页数>,
  "issues": [
    {
      "slide": <页码>,
      "severity": "high|medium|low",
      "type": "overlap|overflow|contrast|spacing|alignment|placeholder|ai_style",
      "description": "<问题描述，如：第3个文本框内的'未来'二字在青色背景上使用了青色字体>",
      "recommendation": "给出具体的建议，如1，颜色：rgb()，2，排版：某部分应该在某一块，如，左栏的内容超出了页面，以左上为坐标原点，向下向右为正方向，应该位于(10%,20%)处"
    }
  ],
  "summary": "<整体评估，1-2句话>"
}`

// SingleTool 是单页 QA 视觉审查工具。
type SingleTool struct {
	op      commandline.Operator
	modelFn func(ctx context.Context) (model.ToolCallingChatModel, error)
	// 复用已初始化的 model，避免每次调用都重新创建
	cachedModel     model.ToolCallingChatModel
	cachedModelInit sync.Once
	cachedModelErr  error
}

// NewSingleTool 创建一个单页 QA Tool 实例。
func NewSingleTool(op commandline.Operator, modelFn func(ctx context.Context) (model.ToolCallingChatModel, error)) tool.InvokableTool {
	return &SingleTool{op: op, modelFn: modelFn}
}

func (t *SingleTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return singleQAToolInfo, nil
}

// runConverter 执行 Python 脚本将 PPTX 转换为图片。
func (t *SingleTool) runConverter(ctx context.Context, wd string) (map[string]any, error) {
	// 优先使用 PROJECT_ROOT 环境变量定位项目根目录（wd 是输出目录，可能在项目外）
	projectRoot := os.Getenv("PROJECT_ROOT")
	var converter string
	if projectRoot != "" {
		converter = filepath.Join(projectRoot, "pkg", "tools", "qa", "pptx_qa_converter.py")
		if _, err := os.Stat(converter); err == nil {
			// 找到则直接使用
		} else {
			converter = ""
		}
	}
	if converter == "" {
		// 回退：向上搜索 converter 脚本，最多搜索 8 级，并尝试可能的子目录结构
		for i := 1; i <= 8; i++ {
			up := strings.Repeat("../", i)
			// 尝试多种可能的路径结构
			for _, subPath := range []string{
				"pkg/tools/qa/pptx_qa_converter.py",
				"backend/pkg/tools/qa/pptx_qa_converter.py",
			} {
				c := filepath.Join(wd, up, subPath)
				if _, err := os.Stat(c); err == nil {
					converter = c
					break
				}
			}
			if converter != "" {
				break
			}
		}
	}
	if converter == "" {
		return nil, fmt.Errorf("找不到 pptx_qa_converter.py（PROJECT_ROOT 未设置，wd=%s）", wd)
	}
	imgDir := filepath.Join(wd, "qa_images")
	if err := os.MkdirAll(imgDir, 0o755); err != nil {
		return nil, fmt.Errorf("创建 QA 图片目录失败: %v", err)
	}

	pythonBin := "/root/pptx_env/bin/python"
	cmd := exec.CommandContext(ctx, pythonBin, converter,
		"--pptx-dir", wd,
		"--output-dir", imgDir,
		"--dpi", "150")
	cmd.Dir = wd
	cmd.Env = append(os.Environ(), "PATH="+os.Getenv("PATH"))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("PPTX 转换失败: %v, stderr: %s", err, stderr.String())
	}

	var result map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("解析转换结果失败: %v", err)
	}
	return result, nil
}

// findTargetImage 查找对应页码的图片文件。
// 支持多种命名格式：
//   - {全局页号}_{PPT内页号}.jpg：例如 1_01.jpg, 2_01.jpg（converter 实际生成的格式）
//   - {stem}_page_{NN}.jpg：旧格式
//   - slide_{N}.jpg / slide_{NN}.jpg
func findTargetImage(imgDir string, slideIdx int) (string, bool) {
	entries, err := os.ReadDir(imgDir)
	if err != nil {
		return "", false
	}

	// 收集所有图片文件并按页码匹配
	var candidates []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
			continue
		}

		// 提取页码：支持新格式 {N}_{M}.jpg 和旧格式 {stem}_page_{N}.jpg
		if pageNum := extractPageNumFromName(name); pageNum == slideIdx {
			candidates = append(candidates, name)
		}
	}

	if len(candidates) > 0 {
		// 按文件名排序，取第一个
		return candidates[0], true
	}
	return "", false
}

// extractPageNumFromName 从图片文件名中提取页码。
// 支持格式（按优先级）：
//   - {全局页号}_{PPT内页号}.jpg：例如 1_01.jpg, 2_01.jpg（converter 实际生成格式）
//   - {stem}_page_{N}.jpg：旧格式
//   - slide_{N}.jpg / slide_{NN}.jpg
func extractPageNumFromName(name string) int {
	// 尝试 {全局页号}_{PPT内页号}.jpg 格式（converter 生成的格式）
	re := regexp.MustCompile(`^(\d+)_(\d+)\.jpe?g$`)
	matches := re.FindStringSubmatch(name)
	if len(matches) >= 2 {
		if n, err := strconv.Atoi(matches[1]); err == nil {
			return n
		}
	}
	// 尝试 _page_{N}.jpg 格式
	re2 := regexp.MustCompile(`_page_(\d+)`)
	matches2 := re2.FindStringSubmatch(name)
	if len(matches2) >= 2 {
		if n, err := strconv.Atoi(matches2[1]); err == nil {
			return n
		}
	}
	// 尝试 slide_{N}.jpg 格式
	re3 := regexp.MustCompile(`slide_(\d+)`)
	matches3 := re3.FindStringSubmatch(name)
	if len(matches3) >= 2 {
		if n, err := strconv.Atoi(matches3[1]); err == nil {
			return n
		}
	}
	return 0
}

// mergeQAResult 将新结果合并到现有 QA 结果中。
// 按 slide 去重：新结果中同一页的内容会覆盖旧结果。
func mergeQAResult(existing *generic.QAResult, new *generic.QAResult) *generic.QAResult {
	if existing == nil {
		return new
	}
	if new == nil {
		return existing
	}

	// 构建新结果页码集合
	newIssuesMap := make(map[int]bool)
	for _, issue := range new.Issues {
		newIssuesMap[issue.Slide] = true
	}

	// 合并 issues：旧结果中页码不在新结果中的保留，新结果的全部追加
	var mergedIssues []generic.QAIssue
	for _, issue := range existing.Issues {
		if !newIssuesMap[issue.Slide] {
			mergedIssues = append(mergedIssues, issue)
		}
	}
	mergedIssues = append(mergedIssues, new.Issues...)

	merged := &generic.QAResult{
		TotalSlides: new.TotalSlides,
		Issues:      mergedIssues,
		HasIssues:   len(mergedIssues) > 0,
	}

	for _, issue := range mergedIssues {
		if issue.Severity == "high" {
			merged.HasHighIssue = true
			break
		}
	}

	// summary 合并：只有当新 summary 是有效 QA 结果时才替换/追加；错误信息不覆盖已有 summary
	newIsError := strings.Contains(new.Summary, "未找到") || strings.Contains(new.Summary, "转换出错")
	if existing.Summary != "" && new.Summary != "" && !newIsError {
		merged.Summary = existing.Summary + "\n" + new.Summary
	} else if new.Summary != "" && !newIsError {
		merged.Summary = new.Summary
	} else if existing.Summary != "" {
		merged.Summary = existing.Summary
	} else {
		merged.Summary = new.Summary
	}

	return merged
}

func (t *SingleTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	wd, ok := params.GetTypedContextParams[string](ctx, params.WorkDirSessionKey)
	if !ok || wd == "" {
		return "", fmt.Errorf("无法获取工作目录")
	}

	// 解析 slide_index 参数
	var args struct {
		SlideIndex int `json:"slide_index"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("解析参数失败: %v", err)
	}
	slideIdx := args.SlideIndex

	// 检查 QA 尝试次数，每页最多 2 次
	attemptCount, _ := generic.IncrementQAAttempt(wd, slideIdx)
	if attemptCount > 2 {
		existingResult, _ := generic.LoadQAResult(wd)
		if existingResult != nil {
			b, _ := json.Marshal(existingResult)
			return string(b), nil
		}
		// 没有历史结果，返回已超过次数限制
		emptyResult := generic.QAResult{
			TotalSlides: 0,
			Issues:      []generic.QAIssue{},
			Summary:     fmt.Sprintf("第 %d 页 QA 已达到最大尝试次数（2次），跳过后续审查", slideIdx),
		}
		b, _ := json.Marshal(emptyResult)
		return string(b), nil
	}

	// Step 1: 将工作目录下所有 PPTX 转换为图片
	convResult, err := t.runConverter(ctx, wd)
	if err != nil {
		emptyResult := generic.QAResult{TotalSlides: 0, Issues: []generic.QAIssue{}, Summary: "转换出错: " + err.Error()}
		b, _ := json.Marshal(emptyResult)
		return string(b), nil
	}

	if errMsg, ok := convResult["error"].(string); ok && errMsg != "" {
		emptyResult := generic.QAResult{TotalSlides: 0, Issues: []generic.QAIssue{}, Summary: "转换出错: " + errMsg}
		b, _ := json.Marshal(emptyResult)
		return string(b), nil
	}

	textContent, _ := convResult["text_content"].(string)
	totalSlides := int(converterResultToFloat(convResult["total_slides"]))

	// Step 2: 找到对应页码的图片（支持 jpg 和 png）
	imgDir := filepath.Join(wd, "qa_images")
	targetImgName, found := findTargetImage(imgDir, slideIdx)

	if !found {
		// 尝试列出目录中的图片文件以便调试
		existingFiles := []string{}
		if entries, err := os.ReadDir(imgDir); err == nil {
			for _, e := range entries {
				if !e.IsDir() {
					existingFiles = append(existingFiles, e.Name())
				}
			}
		}
		emptyResult := generic.QAResult{
			TotalSlides: totalSlides,
			Issues:      []generic.QAIssue{},
			Summary:     fmt.Sprintf("未找到第 %d 页对应的图片文件（目录: %s，文件: %v）", slideIdx, imgDir, existingFiles),
		}
		b, _ := json.Marshal(emptyResult)
		return string(b), nil
	}

	// Step 3: 调用 LLM 进行单页视觉 QA（复用缓存的 model，避免重复初始化）
	t.cachedModelInit.Do(func() {
		t.cachedModel, t.cachedModelErr = t.modelFn(ctx)
	})
	if t.cachedModelErr != nil {
		return "", fmt.Errorf("创建 LLM 失败: %v", t.cachedModelErr)
	}

	result, err := t.doSingleVisualQA(ctx, t.cachedModel, wd, targetImgName, textContent)
	if err != nil {
		if result != nil {
		} else {
			emptyResult := generic.QAResult{
				TotalSlides: totalSlides,
				Issues:      []generic.QAIssue{},
				Summary:     "QA 执行失败: " + err.Error(),
			}
			b, _ := json.Marshal(emptyResult)
			return string(b), nil
		}
	}
	result.TotalSlides = totalSlides

	result.HasIssues = len(result.Issues) > 0
	result.HasHighIssue = false
	for _, issue := range result.Issues {
		if issue.Severity == "high" {
			result.HasHighIssue = true
			break
		}
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("序列化 QA 结果失败: %v", err)
	}

	// Step 4: 追加到现有 QA 结果（而非覆盖）
	existingResult, _ := generic.LoadQAResult(wd)
	mergedResult := mergeQAResult(existingResult, result)
	if err := generic.SaveQAResult(wd, mergedResult); err != nil {
		fmt.Printf("[QA] 警告: 保存 QA 结果失败: %v\n", err)
	}

	return string(resultJSON), nil
}

// converterResultToFloat 安全地将 interface{} 转换为 float64。
func converterResultToFloat(v any) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

// loadImageAsBase64 加载图片并返回 mimeType 和 data URI。
func loadImageAsBase64(imgPath string) (string, string, error) {
	f, err := os.Open(imgPath)
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	fc, err := io.ReadAll(f)
	if err != nil {
		return "", "", err
	}
	mimeType := http.DetectContentType(fc)
	if mimeType == "application/octet-stream" {
		ext := strings.ToLower(filepath.Ext(imgPath))
		if ext == ".jpg" || ext == ".jpeg" {
			mimeType = "image/jpeg"
		} else if ext == ".png" {
			mimeType = "image/png"
		} else {
			mimeType = "image/jpeg"
		}
	}
	return mimeType, fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(fc)), nil
}

// parseSlideIndex 从图片文件名中解析页码。
func parseSlideIndex(imgName string) int {
	return extractPageNumFromName(imgName)
}

// doSingleVisualQA 对单页图片进行视觉 QA。
func (t *SingleTool) doSingleVisualQA(ctx context.Context, model model.ToolCallingChatModel, wd string, imgName string, textContent string) (*generic.QAResult, error) {
	imgDir := filepath.Join(wd, "qa_images")
	imgPath := filepath.Join(imgDir, imgName)

	slideIdx := parseSlideIndex(imgName)

	mimeType, dataURI, err := loadImageAsBase64(imgPath)
	if err != nil {
		return nil, fmt.Errorf("加载图片失败: %v", err)
	}

	var parts []schema.MessageInputPart
	parts = append(parts, schema.MessageInputPart{
		Type: schema.ChatMessagePartTypeText,
		Text: qaSystemPrompt + "\n\n## 幻灯片文本内容（参考）\n" + textContent,
	})
	parts = append(parts, schema.MessageInputPart{
		Type: schema.ChatMessagePartTypeImageURL,
		Image: &schema.MessageInputImage{
			MessagePartCommon: schema.MessagePartCommon{
				URL:      &dataURI,
				MIMEType: mimeType,
			},
			Detail: "",
		},
	})
	parts = append(parts, schema.MessageInputPart{
		Type: schema.ChatMessagePartTypeText,
		Text: fmt.Sprintf("\n请仔细审查第 %d 页幻灯片图片，并以 JSON 格式输出审查结果，不要输出其他内容。", slideIdx),
	})

	msg := &schema.Message{
		Role:                  schema.User,
		UserInputMultiContent: parts,
	}

	resp, err := model.Generate(ctx, []*schema.Message{msg})
	if err != nil {
		return nil, err
	}

	result := parseQAResponse(resp)

	// 解析失败，自动重试一次
	if result != nil && result.Summary != "" && isParseFailed(result.Summary) {
		retryMsg := &schema.Message{
			Role: schema.User,
			UserInputMultiContent: []schema.MessageInputPart{
				{
					Type: schema.ChatMessagePartTypeText,
					Text: "上一次输出格式有误，请重新输出纯 JSON，不要使用 markdown 代码块包裹，确保所有字符为中文字符、英文字母、数字或 ASCII 可见符号，不要出现德语/法语/俄语等西方字母变体（如 ä, é, ö, ß）。",
				},
			},
		}
		respRetry, retryErr := model.Generate(ctx, []*schema.Message{msg, resp, retryMsg})
		if retryErr == nil {
			result = parseQAResponse(respRetry)
		}
	}

	if result == nil {
		return &generic.QAResult{
			Issues:  []generic.QAIssue{},
			Summary: "QA 完成，但 LLM 返回内容无法解析",
		}, nil
	}
	return result, nil
}

// isParseFailed 判断 summary 是否表示解析失败。
func isParseFailed(summary string) bool {
	return strings.Contains(summary, "无法定位") ||
		strings.Contains(summary, "解析结果失败") ||
		strings.Contains(summary, "invalid character")
}

// parseQAResponse 解析 LLM 返回的 QA 结果。
// 注意：解析失败时不返回 error，而是返回带 summary 的 QAResult。
// 调用方通过检查 summary 内容判断是否解析成功。
func parseQAResponse(resp *schema.Message) *generic.QAResult {
	content := resp.Content
	content = strings.TrimSpace(content)

	// 规范化：替换中文引号为 ASCII 双引号
	content = strings.ReplaceAll(content, "\u201c", "\"")
	content = strings.ReplaceAll(content, "\u201d", "\"")
	content = strings.ReplaceAll(content, "\u2018", "'")
	content = strings.ReplaceAll(content, "\u2019", "'")

	// 提取 JSON 块（第一个 { 到最后一个 }）
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start == -1 || end == -1 || end <= start {
		return &generic.QAResult{
			Issues:  []generic.QAIssue{},
			Summary: "QA 完成，但无法定位 JSON 输出: " + truncateForSummary(content),
		}
	}
	jsonStr := content[start : end+1]

	cleaned := cleanJSONString(jsonStr)

	var result generic.QAResult
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		// JSON 不完整，尝试修复：补全截断的字段值
		fixed := fixIncompleteJSON(jsonStr)
		if fixed != jsonStr {
			if err2 := json.Unmarshal([]byte(fixed), &result); err2 == nil {
				return &result
			}
		}
		return &generic.QAResult{
			Issues:  []generic.QAIssue{},
			Summary: "QA 完成，但解析结果失败: " + err.Error() + "\n原始输出: " + truncateForSummary(content),
		}
	}

	return &result
}

// truncateForSummary 将过长的内容截断，避免 summary 过长。
func truncateForSummary(s string) string {
	const maxLen = 500
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// fixIncompleteJSON 尝试修复被截断的 JSON 字符串。
// 常见截断情况：字符串值或数组元素被截断，截断点通常在最后一个完整字段之后。
func fixIncompleteJSON(jsonStr string) string {
	// 策略：从后向前扫描，找到最后一个安全的截断点
	// 安全截断点：处于对象或数组闭合处的 }, 或 ]
	// 通过检查平衡性和字符串状态来判断

	for i := len(jsonStr) - 1; i >= 0; i-- {
		prefix := jsonStr[:i+1]
		if isValidJSONPrefix(prefix) {
			return prefix
		}
	}
	return jsonStr
}

// isValidJSONPrefix 检查给定字符串是否是有效 JSON 的前缀。
// 通过从前往后解析，验证所有已解析部分的结构完整性。
func isValidJSONPrefix(s string) bool {
	if s == "" || s == "{}" || s == "[]" {
		return false
	}
	inStr := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '"' && (i == 0 || s[i-1] != '\\') {
			inStr = !inStr
			continue
		}
		if inStr {
			continue
		}
		switch ch {
		case '{', '[':
			// 对象/数组开始，OK
		case '}', ']':
			// 不允许在有效前缀末尾出现孤立闭合括号
			return false
		}
	}
	// 有效前缀不能以未闭合字符串结束
	return !inStr
}

// cleanJSONString 清理 JSON 字符串，移除非法字符。
// 移除 JSON 结构外的控制字符，但保留制表符、换行、回车（它们可能在字符串值中）。
// 也移除 JSON 结构外侧的中文引号等非 ASCII 标点（已在规范化步骤处理）。
func cleanJSONString(json string) string {
	var out strings.Builder
	for i := 0; i < len(json); i++ {
		ch := json[i]
		// JSON 结构外的非法控制字符：低于 0x20（除 tab/nl/cr）和 0x7F
		if ch < 0x20 && ch != '\t' && ch != '\n' && ch != '\r' {
			continue
		}
		if ch == 0x7F {
			continue
		}
		out.WriteByte(ch)
	}
	return out.String()
}
