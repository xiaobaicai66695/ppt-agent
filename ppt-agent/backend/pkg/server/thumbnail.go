package server

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var thumbCache sync.Map // path → []byte

// GenerateThumbnail converts the first slide of a PPTX to a 150-DPI JPEG thumbnail.
// Results are cached in-memory by PPTX path.
func GenerateThumbnail(pptxPath string) ([]byte, error) {
	if cached, ok := thumbCache.Load(pptxPath); ok {
		return cached.([]byte), nil
	}

	dir, err := os.MkdirTemp("", "ppt-thumb-*")
	if err != nil {
		return nil, fmt.Errorf("mkdir temp: %w", err)
	}
	defer os.RemoveAll(dir)

	// Step 1: PPTX → PDF via LibreOffice headless
	pdfPath := filepath.Join(dir, "slide.pdf")
	lo := exec.Command("libreoffice",
		"--headless", "--norestore", "--invisible",
		"--convert-to", "pdf",
		"--outdir", dir,
		pptxPath,
	)
	if out, err := lo.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("libreoffice: %w (output: %s)", err, string(out))
	}
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		// libreoffice may name the pdf differently; find any .pdf
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			if filepath.Ext(e.Name()) == ".pdf" {
				pdfPath = filepath.Join(dir, e.Name())
				break
			}
		}
	}

	// Step 2: PDF page 1 → JPEG via pdftoppm
	thumbPrefix := filepath.Join(dir, "thumb")
	ppm := exec.Command("pdftoppm", "-jpeg", "-r", "150", "-f", "1", "-l", "1", pdfPath, thumbPrefix)
	if out, err := ppm.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pdftoppm: %w (output: %s)", err, string(out))
	}

	// pdftoppm outputs thumbPrefix-1.jpg
	jpegPath := thumbPrefix + "-1.jpg"
	jpeg, err := os.ReadFile(jpegPath)
	if err != nil {
		return nil, fmt.Errorf("read jpeg: %w", err)
	}

	thumbCache.Store(pptxPath, jpeg)
	return jpeg, nil
}
