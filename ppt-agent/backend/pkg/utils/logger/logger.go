package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

var defaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

var (
	logFile     *os.File
	logFileMu   sync.Mutex
	logFilePath string
)

func LogFilePath() string {
	logFileMu.Lock()
	defer logFileMu.Unlock()
	return logFilePath
}

func Init(json bool) *slog.Logger {
	var handler slog.Handler
	var w io.Writer = os.Stdout
	if path := os.Getenv("LOG_FILE"); path != "" {
		logFileMu.Lock()
		f, err := openLogFile(path)
		if err == nil {
			logFile, logFilePath, w = f, path, io.MultiWriter(os.Stdout, f)
		} else {
			slog.Warn("log_file_open_failed", "path", path, "error", err.Error())
		}
		logFileMu.Unlock()
	}
	if json {
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
	return defaultLogger
}

func openLogFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

func CloseLogFile() {
	logFileMu.Lock()
	defer logFileMu.Unlock()
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
		logFilePath = ""
	}
}
