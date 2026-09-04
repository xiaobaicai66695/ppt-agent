package web

import (
	"os"
	"path/filepath"
	"strings"
)

// safePath 将 baseDir 与 filename 连接并返回清理后的绝对路径。
// 如果结果逃逸出 baseDir，则返回错误（防止路径遍历攻击）。
func safePath(baseDir, filename string) (string, error) {
	cleaned := filepath.Clean(filepath.Join(baseDir, filename))
	if !strings.HasPrefix(cleaned, filepath.Clean(baseDir)+string(filepath.Separator)) {
		return "", os.ErrPermission
	}
	return cleaned, nil
}
