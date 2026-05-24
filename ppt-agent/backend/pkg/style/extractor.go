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

package style

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/prompts"
	"github.com/cloudwego/ppt-agent/pkg/tools/pythonutil"
)

// Extractor 从 PPTX 文本内容中提取用户风格偏好，核心逻辑由 LLM 完成。
// 相比纯规则匹配（ExtractFromTasks / ExtractFromQuery），能更深入地理解内容风格。
type Extractor struct {
	modelFactory func(ctx context.Context) (model.ToolCallingChatModel, error)
}

// NewExtractor 创建一个风格提取器。
// modelFactory 用于创建 LLM 实例（每次调用独立创建，支持上下文复用）。
func NewExtractor(modelFactory func(ctx context.Context) (model.ToolCallingChatModel, error)) *Extractor {
	return &Extractor{modelFactory: modelFactory}
}

type llmExtractionResult struct {
	Themes            []string `json:"themes"`
	Colors            []string `json:"colors"`
	ContentPatterns   []string `json:"content_patterns"`
	LayoutPreferences []string `json:"layout_preferences"`
	LanguageTone     string   `json:"language_tone"`
	TypicalPageCount int      `json:"typical_page_count"`
	SpecialNotes      []string `json:"special_notes"`
}

// ExtractFromPPTX 分析 workDir 中所有 PPTX 文件，提取用户风格偏好。
// 核心由 LLM 完成。
func (e *Extractor) ExtractFromPPTX(ctx context.Context, workDir, query, theme string, tasks []*TaskItemInfo) (*ExtractedStyle, error) {
	// Step 1: 从所有 PPTX 中提取文本
	textContent, totalSlides, err := extractPPTXText(workDir)
	if err != nil {
		logger.Warn("style_extractor_text_failed", "error", err.Error())
		return ExtractFromTasks(tasks, theme), nil
	}

	if textContent == "" || totalSlides == 0 {
		logger.Warn("style_extractor_no_content", "workdir", workDir)
		return ExtractFromTasks(tasks, theme), nil
	}

	// Step 2: 构建 content_type 摘要
	var contentTypes []string
	for _, t := range tasks {
		if t.ContentType != "" {
			contentTypes = append(contentTypes, t.ContentType)
		}
	}
	contentTypeStr := strings.Join(contentTypes, "、")
	if contentTypeStr == "" {
		contentTypeStr = "未知"
	}

	// Step 3: 加载提示词模板
	sysPrompt, userPrompt, err := prompts.RenderStyleExtraction("system.tmpl", "user.tmpl", &prompts.StyleExtractionData{
		UserQuery:    query,
		Theme:        theme,
		PageCount:    totalSlides,
		ContentTypes: contentTypeStr,
		TextContent:  textContent,
	})
	if err != nil {
		logger.Warn("style_extractor_render_failed", "error", err.Error())
		return ExtractFromTasks(tasks, theme), nil
	}

	// Step 4: 调用 LLM
	m, err := e.modelFactory(ctx)
	if err != nil {
		logger.Warn("style_extractor_model_failed", "error", err.Error())
		return ExtractFromTasks(tasks, theme), nil
	}

	resp, err := m.Generate(ctx, []*schema.Message{
		{Role: schema.System, Content: sysPrompt},
		{Role: schema.User, Content: userPrompt},
	})
	if err != nil {
		logger.Warn("style_extractor_llm_failed", "error", err.Error())
		return ExtractFromTasks(tasks, theme), nil
	}

	// Step 5: 解析 LLM 输出
	result := &llmExtractionResult{}
	if err := parseLLMJSONResult(resp.Content, result); err != nil {
		logger.Warn("style_extractor_parse_failed", "content", truncate(resp.Content, 200), "error", err.Error())
		return ExtractFromTasks(tasks, theme), nil
	}

	es := &ExtractedStyle{
		Themes:            result.Themes,
		Colors:            result.Colors,
		ContentPatterns:   result.ContentPatterns,
		LayoutPreferences: result.LayoutPreferences,
		LanguageTone:      result.LanguageTone,
		PageCount:         result.TypicalPageCount,
		SpecialNotes:      result.SpecialNotes,
	}

	if es.PageCount == 0 && totalSlides > 0 {
		es.PageCount = totalSlides
	}

	logger.Info("style_extracted_by_llm",
		"workdir", workDir,
		"themes", strings.Join(es.Themes, ","),
		"language_tone", es.LanguageTone,
		"page_count", es.PageCount)

	return es, nil
}

// extractPPTXText 从 workDir 中的所有 PPTX 文件提取纯文本。
// 通过调用 python-pptx 读取每个 PPTX，提取所有 shape 的文本内容。
func extractPPTXText(workDir string) (string, int, error) {
	pythonBin := pythonutil.GetPythonBinary()

	script := `
import sys
import glob, os, json
from pptx import Presentation

work_dir = sys.argv[1]
all_text = []
total = 0

for pptx_file in sorted(glob.glob(os.path.join(work_dir, "*.pptx"))):
    try:
        prs = Presentation(pptx_file)
        total += len(prs.slides)
        for slide in prs.slides:
            parts = []
            for shape in slide.shapes:
                if hasattr(shape, "text") and shape.text.strip():
                    parts.append(shape.text.strip())
            if parts:
                all_text.append("\n".join(parts))
    except Exception:
        pass

result = {"text": "\n\n".join(all_text), "total": total}
print(json.dumps(result, ensure_ascii=False))
`

	cmd := exec.Command(pythonBin, "-c", script, workDir)
	output, err := cmd.Output()
	if err != nil {
		return "", 0, fmt.Errorf("python extraction failed: %w", err)
	}

	var result struct {
		Text  string `json:"text"`
		Total int    `json:"total"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return "", 0, fmt.Errorf("parse extraction output: %w", err)
	}
	return result.Text, result.Total, nil
}

func parseLLMJSONResult(content string, result *llmExtractionResult) error {
	content = strings.TrimSpace(content)
	// 尝试从 markdown 代码块中提取 JSON
	if idx := strings.Index(content, "```"); idx >= 0 {
		start := idx + 3
		if strings.HasPrefix(content[start:], "json") {
			start += 4
		}
		end := strings.Index(content[start:], "```")
		if end >= 0 {
			content = content[start : start+end]
		} else {
			content = content[start:]
		}
	}
	content = strings.TrimSpace(content)
	return json.Unmarshal([]byte(content), result)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
