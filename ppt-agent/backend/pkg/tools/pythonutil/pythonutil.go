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

package pythonutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// GetPythonBinary returns the configured Python binary path.
// Uses PYTHON_BIN env var if set, falls back to platform-specific defaults.
func GetPythonBinary() string {
	if bin := os.Getenv("PYTHON_BIN"); bin != "" {
		return bin
	}
	if runtime.GOOS == "windows" {
		return "python"
	}
	return "/root/pptx_env/bin/python"
}

// FindConverterPy searches for the pptx_qa_converter.py script.
// It checks PROJECT_ROOT env var first, then walks up to 8 parent directories
// from the given workDir, looking in two conventional sub-paths.
// Returns the absolute path to the converter script, or "" if not found.
func FindConverterPy(workDir string) string {
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot != "" {
		c := filepath.Join(projectRoot, "pkg", "tools", "qa", "pptx_qa_converter.py")
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	for i := 1; i <= 8; i++ {
		up := strings.Repeat("../", i)
		for _, sub := range []string{
			"pkg/tools/qa/pptx_qa_converter.py",
			"backend/pkg/tools/qa/pptx_qa_converter.py",
		} {
			c := filepath.Join(workDir, up, sub)
			// Resolve to absolute path and validate no path traversal
			abs, err := filepath.Abs(c)
			if err != nil {
				continue
			}
			clean := filepath.Clean(abs)
			// Disallow traversal outside an absolute root
			if !strings.HasPrefix(clean, filepath.VolumeName(clean)) {
				continue
			}
			if _, err := os.Stat(clean); err == nil {
				return clean
			}
		}
	}
	return ""
}
