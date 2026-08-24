// Package log_analysis 提供后台日志分析服务。
package log_analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/db"
	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/params"
	"github.com/cloudwego/ppt-agent/pkg/prompts"
	"github.com/cloudwego/ppt-agent/pkg/tools"
)

// Analyzer 日志分析器接口
type Analyzer interface {
	Analyze(ctx context.Context, logs string, taskID string, prevAnalyses []db.TaskErrorAnalysis) (*Result, error)
}

// Result 日志分析结果
type Result struct {
	Analysis   string // 总体分析描述
	RootCause  string // 根因描述（一句话）
	Suggestion string // 修复建议
	TokensUsed int64  // 消耗的 token 数量
	ModelUsed  string // 使用的模型名称
}

// LLMAnalyzer 基于工具调用大模型的日志分析器。
// 绑定 read_file 工具，使 LLM 能够自主读取项目的 prompt 模板和 Python 生成器源码，
// 理解上下文后再给出分析结论。
type LLMAnalyzer struct {
	modelFactory func(ctx context.Context) (model.ToolCallingChatModel, error) // 模型工厂函数
	skillsDir    string                                                        // skills 目录绝对路径
}

// NewLLMAnalyzer 创建日志分析器实例。
// skillsDir 指定 skills 目录的绝对路径，分析器会将其传给 LLM 作为 read_file 工具的可读文件根目录。
func NewLLMAnalyzer(modelFactory func(ctx context.Context) (model.ToolCallingChatModel, error), skillsDir string) *LLMAnalyzer {
	return &LLMAnalyzer{
		modelFactory: modelFactory,
		skillsDir:    skillsDir,
	}
}

// Analyze 对日志片段进行 ReAct 分析。
// LLM 会自主决定是否调用 read_file 工具读取相关源码文件，理解上下文后输出分析结果。
// prevAnalyses 为该任务的历史分析记录，LLM 会参考之前的修复建议避免重复无效尝试。
func (a *LLMAnalyzer) Analyze(ctx context.Context, logs string, taskID string, prevAnalyses []db.TaskErrorAnalysis) (*Result, error) {
	model, err := a.modelFactory(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建分析模型失败: %w", err)
	}

	// 通过模板加载器获取分析 system prompt（自动替换 {{ .SkillsDir }} 占位符）
	systemPrompt, err := prompts.RenderLogAnalysis("analyzer_instruction", &prompts.LogAnalysisData{
		SkillsDir: a.skillsDir,
	})
	if err != nil {
		// 模板加载失败时使用内置降级 prompt
		systemPrompt = buildFallbackPrompt(a.skillsDir)
	}

	// 构建用户消息，附加历史分析上下文
	userMsg := a.buildUserMessage(logs, prevAnalyses)

	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userMsg),
	}

	// 构建 read_file 工具，供 LLM 在 ReAct 循环中调用
	readToolOp := &readFileOperator{workDir: a.skillsDir}
	readTool := tools.NewReadFileTool(readToolOp)

	toolInfo, err := readTool.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取 read_file 工具信息失败: %w", err)
	}

	// 将 read_file 工具绑定到模型
	toolModel, err := model.WithTools([]*schema.ToolInfo{toolInfo})
	if err != nil {
		return nil, fmt.Errorf("绑定工具失败: %w", err)
	}

	// ReAct 循环：generate → 工具调用 → 结果 → generate → … → 完成
	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		resp, err := toolModel.Generate(ctx, messages)
		if err != nil {
			return nil, fmt.Errorf("LLM 分析调用失败: %w", err)
		}

		messages = append(messages, resp)

		// 无工具调用，说明分析已完成
		if len(resp.ToolCalls) == 0 {
			break
		}

		// 执行工具调用
		for _, tc := range resp.ToolCalls {
			var toolResult string
			if tc.Function.Name == "read_file" {
				toolResult = a.executeReadFile(ctx, readToolOp, tc.Function.Arguments)
			} else {
				toolResult = fmt.Sprintf("未知工具: %s", tc.Function.Name)
			}
			messages = append(messages, &schema.Message{
				Role:       schema.Tool,
				ToolCallID: tc.ID,
				Content:    toolResult,
			})
		}
	}

	// 从后往前扫描消息，找到第一条能成功解析出 root_cause 或 suggestion 的助手消息
	// （LLM 的最终分析可能不在最后一条消息中，比如最后一条是 tool 结果）
	var analysis, rootCause, suggestion string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == schema.Assistant && messages[i].Content != "" {
			analysis, rootCause, suggestion = parseAnalysisResult(messages[i].Content)
			if rootCause != "" || suggestion != "" || analysis != "" {
				break
			}
		}
	}

	modelName := ArkModelName(model)

	return &Result{
		Analysis:   analysis,
		RootCause:  rootCause,
		Suggestion: suggestion,
		ModelUsed:  modelName,
	}, nil
}

// buildUserMessage 构建用户消息，附加历史分析上下文。
// 当有历史分析时，会在消息中列出之前的修复建议，帮助 LLM 避免重复无效尝试。
func (a *LLMAnalyzer) buildUserMessage(logs string, prevAnalyses []db.TaskErrorAnalysis) string {
	var sb strings.Builder
	sb.WriteString("## 错误日志\n\n以下是来自 ppt-agent 系统的日志片段（ERROR 和 DEBUG 级别），请分析：\n\n")
	sb.WriteString(logs)

	// 附加历史分析记录，帮助 LLM 避免重复之前的无效修复尝试
	if len(prevAnalyses) > 0 {
		sb.WriteString("\n\n## 历史分析记录\n\n")
		sb.WriteString("该任务之前已有 ")
		sb.WriteString(fmt.Sprintf("%d", len(prevAnalyses)))
		sb.WriteString(" 次分析，以下是历史修复建议（请参考以避免重复相同的无效尝试）：\n\n")
		for i, prev := range prevAnalyses {
			sb.WriteString(fmt.Sprintf("### 历史分析 #%d（触发类型: %s）\n", i+1, prev.TriggerType))
			if prev.RootCause != "" {
				sb.WriteString(fmt.Sprintf("- 根因: %s\n", prev.RootCause))
			}
			if prev.Suggestion != "" {
				sb.WriteString(fmt.Sprintf("- 修复建议: %s\n", prev.Suggestion))
			}
			if prev.Analysis != "" {
				analysis := prev.Analysis
				if len(analysis) > 300 {
					analysis = analysis[:300] + "...(已截断)"
				}
				sb.WriteString(fmt.Sprintf("- 分析摘要: %s\n", analysis))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n请先用 read_file 读取你认为相关的项目文件（如 prompt 模板或 generator 源码），理解上下文后再给出分析。")
	return sb.String()
}

// executeReadFile 解析工具调用参数并读取文件内容。
func (a *LLMAnalyzer) executeReadFile(ctx context.Context, op *readFileOperator, argsJSON string) string {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return fmt.Sprintf("参数解析失败: %v", err)
	}
	if args.Path == "" {
		return "错误: path 参数为空"
	}

	content, err := op.ReadFile(ctx, args.Path)
	if err != nil {
		return fmt.Sprintf("文件读取失败: %v", err)
	}
	// 超长文件截断，避免消耗过多上下文
	if len(content) > 50000 {
		content = content[:50000] + "\n... (内容过长，已截断)"
	}
	return fmt.Sprintf("File: %s\nContent:\n%s", args.Path, content)
}

// parseAnalysisResult 从 LLM 回复中按【标记】解析分析结果。
// 格式：【分析】...【根本原因】...【修复建议】...
func parseAnalysisResult(text string) (analysis, rootCause, suggestion string) {
	analysis = extractTag(text, "【分析】", "【根本原因】")
	rootCause = extractTag(text, "【根本原因】", "【修复建议】")
	suggestion = extractTag(text, "【修复建议】", "")

	if rootCause == "" && suggestion == "" {
		// 解析失败：全文作为 analysis，root_cause/suggestion 留空
		return text, "", ""
	}
	return analysis, rootCause, suggestion
}

// extractTag 从文本中提取两个标记之间的内容。
func extractTag(text, startTag, endTag string) string {
	idx := strings.Index(text, startTag)
	if idx < 0 {
		return ""
	}
	inner := text[idx+len(startTag):]
	if endTag == "" {
		return strings.TrimSpace(inner)
	}
	end := strings.Index(inner, endTag)
	if end < 0 {
		return strings.TrimSpace(inner)
	}
	return strings.TrimSpace(inner[:end])
}

// buildFallbackPrompt 当模板文件加载失败时使用的降级 system prompt。
func buildFallbackPrompt(skillsDir string) string {
	return fmt.Sprintf(`你是一个专业的 Go 后端系统调试专家，同时也是 Python PPT 生成（python-pptx）领域的专家。

## 你的任务

分析来自 ppt-agent 系统的错误日志，找出问题的根本原因，并给出修复建议。

## 项目背景

ppt-agent 是一个 AI PPT 生成系统：
- Go 后端使用“意图分类 → Planner LLM → DeckSpec/tasks.json → Eino workflow → renderer worker pool”的生成链路
- Python PPT 生成使用 python-pptx 库，代码位于 skills/ppt-deck-planner/generators/ 目录
- Planner 只调用 read_file、update_tasks_manifest、search 准备 DeckSpec；页面渲染由 render_task.py 按 task_id 确定性完成

## 日志分析方法

1. **ERROR 日志**：重点关注 "ERROR"、"panic"、"failed"、"exception" 等关键词
2. **DEBUG 日志**：上下文信息，理解执行流程
3. **语言判断**：
   - Go 日志（pkg/、github.com/、eino/）→ Go 后端问题（API 调用失败、超时、资源不足、逻辑错误等）
   - Python Traceback（Traceback、python-pptx、.py）→ PPT 生成器代码问题（API 使用错误、文件路径问题、依赖缺失等）
4. **文件关联**：遇到错误时，用 read_file 工具读取相关源代码文件来理解上下文

## 历史分析参考

如果用户消息中存在 ## 历史分析记录 部分，**必须先阅读历史记录**：
- 如果之前已给出修复建议且该错误仍然出现，说明之前的建议未被正确执行或方向有误，需要重新分析
- 避免重复推荐相同的无效修复方案
- 关注之前的 root_cause 与当前日志是否指向同一根本问题

## 可用文件路径

- /ppt/ppt-agent/backend/pkg/prompts
- 组件契约：%s/ppt-deck-planner/templates/component_contracts.json
- Python 生成器：%s/ppt-deck-planner/generators/base.py
- Python 生成器：%s/ppt-deck-planner/generators/*_generator.py

## 输出格式

按以下**严格格式**输出分析结果（不要输出 JSON，不要输出其他代码块）：

【分析】
总体分析描述，包含具体的代码行号、函数名、文件路径等定位信息，100-300字

【根本原因】
简短的根因描述，一句话，必须精确到具体文件/函数/行

【修复建议】
具体可操作的修复建议，必须是有把握的修复方案，100-200字，包含文件路径和具体修改内容

**格式要求：**
- 必须使用中文全角方括号【】作为每个字段的起始标记，不要用 JSON
- 每个标记独占一行，标记后面紧跟对应内容
- 三个字段缺一不可`, skillsDir, skillsDir, skillsDir)
}

// readFileOperator 是 read_file 工具的命令行操作实现。
// 仅实现 ReadFile，写操作 panic（分析器只需读取文件）。
type readFileOperator struct {
	workDir string
}

func (r *readFileOperator) ReadFile(ctx context.Context, path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (r *readFileOperator) WriteFile(ctx context.Context, path string, content string) error {
	panic("readFileOperator: WriteFile 不应被调用")
}

func (r *readFileOperator) IsDirectory(ctx context.Context, path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func (r *readFileOperator) Exists(ctx context.Context, path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (r *readFileOperator) RunCommand(ctx context.Context, command []string) (*commandline.CommandOutput, error) {
	panic("readFileOperator: RunCommand 不应被调用")
}

func (r *readFileOperator) GetWorkDir(ctx context.Context) string {
	if wd, ok := params.GetTypedContextParams[string](ctx, params.WorkDirSessionKey); ok && wd != "" {
		return wd
	}
	return r.workDir
}

func (r *readFileOperator) SetWorkDir(ctx context.Context, dir string) context.Context {
	r.workDir = dir
	return params.SetTypedContextParams(ctx, params.WorkDirSessionKey, dir)
}

// ArkModelName 从模型实例中提取模型名称。
func ArkModelName(m model.ToolCallingChatModel) string {
	type namedModel interface{ String() string }
	if nm, ok := m.(namedModel); ok {
		return nm.String()
	}
	type primaryModel interface{ PrimaryModelName() string }
	if pm, ok := m.(primaryModel); ok {
		return pm.PrimaryModelName()
	}
	return ""
}

// ─────────────────────────────────────────────────────────────────────────────
// Service 后台日志分析服务
// ─────────────────────────────────────────────────────────────────────────────

// Service 管理后台日志分析任务（空闲分析 + 失败分析）。
type Service struct {
	analyzer     Analyzer                                                      // 日志分析器
	modelFactory func(ctx context.Context) (model.ToolCallingChatModel, error) // 模型工厂
	skillsDir    string                                                        // skills 目录
	idleInterval time.Duration                                                 // 空闲分析间隔
	logLines     int                                                           // 每次读取的日志行数

	mu          sync.Mutex
	stopCh      chan struct{}
	pendingTask chan taskRequest // 待处理任务队列
	running     bool

	lastFileOffset int64 // 已处理日志文件偏移量（字节）
}

type taskRequest struct {
	TaskID      string // 任务 ID，"system-idle" 表示空闲分析
	TriggerType string // 触发类型："idle" 或 "failed"
	Logs        string // 日志内容（外部传入时非空）
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	// ModelFactory 创建分析用聊天模型的工厂函数
	ModelFactory func(ctx context.Context) (model.ToolCallingChatModel, error)
	// IdleInterval 空闲时分析日志的间隔。默认 5 分钟，设为 0 可禁用空闲分析。
	IdleInterval time.Duration
	// LogLines 每次读取的日志行数。默认 300 行。
	LogLines int
	// SkillsDir skills 目录的绝对路径。用于告知分析器可读取的文件位置。
	// 为空时只分析日志文本，不读取源码文件。
	SkillsDir string
}

// NewService 创建后台日志分析服务。
func NewService(cfg *ServiceConfig) *Service {
	idleInterval := 5 * time.Minute
	if cfg.IdleInterval > 0 {
		idleInterval = cfg.IdleInterval
	}
	logLines := 300
	if cfg.LogLines > 0 {
		logLines = cfg.LogLines
	}

	var analyzer Analyzer
	if cfg.ModelFactory != nil {
		analyzer = NewLLMAnalyzer(cfg.ModelFactory, cfg.SkillsDir)
	}

	return &Service{
		analyzer:       analyzer,
		modelFactory:   cfg.ModelFactory,
		skillsDir:      cfg.SkillsDir,
		idleInterval:   idleInterval,
		logLines:       logLines,
		stopCh:         make(chan struct{}),
		pendingTask:    make(chan taskRequest, 10),
		lastFileOffset: 0,
	}
}

// Start 启动后台分析循环（应只调用一次，在服务启动后）。
func (s *Service) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	logger.Info("log_analysis_service_started",
		"idle_interval", s.idleInterval.String(),
		"log_lines", s.logLines,
		"skills_dir", s.skillsDir)

	go s.loop()
}

// Stop 优雅停止后台服务。
func (s *Service) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopCh)
	logger.Info("log_analysis_service_stopped")
}

// Trigger 将一个日志分析任务加入队列。triggerType 为 "idle" 或 "failed"。
func (s *Service) Trigger(taskID, triggerType string) {
	select {
	case s.pendingTask <- taskRequest{TaskID: taskID, TriggerType: triggerType}:
		logger.Debug("log_analysis_triggered", "task_id", taskID, "type", triggerType)
	default:
		logger.Warn("log_analysis_queue_full", "task_id", taskID)
	}
}

func (s *Service) loop() {
	idleTicker := time.NewTicker(s.idleInterval)
	defer idleTicker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case req := <-s.pendingTask:
			s.runAnalysis(req)
		case <-idleTicker.C:
			s.runIdleAnalysis()
		}
	}
}

func (s *Service) runAnalysis(req taskRequest) {
	if s.analyzer == nil {
		logger.Warn("log_analysis_disabled", "task_id", req.TaskID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var logs string
	var err error
	if req.Logs != "" {
		// 日志内容由外部传入（如任务失败时）
		logs = req.Logs
	} else {
		// 读取 ERROR + DEBUG + INFO，以覆盖 bug 日志和 user_feedback 等诊断信息
		logs, err = logger.ReadLastNLogLinesByLevel(s.logLines, logger.LogLevelError|logger.LogLevelDebug|logger.LogLevelInfo)
		if err != nil {
			logger.Error("log_analysis_read_failed", "task_id", req.TaskID, "error", err.Error())
			return
		}
	}
	if logs == "" {
		logger.Debug("log_analysis_no_logs", "task_id", req.TaskID)
		return
	}

	// 获取该任务的历史分析记录，用于避免重复无效的修复尝试
	var prevAnalyses []db.TaskErrorAnalysis
	if req.TaskID != "system-idle" {
		prevAnalyses, _ = db.GetTaskErrorAnalysis(req.TaskID)
	}

	logger.Info("log_analysis_starting",
		"task_id", req.TaskID,
		"trigger", req.TriggerType,
		"log_bytes", len(logs),
		"prev_analyses", len(prevAnalyses))

	result, err := s.analyzer.Analyze(ctx, logs, req.TaskID, prevAnalyses)
	if err != nil {
		logger.Error("log_analysis_llm_failed", "task_id", req.TaskID, "error", err.Error())
		return
	}

	analysis := &db.TaskErrorAnalysis{
		TaskID:      req.TaskID,
		TriggerType: req.TriggerType,
		LogSnippet:  truncateForDB(logs, 65535),
		Analysis:    result.Analysis,
		RootCause:   result.RootCause,
		Suggestion:  result.Suggestion,
		TokensUsed:  result.TokensUsed,
		ModelUsed:   result.ModelUsed,
		CreatedAt:   time.Now(),
	}

	if err := db.CreateTaskErrorAnalysis(analysis); err != nil {
		logger.Error("log_analysis_db_save_failed", "task_id", req.TaskID, "error", err.Error())
	} else {
		logger.Info("log_analysis_completed",
			"task_id", req.TaskID,
			"root_cause", result.RootCause,
			"tokens", result.TokensUsed)
	}
}

// HasRunningTasksFunc 外部注入的回调函数，用于判断当前是否有运行中的任务。
// 为 nil 时，空闲分析不检查任务状态。
var HasRunningTasksFunc func() bool

func (s *Service) runIdleAnalysis() {
	s.mu.Lock()
	running := s.running
	currentOffset := s.lastFileOffset
	s.mu.Unlock()
	if !running {
		return
	}

	// 有任务在运行时跳过本次空闲分析
	if HasRunningTasksFunc != nil && HasRunningTasksFunc() {
		return
	}

	// 读取 ERROR + DEBUG + INFO，以覆盖 bug 日志和 user_feedback 等诊断信息
	logs, newOffset, err := logger.ReadFromOffset(currentOffset, logger.LogLevelError|logger.LogLevelDebug|logger.LogLevelInfo)
	if err != nil {
		logger.Warn("idle_log_read_failed", "error", err.Error())
		return
	}
	if logs == "" {
		return
	}

	s.mu.Lock()
	s.lastFileOffset = newOffset
	s.mu.Unlock()

	s.runAnalysis(taskRequest{
		TaskID:      "system-idle",
		TriggerType: "idle",
		Logs:        logs,
	})
}

func truncateForDB(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// GetRecentAnalyses 从数据库获取最近的分析记录。
func GetRecentAnalyses(limit int) ([]db.TaskErrorAnalysis, error) {
	return db.ListRecentErrorAnalyses(limit)
}

// GetTaskAnalyses 获取指定任务的所有分析记录。
func GetTaskAnalyses(taskID string) ([]db.TaskErrorAnalysis, error) {
	return db.GetTaskErrorAnalysis(taskID)
}
