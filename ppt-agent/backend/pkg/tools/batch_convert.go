package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/tools/pythonutil"
)

var batchConvertToolInfo = &schema.ToolInfo{
	Name: "batch_convert",
	Desc: `批量转换工具。将指定工作目录下的所有 PPTX 文件转换为图片（JPG）和合并 PDF。

该工具会：
1. 将所有 PPTX 合并为一个 PPTX（只启动一次 LibreOffice）
2. 转换为 merged.pdf 持久保存到工作目录（供 batch_pdf_review 使用）
3. 单独转换为 JPG 图片保存到 qa_images/ 目录（供前端展示）

建议在每批幻灯片生成完成后立即调用此工具，提前准备好图片和 PDF。`,
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
}

// BatchConvertTool 将 workDir 中的所有 PPTX 转换为 JPG 图片和合并 PDF。
type BatchConvertTool struct {
	op commandline.Operator
}

func NewBatchConvertTool(op commandline.Operator) tool.InvokableTool {
	return &BatchConvertTool{op: op}
}

func (t *BatchConvertTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return batchConvertToolInfo, nil
}

func (t *BatchConvertTool) InvokableRun(ctx context.Context, _ string, _ ...tool.Option) (string, error) {
	type workDirGetter interface{ GetWorkDir(context.Context) string }
	wd := ""
	if getter, ok := t.op.(workDirGetter); ok {
		wd = getter.GetWorkDir(ctx)
	}
	if wd == "" {
		return "", fmt.Errorf("无法获取工作目录")
	}

	result, err := RunBatchConvert(ctx, wd)
	if err != nil {
		return "", err
	}
	b, _ := json.Marshal(result)
	return string(b), nil
}

// convertMu 序列化对同一 workDir 的并发调用，防止文件系统竞态
// （两个都写入 qa_images/，都覆盖 merged.pdf）。
var convertMu sync.Map // key: workDir string -> struct{}

func convertLock(wd string) func() {
	v, _ := convertMu.LoadOrStore(wd, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	mu.Lock()
	return func() { mu.Unlock() }
}

// RunBatchConvert 将 workDir 中的所有 PPTX 转换为合并 PDF 和 JPG 图片。
// 它进行单个 Python 调用，在一次 LibreOffice 启动中处理 PDF 和 JPG
// （脚本先合并 PPTX，然后在同一次运行中转换为 PDF 和 JPG），
// 避免了之前的两步 → 两次 LibreOffice 模式。
func RunBatchConvert(ctx context.Context, wd string) (map[string]any, error) {
	defer convertLock(wd)()

	converter := pythonutil.FindConverterPy(wd)
	if converter == "" {
		return nil, fmt.Errorf("找不到 pptx_qa_converter.py")
	}

	qaDir := filepath.Join(wd, "qa_images")
	if err := os.MkdirAll(qaDir, 0755); err != nil {
		return nil, fmt.Errorf("创建 qa_images 目录失败: %v", err)
	}

	pythonBin := pythonutil.GetPythonBinary()

	// 单次 Python 调用：合并所有 PPTX，运行一次 LibreOffice，
	// 同时生成 merged.pdf（持久化到 workDir）和 qa_images/*.jpg。
	// --pdf-out 标志告诉脚本将 PDF 保存到哪里。
	pdfOut := filepath.Join(wd, "merged.pdf")
	cmdArgs := []string{
		converter,
		"--pptx-dir", wd,
		"--output-dir", qaDir,
		"--pdf-only",
		"--pdf-out", pdfOut,
	}
	cmd := exec.CommandContext(ctx, pythonBin, cmdArgs...)
	cmd.Dir = wd

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("转换失败: %v, stderr: %s", err, stderr.String())
	}

	var result map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("解析转换结果失败: %v", err)
	}

	if errMsg, ok := result["error"].(string); ok && errMsg != "" {
		return nil, fmt.Errorf("%s", errMsg)
	}

	// 现在在第二次调用中生成 JPG 图片（增量——仅缺失的）。
	// --files 标志确保脚本先只合并这些文件，
	// 然后用一次 LibreOffice 启动进行转换。
	missing := findMissingJPGs(wd, qaDir)
	if len(missing) > 0 {
		jpgArgs := []string{
			converter,
			"--pptx-dir", wd,
			"--output-dir", qaDir,
			"--dpi", "150",
			"--files",
		}
		jpgArgs = append(jpgArgs, missing...)

		cmd2 := exec.CommandContext(ctx, pythonBin, jpgArgs...)
		cmd2.Dir = wd

		var jpgStdout, jpgStderr bytes.Buffer
		cmd2.Stdout = &jpgStdout
		cmd2.Stderr = &jpgStderr

		if err := cmd2.Run(); err != nil {
			return nil, fmt.Errorf("JPG 转换失败: %v, stderr: %s", err, jpgStderr.String())
		}

		var jpgResult map[string]any
		if err := json.Unmarshal(jpgStdout.Bytes(), &jpgResult); err != nil {
			return nil, fmt.Errorf("解析 JPG 转换结果失败: %v", err)
		}
		result["slide_images"] = jpgResult["slide_images"]
	}

	return map[string]any{
		"pdf_path":     result["pdf_path"],
		"slide_count":  result["slide_count"],
		"slide_images": result["slide_images"],
		"text_content": result["text_content"],
	}, nil
}

func findMissingJPGs(workDir, qaDir string) []string {
	var missing []string
	entries, _ := os.ReadDir(workDir)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".pptx") {
			continue
		}
		jpgName := strings.TrimSuffix(e.Name(), ".pptx") + ".jpg"
		if _, err := os.Stat(filepath.Join(qaDir, jpgName)); os.IsNotExist(err) {
			missing = append(missing, e.Name())
		}
	}
	return missing
}
