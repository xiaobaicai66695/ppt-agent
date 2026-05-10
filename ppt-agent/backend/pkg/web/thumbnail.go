package web

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var thumbCache sync.Map

// GenerateThumbnail reads a pre-generated JPEG from qa_images/.
// The JPEG is generated eagerly when the .pptx file is first detected (by pollProgress)
// or during QA phase. No on-demand LibreOffice here.
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
	if err != nil {
		// Not yet generated — eagerly convert all PPTX to JPG now
		GenerateQAImages(workDir)
		jpeg, err = os.ReadFile(jpegPath)
		if err != nil {
			return nil, fmt.Errorf("thumbnail not yet generated: %w", err)
		}
	}

	thumbCache.Store(pptxPath, jpeg)
	return jpeg, nil
}

// GenerateQAImages runs the Python PPTX→JPG converter for all PPTX files in the workDir.
// Call this eagerly when new files are detected so thumbnails and QA reuse the same images.
func GenerateQAImages(workDir string) {
	qaDir := filepath.Join(workDir, "qa_images")
	os.MkdirAll(qaDir, 0755)

	converter := findConverterPy(workDir)
	if converter == "" {
		return
	}

	// Check if all PPTX files already have matching JPGs
	if allJPGsExist(workDir, qaDir) {
		return
	}

	pythonBin := "/root/pptx_env/bin/python"
	cmd := exec.Command(pythonBin, converter,
		"--pptx-dir", workDir,
		"--output-dir", qaDir,
		"--dpi", "150")
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("[Thumbnail] QA 图片生成失败: %v (输出: %s)", err, string(out))
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
	return len(entries) > 0
}

func findConverterPy(wd string) string {
	for i := 1; i <= 8; i++ {
		up := ""
		for j := 0; j < i; j++ {
			up += "../"
		}
		for _, sub := range []string{
			"pkg/tools/qa/pptx_qa_converter.py",
			"backend/pkg/tools/qa/pptx_qa_converter.py",
		} {
			p := filepath.Join(wd, up, sub)
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	return ""
}
