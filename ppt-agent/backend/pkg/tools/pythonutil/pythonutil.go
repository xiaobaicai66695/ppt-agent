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
	"strings"
)

// GetPythonBinary 返回配置的 Python 二进制文件路径。
// 如果设置了 PYTHON_BIN 环境变量则使用该值；否则按常见 Linux 部署路径探测。
func GetPythonBinary() string {
	if bin := os.Getenv("PYTHON_BIN"); bin != "" {
		return bin
	}
	candidates := []string{
		"/root/pptx_env/bin/python",
		"/home/ubuntu/.openclaw/workspace/venvs/pptx-env/bin/python",
		"/usr/bin/python3",
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return candidates[0]
}

// FindConverterPy locates the shared PPTX-to-JPEG converter used by thumbnail delivery.
func FindConverterPy(workDir string) string {
	if projectRoot := os.Getenv("PROJECT_ROOT"); projectRoot != "" {
		candidate := filepath.Join(projectRoot, "pkg", "tools", "qa", "pptx_qa_converter.py")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	for i := 1; i <= 8; i++ {
		up := strings.Repeat("../", i)
		for _, sub := range []string{"pkg/tools/qa/pptx_qa_converter.py", "backend/pkg/tools/qa/pptx_qa_converter.py"} {
			candidate, err := filepath.Abs(filepath.Join(workDir, up, sub))
			if err != nil {
				continue
			}
			if _, err := os.Stat(filepath.Clean(candidate)); err == nil {
				return filepath.Clean(candidate)
			}
		}
	}
	return ""
}
