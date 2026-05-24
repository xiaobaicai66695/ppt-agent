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
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/generic"
	"github.com/cloudwego/ppt-agent/pkg/metrics"
	"github.com/cloudwego/ppt-agent/pkg/params"
	"github.com/cloudwego/ppt-agent/pkg/tools/pythonutil"
)

var batchPDFToolInfo = &schema.ToolInfo{
	Name: "batch_pdf_review",
	Desc: `批量 PDF 视觉质量审查。将所有幻灯片合并后的 PDF 一次性发给视觉 AI 模型审查。

该工具会：
1. 将 workDir 下的所有 PPTX 合并并转换为单个 PDF
2. 将 PDF 作为附件发给多模态视觉模型，一次性审查所有页
3. 返回按页分组的 QA 报告，包含问题描述和具体修复建议

推荐优先使用此工具，一次 LLM 调用替代逐页审查的多次调用。`,
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
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

type conversionCacheEntry struct {
	Result    map[string]any
	ExpiredAt time.Time
}

// NewBatchPDFTool 创建一个批量 PDF QA Tool 实例。
func NewBatchPDFTool(op commandline.Operator, modelFn func(ctx context.Context) (model.ToolCallingChatModel, error)) tool.InvokableTool {
	return &BatchPDFTool{op: op, modelFn: modelFn}
}

func (t *BatchPDFTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return batchPDFToolInfo, nil
}

// runPDFConverter generates merged PDF for the workDir using the converter script.
// Caches the result to avoid redundant conversion on concurrent calls.
func (t *BatchPDFTool) runPDFConverter(ctx context.Context, wd string) (string, string, error) {
	t.conversionCacheMu.Lock()
	if t.conversionCache == nil {
		t.conversionCache = make(map[string]*conversionCacheEntry)
	}
	entry, ok := t.conversionCache[wd]
	if ok && time.Now().Before(entry.ExpiredAt) {
		t.conversionCacheMu.Unlock()
		pdfPath, _ := entry.Result["pdf_path"].(string)
		textContent, _ := entry.Result["text_content"].(string)
		return pdfPath, textContent, nil
	}
	t.conversionCacheMu.Unlock()

	// Find converter script using shared utility
	converter := pythonutil.FindConverterPy(wd)
	if converter == "" {
		return "", "", fmt.Errorf("找不到 pptx_qa_converter.py")
	}

	// Check if PDF already exists
	existingPDF := filepath.Join(wd, "merged.pdf")
	if _, err := os.Stat(existingPDF); err == nil {
		textContent := t.extractTextFromPPTXFiles(wd)
		t.conversionCacheMu.Lock()
		t.conversionCache[wd] = &conversionCacheEntry{
			Result:    map[string]any{"pdf_path": existingPDF, "text_content": textContent},
			ExpiredAt: time.Now().Add(conversionCacheTTL),
		}
		t.conversionCacheMu.Unlock()
		return existingPDF, textContent, nil
	}

	pythonBin := pythonutil.GetPythonBinary()
	cmdArgs := []string{
		converter,
		"--pptx-dir", wd,
		"--output-dir", wd,
		"--pdf-only",
	}
	cmd := exec.CommandContext(ctx, pythonBin, cmdArgs...)
	cmd.Dir = wd

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("PDF 生成失败: %v, stderr: %s", err, stderr.String())
	}

	var result map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return "", "", fmt.Errorf("解析 PDF 生成结果失败: %v", err)
	}

	if errMsg, ok := result["error"].(string); ok && errMsg != "" {
		return "", "", fmt.Errorf("%s", errMsg)
	}

	pdfPath, _ := result["pdf_path"].(string)
	textContent, _ := result["text_content"].(string)

	t.conversionCacheMu.Lock()
	t.conversionCache[wd] = &conversionCacheEntry{
		Result:    result,
		ExpiredAt: time.Now().Add(conversionCacheTTL),
	}
	t.conversionCacheMu.Unlock()

	return pdfPath, textContent, nil
}

// extractTextFromPPTXFiles extracts text content from all PPTX files in the directory.
// Uses a single Python subprocess call for efficiency (avoids N process spawns for N files).
func (t *BatchPDFTool) extractTextFromPPTXFiles(wd string) string {
	entries, err := os.ReadDir(wd)
	if err != nil {
		return ""
	}

	var pptxFiles []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".pptx") {
			continue
		}
		pptxFiles = append(pptxFiles, filepath.Join(wd, e.Name()))
	}

	if len(pptxFiles) == 0 {
		return ""
	}

	// Build a single Python script that processes all files at once
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
	cmd.Dir = wd
	out, _ := cmd.Output()
	if len(out) == 0 {
		return ""
	}

	var results []struct {
		Path string `json:"path"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(out, &results); err != nil {
		return ""
	}

	var allText []string
	for _, r := range results {
		if r.Text == "" {
			continue
		}
		stem := filepath.Base(r.Path)
		stem = strings.TrimSuffix(stem, ".pptx")
		allText = append(allText, fmt.Sprintf("[%s]\n%s", stem, r.Text))
	}
	return strings.Join(allText, "\n\n")
}

// parseScore extracts the 1-5 score from the QA report text.
// Uses LastIndex to find the final score line, avoiding false matches
// from scoring criteria quoted in fix suggestions.
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

// doBatchVisualQA sends the merged PDF as a base64 data URI to the multimodal LLM
// for a single-shot review of all slides.
func (t *BatchPDFTool) doBatchVisualQA(ctx context.Context, model model.ToolCallingChatModel, wd string) (*generic.QAResult, error) {
	pdfPath, textContent, err := t.runPDFConverter(ctx, wd)
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

` + textContent

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

	report := strings.TrimSpace(resp.Content)
	if report == "" {
		return &generic.QAResult{
			Reports: []string{"merged.pdf|（LLM 返回为空）"},
			Summary: "LLM 返回为空",
		}, nil
	}

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
		Reports:      []string{"merged.pdf|" + report},
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

	return result, nil
}

func strPtr(s string) *string { return &s }

// InvokableRun implements tool.InvokableTool for BatchPDFTool.
func (t *BatchPDFTool) InvokableRun(ctx context.Context, _ string, _ ...tool.Option) (string, error) {
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

	model, err := t.getOrCreateModelPDF(ctx)
	if err != nil {
		return "", fmt.Errorf("创建 QA 模型失败: %v", err)
	}

	result, err := t.doBatchVisualQA(ctx, model, wd)
	if err != nil {
		return "", err
	}

	b, _ := json.Marshal(result)
	return string(b), nil
}

func (t *BatchPDFTool) getOrCreateModelPDF(ctx context.Context) (model.ToolCallingChatModel, error) {
	t.cachedModelMu.Lock()
	defer t.cachedModelMu.Unlock()

	// If a cached model exists and is not in an error state, check TTL.
	if t.cachedModel != nil && t.cachedModelErr == nil {
		if time.Since(t.cachedModelAt) < modelCacheTTL {
			return t.cachedModel, nil
		}
	}

	// Cache expired or errored — clear stale model before retrying to avoid
	// tight retry loops when the model factory persistently fails.
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
