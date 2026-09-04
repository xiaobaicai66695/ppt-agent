package web

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/ppt-agent/pkg/runtime/task"
	"github.com/cloudwego/ppt-agent/pkg/tools/pythonutil"
	"github.com/cloudwego/ppt-agent/pkg/utils/logger"
)

// thumbMu 序列化每个 workDir 的转换调用，防止并发缩略图请求争夺同一个 soffice 进程。
var thumbMu sync.Map // key: workDir string -> *sync.Mutex

func thumbLock(workDir string) func() {
	v, _ := thumbMu.LoadOrStore(workDir, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	mu.Lock()
	return func() { mu.Unlock() }
}

var thumbCache sync.Map

type thumbnailConverter func(workDir, qaDir string, files []string) error

// GenerateThumbnail 从 qa_images/ 读取预生成的 JPEG。
// 首次请求时，增量转换缺失的文件（单次增量转换）。
// 调用者负责传递完整的 .pptx 文件路径。
func GenerateThumbnail(pptxPath string) ([]byte, error) {
	return generateThumbnail(pptxPath, func(workDir, qaDir string, files []string) error {
		converter := findConverterPy(workDir)
		if converter == "" {
			return fmt.Errorf("thumbnail converter not found from work directory %s", workDir)
		}
		return runConverter(converter, workDir, qaDir, files)
	})
}

func generateThumbnail(pptxPath string, convert thumbnailConverter) ([]byte, error) {
	if cached, ok := thumbCache.Load(pptxPath); ok {
		return cached.([]byte), nil
	}

	workDir := filepath.Dir(pptxPath)
	qaDir := filepath.Join(workDir, "qa_images")

	base := filepath.Base(pptxPath)
	ext := filepath.Ext(base)
	stem := base[:len(base)-len(ext)]
	jpegPath := filepath.Join(qaDir, stem+".jpg")

	jpeg, err := os.ReadFile(jpegPath)
	if err == nil {
		thumbCache.Store(pptxPath, jpeg)
		return jpeg, nil
	}

	// 浏览器请求和后台预生成可能同时到达。拿锁后必须再次检查，
	// 避免前一个调用已经生成图片后再次启动转换进程。
	release := thumbLock(workDir)
	defer release()
	if cached, ok := thumbCache.Load(pptxPath); ok {
		return cached.([]byte), nil
	}
	if jpeg, err = os.ReadFile(jpegPath); err == nil {
		thumbCache.Store(pptxPath, jpeg)
		return jpeg, nil
	}

	if err := os.MkdirAll(qaDir, 0755); err != nil {
		return nil, fmt.Errorf("create thumbnail directory: %w", err)
	}
	if err := convert(workDir, qaDir, []string{base}); err != nil {
		return nil, fmt.Errorf("thumbnail conversion failed: %w", err)
	}

	jpeg, err = os.ReadFile(jpegPath)
	if err != nil {
		return nil, fmt.Errorf("thumbnail not yet generated: %w", err)
	}
	thumbCache.Store(pptxPath, jpeg)
	return jpeg, nil
}

func (s *Server) prepareThumbnail(taskID, workDir, filename string) {
	pptxPath := filepath.Join(workDir, filename)
	ts := s.tasks.GetTaskState(taskID)
	if ts == nil {
		return
	}

	if _, err := GenerateThumbnail(pptxPath); err != nil {
		logger.Warn("thumbnail_prepare_failed", "task_id", taskID, "file", filename, "error", err.Error())
		ts.Broadcast(task.SSERichEvent{
			Type:  "thumbnail_error",
			Files: []string{filename},
			Error: err.Error(),
		})
		return
	}
	ts.Broadcast(task.SSERichEvent{
		Type:  "thumbnail_ready",
		Files: []string{filename},
	})
}

// GenerateQAImages 对 workDir 中所有 PPTX 文件运行 Python PPTX→JPG 转换器。
// 使用 per-workDir 锁防止并发进程竞争。
// 在开始 QA 或需要完整批量转换时调用此函数。
func GenerateQAImages(workDir string) {
	release := thumbLock(workDir)
	defer release()

	qaDir := filepath.Join(workDir, "qa_images")
	os.MkdirAll(qaDir, 0755)

	converter := findConverterPy(workDir)
	if converter == "" {
		logger.Warn("thumbnail_converter_not_found", "workDir", workDir)
		return
	}

	missing := findMissingJPGs(workDir, qaDir)
	if len(missing) > 0 {
		if err := runConverter(converter, workDir, qaDir, missing); err != nil {
			return
		}
	}

	// 如果现在全部已转换，则完成。否则回退到完整批量。
	if allJPGsExist(workDir, qaDir) {
		return
	}
	_ = runConverter(converter, workDir, qaDir, nil)
}

// findMissingJPGs 返回在 qaDir 中没有对应 JPG 的 PPTX 文件名列表（不含路径）。
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

// convertMissingFiles 仅转换指定的 PPTX 文件名。
// 调用者必须在调用前持有 thumbLock(workDir)。
func convertMissingFiles(workDir string, pptxFilenames []string) {
	release := thumbLock(workDir)
	defer release()

	converter := findConverterPy(workDir)
	if converter == "" {
		logger.Warn("thumbnail_converter_not_found", "workDir", workDir)
		return
	}
	qaDir := filepath.Join(workDir, "qa_images")
	_ = runConverter(converter, workDir, qaDir, pptxFilenames)
}

// runConverter 调用 Python 转换器脚本。
// 如果 files 为 nil，转换 workDir 中的所有 PPTX（完整批量）。
// 如果 files 非 nil，仅转换指定文件（增量，先合并）。
// 应用 120 秒超时以防止挂起。
func runConverter(converter, workDir, qaDir string, files []string) error {
	cmdArgs := []string{converter, "--pptx-dir", workDir, "--output-dir", qaDir, "--dpi", "150"}
	if len(files) > 0 {
		cmdArgs = append(cmdArgs, "--files")
		cmdArgs = append(cmdArgs, files...)
	}

	pythonBin := pythonutil.GetPythonBinary()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, pythonBin, cmdArgs...)
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Warn("qa_image_generation_timeout",
				"workDir", workDir,
				"files", len(files))
			return fmt.Errorf("converter timeout after 120s")
		} else {
			logger.Warn("qa_image_generation_failed",
				"err", err.Error(),
				"output", string(out))
			return fmt.Errorf("converter exited with %v: %s", err, strings.TrimSpace(string(out)))
		}
	}
	return nil
}

func allJPGsExist(workDir, qaDir string) bool {
	entries, _ := os.ReadDir(workDir)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".pptx") {
			continue
		}
		jpgName := strings.TrimSuffix(e.Name(), ".pptx") + ".jpg"
		if _, err := os.Stat(filepath.Join(qaDir, jpgName)); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func findConverterPy(wd string) string {
	return pythonutil.FindConverterPy(wd)
}
