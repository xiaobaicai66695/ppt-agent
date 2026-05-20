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

// Package prompts loads agent prompt templates from .tmpl files on disk.
// Prompt files are organized by execution mode under the prompts/ directory:
//
//	prompts/
//	├── prompts.go                # 本文件，通用加载函数
//	├── executor_user_prompt.tmpl # PlanExecute executor 用户提示词（根目录）
//	├── deep/                     # DeepAgent 模式模板
//	│   ├── master_instruction.tmpl
//	│   ├── slide_executor_instruction.tmpl
//	│   ├── reviewer_instruction.tmpl
//	│   └── fixer_instruction.tmpl
//	└── planexecute/              # PlanExecute 模式模板
//	    ├── planner_system.tmpl
//	    ├── planner_user.tmpl
//	    ├── executor_system.tmpl
//	    ├── executor_user.tmpl
//	    ├── replanner_system.tmpl
//	    └── replanner_user.tmpl
//
// 每个模板通过 Render*(data) 系列函数加载并渲染。
package prompts

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed *.tmpl
//go:embed deep/*.tmpl
//go:embed planexecute/*.tmpl
var FS embed.FS

// TemplateData holds the data fields used across prompt templates.
type TemplateData struct {
	SystemPrompt      string // Injected system-level instructions (skills, rules)
	Input            string // User's original request
	ExecutorContext  string // Current execution state summary
	Step             string // The current step/task to execute
	Skills           string // Loaded skill contents
	WorkDir          string // Working directory absolute path
	SkillsDir        string // Skills directory absolute path
	TmplDir          string // Template directory absolute path
	TasksJSON        string // tasks.json absolute path
	TemplateCatalog  string // Inline template catalog table (for deep agent)
	UserQuery        string // User query (for planner/replanner)
	CurrentTime      string // Current time
	ExecutedCount    string // Number of executed slides
	TotalCount       string // Total number of slides
	RemainingPlan    string // Remaining slides plan
	QASummary        string // QA result summary
	HasOutline       bool   // True if user provided structured outline (skip planning)
}

// Render executes a named template with the given data and returns the rendered string.
func Render(name string, data *TemplateData) (string, error) {
	tmpl, err := template.ParseFS(FS, name+".tmpl")
	if err != nil {
		return "", fmt.Errorf("prompts: parse %s.tmpl: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("prompts: execute %s.tmpl: %w", name, err)
	}

	return buf.String(), nil
}

// RenderExecutorUserPrompt renders the executor's user prompt template (root level).
func RenderExecutorUserPrompt(data *TemplateData) (string, error) {
	return Render("executor_user_prompt", data)
}

// RenderDeepAgent renders deep agent templates.
func RenderDeepAgent(name string, data *TemplateData) (string, error) {
	return Render("deep/"+name, data)
}

// RenderPlanExecute renders planexecute agent templates.
func RenderPlanExecute(name string, data *TemplateData) (string, error) {
	return Render("planexecute/"+name, data)
}
