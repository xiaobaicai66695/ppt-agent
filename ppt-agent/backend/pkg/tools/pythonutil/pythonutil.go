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

// GetPythonBinary 返回配置的 Python 二进制文件路径。
// 如果设置了 PYTHON_BIN 环境变量则使用该值，否则回退到平台特定的默认值。
func GetPythonBinary() string {
	if bin := os.Getenv("PYTHON_BIN"); bin != "" {
		return bin
	}
	if runtime.GOOS == "windows" {
		return "python"
	}
	return "/root/pptx_env/bin/python"
}

// FindConverterPy 搜索 pptx_qa_converter.py 脚本。
// 它首先检查 PROJECT_ROOT 环境变量，然后从给定 workDir 向上搜索最多 8 级父目录，
// 在两个常规子路径中查找。
// 返回转换器脚本的绝对路径，如果未找到则返回空字符串。
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
			// 解析为绝对路径并验证无路径遍历
			abs, err := filepath.Abs(c)
			if err != nil {
				continue
			}
			clean := filepath.Clean(abs)
			// 禁止遍历到绝对根目录之外
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
