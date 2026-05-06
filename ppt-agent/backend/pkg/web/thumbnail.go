package web

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var thumbCache sync.Map

func GenerateThumbnail(pptxPath string) ([]byte, error) {
	if cached, ok := thumbCache.Load(pptxPath); ok {
		return cached.([]byte), nil
	}

	dir, err := os.MkdirTemp("", "ppt-thumb-*")
	if err != nil {
		return nil, fmt.Errorf("mkdir temp: %w", err)
	}
	defer os.RemoveAll(dir)

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
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			if filepath.Ext(e.Name()) == ".pdf" {
				pdfPath = filepath.Join(dir, e.Name())
				break
			}
		}
	}

	thumbPrefix := filepath.Join(dir, "thumb")
	ppm := exec.Command("pdftoppm", "-jpeg", "-r", "150", "-f", "1", "-l", "1", pdfPath, thumbPrefix)
	if out, err := ppm.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pdftoppm: %w (output: %s)", err, string(out))
	}

	jpegPath := thumbPrefix + "-1.jpg"
	jpeg, err := os.ReadFile(jpegPath)
	if err != nil {
		return nil, fmt.Errorf("read jpeg: %w", err)
	}

	thumbCache.Store(pptxPath, jpeg)
	return jpeg, nil
}
