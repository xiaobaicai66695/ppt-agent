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
)

var defaultLogger *slog.Logger

func init() {
	defaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// Init initializes the global logger.
// If json is true, outputs JSON lines. Otherwise outputs human-readable text.
// Returns the created logger.
func Init(json bool) *slog.Logger {
	var handler slog.Handler
	var w io.Writer = os.Stdout
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

// Default returns the global logger.
func Default() *slog.Logger {
	return defaultLogger
}

// Info logs an informational message.
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Debug logs a debug message.
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Warn logs a warning message.
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Error logs an error message.
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// Fatal logs a fatal message and exits.
func Fatal(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
	os.Exit(1)
}

// With returns a logger with the given attributes.
func With(args ...any) *slog.Logger {
	return defaultLogger.With(args...)
}

// WithContext returns a context-aware logger (slog context not yet in stdlib,
// but we provide the pattern so callers can pass structured fields).
func WithContext(ctx context.Context) *slog.Logger {
	return defaultLogger
}
