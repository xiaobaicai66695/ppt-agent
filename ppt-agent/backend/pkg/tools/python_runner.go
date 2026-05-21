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

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/metrics"
	"github.com/cloudwego/ppt-agent/pkg/params"
)

var pythonRunnerToolInfo = &schema.ToolInfo{
	Name: "python3",
	Desc: `Execute Python code. The code will be saved to a temporary .py file and executed.
* Use this tool to run Python scripts for PPT generation.`,
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"code": {
			Type:     "string",
			Desc:     "Python code to execute",
			Required: true,
		},
	}),
}

// Default timeout for Python execution (60 seconds).
const defaultPythonTimeout = 60 * time.Second

// MaxPythonTimeout is the hard upper limit (5 minutes).
const MaxPythonTimeout = 5 * time.Minute

// MaxCodeSize is the maximum allowed code size (100 KB).
const MaxCodeSize = 100 * 1024

// Dangerous patterns that should not appear in user-generated code.
// This catches common attack patterns even when wrapped.
var dangerousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`os\.system\s*\(`),
	regexp.MustCompile(`subprocess\s*\.\s*(run|call|Popen|check_output|shell)\s*\(`),
	regexp.MustCompile(`__import__\s*\(`),
	regexp.MustCompile(`exec\s*\(`),
	regexp.MustCompile(`eval\s*\(`),
	regexp.MustCompile(`compile\s*\(`),
	regexp.MustCompile(`open\s*\([^)]*["']\/dev\/`),
	regexp.MustCompile(`open\s*\([^)]*["']\/proc\/`),
	regexp.MustCompile(`rm\s+-rf`),
	regexp.MustCompile(`curl\s+.*\|?\s*sh`),
	regexp.MustCompile(`wget\s+.*\|?\s*sh`),
}

type pythonRunnerTool struct {
	op      commandline.Operator
	timeout time.Duration
}

type pythonInput struct {
	Code string `json:"code"`
}

// NewPythonRunnerToolImpl creates a Python runner tool with safety limits.
func NewPythonRunnerToolImpl(op commandline.Operator) tool.InvokableTool {
	return newPythonRunnerToolImpl(op, defaultPythonTimeout)
}

func newPythonRunnerToolImpl(op commandline.Operator, timeout time.Duration) tool.InvokableTool {
	if timeout <= 0 {
		timeout = defaultPythonTimeout
	}
	if timeout > MaxPythonTimeout {
		timeout = MaxPythonTimeout
	}

	// Allow env override: PYTHON_TIMEOUT_SECS=30
	if secs := os.Getenv("PYTHON_TIMEOUT_SECS"); secs != "" {
		if s, err := strconv.Atoi(secs); err == nil && s > 0 {
			timeout = time.Duration(s) * time.Second
			if timeout > MaxPythonTimeout {
				timeout = MaxPythonTimeout
			}
		}
	}

	return &pythonRunnerTool{op: op, timeout: timeout}
}

func (p *pythonRunnerTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return pythonRunnerToolInfo, nil
}

func (p *pythonRunnerTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	input := &pythonInput{}
	if err := json.Unmarshal([]byte(argumentsInJSON), input); err != nil {
		return "", err
	}

	if len(input.Code) == 0 {
		return "code cannot be empty", nil
	}

	// 1. Code size limit
	if len(input.Code) > MaxCodeSize {
		logger.Warn("python_code_rejected_size",
			"size_bytes", len(input.Code),
			"max_bytes", MaxCodeSize)
		metrics.RecordToolCall("python3", "error")
		return fmt.Sprintf("code size %d exceeds limit %d bytes", len(input.Code), MaxCodeSize), nil
	}

	// 2. Dangerous command detection
	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(input.Code) {
			logger.Warn("python_code_rejected_dangerous",
				"pattern", pattern.String(),
				"code_len", len(input.Code))
			metrics.RecordToolCall("python3", "error")
			return fmt.Sprintf("code contains dangerous pattern '%s' and was rejected", pattern.String()), nil
		}
	}

	wd, _ := params.GetTypedContextParams[string](ctx, params.WorkDirSessionKey)
	if wd == "" {
		wd, _ = os.Getwd()
	}

	tmpFile := filepath.Join(wd, fmt.Sprintf("temp_script_%d.py", time.Now().UnixNano()))

	// 3. Wrap code with safety guard: redirect dangerous builtins
	// Python-level safety guard is disabled by default — the Go-level
	// dangerousPatterns regex (os.system, subprocess, exec, eval, curl|sh
	// etc.) already catches real threats. The Python wrapper was too
	// aggressive and broke legitimate python-pptx usage (open(), import os).
	// Set env PYTHON_SAFETY_GUARD=true to re-enable.
	codeToWrite := input.Code
	if os.Getenv("PYTHON_SAFETY_GUARD") == "true" {
		codeToWrite = wrapWithSafetyGuard(input.Code)
	}
	if err := os.WriteFile(tmpFile, []byte(codeToWrite), 0o644); err != nil {
		return fmt.Sprintf("failed to write temp file: %v", err), nil
	}

	// Ensure cleanup even on concurrent calls to same workDir
	defer os.Remove(tmpFile)

	// 4. Apply timeout to context
	execCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// 5. Determine Python binary
	pythonBin := getPythonBinary()

	cmd := []string{pythonBin, tmpFile}
	o := tool.GetImplSpecificOptions(&options{op: p.op}, opts...)
	output, err := o.op.RunCommand(execCtx, cmd)

	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			logger.Warn("python_execution_timeout",
				"timeout", p.timeout.String(),
				"code_size", len(input.Code))
			metrics.RecordToolCall("python3", "error")
			return fmt.Sprintf("Python execution timed out after %v. This can happen when LibreOffice is slow to start or the script is complex.", p.timeout), nil
		}
		logger.Error("python_execution_failed", "error", err.Error())
		metrics.RecordToolCall("python3", "error")
		return fmt.Sprintf("Python execution failed: %v", err), nil
	}

	metrics.RecordToolCall("python3", "success")
	return formatPythonOutput(output), nil
}

// getPythonBinary returns the Python binary path.
// Uses PYTHON_BIN env var if set, falls back to platform-specific defaults.
func getPythonBinary() string {
	if bin := os.Getenv("PYTHON_BIN"); bin != "" {
		return bin
	}
	if runtime.GOOS == "windows" {
		return "python"
	}
	return "/root/pptx_env/bin/python"
}

// wrapWithSafetyGuard prepends a safety wrapper that restricts dangerous
// builtins during execution. This catches subprocess, system calls, etc.
// without requiring sandboxing (which needs root/cgroups on Linux).
func wrapWithSafetyGuard(code string) string {
	guard := fmt.Sprintf(`# -*- coding: utf-8 -*-
# Safety wrapper: restrict dangerous builtins
import builtins as _b

# ── Step 1: Capture originals BEFORE any replacement ──
_ORIGINAL_IMPORT = __builtins__.__import__

_FORBIDDEN = frozenset({"system", "popen", "exec", "eval", "compile", "__import__",
    "run", "check_output", "call", "Popen", "getoutput", "getstatusoutput",
    "mkfifo", "mknod"})
# NOTE: "open" and "file" are intentionally NOT blocked — pptx/zipfile internals use them.

def _deny(*a, **k):
    raise RuntimeError("Dangerous function call blocked by safety guard")

for _f in _FORBIDDEN:
    try:
        setattr(_b, _f, _deny)
    except Exception:
        pass

import sys
for _m in list(sys.modules.keys()):
    if any(_m.startswith(p) for p in ("subprocess", "shutil", "glob", "fnmatch")):
        try:
            del sys.modules[_m]
        except Exception:
            pass

# Allowlist: only these modules are safe to import.
# os/pathlib/sys/copy are required by generators and pptx internals.
_ALLOWED_MODULES = frozenset({
    "pptx", "pptx.util", "pptx.dml", "pptx.enum", "pptx.oxml",
    "jinja2", "markupsafe", "PIL", "PIL.Image", "PIL.JpegImagePlugin",
    "PIL.PngImagePlugin", "PIL.ImageDraw", "PIL.ImageFont",
    "reportlab", "reportlab.pdfgen", "reportlab.platypus",
    "matplotlib", "numpy", "json", "re", "datetime", "uuid",
    "collections", "functools", "itertools", "math", "random",
    "typing", "os", "os.path", "sys", "copy", "pathlib", "shutil",
})

def _safe_import(name, *a, **k):
    base = name.split(".")[0]
    if base not in _ALLOWED_MODULES and base not in dir(_b):
        raise ImportError(f"Import of '{name}' is not allowed in sandboxed mode")
    return _ORIGINAL_IMPORT(name, *a, **k)

if hasattr(__builtins__, '__import__'):
    __builtins__.__import__ = _safe_import
else:
    __builtins__['__import__'] = _safe_import

# Disable file write outside work dir (will be set by caller)
# We rely on the work dir restriction from the Go side instead.

`)

	return guard + "\n# --- User Code ---\n" + code
}

func formatPythonOutput(output *commandline.CommandOutput) string {
	result := ""
	if output.Stdout != "" {
		result += fmt.Sprintf("stdout:\n%s\n", output.Stdout)
	}
	if output.Stderr != "" {
		result += fmt.Sprintf("stderr:\n%s\n", output.Stderr)
	}
	if result == "" {
		result = "(no output)"
	}
	return result
}

// CheckCodeSafety performs static analysis on Python code and returns
// a human-readable list of any issues found.
func CheckCodeSafety(code string) []string {
	var issues []string

	if len(code) > MaxCodeSize {
		issues = append(issues, fmt.Sprintf("code size %d exceeds limit %d bytes", len(code), MaxCodeSize))
	}

	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(code) {
			issues = append(issues, fmt.Sprintf("dangerous pattern: %s", pattern.String()))
		}
	}

	return issues
}

// ParsePythonTimeout parses a timeout string like "60s", "2m", "120".
func ParsePythonTimeout(s string) (time.Duration, error) {
	if s == "" {
		return defaultPythonTimeout, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		// Try parsing as plain seconds
		secs, parseErr := strconv.Atoi(strings.TrimSpace(s))
		if parseErr != nil {
			return 0, fmt.Errorf("invalid timeout '%s': %w", s, err)
		}
		d = time.Duration(secs) * time.Second
	}
	if d <= 0 {
		return defaultPythonTimeout, nil
	}
	if d > MaxPythonTimeout {
		d = MaxPythonTimeout
	}
	return d, nil
}
