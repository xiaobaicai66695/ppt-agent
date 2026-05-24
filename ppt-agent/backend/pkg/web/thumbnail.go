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

	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/tools/pythonutil"
)

// thumbMu serializes conversion calls per workDir to prevent concurrent
// thumbnail requests from fighting over the same soffice process.
var thumbMu sync.Map // key: workDir string -> *sync.Mutex

func thumbLock(workDir string) func() {
	v, _ := thumbMu.LoadOrStore(workDir, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	mu.Lock()
	return func() { mu.Unlock() }
}

var thumbCache sync.Map

// GenerateThumbnail reads a pre-generated JPEG from qa_images/.
// On first request, converts only the missing files (incremental, single-pass).
// The caller is responsible for passing the full path to a .pptx file.
func GenerateThumbnail(pptxPath string) ([]byte, error) {
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

	// Not cached and not on disk — convert, serialized per workDir.
	// No longer runs two passes; one incremental conversion attempt is sufficient.
	convertMissingFiles(workDir, []string{base})

	jpeg, err = os.ReadFile(jpegPath)
	if err != nil {
		return nil, fmt.Errorf("thumbnail not yet generated: %w", err)
	}
	thumbCache.Store(pptxPath, jpeg)
	return jpeg, nil
}

// GenerateQAImages runs the Python PPTX→JPG converter for all PPTX files in the workDir.
// Uses a per-workDir lock to prevent concurrent processes from fighting.
// Call this when starting QA or when a full batch conversion is needed.
func GenerateQAImages(workDir string) {
	release := thumbLock(workDir)
	defer release()

	qaDir := filepath.Join(workDir, "qa_images")
	os.MkdirAll(qaDir, 0755)

	converter := findConverterPy(workDir)
	if converter == "" {
		return
	}

	missing := findMissingJPGs(workDir, qaDir)
	if len(missing) > 0 {
		runConverter(converter, workDir, qaDir, missing)
	}

	// If everything is now converted, done. Otherwise fall back to full batch.
	if allJPGsExist(workDir, qaDir) {
		return
	}
	runConverter(converter, workDir, qaDir, nil)
}

// findMissingJPGs returns the list of PPTX filenames (without path) that have
// no corresponding JPG in qaDir.
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

// convertMissingFiles converts only the specified PPTX filenames.
// The caller must hold thumbLock(workDir) before calling.
func convertMissingFiles(workDir string, pptxFilenames []string) {
	release := thumbLock(workDir)
	defer release()

	converter := findConverterPy(workDir)
	if converter == "" {
		return
	}
	qaDir := filepath.Join(workDir, "qa_images")
	runConverter(converter, workDir, qaDir, pptxFilenames)
}

// runConverter invokes the Python converter script.
// If files is nil, converts all PPTX in workDir (full batch).
// If files is non-nil, converts only those specified files (incremental, merged first).
// A 120-second timeout is applied to prevent hanging.
func runConverter(converter, workDir, qaDir string, files []string) {
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
		} else {
			logger.Warn("qa_image_generation_failed",
				"err", err.Error(),
				"output", string(out))
		}
	}
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
