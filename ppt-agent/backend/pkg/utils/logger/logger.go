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

package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var defaultLogger *slog.Logger

func init() {
	defaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

var (
	logFile     *os.File
	logFileMu   sync.Mutex
	logFilePath string
)

// LogFilePath 返回当前日志文件路径（如果有）
func LogFilePath() string {
	return logFilePath
}

// Init 初始化全局日志记录器
// 如果 json 为 true，则输出 JSON 格式日志；否则输出人类可读的文本格式
// 如果设置了 LOG_FILE 环境变量，同时写入该文件
func Init(json bool) *slog.Logger {
	var handler slog.Handler
	var w io.Writer = os.Stdout

	if path := os.Getenv("LOG_FILE"); path != "" {
		logFileMu.Lock()
		var err error
		logFile, err = openLogFile(path)
		if err == nil {
			logFilePath = path
			w = io.MultiWriter(os.Stdout, logFile)
		} else {
			slog.Warn("log_file_open_failed", "path", path, "error", err.Error())
		}
		logFileMu.Unlock()
	}

	if json {
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		handler = slog.NewTextHandler(w, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
	return defaultLogger
}

// openLogFile 打开（或创建）日志文件，如果已存在则追加写入
func openLogFile(path string) (*os.File, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// CloseLogFile 关闭日志文件。在服务器关闭时调用
func CloseLogFile() {
	logFileMu.Lock()
	defer logFileMu.Unlock()
	if logFile != nil {
		logFile.Close()
		logFile = nil
		logFilePath = ""
	}
}

// ReadLastNLogLines 读取日志文件最后 n 行。如果未配置日志文件则返回空字符串
func ReadLastNLogLines(n int) (string, error) {
	logFileMu.Lock()
	path := logFilePath
	logFileMu.Unlock()
	if path == "" {
		return "", nil
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	if info.Size() == 0 {
		return "", nil
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	// 始终将日志文件视为 UTF-8 编码，用 Unicode 替换字符替换无效序列以保持字符串有效
	text := string(data)
	if !utf8.Valid(data) {
		text = strings.ToValidUTF8(text, "\uFFFD")
	}

	lines := splitLines(text)
	if len(lines) <= n {
		return text, nil
	}

	result := joinLines(lines[len(lines)-n:])
	return result, nil
}

// LogLevel 日志级别过滤器类型
type LogLevel int

const (
	LogLevelError LogLevel = 1 << iota
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

// ReadLastNLogLinesByLevel 读取最后 n 行并按日志级别过滤
// 只包含 JSON 中 "level" 字段匹配给定级别的行
// 如果未配置日志文件则返回空字符串
func ReadLastNLogLinesByLevel(n int, level LogLevel) (string, error) {
	logFileMu.Lock()
	path := logFilePath
	logFileMu.Unlock()
	if path == "" {
		return "", nil
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	if info.Size() == 0 {
		return "", nil
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	text := string(data)
	if !utf8.Valid(data) {
		text = strings.ToValidUTF8(text, "\uFFFD")
	}

	lines := splitLines(text)
	// Take the last n lines first, then filter
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	var filtered []string
	for _, line := range lines {
		if lineContainsLevel(line, level) {
			filtered = append(filtered, line)
		}
	}

	return joinLines(filtered), nil
}

// ReadFromOffset 从字节偏移量开始读取新的日志内容
// 返回的偏移量是读取时的文件大小（供下次使用）
// 如果文件被截断（大小 < 偏移量），则重置并从头开始读取
// 如果没有新内容，返回空字符串且偏移量不变
// level 过滤要包含的日志级别（使用 LogLevelError | LogLevelDebug 等组合）
func ReadFromOffset(offset int64, level LogLevel) (content string, newOffset int64, err error) {
	logFileMu.Lock()
	path := logFilePath
	logFileMu.Unlock()
	if path == "" {
		return "", 0, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return "", 0, err
	}

	// 文件被截断（日志轮转），重置到开头
	if info.Size() < offset {
		offset = 0
	}

	// 没有新内容
	if info.Size() == offset {
		return "", offset, nil
	}

	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return "", 0, err
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return "", 0, err
	}

	text := string(data)
	if !utf8.Valid(data) {
		text = strings.ToValidUTF8(text, "\uFFFD")
	}

	lines := splitLines(text)
	var filtered []string
	for _, line := range lines {
		if lineContainsLevel(line, level) {
			filtered = append(filtered, line)
		}
	}

	return joinLines(filtered), info.Size(), nil
}

// lineContainsLevel 如果 JSON 日志行的 "level" 字段匹配任何请求的级别则返回 true
func lineContainsLevel(line string, level LogLevel) bool {
	// JSON 日志格式: {"time":"...","level":"INFO","msg":"..."}
	levelIdx := strings.Index(line, `"level":"`)
	if levelIdx >= 0 {
		start := levelIdx + len(`"level":"`)
		end := start
		for end < len(line) && line[end] != '"' {
			end++
		}
		levelVal := strings.ToUpper(line[start:end])

		if level&LogLevelError != 0 && (levelVal == "ERROR" || levelVal == "ERR") {
			return true
		}
		if level&LogLevelWarn != 0 && (levelVal == "WARN" || levelVal == "WARNING") {
			return true
		}
		if level&LogLevelInfo != 0 && levelVal == "INFO" {
			return true
		}
		if level&LogLevelDebug != 0 && levelVal == "DEBUG" {
			return true
		}
		return false
	}

	// 文本日志格式: time=... level=ERROR msg=...
	lower := strings.ToLower(line)
	if level&LogLevelError != 0 && strings.Contains(lower, "level=error") {
		return true
	}
	if level&LogLevelWarn != 0 && strings.Contains(lower, "level=warn") {
		return true
	}
	if level&LogLevelInfo != 0 && strings.Contains(lower, "level=info") {
		return true
	}
	if level&LogLevelDebug != 0 && strings.Contains(lower, "level=debug") {
		return true
	}
	return false
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			if i > 0 && s[i-1] != '\r' {
				lines = append(lines, s[start:i])
			} else if i > 1 {
				lines = append(lines, s[start:i-1])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func joinLines(lines []string) string {
	result := ""
	for i, l := range lines {
		result += l
		if i < len(lines)-1 {
			result += "\n"
		}
	}
	return result
}

// Default 返回全局日志记录器
func Default() *slog.Logger {
	return defaultLogger
}

// Info 记录一条信息日志
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Debug 记录一条调试日志
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Warn 记录一条警告日志
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Error 记录一条错误日志
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// Fatal 记录一条致命日志并退出程序
func Fatal(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
	os.Exit(1)
}

// With 返回一个带有给定属性的日志记录器
func With(args ...any) *slog.Logger {
	return defaultLogger.With(args...)
}

// WithContext 返回一个上下文感知的日志记录器
func WithContext(ctx context.Context) *slog.Logger {
	return defaultLogger
}

// Sleep 后台工作线程使用的辅助睡眠函数
func Sleep(d time.Duration) {
	time.Sleep(d)
}
