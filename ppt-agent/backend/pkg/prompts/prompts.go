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
// Prompt files are organized under the prompts/ directory:
//
//	prompts/
//	├── prompts.go                       # 本文件，通用加载函数
//	├── deep/                            # DeepAgent 模式模板
//	│   ├── master_instruction.tmpl
//	│   ├── slide_executor_instruction.tmpl
//	│   ├── reviewer_instruction.tmpl
//	│   ├── fixer_instruction.tmpl
//	│   └── slide_executor_continue_instruction.tmpl
//	└── style/                          # Style extraction templates
//
// 每个模板通过 Render*(data) 系列函数加载并渲染。
package prompts

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed deep/*.tmpl style/*.tmpl
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
	OutlineQuery     string // User's original topic query (used when HasOutline=true for content generation)
	OutlineTemplate  string // Template name from user's outline (used when HasOutline=true)
	OutlineTheme    string // Theme name from user's outline (used when HasOutline=true)
	OutlineTitle    string // PPT title from user's outline (used when HasOutline=true)
	StyleContext    string // User style preference context for personalized generation
	UserMessage     string // User's fix/repair request message (for Fixer agent)
	TargetPages     []int  // Specific page indices to process (for continue mode)
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

// RenderDeepAgent renders deep agent templates.
func RenderDeepAgent(name string, data *TemplateData) (string, error) {
	return Render("deep/"+name, data)
}

// RenderSlideExecutorContinueInstruction renders the continue-mode slide executor instruction.
func RenderSlideExecutorContinueInstruction(data *TemplateData) (string, error) {
	return Render("deep/slide_executor_continue_instruction", data)
}

// RenderStyleExtraction renders style extraction prompts.
func RenderStyleExtraction(system, user string, data *StyleExtractionData) (sysOut, userOut string, err error) {
	sysTmpl, err := template.ParseFS(FS, "style/"+system)
	if err != nil {
		return "", "", fmt.Errorf("prompts: parse style/%s: %w", system, err)
	}
	userTmpl, err := template.ParseFS(FS, "style/"+user)
	if err != nil {
		return "", "", fmt.Errorf("prompts: parse style/%s: %w", user, err)
	}

	var sysBuf, userBuf bytes.Buffer
	if err := sysTmpl.Execute(&sysBuf, data); err != nil {
		return "", "", fmt.Errorf("prompts: execute style/%s: %w", system, err)
	}
	if err := userTmpl.Execute(&userBuf, data); err != nil {
		return "", "", fmt.Errorf("prompts: execute style/%s: %w", user, err)
	}
	return sysBuf.String(), userBuf.String(), nil
}

// StyleExtractionData holds template data for style extraction prompts.
type StyleExtractionData struct {
	UserQuery      string
	Theme          string
	PageCount      int
	ContentTypes   string
	TextContent    string
}
