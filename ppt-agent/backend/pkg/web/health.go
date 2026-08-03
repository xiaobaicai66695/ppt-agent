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
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cloudwego/ppt-agent/pkg/db"
	"github.com/cloudwego/ppt-agent/pkg/tools/pythonutil"
)

// HealthStatus 表示单个健康检查的结果。
type HealthStatus struct {
	Status  string `json:"status"` // "ok" or "error"
	Message string `json:"message,omitempty"`
}

// HealthReport 是完整的健康检查响应。
type HealthReport struct {
	Status     string                  `json:"status"` // "ok", "degraded", or "error"
	Version    string                  `json:"version"`
	Uptime     string                  `json:"uptime"`
	Components map[string]HealthStatus `json:"components"`
}

// StartTime 由 main.go 设置，以便健康检查可以报告服务运行时间。
var StartTime = time.Now()

// Version 是应用版本号，由 main.go 设置。
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

	// 检查 Python 运行时
	pyStatus := s.checkPython(ctx)
	report.Components["python"] = pyStatus
	if pyStatus.Status != "ok" {
		report.Status = "error"
	}

	// 检查 LibreOffice（用于 QA 图片生成）
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

	// 使用 2 秒超时进行 Ping
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		return HealthStatus{Status: "error", Message: "ping failed: " + err.Error()}
	}

	return HealthStatus{Status: "ok"}
}

func (s *Server) checkPython(ctx context.Context) HealthStatus {
	pythonBin := getPythonBin()

	// 检查文件是否存在
	if _, err := os.Stat(pythonBin); err != nil {
		return HealthStatus{Status: "error", Message: "python binary not found at " + pythonBin}
	}

	// 快速版本检查
	cmd := exec.CommandContext(ctx, pythonBin, "-c", "import sys; print(sys.version_info[:2])")
	out, err := cmd.Output()
	if err != nil {
		return HealthStatus{Status: "error", Message: "python version check failed: " + err.Error()}
	}

	return HealthStatus{Status: "ok", Message: string(out)}
}

func (s *Server) checkLibreOffice(ctx context.Context) HealthStatus {
	// 项目 CLI 仅支持 Linux，检查 PATH 和常见 Linux 安装路径。
	paths := []string{
		"libreoffice",
		"soffice",
		"/usr/bin/libreoffice",
		"/usr/bin/soffice",
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

// getPythonBin 返回配置的 Python 二进制路径。
func getPythonBin() string {
	return pythonutil.GetPythonBinary()
}
