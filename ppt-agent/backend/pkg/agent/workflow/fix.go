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
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/logger"
)

// buildFixModelAndTools creates the Fix model and tools node.
func buildFixModelAndTools(ctx context.Context, state *workflowState) (
	model.BaseChatModel, *compose.ToolsNode, error,
) {
	cm, err := agentutils.NewFallbackToolCallingChatModel(ctx,
		agentutils.WithMaxTokens(32768),
		agentutils.WithTemperature(0),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create fix model: %w", err)
	}

	toolsCfg := compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{
			&pythonRunnerTool{state: state},
			&readFileTool{state: state},
		},
	}
	toolsNode, err := compose.NewToolNode(ctx, &toolsCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("create fix tools node: %w", err)
	}

	return cm, toolsNode, nil
}

// fixPreHandler prepares messages for the Fix LLM.
func fixPreHandler(state *workflowState) compose.StatePreHandler[[]*schema.Message, *workflowState] {
	return func(ctx context.Context, input []*schema.Message, s *workflowState) ([]*schema.Message, error) {
		s.mu.Lock()
		defer s.mu.Unlock()

		// Re-read manifest.
		manifest, err := ReadTasksManifest(s.WorkDir)
		if err == nil && manifest != nil {
			s.Manifest = manifest
		}

		prompt := buildFixPrompt(s)

		msgs := make([]*schema.Message, 0, 2)
		msgs = append(msgs, schema.SystemMessage(prompt))
		for _, m := range input {
			msgs = append(msgs, m)
		}
		return msgs, nil
	}
}

func buildFixPrompt(s *workflowState) string {
	tmplDir := s.SkillsDir + "/visual_designer/templates"

	var sb strings.Builder
	sb.WriteString("你是一个PPT修复专家。请根据QA发现的问题定点修复幻灯片。\n\n")
	sb.WriteString("## 工作目录\n")
	sb.WriteString(s.WorkDir + "\n\n")
	sb.WriteString("## 模板目录\n")
	sb.WriteString(tmplDir + "\n\n")
	sb.WriteString("## 可用工具\n")
	sb.WriteString("- read_file: 读取模板或文件\n")
	sb.WriteString("- python3: 执行Python脚本修复PPTX文件\n\n")

	if s.Manifest != nil {
		issues := s.Manifest.NeedsFix()
		if len(issues) > 0 {
			sb.WriteString(fmt.Sprintf("## 需要修复的幻灯片（共%d页）\n", len(issues)))
			for _, t := range issues {
				sb.WriteString(fmt.Sprintf("- 第%d页: %s (%s)\n", t.PageIndex, t.Title, t.ContentType))
				if t.QAReport != "" {
					sb.WriteString(fmt.Sprintf("  问题: %s\n", t.QAReport))
				}
			}
			sb.WriteString("\n请先读取模板了解设计意图，然后使用python3定点修复这些问题。\n")
			sb.WriteString("修复后请更新 tasks.json 状态为 fixed。\n")
		}
	}

	if s.StyleContext != "" {
		sb.WriteString("\n## 用户风格偏好\n")
		sb.WriteString(s.StyleContext)
		sb.WriteString("\n")
	}

	return sb.String()
}

// fixPostHandler processes Fix output.
func fixPostHandler(_ *workflowState) compose.StatePostHandler[*schema.Message, *workflowState] {
	return func(ctx context.Context, output *schema.Message, s *workflowState) (*schema.Message, error) {
		s.mu.Lock()
		defer s.mu.Unlock()

		if output != nil && output.Content != "" {
			s.OutputMessages = append(s.OutputMessages, output.Content)
		}
		return output, nil
	}
}

// fixNode processes Fix results and updates manifest.
func fixNode(state *workflowState) func(ctx context.Context, in []*schema.Message) ([]*schema.Message, error) {
	return func(ctx context.Context, in []*schema.Message) ([]*schema.Message, error) {
		state.mu.Lock()
		manifest := state.Manifest
		state.mu.Unlock()

		if manifest == nil {
			return in, nil
		}

		// Re-read manifest.
		updated, err := ReadTasksManifest(state.WorkDir)
		if err == nil && updated != nil {
			state.mu.Lock()
			state.Manifest = updated
			manifest = updated
			state.mu.Unlock()
		}

		// Update status for tasks that needed fixing.
		now := time.Now().Format(time.RFC3339)
		changed := false
		for _, t := range manifest.Tasks {
			if t.Status == StatusQADone && t.QAReport != "" {
				// Increment fix attempts.
				t.FixAttempts++
				t.Status = StatusFixed
				t.CreatedAt = now
				changed = true
			}
		}

		if changed {
			WriteTasksManifest(state.WorkDir, manifest)
			state.mu.Lock()
			state.Manifest = manifest
			state.mu.Unlock()
		}

		logger.Info("fix_node_done")
		return in, nil
	}
}
