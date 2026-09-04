package web

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestGenerateThumbnailRechecksCacheAfterLock(t *testing.T) {
	workDir := t.TempDir()
	pptxPath := filepath.Join(workDir, "1_intro.pptx")
	if err := os.WriteFile(pptxPath, []byte("pptx"), 0600); err != nil {
		t.Fatal(err)
	}
	thumbCache.Delete(pptxPath)

	var conversions atomic.Int32
	conversionStarted := make(chan struct{})
	allowConversion := make(chan struct{})
	convert := func(_ string, qaDir string, _ []string) error {
		if conversions.Add(1) == 1 {
			close(conversionStarted)
		}
		<-allowConversion
		if err := os.WriteFile(filepath.Join(qaDir, "1_intro.jpg"), []byte("jpeg"), 0600); err != nil {
			t.Errorf("write thumbnail: %v", err)
		}
		return nil
	}

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	load := func() {
		defer wg.Done()
		_, err := generateThumbnail(pptxPath, convert)
		errs <- err
	}
	wg.Add(1)
	go load()
	<-conversionStarted
	wg.Add(1)
	go load()
	close(allowConversion)
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("generateThumbnail returned %v", err)
		}
	}
	if got := conversions.Load(); got != 1 {
		t.Fatalf("conversion count = %d, want 1", got)
	}
}

func TestGenerateThumbnailPropagatesConverterError(t *testing.T) {
	workDir := t.TempDir()
	pptxPath := filepath.Join(workDir, "1_intro.pptx")
	if err := os.WriteFile(pptxPath, []byte("pptx"), 0600); err != nil {
		t.Fatal(err)
	}
	thumbCache.Delete(pptxPath)

	_, err := generateThumbnail(pptxPath, func(_, _ string, _ []string) error {
		return errors.New("lock permission denied")
	})
	if err == nil || !strings.Contains(err.Error(), "lock permission denied") {
		t.Fatalf("error = %v, want converter cause", err)
	}
}
