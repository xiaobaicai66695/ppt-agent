package main

import (
	"encoding/json"
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

func TestBenchmarkDatasetCoverage(t *testing.T) {
	const minimumCasesPerSuite = 10
	suites := []string{"router", "planner", "reviewer", "fixer"}
	datasets := []string{"test", "validation"}
	seenIDs := make(map[string]string)
	seenRequests := make(map[string]string)

	for _, dataset := range datasets {
		for _, suite := range suites {
			casesRoot := filepath.Join("..", "..", "..", "benchmark", "cases", suite)
			if dataset == "validation" {
				casesRoot = filepath.Join("..", "..", "..", "benchmark", "validation_cases", suite)
			}
			cases, err := loadCases(suite, options{dataset: dataset, suite: suite, casesPath: casesRoot})
			if err != nil {
				t.Fatalf("load %s/%s cases: %v", dataset, suite, err)
			}
			if len(cases) < minimumCasesPerSuite {
				t.Fatalf("%s/%s has %d cases, want at least %d", dataset, suite, len(cases), minimumCasesPerSuite)
			}
			for _, c := range cases {
				if strings.TrimSpace(c.ID) == "" {
					t.Fatalf("%s/%s contains an empty case id", dataset, suite)
				}
				if previous, exists := seenIDs[c.ID]; exists {
					t.Fatalf("duplicate case id %q in %s and %s/%s", c.ID, previous, dataset, suite)
				}
				seenIDs[c.ID] = dataset + "/" + suite

				request := benchmarkCaseRequest(c.Input)
				if request == "" {
					t.Fatalf("%s/%s case %q has no user request/message", dataset, suite, c.ID)
				}
				key := suite + "\x00" + request
				if previous, exists := seenRequests[key]; exists && previous != dataset {
					t.Fatalf("%s/%s reuses an exact request across test and validation: %q", suite, dataset, request)
				}
				seenRequests[key] = dataset
			}
		}
	}
}

func benchmarkCaseRequest(raw json.RawMessage) string {
	var input struct {
		UserRequest string `json:"user_request"`
		UserMessage string `json:"user_message"`
	}
	if err := json.Unmarshal(raw, &input); err != nil {
		return ""
	}
	return strings.TrimSpace(firstNonEmpty(input.UserRequest, input.UserMessage))
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
