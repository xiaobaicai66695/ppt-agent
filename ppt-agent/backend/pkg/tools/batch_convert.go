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

// BatchConvertTool converts all PPTX in a workDir to JPG images and a merged PDF.
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

// convertMu serializes concurrent calls to the same workDir to prevent
// filesystem races (both writing to qa_images/, both overwriting merged.pdf).
var convertMu sync.Map // key: workDir string -> struct{}

func convertLock(wd string) func() {
	v, _ := convertMu.LoadOrStore(wd, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	mu.Lock()
	return func() { mu.Unlock() }
}

// RunBatchConvert converts all PPTX in workDir to merged PDF + JPG images.
// It makes a single Python call that handles both PDF and JPG in one
// LibreOffice invocation (the script merges PPTX, then converts to PDF and JPG
// in the same run), avoiding the previous two-step → two-LibreOffice pattern.
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

	// Single Python invocation: it merges all PPTX, runs LibreOffice once,
	// and produces both merged.pdf (persisted to workDir) and qa_images/*.jpg.
	// The --pdf-out flag tells the script where to persist the PDF.
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

	// Now generate JPG images in a second call (incremental — only missing ones).
	// The --files flag ensures the script merges only those files first,
	// then converts with a single LibreOffice startup.
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
