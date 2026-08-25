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
//	├── planner/                         # PPT Planner 模板
//	│   └── master_instruction.tmpl
//	└── style/                          # Style extraction templates
//
// 每个模板通过 Render*(data) 系列函数加载并渲染。
package prompts

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"text/template"
)

//go:embed planner/*.tmpl reviewer/*.tmpl fixer/*.tmpl log_analysis/*.tmpl
var FS embed.FS

// templateFuncs 提供给所有模板的函数映射。
var templateFuncs = template.FuncMap{
	"mul": func(a, b int) int { return a * b },
}

// parseWithFuncs 解析嵌入文件中的模板，并注入自定义函数。
// 注意：Funcs 必须在解析模板之前调用，否则模板中的自定义函数会报 "not defined"。
// root 模板名使用 filepath.Base(pattern)，与 ParseFS 内部创建的模板名一致，确保 Execute() 能正确执行。
func parseWithFuncs(pattern string) (*template.Template, error) {
	tmpl := template.New(filepath.Base(pattern)).Funcs(templateFuncs)
	_, err := tmpl.ParseFS(FS, pattern)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

// TemplateData 跨提示模板使用的数据字段。
type TemplateData struct {
	SkillsDir            string
	TasksJSON            string
	HasOutline           bool
	OutlineQuery         string
	SuggestedPageCount   int
	StyleContext         string
	ImageSearchAvailable bool
}

// Render 使用给定数据执行命名模板并返回渲染后的字符串。
func Render(name string, data *TemplateData) (string, error) {
	tmpl, err := parseWithFuncs(name + ".tmpl")
	if err != nil {
		return "", fmt.Errorf("prompts: parse %s.tmpl: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("prompts: execute %s.tmpl: %w", name, err)
	}

	return buf.String(), nil
}

// RenderPlanner 渲染 PPT Planner 模板。
func RenderPlanner(name string, data *TemplateData) (string, error) {
	return Render("planner/"+name, data)
}

// RenderReviewer 渲染 DeckSpec Reviewer 模板。
func RenderReviewer(name string, data *TemplateData) (string, error) {
	return Render("reviewer/"+name, data)
}

// RenderFixer 渲染生成后定点修复模板。
func RenderFixer(name string, data *TemplateData) (string, error) {
	return Render("fixer/"+name, data)
}

// RenderLogAnalysis 渲染日志分析模板。
func RenderLogAnalysis(name string, data *LogAnalysisData) (string, error) {
	tmpl, err := parseWithFuncs("log_analysis/" + name + ".tmpl")
	if err != nil {
		return "", fmt.Errorf("prompts: parse log_analysis/%s.tmpl: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("prompts: execute log_analysis/%s.tmpl: %w", name, err)
	}

	return buf.String(), nil
}

// LogAnalysisData 日志分析模板数据结构
type LogAnalysisData struct {
	SkillsDir string // skills 目录的绝对路径，用于告知 LLM 可读取的文件路径
}
