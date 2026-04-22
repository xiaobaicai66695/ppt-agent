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

package command

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/cloudwego/eino-ext/components/tool/commandline"

	"github.com/cloudwego/ppt-agent/pkg/params"
)

// WorkDirFunc 定义获取工作目录的函数类型
type WorkDirFunc func(ctx context.Context) string

// WorkDirBackend 是一个支持动态工作目录的 local backend 包装器
type WorkDirBackend struct {
	workDirFunc WorkDirFunc
}


// NewWorkDirBackend 创建支持工作目录的 backend
func NewWorkDirBackend(_ context.Context, _ *struct{}) (*WorkDirBackend, error) {
	return &WorkDirBackend{}, nil
}

// SetWorkDirFunc 设置获取工作目录的函数
func (b *WorkDirBackend) SetWorkDirFunc(fn WorkDirFunc) {
	b.workDirFunc = fn
}

// getWorkDir 从 context 获取工作目录
func (b *WorkDirBackend) getWorkDir(ctx context.Context) string {
	if b.workDirFunc != nil {
		if wd := b.workDirFunc(ctx); wd != "" {
			return wd
		}
	}
	if wd, ok := params.GetTypedContextParams[string](ctx, params.WorkDirSessionKey); ok {
		return wd
	}
	if dir, err := os.Getwd(); err == nil {
		return dir
	}
	return ""
}

// LocalOperator 本地命令行操作实现
type LocalOperator struct{}

// 确保 LocalOperator 实现了 commandline.Operator 接口
var _ commandline.Operator = (*LocalOperator)(nil)

// ReadFile 读取文件内容
func (l *LocalOperator) ReadFile(ctx context.Context, path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return err.Error(), nil
	}
	return string(b), nil
}

// WriteFile 写入文件内容
func (l *LocalOperator) WriteFile(ctx context.Context, path, content string) error {
	return os.WriteFile(path, []byte(content), 0o666)
}

// IsDirectory 检查路径是否为目录
func (l *LocalOperator) IsDirectory(ctx context.Context, path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// Exists 检查路径是否存在
func (l *LocalOperator) Exists(ctx context.Context, path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// RunCommand 执行命令
func (l *LocalOperator) RunCommand(ctx context.Context, command []string) (*commandline.CommandOutput, error) {
	wd := l.GetWorkDir(ctx)

	var shellCmd []string
	switch runtime.GOOS {
	case "windows":
		shellCmd = append([]string{"cmd.exe", "/C"}, command...)
	default:
		shellCmd = []string{"/bin/sh", "-c", strings.Join(command, " ")}
	}

	cmd := exec.CommandContext(ctx, shellCmd[0], shellCmd[1:]...)
	if wd != "" {
		cmd.Dir = wd
	}

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.Stdout = outBuf
	cmd.Stderr = errBuf
	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("internal error:\ncommand: %v\n\nerr: %v\n\nexec error: %v", cmd.String(), err, errBuf.String())
		return nil, err
	}
	return &commandline.CommandOutput{
		Stdout: outBuf.String(),
		Stderr: errBuf.String(),
	}, nil
}

// GetWorkDir 获取工作目录
func (l *LocalOperator) GetWorkDir(ctx context.Context) string {
	wd, ok := params.GetTypedContextParams[string](ctx, params.WorkDirSessionKey)
	if ok {
		return wd
	}
	// fallback 到当前目录
	if dir, err := os.Getwd(); err == nil {
		return dir
	}
	return ""
}

// SetWorkDir 设置工作目录到 context
func (l *LocalOperator) SetWorkDir(ctx context.Context, dir string) context.Context {
	return params.SetTypedContextParams(ctx, params.WorkDirSessionKey, dir)
}
