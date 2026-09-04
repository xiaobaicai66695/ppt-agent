package logger

import (
	"context"
	"log/slog"
	"os"
	"time"
)

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
