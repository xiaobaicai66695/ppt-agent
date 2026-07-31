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
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/generic"
	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/metrics"
	"github.com/cloudwego/ppt-agent/pkg/params"
	"github.com/cloudwego/ppt-agent/pkg/tools/pythonutil"
)

var batchPDFToolInfo = &schema.ToolInfo{
	Name: "batch_pdf_review",
	Desc: "批量 PDF 视觉质量审查工具。每次处理一批幻灯片（每批 5 页），独立转换并审查。\n\n【参数】batch_idx（integer）：当前批次索引，从 0 开始，用于生成唯一的 PDF 文件名（如 1-5.pdf）。\n\n工具会：\n1. 按文件名顺序取该批次的 PPTX 文件\n2. 独立转换为 1-5.pdf\n3. 将该批 PDF 发送给多模态视觉模型审查\n4. 返回按页分组的 QA 报告",
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"batch_idx": {
			Type:     "integer",
			Desc:     "当前批次索引（从 0 开始），如第 1 批传 0，第 2 批传 1，用于生成唯一的 PDF 文件名",
			Required: true,
		},
	}),
}

// batchPageRange 从一批 PPTX 文件名中提取页码范围，返回如 "1-5"、"6-10"。
// 文件名格式通常为 "1_标题.pptx"，提取开头的数字作为页码。
func batchPageRange(files []string) string {
	if len(files) == 0 {
		return "0-0"
	}
	pages := make([]int, 0, len(files))
	for _, f := range files {
		name := filepath.Base(f)
		name = strings.TrimSuffix(name, ".pptx")
		parts := strings.SplitN(name, "_", 2)
		if n, err := strconv.Atoi(parts[0]); err == nil && n > 0 {
			pages = append(pages, n)
		}
	}
	if len(pages) == 0 {
		return "unknown"
	}
	sort.Ints(pages)
	return fmt.Sprintf("%d-%d", pages[0], pages[len(pages)-1])
}

// BatchPDFTool 是批量 PDF 视觉审查工具。
// 将所有幻灯片合并为 PDF 后，一次发给多模态 LLM 审查（一次调用替代 N 次）。
type BatchPDFTool struct {
	op      commandline.Operator
	modelFn func(ctx context.Context) (model.ToolCallingChatModel, error)

	cachedModel    model.ToolCallingChatModel
	cachedModelMu  sync.Mutex
	cachedModelAt  time.Time
	cachedModelErr error

	conversionCache   map[string]*conversionCacheEntry
	conversionCacheMu sync.Mutex
}

// modelCacheTTL 模型缓存有效期（超时后重新创建，防止 token 过期或连接断开导致 QA 永久失效）
const modelCacheTTL = 10 * time.Minute

// conversionCacheTTL 转换结果缓存有效期（5 分钟足够完成一次批量 QA）
const conversionCacheTTL = 5 * time.Minute

// batchQASize 每批质检的幻灯片数量
const batchQASize = 5

type conversionCacheEntry struct {
	PptxFiles  []string          // 按文件名排序的所有 PPTX 文件路径
	TextByFile map[string]string // key: 文件名, value: 该文件所有幻灯片的文本
	ExpiredAt  time.Time
}

// NewBatchPDFTool 创建一个批量 PDF QA Tool 实例。
func NewBatchPDFTool(op commandline.Operator, modelFn func(ctx context.Context) (model.ToolCallingChatModel, error)) tool.InvokableTool {
	return &BatchPDFTool{op: op, modelFn: modelFn}
}

func (t *BatchPDFTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return batchPDFToolInfo, nil
}

// listPPTXAndExtractText 返回 workDir 下所有 PPTX 文件（按文件名排序）及其文本内容。
// 结果会被缓存，避免重复遍历目录和提取文本。
func (t *BatchPDFTool) listPPTXAndExtractText(wd string) ([]string, map[string]string, error) {
	t.conversionCacheMu.Lock()
	defer t.conversionCacheMu.Unlock()

	if t.conversionCache == nil {
		t.conversionCache = make(map[string]*conversionCacheEntry)
	}
	entry, ok := t.conversionCache[wd]
	if ok && time.Now().Before(entry.ExpiredAt) {
		return entry.PptxFiles, entry.TextByFile, nil
	}

	entries, err := os.ReadDir(wd)
	if err != nil {
		return nil, nil, fmt.Errorf("读取目录失败: %w", err)
	}

	var pptxFiles []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".pptx") {
			continue
		}
		pptxFiles = append(pptxFiles, filepath.Join(wd, e.Name()))
	}
	if len(pptxFiles) == 0 {
		return nil, nil, fmt.Errorf("目录中未找到 PPTX 文件")
	}

	// 按文件名排序，保证分批顺序稳定
	sort.Strings(pptxFiles)

	textByFile, err := t.extractTextFromPPTXFilesMap(pptxFiles)
	if err != nil {
		return nil, nil, err
	}

	t.conversionCache[wd] = &conversionCacheEntry{
		PptxFiles:  pptxFiles,
		TextByFile: textByFile,
		ExpiredAt:  time.Now().Add(conversionCacheTTL),
	}

	return pptxFiles, textByFile, nil
}

// convertBatchToPDF 将指定批次的 PPTX 文件转换为合并 PDF，返回 PDF 路径。
func (t *BatchPDFTool) convertBatchToPDF(ctx context.Context, wd string, batchFiles []string, batchIdx int) (string, error) {
	converter := pythonutil.FindConverterPy(wd)
	if converter == "" {
		return "", fmt.Errorf("找不到 pptx_qa_converter.py")
	}

	// 从文件名提取页码范围，生成描述性文件名，如 "1-5.pdf"、"6-10.pdf"
	pageRange := batchPageRange(batchFiles)
	outName := fmt.Sprintf("%s.pdf", pageRange)
	outPDF := filepath.Join(wd, outName)

	// 检查是否已有该批次 PDF（缓存命中）
	if _, err := os.Stat(outPDF); err == nil {
		return outPDF, nil
	}

	// 只传文件名（不含路径），converter 会在 pptx_dir 下拼接
	basenames := make([]string, len(batchFiles))
	for i, f := range batchFiles {
		basenames[i] = filepath.Base(f)
	}

	pythonBin := pythonutil.GetPythonBinary()
	cmdArgs := []string{
		converter,
		"--pptx-dir", wd,
		"--output-dir", wd,
		"--pdf-only",
		"--files",
	}
	cmdArgs = append(cmdArgs, basenames...)
	cmd := exec.CommandContext(ctx, pythonBin, cmdArgs...)
	cmd.Dir = wd

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("批次 %d PDF 生成失败: %v, stderr: %s", batchIdx, err, stderr.String())
	}

	var result map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return "", fmt.Errorf("解析批次 %d PDF 生成结果失败: %v", batchIdx, err)
	}

	if errMsg, ok := result["error"].(string); ok && errMsg != "" {
		return "", fmt.Errorf("批次 %d: %s", batchIdx, errMsg)
	}

	// converter 可能把 PDF 输出到不同位置，尝试多种可能
	candidates := []string{
		outPDF,
		filepath.Join(wd, fmt.Sprintf("merged_%s.pdf", pageRange)),
		filepath.Join(wd, "merged.pdf"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			if p != outPDF {
				os.Rename(p, outPDF)
			}
			return outPDF, nil
		}
	}

	// 尝试从 converter 输出中找实际生成的 PDF 路径
	if actualPath, ok := result["pdf_path"].(string); ok && actualPath != "" {
		if _, err := os.Stat(actualPath); err == nil {
			if actualPath != outPDF {
				os.Rename(actualPath, outPDF)
			}
			return outPDF, nil
		}
	}

	return "", fmt.Errorf("批次 %d 未找到生成的 PDF 文件", batchIdx)
}

// extractTextFromPPTXFilesMap 从给定的 PPTX 文件中提取文本内容，
// 返回以完整文件路径为键的映射。
func (t *BatchPDFTool) extractTextFromPPTXFilesMap(pptxFiles []string) (map[string]string, error) {
	if len(pptxFiles) == 0 {
		return nil, nil
	}

	script := `
import sys
import json
from pptx import Presentation

results = []
for pptx_path in sys.argv[1:]:
    try:
        prs = Presentation(pptx_path)
        texts = []
        for slide in prs.slides:
            for shape in slide.shapes:
                if hasattr(shape, "text") and shape.text.strip():
                    texts.append(shape.text.strip())
        results.append({"path": pptx_path, "text": "\n".join(texts)})
    except Exception:
        results.append({"path": pptx_path, "text": ""})

print(json.dumps(results, ensure_ascii=False))
`

	cmdArgs := make([]string, 0, 2+len(pptxFiles))
	cmdArgs = append(cmdArgs, "-c", script)
	cmdArgs = append(cmdArgs, pptxFiles...)
	cmd := exec.Command(pythonutil.GetPythonBinary(), cmdArgs...)
	out, _ := cmd.Output()
	if len(out) == 0 {
		return nil, nil
	}

	var results []struct {
		Path string `json:"path"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(out, &results); err != nil {
		return nil, nil
	}

	textByFile := make(map[string]string)
	for _, r := range results {
		textByFile[r.Path] = r.Text
	}
	return textByFile, nil
}

// parseScore 从 QA 报告文本中提取 1-5 分评分。
// 使用 LastIndex 查找最后一个评分行，避免修复建议中引用评分标准时产生的误匹配。
func parseScore(report string) int {
	patterns := []string{"【评分】", "评分：", "评分:", "score:", "Score:"}
	for _, prefix := range patterns {
		idx := strings.LastIndex(report, prefix)
		if idx < 0 {
			continue
		}
		rest := report[idx+len(prefix):]
		rest = strings.TrimSpace(rest)
		if len(rest) > 0 {
			c := rest[0]
			if c >= '1' && c <= '5' {
				return int(c - '0')
			}
		}
	}
	return 0
}

// doBatchVisualQA 将当前批次的 PPTX 文件转换为 PDF 并发送给多模态 LLM 进行视觉审查。
func (t *BatchPDFTool) doBatchVisualQA(ctx context.Context, model model.ToolCallingChatModel, wd string, batchIdx int) (*generic.QAResult, error) {
	pdfFiles, textByFile, err := t.listPPTXAndExtractText(wd)
	if err != nil {
		return nil, err
	}

	start := batchIdx * batchQASize
	end := start + batchQASize
	if start >= len(pdfFiles) {
		return &generic.QAResult{
			Reports: []string{fmt.Sprintf("batch_%02d.pdf|（无文件）", batchIdx)},
			Summary: "该批次无文件",
		}, nil
	}
	if end > len(pdfFiles) {
		end = len(pdfFiles)
	}
	batchFiles := pdfFiles[start:end]
	pageRange := batchPageRange(batchFiles)

	pdfPath, err := t.convertBatchToPDF(ctx, wd, batchFiles, batchIdx)
	if err != nil {
		return nil, err
	}

	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("读取 PDF 文件失败: %v", err)
	}
	pdfBase64 := base64.StdEncoding.EncodeToString(pdfBytes)

	batchPrompt := `你是 PPT 视觉质量审查专家，负责对一份包含多页幻灯片的 PDF 进行严格检查，并给出可执行的修复指令。

## 你的职责

你的审查结果将直接交给另一个 AI Agent 执行修复。因此：
- 如果发现问题，给出**具体、可执行**的修复指令
- 如果没有问题，明确说明"该页检查通过"

## 必须检查的问题类型

1. **overlap（重叠）** — 文字与形状/图片重叠、线条穿过文字
2. **overflow（溢出）** — 文字超出文本框边界被截断、超出幻灯片边界
3. **contrast（对比度）** — 浅色文字在浅色背景上、深色文字在深色背景上
4. **spacing（间距）** — 元素间距不一致、元素过于靠近（<0.3英寸）、边距不足（<0.5英寸）
5. **alignment（对齐）** — 同一列的元素没有对齐、视觉重心不稳
6. **placeholder（占位符残留）** — 包含 "xxxx"、"lorem"、"ipsum"、"placeholder" 等占位符文本
7. **ai_style（AI感特征）** — 标题下有装饰线、紫色渐变科技风、过于均匀的配色
8. **layout_monotony（布局单调）** — 所有元素机械式均匀排列，缺乏视觉节奏变化
9. **content_emptiness（内容空洞）** — 文字内容缺乏深度，仅有标题级别的空洞罗列

## 严重程度定义

- **high** — 明显影响阅读或观感，必须修复（如文字被截断、重叠、内容空洞无实质信息）
- **medium** — 视觉上不够精致，建议修复（如间距不均、对比度略低）
- **low** — 微小瑕疵，不影响整体

## 综合评分（必须）

在审查结束后，为整体打出 1-5 分：
- **1 分** — 不可用：内容空洞或存在多处 high 问题
- **2 分** — 较差：有 high 问题或内容明显单薄
- **3 分** — 及格：无 high 问题，内容基本充实
- **4 分** — 良好：布局合理，内容充实
- **5 分** — 优秀：设计精良，无任何问题

## 输出格式

必须为**每一页**单独输出审查结果：

【第 N 页】
- 状态：pass / fail
- 问题（若有）：<问题描述> | 严重度：high/medium/low | 修复：<具体可执行的修复指令>
- 评分：X/5

【整体评分】X/5

以下是该 PDF 各页的文本内容（仅供参考）：

` + t.buildTextContent(batchFiles, textByFile)

	msg := &schema.Message{
		Role: schema.User,
		UserInputMultiContent: []schema.MessageInputPart{
			{
				Type: schema.ChatMessagePartTypeText,
				Text: batchPrompt,
			},
			{
				Type: schema.ChatMessagePartTypeImageURL,
				Image: &schema.MessageInputImage{
					MessagePartCommon: schema.MessagePartCommon{
						URL:      strPtr("data:application/pdf;base64," + pdfBase64),
						MIMEType: "application/pdf",
					},
					Detail: "",
				},
			},
		},
	}

	resp, err := model.Generate(ctx, []*schema.Message{msg})
	if err != nil {
		return nil, err
	}

	// 临时失败时重试一次（空响应或非常短的响应）。
	report := strings.TrimSpace(resp.Content)
	if report == "" || len(report) < 20 {
		logger.Warn("qa_batch_retry", "batch_idx", batchIdx, "reason", "empty_or_short_response")
		resp, err = model.Generate(ctx, []*schema.Message{msg})
		if err != nil {
			return nil, err
		}
		report = strings.TrimSpace(resp.Content)
	}
	if report == "" {
		return &generic.QAResult{
			Reports: []string{fmt.Sprintf("%s.pdf|（LLM 返回为空）", pageRange)},
			Summary: "LLM 返回为空",
		}, nil
	}

	reportFileName := fmt.Sprintf("%s.pdf", pageRange)
	hasHigh := strings.Contains(report, "high")
	hasMedium := strings.Contains(report, "medium")
	score := parseScore(report)

	var hasIssues bool
	if score > 0 {
		hasIssues = score <= 3 || hasHigh
	} else {
		hasIssues = hasHigh || hasMedium
	}

	result := &generic.QAResult{
		Reports:      []string{reportFileName + "|" + report},
		HasIssues:    hasIssues,
		HasHighIssue: hasHigh,
		Score:        score,
	}

	if hasHigh {
		metrics.RecordQAIssue("high")
	}
	if hasMedium {
		metrics.RecordQAIssue("medium")
	}
	if hasIssues && !hasHigh && !hasMedium {
		metrics.RecordQAIssue("low")
	}
	if score > 0 {
		metrics.RecordSlideScore(float64(score), "batch_pdf")
	}
	if meta := agentutils.RuntimeMetaFromContext(ctx); meta != nil {
		low := 0
		if hasIssues && !hasHigh && !hasMedium {
			low = 1
		}
		meta.RecordQAIssues(boolToInt(hasHigh), boolToInt(hasMedium), low)
	}

	return result, nil
}

func strPtr(s string) *string { return &s }

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

// buildTextContent 从 textByFile 构建批次的文本内容摘要。
func (t *BatchPDFTool) buildTextContent(batchFiles []string, textByFile map[string]string) string {
	var parts []string
	for _, f := range batchFiles {
		stem := filepath.Base(f)
		stem = strings.TrimSuffix(stem, ".pptx")
		if text, ok := textByFile[f]; ok && text != "" {
			parts = append(parts, fmt.Sprintf("[%s]\n%s", stem, text))
		}
	}
	return strings.Join(parts, "\n\n")
}

// InvokableRun 实现 tool.InvokableTool 接口。
func (t *BatchPDFTool) InvokableRun(ctx context.Context, input string, _ ...tool.Option) (string, error) {
	wd, ok := params.GetTypedContextParams[string](ctx, params.WorkDirSessionKey)
	if !ok || wd == "" {
		type workDirGetter interface{ GetWorkDir(context.Context) string }
		if getter, ok := t.op.(workDirGetter); ok {
			wd = getter.GetWorkDir(ctx)
		}
	}
	if wd == "" {
		return "", fmt.Errorf("无法获取工作目录")
	}

	batchIdx := 0
	if input != "" {
		var args struct {
			BatchIdx int `json:"batch_idx"`
		}
		if err := json.Unmarshal([]byte(input), &args); err == nil {
			batchIdx = args.BatchIdx
		}
	}

	model, err := t.getOrCreateModelPDF(ctx)
	if err != nil {
		return "", fmt.Errorf("创建 QA 模型失败: %v", err)
	}

	result, err := t.doBatchVisualQA(ctx, model, wd, batchIdx)
	if err != nil {
		return "", err
	}

	b, _ := json.Marshal(result)
	return string(b), nil
}

func (t *BatchPDFTool) getOrCreateModelPDF(ctx context.Context) (model.ToolCallingChatModel, error) {
	t.cachedModelMu.Lock()
	defer t.cachedModelMu.Unlock()

	// 如果存在缓存的模型且未处于错误状态，检查 TTL。
	if t.cachedModel != nil && t.cachedModelErr == nil {
		if time.Since(t.cachedModelAt) < modelCacheTTL {
			return t.cachedModel, nil
		}
	}

	// 缓存过期或出错——清除陈旧模型后再重试，避免模型工厂持续失败时的紧密重试循环。
	if t.cachedModel != nil && t.cachedModelErr != nil {
		t.cachedModel = nil
		t.cachedModelErr = nil
	}

	t.cachedModel, t.cachedModelErr = t.modelFn(ctx)
	if t.cachedModelErr == nil {
		t.cachedModelAt = time.Now()
	}
	return t.cachedModel, t.cachedModelErr
}
