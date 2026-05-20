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

package web

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cloudwego/ppt-agent/pkg/db"
)

// HealthStatus represents the result of a single health check.
type HealthStatus struct {
	Status  string `json:"status"`  // "ok" or "error"
	Message string `json:"message,omitempty"`
}

// HealthReport is the full health check response.
type HealthReport struct {
	Status      string                  `json:"status"` // "ok", "degraded", or "error"
	Version     string                  `json:"version"`
	Uptime     string                  `json:"uptime"`
	Components  map[string]HealthStatus `json:"components"`
}

// StartTime is set from main.go so health can report uptime.
var StartTime = time.Now()

// Version is the application version, set from main.go.
var Version = "dev"

func (s *Server) handleHealthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	report := HealthReport{
		Status:     "ok",
		Version:    Version,
		Uptime:     time.Since(StartTime).Round(time.Second).String(),
		Components: make(map[string]HealthStatus),
	}

	// Check MySQL
	dbStatus := s.checkDB(ctx)
	report.Components["mysql"] = dbStatus
	if dbStatus.Status != "ok" {
		report.Status = "degraded"
	}

	// Check Python binary
	pyStatus := s.checkPython(ctx)
	report.Components["python"] = pyStatus
	if pyStatus.Status != "ok" {
		report.Status = "error"
	}

	// Check LibreOffice (for QA images)
	loStatus := s.checkLibreOffice(ctx)
	report.Components["libreoffice"] = loStatus
	if loStatus.Status != "ok" {
		report.Status = "degraded"
	}

	statusCode := http.StatusOK
	if report.Status == "error" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, report)
}

func (s *Server) checkDB(ctx context.Context) HealthStatus {
	if db.DB == nil {
		return HealthStatus{Status: "error", Message: "database not configured"}
	}

	sqlDB, err := db.DB.DB()
	if err != nil {
		return HealthStatus{Status: "error", Message: "failed to get underlying sql.DB"}
	}

	// Ping with 2s timeout
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		return HealthStatus{Status: "error", Message: "ping failed: " + err.Error()}
	}

	return HealthStatus{Status: "ok"}
}

func (s *Server) checkPython(ctx context.Context) HealthStatus {
	pythonBin := getPythonBin()

	// Check file exists
	if _, err := os.Stat(pythonBin); err != nil {
		return HealthStatus{Status: "error", Message: "python binary not found at " + pythonBin}
	}

	// Quick version check
	cmd := exec.CommandContext(ctx, pythonBin, "-c", "import sys; print(sys.version_info[:2])")
	out, err := cmd.Output()
	if err != nil {
		return HealthStatus{Status: "error", Message: "python version check failed: " + err.Error()}
	}

	return HealthStatus{Status: "ok", Message: string(out)}
}

func (s *Server) checkLibreOffice(ctx context.Context) HealthStatus {
	// Try common LibreOffice paths
	paths := []string{
		"libreoffice",
		"soffice",
		"/usr/bin/libreoffice",
		"/usr/bin/soffice",
		"/Applications/LibreOffice.app/Contents/MacOS/soffice",
		"C:\\Program Files\\LibreOffice\\program\\soffice.exe",
		"C:\\Program Files (x86)\\LibreOffice\\program\\soffice.exe",
	}

	if runtime.GOOS == "windows" {
		paths = []string{
			"soffice",
			"C:\\Program Files\\LibreOffice\\program\\soffice.exe",
			"C:\\Program Files (x86)\\LibreOffice\\program\\soffice.exe",
		}
	}

	var lastErr error
	for _, p := range paths {
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, p, "--version")
		_, err := cmd.Output()
		if err == nil {
			return HealthStatus{Status: "ok"}
		}
		lastErr = err
	}

	return HealthStatus{
		Status:  "error",
		Message: "libreoffice not found in PATH: " + lastErr.Error(),
	}
}

// getPythonBin returns the configured Python binary path.
func getPythonBin() string {
	// Use env override if set
	if bin := os.Getenv("PYTHON_BIN"); bin != "" {
		return bin
	}
	if runtime.GOOS == "windows" {
		return "python"
	}
	return "/root/pptx_env/bin/python"
}
