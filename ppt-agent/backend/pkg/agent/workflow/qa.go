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

package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/logger"
)

// buildQAModelAndTools creates the QA model and tools node.
func buildQAModelAndTools(ctx context.Context, state *workflowState) (
	model.BaseChatModel, *compose.ToolsNode, error,
) {
	if state.QAModelFn == nil {
		return nil, nil, fmt.Errorf("QA model function not provided")
	}

	cm, err := state.QAModelFn(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("create QA model: %w", err)
	}

	toolsCfg := compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{
			&qaPDFTool{state: state},
			&readFileTool{state: state},
		},
	}
	toolsNode, err := compose.NewToolNode(ctx, &toolsCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("create QA tools node: %w", err)
	}

	return cm, toolsNode, nil
}

// qaPDFTool converts PPTX to PDF then JPEG for visual QA.
type qaPDFTool struct {
	state *workflowState
}

func (t *qaPDFTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "batch_pdf_qa",
		Desc: "批量将PPTX文件转换为PDF/JPEG以便进行视觉质检",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"file_paths": {
				Type: "string",
				Desc: "PPTX文件路径（支持多个，逗号分隔）",
			},
		}),
	}, nil
}

func (t *qaPDFTool) InvokableRun(ctx context.Context, params string, opts ...tool.Option) (string, error) {
	var args map[string]any
	if err := json.Unmarshal([]byte(params), &args); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	pathsStr, _ := args["file_paths"].(string)
	if pathsStr == "" {
		return "错误: file_paths参数缺失", nil
	}

	paths := strings.Split(pathsStr, ",")
	var results []string

	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		// Resolve path.
		if !filepath.IsAbs(p) {
			p = filepath.Join(t.state.WorkDir, p)
		}

		// Check if file exists.
		if _, err := os.Stat(p); os.IsNotExist(err) {
			results = append(results, fmt.Sprintf("文件不存在: %s", p))
			continue
		}

		// Convert to JPEG for QA review.
		jpegPath, err := convertPPTXToJPEG(p)
		if err != nil {
			results = append(results, fmt.Sprintf("转换失败 %s: %v", p, err))
			continue
		}
		results = append(results, fmt.Sprintf("已转换: %s → %s", filepath.Base(p), jpegPath))
	}

	return strings.Join(results, "\n"), nil
}

// convertPPTXToJPEG converts a PPTX file to JPEG images using LibreOffice.
func convertPPTXToJPEG(pptxPath string) (string, error) {
	dir := filepath.Dir(pptxPath)
	base := strings.TrimSuffix(filepath.Base(pptxPath), ".pptx")
	outputDir := filepath.Join(dir, base+"_qa")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("mkdir output dir: %w", err)
	}

	// Try LibreOffice first.
	var cmd *exec.Cmd
	loPath := os.Getenv("LIBREOFFICE_PATH")
	if loPath != "" {
		cmd = exec.Command(loPath, "--headless", "--convert-to", "jpg", "--outdir", outputDir, pptxPath)
	} else {
		cmd = exec.Command("libreoffice", "--headless", "--convert-to", "jpg", "--outdir", outputDir, pptxPath)
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd = exec.CommandContext(timeoutCtx, cmd.Path, cmd.Args[1:]...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("libreoffice convert: %w (output: %s)", err, string(out))
	}

	// Return path to first jpeg.
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return "", fmt.Errorf("read output dir: %w", err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".jpg") || strings.HasSuffix(e.Name(), ".jpeg") {
			return filepath.Join(outputDir, e.Name()), nil
		}
	}
	return outputDir, nil
}

// qaPreHandler prepares messages for the QA LLM.
func qaPreHandler(state *workflowState) compose.StatePreHandler[[]*schema.Message, *workflowState] {
	return func(ctx context.Context, input []*schema.Message, s *workflowState) ([]*schema.Message, error) {
		s.mu.Lock()
		defer s.mu.Unlock()

		// Re-read manifest.
		manifest, err := ReadTasksManifest(s.WorkDir)
		if err == nil && manifest != nil {
			s.Manifest = manifest
		}

		prompt := buildQAPrompt(s)

		msgs := make([]*schema.Message, 0, 2)
		msgs = append(msgs, schema.SystemMessage(prompt))
		for _, m := range input {
			msgs = append(msgs, m)
		}
		return msgs, nil
	}
}

func buildQAPrompt(s *workflowState) string {
	var sb strings.Builder
	sb.WriteString("你是一个视觉质量审查专家。请审查已生成的PPT幻灯片是否存在以下问题：\n\n")
	sb.WriteString("## 质量问题检查项\n")
	sb.WriteString("1. 文字溢出：文本是否超出边界\n")
	sb.WriteString("2. 元素重叠：各元素是否重叠\n")
	sb.WriteString("3. 对比度：文字与背景对比度是否足够\n")
	sb.WriteString("4. 字体大小：字号是否合适（标题≥28pt，正文≥18pt）\n")
	sb.WriteString("5. 图片质量：图片是否清晰、无压缩痕迹\n")
	sb.WriteString("6. 排版对齐：元素是否对齐\n\n")

	sb.WriteString("## 工作目录\n")
	sb.WriteString(s.WorkDir + "\n\n")

	// List done tasks.
	if s.Manifest != nil {
		var doneTasks []*TaskItem
		for _, t := range s.Manifest.Tasks {
			if t.Status == StatusDone {
				doneTasks = append(doneTasks, t)
			}
		}
		if len(doneTasks) > 0 {
			sb.WriteString(fmt.Sprintf("## 已完成的幻灯片（共%d页）\n", len(doneTasks)))
			for _, t := range doneTasks {
				sb.WriteString(fmt.Sprintf("- 第%d页: %s (%s) → %s\n",
					t.PageIndex, t.Title, t.ContentType, t.OutputFile))
			}
			sb.WriteString("\n请使用 batch_pdf_qa 工具将这些PPTX转换为JPEG图片进行审查。\n")
		}
	}

	sb.WriteString("\n## 输出格式\n")
	sb.WriteString("审查完成后，请按以下JSON格式输出结果：\n")
	sb.WriteString("{\"needs_fix\": true/false, \"issues\": [{\"page\": 1, \"problem\": \"...\", \"severity\": \"high/medium/low\"}]}\n")
	sb.WriteString("如果所有页面都没有问题，设置 needs_fix: false。\n")

	return sb.String()
}

// qaPostHandler processes QA output and updates manifest.
func qaPostHandler(_ *workflowState) compose.StatePostHandler[*schema.Message, *workflowState] {
	return func(ctx context.Context, output *schema.Message, s *workflowState) (*schema.Message, error) {
		s.mu.Lock()
		defer s.mu.Unlock()

		if output != nil && output.Content != "" {
			s.OutputMessages = append(s.OutputMessages, output.Content)
		}
		return output, nil
	}
}

// qaNode processes QA results and updates manifest with issues.
func qaNode(state *workflowState) func(ctx context.Context, in []*schema.Message) ([]*schema.Message, error) {
	return func(ctx context.Context, in []*schema.Message) ([]*schema.Message, error) {
		state.mu.Lock()
		manifest := state.Manifest
		state.mu.Unlock()

		if manifest == nil {
			return in, nil
		}

		// Re-read manifest to pick up any changes.
		updated, err := ReadTasksManifest(state.WorkDir)
		if err == nil && updated != nil {
			state.mu.Lock()
			state.Manifest = updated
			manifest = updated
			state.mu.Unlock()
		}

		// Parse QA issues from the last LLM output.
		var qaResult struct {
			NeedsFix bool `json:"needs_fix"`
			Issues   []struct {
				Page     int    `json:"page"`
				Problem  string `json:"problem"`
				Severity string `json:"severity"`
			} `json:"issues"`
		}

		for _, msg := range state.OutputMessages {
			if strings.Contains(msg, "needs_fix") || strings.Contains(msg, "issues") {
				// Try to extract JSON.
				start := strings.Index(msg, "{")
				end := strings.LastIndex(msg, "}")
				if start >= 0 && end > start {
					sub := msg[start : end+1]
					if err2 := json.Unmarshal([]byte(sub), &qaResult); err2 == nil {
						break
					}
				}
			}
		}

		if qaResult.NeedsFix && len(qaResult.Issues) > 0 {
			// Map issues to tasks.
			issueMap := make(map[int]string)
			for _, issue := range qaResult.Issues {
				if issue.Page > 0 && issue.Page <= len(manifest.Tasks) {
					issueMap[issue.Page] = fmt.Sprintf("[%s] %s", issue.Severity, issue.Problem)
				}
			}

			now := time.Now().Format(time.RFC3339)
			for i, t := range manifest.Tasks {
				if issue, ok := issueMap[t.PageIndex]; ok {
					if manifest.Tasks[i].QAReport == "" {
						manifest.Tasks[i].QAReport = issue
					} else {
						manifest.Tasks[i].QAReport += "\n" + issue
					}
					manifest.Tasks[i].Status = StatusQADone
					manifest.Tasks[i].CreatedAt = now
				}
			}

			WriteTasksManifest(state.WorkDir, manifest)
			state.mu.Lock()
			state.Manifest = manifest
			state.mu.Unlock()

			logger.Info("qa_issues_found", "count", len(qaResult.Issues))
		} else {
			// All pages pass — mark them as QA done.
			now := time.Now().Format(time.RFC3339)
			changed := false
			for i, t := range manifest.Tasks {
				if t.Status == StatusDone {
					manifest.Tasks[i].Status = StatusQADone
					manifest.Tasks[i].QAReport = ""
					manifest.Tasks[i].CreatedAt = now
					changed = true
				}
			}
			if changed {
				WriteTasksManifest(state.WorkDir, manifest)
				state.mu.Lock()
				state.Manifest = manifest
				state.mu.Unlock()
			}
			logger.Info("qa_all_pass")
		}

		return in, nil
	}
}

