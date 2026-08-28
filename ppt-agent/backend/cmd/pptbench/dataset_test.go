package main

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultCasesPathSeparatesTestAndValidation(t *testing.T) {
	testCases := defaultCasesPath("test", "planner")
	validation := defaultCasesPath("validation", "planner")
	if testCases == validation {
		t.Fatal("test and validation cases must not share a directory")
	}
	if !strings.HasSuffix(testCases, filepath.Join("benchmark", "cases", "planner")) {
		t.Fatalf("unexpected test path: %s", testCases)
	}
	if !strings.HasSuffix(validation, filepath.Join("benchmark", "validation_cases", "planner")) {
		t.Fatalf("unexpected validation path: %s", validation)
	}
}

func TestDefaultOutPathIncludesDatasetAndSuite(t *testing.T) {
	path := defaultOutPath("validation", "fixer")
	base := filepath.Base(path)
	if !strings.HasSuffix(base, "-validation-fixer") {
		t.Fatalf("run directory must identify validation fixer run: %s", base)
	}
	if _, err := time.Parse("20060102-150405", strings.TrimSuffix(base, "-validation-fixer")); err != nil {
		t.Fatalf("run directory must start with a timestamp: %s (%v)", base, err)
	}
}

func TestRunConfigRecordsDataset(t *testing.T) {
	config := runConfig(options{dataset: "validation", suite: "all", step: "model", timeout: time.Minute})
	if got := config["dataset"]; got != "validation" {
		t.Fatalf("dataset = %v, want validation", got)
	}
}
