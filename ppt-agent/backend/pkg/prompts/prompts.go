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

//go:embed deep/*.tmpl style/*.tmpl log_analysis/*.tmpl
var FS embed.FS

// TemplateData 跨提示模板使用的数据字段。
type TemplateData struct {
	SystemPrompt      string // 注入的系统级指令（skills、rules）
	Input            string // 用户原始请求
	ExecutorContext  string // 当前执行状态摘要
	Step             string // 当前步骤/任务
	Skills           string // 加载的 skill 内容
	WorkDir          string // 工作目录绝对路径
	SkillsDir        string // Skills 目录绝对路径
	TmplDir          string // 模板目录绝对路径
	TasksJSON        string // tasks.json 绝对路径
	TemplateCatalog  string // 内联模板目录表（用于 deep agent）
	UserQuery        string // 用户查询（用于 planner/replanner）
	CurrentTime      string // 当前时间
	ExecutedCount    string // 已执行的幻灯片数量
	TotalCount       string // 总幻灯片数量
	RemainingPlan    string // 剩余幻灯片计划
	QASummary        string // QA 结果摘要
	HasOutline       bool   // 用户是否提供了结构化大纲（跳过规划）
	OutlineQuery     string // 用户原始主题查询（HasOutline=true 时用于内容生成）
	OutlineTemplate  string // 用户大纲中的模板名称（HasOutline=true 时使用）
	OutlineTheme    string // 用户大纲中的主题名称（HasOutline=true 时使用）
	OutlineTitle    string // 用户大纲中的 PPT 标题（HasOutline=true 时使用）
	StyleContext    string // 用户风格偏好上下文，用于个性化生成
	EnableQA       bool   // 是否启用 QA 质量检查
	Concurrency    int    // 每批最大并发页数（来自路由决策，默认 5）
	UserMessage     string // 用户修复请求消息（用于 Fixer agent）
	TargetPages     []int  // 要处理的特定页面索引（用于继续模式）
	UserPreferences *UserPreferences // 用户学习到的偏好，用于个性化生成
}

// UserPreferences 用户学习到的偏好，用于个性化 PPT 生成
type UserPreferences struct {
	LanguageTone      string // 例如："专业商务"、"轻松活泼"
	PreferredColors   []string // 例如：["charcoal_light", "ocean_soft"]
	PreferredLayouts  []string // 例如：["left-image", "two-column"]
	PreferredFonts    []string // 例如：["微软雅黑", "思源黑体"]
	PreferredThemes  []string // 例如：["tech-intro", "pitch-deck"]
	AnimationLevel   string // 例如："minimal"、"moderate"、"rich"
	TypicalPageCount int // 用户的典型页数
	SuccessPatterns  []SuccessPatternInfo // 历史成功模式
	BrandElements    *BrandPreferenceInfo // 品牌元素
	ChartPreferences *ChartPreferenceInfo // 图表偏好
}

// SuccessPatternInfo 用户历史中的成功模式
type SuccessPatternInfo struct {
	Domain         string  // 例如："business"、"technical"
	Template       string  // 模板名称
	Theme          string  // 主题名称
	AvgScore       float64 // 平均质量评分
}

// BrandPreferenceInfo 用户品牌元素偏好
type BrandPreferenceInfo struct {
	LogoPosition string // 例如："bottom-right"、"top-left"
	FooterText   string // 页脚文字
	WatermarkText string // 水印文字
}

// ChartPreferenceInfo 用户图表偏好
type ChartPreferenceInfo struct {
	PreferredTypes []string // 例如：["bar"、"line"]
	Use3D         bool
	ColorScheme   string
}

// Render 使用给定数据执行命名模板并返回渲染后的字符串。
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

// RenderDeepAgent 渲染 deep agent 模板。
func RenderDeepAgent(name string, data *TemplateData) (string, error) {
	return Render("deep/"+name, data)
}

// RenderSlideExecutorContinueInstruction 渲染继续模式的幻灯片执行器指令。
func RenderSlideExecutorContinueInstruction(data *TemplateData) (string, error) {
	return Render("deep/slide_executor_continue_instruction", data)
}

// RenderLogAnalysis 渲染日志分析模板。
func RenderLogAnalysis(name string, data *LogAnalysisData) (string, error) {
	tmpl, err := template.ParseFS(FS, "log_analysis/"+name+".tmpl")
	if err != nil {
		return "", fmt.Errorf("prompts: parse log_analysis/%s.tmpl: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("prompts: execute log_analysis/%s.tmpl: %w", name, err)
	}

	return buf.String(), nil
}

// RenderStyleExtraction 渲染风格提取提示。
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

// StyleExtractionData 风格提取提示的模板数据。
type StyleExtractionData struct {
	UserQuery      string
	Theme          string
	PageCount      int
	ContentTypes   string
	TextContent    string
}

// LogAnalysisData 日志分析模板数据结构
type LogAnalysisData struct {
	SkillsDir string // skills 目录的绝对路径，用于告知 LLM 可读取的文件路径
}
