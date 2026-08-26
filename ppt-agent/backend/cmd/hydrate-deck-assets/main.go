package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
	"github.com/joho/godotenv"
)

func main() {
	workDir := flag.String("work-dir", "", "absolute task work directory containing tasks.json")
	flag.Parse()
	if strings.TrimSpace(*workDir) == "" {
		fatal("--work-dir is required")
	}
	absWorkDir, err := filepath.Abs(*workDir)
	if err != nil {
		fatal(err.Error())
	}
	if currentDir, cwdErr := os.Getwd(); cwdErr == nil {
		if loadErr := godotenv.Load(filepath.Join(currentDir, ".env")); loadErr != nil && !os.IsNotExist(loadErr) {
			fatal(loadErr.Error())
		}
	}
	manifest, err := deck.ReadTasksManifest(absWorkDir)
	if err != nil {
		fatal(err.Error())
	}
	counts, err := deck.MaterializePlannedDeckAssets(context.Background(), absWorkDir, manifest)
	if err != nil {
		fatal(err.Error())
	}
	if counts.Backgrounds > 0 || counts.Images > 0 {
		if err := deck.WriteTasksManifest(absWorkDir, manifest); err != nil {
			fatal(err.Error())
		}
	}
	payload, _ := json.Marshal(map[string]any{
		"ok":                       true,
		"materialized_backgrounds": counts.Backgrounds,
		"materialized_images":      counts.Images,
	})
	fmt.Println(string(payload))
}

func fatal(message string) {
	payload, _ := json.Marshal(map[string]any{"ok": false, "error": message})
	fmt.Fprintln(os.Stderr, string(payload))
	os.Exit(1)
}
