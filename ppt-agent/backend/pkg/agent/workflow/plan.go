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
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/logger"
)

// buildPlanModel creates the LLM for the Plan node.
func buildPlanModel(ctx context.Context, _ *workflowState) (model.BaseChatModel, error) {
	cm, err := agentutils.NewFallbackToolCallingChatModel(ctx,
		agentutils.WithMaxTokens(32768),
		agentutils.WithTemperature(0),
	)
	if err != nil {
		return nil, fmt.Errorf("create plan model: %w", err)
	}
	return cm, nil
}

// planPreHandler prepares messages for the planner LLM.
// It injects the system prompt and user query into the message history.
func planPreHandler(state *workflowState) compose.StatePreHandler[[]*schema.Message, *workflowState] {
	return func(ctx context.Context, input []*schema.Message, s *workflowState) ([]*schema.Message, error) {
		s.mu.Lock()
		defer s.mu.Unlock()

		s.UserQuery = state.UserQuery
		s.Skills = state.Skills
		s.StyleContext = state.StyleContext
		s.EnhancedProfile = state.EnhancedProfile
		s.WorkDir = state.WorkDir

		var sb strings.Builder
		sb.WriteString("你是PPT大纲规划专家。根据用户的需求，生成一份详细的PPT大纲规划。\n\n")
		sb.WriteString("## 用户需求\n")
		sb.WriteString(state.UserQuery)
		sb.WriteString("\n\n")

		if state.StyleContext != "" {
			sb.WriteString("## 用户风格偏好\n")
			sb.WriteString(state.StyleContext)
			sb.WriteString("\n\n")
		}

		sb.WriteString("请生成JSON格式的PPT大纲（直接输出JSON，不要markdown代码块）：\n")
		sb.WriteString("格式：{\"title\":\"...\",\"theme\":\"...\",\"slides\":[{\"title\":\"...\",\"content_type\":\"...\",\"description\":\"...\",\"background\":\"...\"}]}\n")
		sb.WriteString("- slides数量6-20页\n")
		sb.WriteString("- 内容充实具体\n")
		sb.WriteString("- content_type可选: title_slide, section_slide, content_slide, two_column_slide, image_content_slide, summary_slide\n")
		sb.WriteString("- theme可选: minimalist_blue, tech_blue, business_blue, nature_green, warm_orange, charcoal_light, ocean_soft\n")

		msgs := make([]*schema.Message, 0, 2)
		msgs = append(msgs, schema.SystemMessage(sb.String()))
		for _, m := range input {
			msgs = append(msgs, m)
		}
		return msgs, nil
	}
}

// planPostHandler processes the planner LLM output.
// It extracts JSON from the response and writes tasks.json.
func planPostHandler(_ *workflowState) compose.StatePostHandler[*schema.Message, *workflowState] {
	return func(ctx context.Context, output *schema.Message, s *workflowState) (*schema.Message, error) {
		s.mu.Lock()
		defer s.mu.Unlock()

		if output == nil || output.Content == "" {
			return output, nil
		}

		s.OutputMessages = append(s.OutputMessages, output.Content)

		// Try to extract JSON from the response.
		content := strings.TrimSpace(output.Content)

		// Strip markdown code fences.
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)

		// Try to find JSON in the content.
		var outline TaskOutline
		if err := json.Unmarshal([]byte(content), &outline); err != nil {
			// Try to find JSON substring.
			start := strings.Index(content, "{")
			end := strings.LastIndex(content, "}")
			if start >= 0 && end > start {
				sub := content[start : end+1]
				if err2 := json.Unmarshal([]byte(sub), &outline); err2 != nil {
					logger.Warn("plan_parse_failed", "error", err2.Error())
					return output, nil
				}
			} else {
				logger.Warn("plan_no_json", "content_len", len(content))
				return output, nil
			}
		}

		// Convert outline to manifest and write tasks.json.
		manifest := outlineToManifest(&outline, s.WorkDir)
		if err := WriteTasksManifest(s.WorkDir, manifest); err != nil {
			return output, fmt.Errorf("write tasks manifest: %w", err)
		}

		s.Manifest = manifest
		logger.Info("plan_complete", "slides", len(outline.Slides), "title", outline.Title)
		return output, nil
	}
}

// planNode is a no-op node that just passes through after planner wrote tasks.json.
func planNode(_ *workflowState) func(ctx context.Context, in []*schema.Message) ([]*schema.Message, error) {
	return func(ctx context.Context, in []*schema.Message) ([]*schema.Message, error) {
		return in, nil
	}
}

// outlineToManifest converts a TaskOutline to a TasksManifest.
func outlineToManifest(outline *TaskOutline, _ string) *TasksManifest {
	tasks := make([]*TaskItem, 0, len(outline.Slides))
	for i, slide := range outline.Slides {
		safeTitle := sanitizeFilename(slide.Title)
		item := &TaskItem{
			TaskID:      fmt.Sprintf("slide-%d", i+1),
			PageIndex:   i + 1,
			Title:       slide.Title,
			ContentType: slide.ContentType,
			Description: slide.Description,
			OutputFile:  fmt.Sprintf("%d_%s.pptx", i+1, safeTitle),
			Status:      StatusPending,
			Background:  slide.Background,
		}
		tasks = append(tasks, item)
	}
	return &TasksManifest{
		Title:    outline.Title,
		Theme:    outline.Theme,
		Template: outline.Template,
		Tasks:    tasks,
	}
}

func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_", "?", "_",
		"\"", "_", "<", "_", ">", "_", "|", "_", "\n", "_",
	)
	return replacer.Replace(name)
}
