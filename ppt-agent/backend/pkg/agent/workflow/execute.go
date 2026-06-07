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
	"regexp"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/logger"
)

// buildExecutorModelAndTools creates the executor LLM and tools node.
func buildExecutorModelAndTools(ctx context.Context, state *workflowState) (
	model.ToolCallingChatModel, *compose.ToolsNode, error,
) {
	cm, err := agentutils.NewFallbackToolCallingChatModel(ctx,
		agentutils.WithMaxTokens(32768),
		agentutils.WithTemperature(0),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create executor model: %w", err)
	}

	compressor, err := agentutils.NewFallbackToolCallingChatModel(ctx,
		agentutils.WithTextModel(),
		agentutils.WithMaxTokens(4096),
		agentutils.WithTemperature(0),
	)
	if err == nil {
		cm = agentutils.NewChatModelCompressor(cm, compressor,
			agentutils.WithCompressThreshold(60),
			agentutils.WithTokenThreshold(200000),
			agentutils.WithPreserveCount(8),
		)
	}

	toolsCfg := compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{
			newPythonRunnerTool(state),
			newReadFileTool(state),
			newSearchTool(),
		},
	}
	toolsNode, err := compose.NewToolNode(ctx, &toolsCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("create executor tools node: %w", err)
	}

	return cm, toolsNode, nil
}

// execPreHandler prepares messages for the executor LLM.
func execPreHandler(state *workflowState) compose.StatePreHandler[[]*schema.Message, *workflowState] {
	return func(ctx context.Context, input []*schema.Message, s *workflowState) ([]*schema.Message, error) {
		s.mu.Lock()
		defer s.mu.Unlock()

		manifest, err := ReadTasksManifest(s.WorkDir)
		if err == nil && manifest != nil {
			s.Manifest = manifest
		}

		prompt := buildExecutorPrompt(s)

		msgs := make([]*schema.Message, 0, 2)
		msgs = append(msgs, schema.SystemMessage(prompt))
		for _, m := range input {
			msgs = append(msgs, m)
		}
		return msgs, nil
	}
}

func buildExecutorPrompt(s *workflowState) string {
	tasksJSON := filepath.Join(s.WorkDir, "tasks.json")
	tmplDir := filepath.Join(s.SkillsDir, "visual_designer", "templates", "full-decks")

	var pendingTasks []*TaskItem
	if s.Manifest != nil {
		for _, t := range s.Manifest.Tasks {
			if t.Status == StatusPending || t.Status == StatusGenerating {
				pendingTasks = append(pendingTasks, t)
				if len(pendingTasks) >= s.Concurrency {
					break
				}
			}
		}
	}

	var sb strings.Builder
	sb.WriteString("你是一个PPT幻灯片生成专家。请读取任务清单并生成幻灯片。\n\n")
	sb.WriteString("## 任务清单\n")
	sb.WriteString("tasks.json路径: " + tasksJSON + "\n")
	sb.WriteString("模板目录: " + tmplDir + "\n\n")
	sb.WriteString("## 可用工具\n")
	sb.WriteString("- read_file: 读取模板或任务清单文件\n")
	sb.WriteString("- python3: 执行Python脚本生成PPTX文件\n")
	sb.WriteString("- search: 搜索网络获取真实信息\n\n")

	if len(pendingTasks) > 0 {
		sb.WriteString(fmt.Sprintf("## 本次生成任务（最多%d页并发）\n", s.Concurrency))
		for _, t := range pendingTasks {
			sb.WriteString(fmt.Sprintf("- 第%d页: %s (%s)\n", t.PageIndex, t.Title, t.ContentType))
		}
	} else {
		sb.WriteString("## 当前没有待生成的幻灯片，任务已完成。\n")
	}

	sb.WriteString("\n## 执行步骤\n")
	sb.WriteString("1. 读取 tasks.json 了解所有页面\n")
	sb.WriteString("2. 为待生成页面读取对应模板\n")
	sb.WriteString("3. 使用 python3 执行模板生成 PPT 文件\n")
	sb.WriteString("4. 更新 tasks.json 中的状态为 done\n")
	sb.WriteString("5. 完成后输出 'DONE'\n\n")

	if s.StyleContext != "" {
		sb.WriteString("## 用户风格偏好\n")
		sb.WriteString(s.StyleContext)
		sb.WriteString("\n")
	}

	return sb.String()
}

// execPostHandler processes executor output.
func execPostHandler(_ *workflowState) compose.StatePostHandler[*schema.Message, *workflowState] {
	return func(ctx context.Context, output *schema.Message, s *workflowState) (*schema.Message, error) {
		s.mu.Lock()
		defer s.mu.Unlock()
		if output != nil && output.Content != "" {
			s.OutputMessages = append(s.OutputMessages, output.Content)
		}
		return output, nil
	}
}

// executeNode updates task status based on generated files.
func executeNode(state *workflowState) func(ctx context.Context, in []*schema.Message) ([]*schema.Message, error) {
	return func(ctx context.Context, in []*schema.Message) ([]*schema.Message, error) {
		state.mu.Lock()
		manifest := state.Manifest
		state.mu.Unlock()

		if manifest == nil {
			return in, nil
		}

		updated, err := ReadTasksManifest(state.WorkDir)
		if err == nil && updated != nil {
			state.mu.Lock()
			state.Manifest = updated
			manifest = updated
			state.mu.Unlock()
		}

		entries, err := os.ReadDir(state.WorkDir)
		if err == nil {
			now := time.Now().Format(time.RFC3339)
			for _, entry := range entries {
				if entry.IsDir() || filepath.Ext(entry.Name()) != ".pptx" {
					continue
				}
				for _, t := range manifest.Tasks {
					if t.OutputFile == entry.Name() && (t.Status == StatusPending || t.Status == StatusGenerating) {
						t.Status = StatusDone
						t.CreatedAt = now
					}
				}
			}
			WriteTasksManifest(state.WorkDir, manifest)
			state.mu.Lock()
			state.Manifest = manifest
			state.mu.Unlock()
		}

		logger.Info("execute_node_done", "manifest_tasks", len(manifest.Tasks))
		return in, nil
	}
}

// toolPreHandler passes tool messages through.
func toolPreHandler(_ *workflowState) compose.StatePreHandler[*schema.Message, *workflowState] {
	return func(ctx context.Context, input *schema.Message, s *workflowState) (*schema.Message, error) {
		return input, nil
	}
}

// ── Tool implementations ─────────────────────────────────────────────────────────

// pythonRunnerTool executes Python scripts to generate PPTX files.
type pythonRunnerTool struct {
	state *workflowState
}

func newPythonRunnerTool(state *workflowState) *pythonRunnerTool {
	return &pythonRunnerTool{state: state}
}

func (t *pythonRunnerTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "python3",
		Desc: "执行Python脚本生成PPTX文件。参数: script (完整的Python代码)",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"script": {
				Type: "string",
				Desc: "完整的Python脚本代码",
			},
		}),
	}, nil
}

func (t *pythonRunnerTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args map[string]any
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	script, _ := args["script"].(string)
	if script == "" {
		return "错误: script参数缺失", nil
	}

	scriptFile := filepath.Join(t.state.WorkDir, "gen_script_"+fmt.Sprintf("%d", time.Now().UnixNano())+".py")
	if err := os.WriteFile(scriptFile, []byte(script), 0644); err != nil {
		return "", fmt.Errorf("write script: %w", err)
	}
	defer os.Remove(scriptFile)

	pythonPath := os.Getenv("PYTHON_PATH")
	if pythonPath == "" {
		pythonPath = "python"
	}

	var stderr strings.Builder
	cmd := exec.Command(pythonPath, scriptFile)
	cmd.Dir = t.state.WorkDir
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	result := string(output)
	if stderr.Len() > 0 {
		result += "\nSTDERR: " + stderr.String()
	}
	if err != nil {
		result += "\nERROR: " + err.Error()
	}

	t.updateTaskStatus(t.state.WorkDir, result)
	return result, nil
}

func (t *pythonRunnerTool) updateTaskStatus(workDir, output string) {
	manifest, err := ReadTasksManifest(workDir)
	if err != nil || manifest == nil {
		return
	}
	filePattern := regexp.MustCompile(`[\w\-]+\.pptx`)
	files := filePattern.FindAllString(output, -1)
	now := time.Now().Format(time.RFC3339)
	for _, fname := range files {
		for _, task := range manifest.Tasks {
			if task.OutputFile == fname && task.Status == StatusPending {
				task.Status = StatusDone
				task.CreatedAt = now
			}
		}
	}
	WriteTasksManifest(workDir, manifest)
}

// readFileTool reads file content.
type readFileTool struct {
	state *workflowState
}

func newReadFileTool(state *workflowState) *readFileTool {
	return &readFileTool{state: state}
}

func (t *readFileTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "read_file",
		Desc: "读取文件内容。参数: path (文件路径)",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type: "string",
				Desc: "文件路径",
			},
		}),
	}, nil
}

func (t *readFileTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args map[string]any
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	path, _ := args["path"].(string)
	if path == "" {
		return "错误: path参数缺失", nil
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(t.state.WorkDir, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	return string(data), nil
}

// searchTool provides web search capabilities.
type searchTool struct{}

func newSearchTool() *searchTool { return &searchTool{} }

func (t *searchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "search",
		Desc: "搜索网络获取真实信息。参数: query (搜索关键词)",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type: "string",
				Desc: "搜索查询",
			},
		}),
	}, nil
}

func (t *searchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args map[string]any
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	query, _ := args["query"].(string)
	if query == "" {
		return "错误: query参数缺失", nil
	}
	return fmt.Sprintf("[搜索结果: 关于 %s 的相关信息。注: 生产环境请接入真实搜索API]", query), nil
}
