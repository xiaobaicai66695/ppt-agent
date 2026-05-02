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
1. 查找指定 PPTX 文件对应的图片
2. 使用视觉 AI 模型审查该页幻灯片
3. 返回该页的 QA 报告，包含问题描述和具体修复建议（必须包含 python-pptx 代码片段）

传入参数 pptx_filename（PPTX 文件名），工具会自动查找对应的图片文件。`,
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"pptx_filename": {
			Type:     schema.String,
			Desc:     "PPTX 文件名，如 4_标题页.pptx 或 4.1_金融与法律.pptx",
			Required: true,
		},
	}),
}

// qaSystemPrompt 是视觉 QA 的系统提示词。
const qaSystemPrompt = `你是 PPT 视觉质量审查专家，负责对幻灯片进行严格检查，并给出可执行的修复指令。

## 你的职责

你的审查结果将直接交给另一个 AI Agent 执行修复。因此：
- 如果发现问题，给出**具体、可执行**的修复指令
- 如果没有发现问题，明确说明"该页检查通过"
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
3. 给出具体建议，如颜色 rgb()、排版位置坐标（以左上为原点，向下向右为正方向）
4. 参照模板规范检查：每种 content_type 对应的模板（templates/single-page/*.json）有明确的元素规范和 NEVER 清单

## 输出格式

请用**自然语言**输出审查结果，格式如下：

如果有问题：
【问题1 - high】
- 页面：<页码标识>
- 描述：<问题描述>
- 修复：<具体可执行的修复指令，包含 python-pptx 代码>

【问题2 - medium】
...

如果无问题：
【审查结果】该页检查通过，无视觉问题。`

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

// findTargetImage 查找对应 PPTX 文件名的图片。
// 支持多种命名格式：
//   - {N.M}.jpg：例如 4.1.jpg（子页，N.M 格式）
//   - {N_MM}.jpg：例如 4_01.jpg（普通页，N_MM 格式）
//   - {stem}.jpg：直接匹配文件名（不含扩展名）
func findTargetImage(imgDir string, pptxFilename string) (string, bool) {
	entries, err := os.ReadDir(imgDir)
	if err != nil {
		return "", false
	}

	// 提取待查找的 stem（如 "4.1"、"4_01"）
	stem := strings.TrimSuffix(pptxFilename, ".pptx")

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
			continue
		}

		// 匹配：去掉扩展名后等于 stem
		candidateStem := strings.TrimSuffix(name, filepath.Ext(name))
		if candidateStem == stem {
			return name, true
		}
	}

	return "", false
}

// extractPageNumFromName 从图片文件名中提取页码标识（用于 QA 结果中的 slide 字段）。
// 支持格式：
//   - {N.M}.jpg：子页，如 4.1
//   - {N_MM}.jpg：普通页，如 4_01
//   - {stem}.jpg：直接返回 stem
func extractPageNumFromName(name string) string {
	stem := strings.TrimSuffix(name, filepath.Ext(name))
	return stem
}

// mergeQAResult 将新结果合并到现有 QA 结果中。
// 直接追加 Reports 列表。
func mergeQAResult(existing *generic.QAResult, new *generic.QAResult) *generic.QAResult {
	if existing == nil {
		return new
	}
	if new == nil {
		return existing
	}

	merged := &generic.QAResult{
		TotalSlides: new.TotalSlides,
		Reports:     append(existing.Reports, new.Reports...),
		HasIssues:   existing.HasIssues || new.HasIssues,
		HasHighIssue: existing.HasHighIssue || new.HasHighIssue,
	}

	return merged
}

func (t *SingleTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	wd, ok := params.GetTypedContextParams[string](ctx, params.WorkDirSessionKey)
	if !ok || wd == "" {
		return "", fmt.Errorf("无法获取工作目录")
	}

	// 解析 pptx_filename 参数
	var args struct {
		PPTXFilename string `json:"pptx_filename"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("解析参数失败: %v", err)
	}
	pptxFilename := args.PPTXFilename
	if pptxFilename == "" {
		return "", fmt.Errorf("pptx_filename 参数不能为空")
	}

	// 检查 QA 尝试次数，每页最多 2 次
	attemptCount, _ := generic.IncrementQAAttempt(wd, pptxFilename)
	if attemptCount > 2 {
		existingResult, _ := generic.LoadQAResult(wd)
		if existingResult != nil {
			b, _ := json.Marshal(existingResult)
			return string(b), nil
		}
		// 没有历史结果，返回已超过次数限制
		emptyResult := generic.QAResult{
			TotalSlides: 0,
			Reports:     []string{},
			Summary:     fmt.Sprintf("%s QA 已达到最大尝试次数（2次），跳过后续审查", pptxFilename),
		}
		b, _ := json.Marshal(emptyResult)
		return string(b), nil
	}

	// Step 1: 将工作目录下所有 PPTX 转换为图片
	convResult, err := t.runConverter(ctx, wd)
	if err != nil {
		emptyResult := generic.QAResult{TotalSlides: 0, Reports: []string{}, Summary: "转换出错: " + err.Error()}
		b, _ := json.Marshal(emptyResult)
		return string(b), nil
	}

	if errMsg, ok := convResult["error"].(string); ok && errMsg != "" {
		emptyResult := generic.QAResult{TotalSlides: 0, Reports: []string{}, Summary: "转换出错: " + errMsg}
		b, _ := json.Marshal(emptyResult)
		return string(b), nil
	}

	textContent, _ := convResult["text_content"].(string)
	totalSlides := int(converterResultToFloat(convResult["total_slides"]))

	// Step 2: 找到对应 PPTX 文件名的图片
	imgDir := filepath.Join(wd, "qa_images")
	targetImgName, found := findTargetImage(imgDir, pptxFilename)

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
			Reports:     []string{},
			Summary:     fmt.Sprintf("未找到 %s 对应的图片文件（目录: %s，文件: %v）", pptxFilename, imgDir, existingFiles),
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
				Reports:     []string{},
				Summary:     "QA 执行失败: " + err.Error(),
			}
			b, _ := json.Marshal(emptyResult)
			return string(b), nil
		}
	}
	result.TotalSlides = totalSlides

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

// parseSlideIndex 从图片文件名中解析页码标识（用于 QA 结果中的 slide 字段）。
func parseSlideIndex(imgName string) string {
	return extractPageNumFromName(imgName)
}

// doSingleVisualQA 对单页图片进行视觉 QA。
func (t *SingleTool) doSingleVisualQA(ctx context.Context, model model.ToolCallingChatModel, wd string, imgName string, textContent string) (*generic.QAResult, error) {
	imgDir := filepath.Join(wd, "qa_images")
	imgPath := filepath.Join(imgDir, imgName)

	slideKey := parseSlideIndex(imgName) // 现在返回 PPTX 文件名 stem

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
		Text: fmt.Sprintf("\n请仔细审查幻灯片图片 %s，以自然语言输出审查结果。", slideKey),
	})

	msg := &schema.Message{
		Role:                  schema.User,
		UserInputMultiContent: parts,
	}

	resp, err := model.Generate(ctx, []*schema.Message{msg})
	if err != nil {
		return nil, err
	}

	report := strings.TrimSpace(resp.Content)
	if report == "" {
		return &generic.QAResult{
			Reports: []string{slideKey + "|（LLM 返回为空）"},
			Summary: "LLM 返回为空",
		}, nil
	}

	// 判断是否有问题：通过关键词检测
	hasHigh := strings.Contains(report, "high") || strings.Contains(report, "【问题")
	hasMedium := strings.Contains(report, "medium")
	hasIssues := hasHigh || hasMedium || !strings.Contains(report, "检查通过") && !strings.Contains(report, "无视觉问题")

	result := &generic.QAResult{
		Reports:      []string{slideKey + "|" + report},
		HasIssues:    hasIssues,
		HasHighIssue: hasHigh,
	}

	return result, nil
}

