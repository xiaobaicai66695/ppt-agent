package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
)

func projectRoot() string {
	wd, _ := os.Getwd()
	if filepath.Base(wd) == "backend" {
		return filepath.Clean(filepath.Join(wd, ".."))
	}
	if filepath.Base(wd) == "ppt-agent" {
		return wd
	}
	return filepath.Join(wd, "ppt-agent")
}

func loadEnv() {
	root := projectRoot()
	_ = godotenv.Load(filepath.Join(root, ".env"))
	_ = godotenv.Load(filepath.Join(root, "backend", ".env"))
}

func defaultOutPath(dataset, suite string) string {
	runName := time.Now().Format("20060102-150405") + "-" + safePathName(dataset) + "-" + safePathName(suite)
	return filepath.Join(projectRoot(), "benchmark", "runs", runName)
}

func defaultCasesPath(dataset, suite string) string {
	if dataset == "validation" {
		return filepath.Join(projectRoot(), "benchmark", "validation_cases", suite)
	}
	return filepath.Join(projectRoot(), "benchmark", "cases", suite)
}

func readJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func cloneManifest(manifest *deck.TasksManifest) *deck.TasksManifest {
	if manifest == nil {
		return nil
	}
	data, _ := json.Marshal(manifest)
	var out deck.TasksManifest
	_ = json.Unmarshal(data, &out)
	return &out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstEventError(events []deck.AgentEvent) string {
	for _, event := range events {
		if strings.TrimSpace(event.Error) != "" {
			return strings.TrimSpace(event.Error)
		}
	}
	return ""
}

func inferAllowedPages(message string) []int {
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(message, -1)
	seen := map[int]bool{}
	var pages []int
	for _, match := range matches {
		page, err := strconv.Atoi(match)
		if err == nil && page > 0 && page <= 50 && !seen[page] {
			seen[page] = true
			pages = append(pages, page)
		}
	}
	if len(pages) == 0 {
		return []int{1}
	}
	return pages
}

func safePathName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "case"
	}
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}
