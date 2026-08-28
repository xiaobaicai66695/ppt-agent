package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func runCase(suite string, c benchCase, opt options) (caseSummary, error) {
	started := time.Now()
	caseDir := filepath.Join(opt.outPath, suite, safePathName(c.ID))
	if err := os.MkdirAll(caseDir, 0o755); err != nil {
		return caseSummary{}, err
	}
	if err := os.WriteFile(filepath.Join(caseDir, "case.json"), c.Raw, 0o644); err != nil {
		return caseSummary{}, err
	}

	var output modelOutput
	if opt.step == "model" || opt.step == "all" {
		ctx, cancel := context.WithTimeout(context.Background(), opt.timeout)
		defer cancel()
		agentOutput := executeAgent(ctx, suite, c, opt, caseDir)
		agentOutput.DurationMS = time.Since(started).Milliseconds()
		output = compactModelOutput(agentOutput)
		if err := writeJSON(filepath.Join(caseDir, "model_output.json"), output); err != nil {
			return caseSummary{}, err
		}
		if err := writeJSON(filepath.Join(caseDir, "trace.json"), agentOutput); err != nil {
			return caseSummary{}, err
		}
	} else {
		if err := readJSON(filepath.Join(caseDir, "model_output.json"), &output); err != nil {
			return caseSummary{}, err
		}
	}

	cs := caseSummary{CaseID: c.ID, Suite: suite, Error: output.Error}
	if opt.step == "model" {
		return cs, nil
	}
	ji, err := buildJudgeInput(suite, c, output)
	if err != nil {
		return cs, err
	}
	if err := writeJSON(filepath.Join(caseDir, "judge_input.json"), ji); err != nil {
		return cs, err
	}
	jr := callJudge(context.Background(), ji, opt)
	if jr.CaseID == "" {
		jr.CaseID = c.ID
	}
	if jr.Suite == "" {
		jr.Suite = suite
	}
	if err := writeJSON(filepath.Join(caseDir, "score.json"), jr); err != nil {
		return cs, err
	}
	cs.Score = jr.Score
	cs.Pass = jr.Pass
	cs.Judged = true
	cs.Error = firstNonEmpty(output.Error, jr.Error)
	return cs, nil
}

func compactModelOutput(output agentOutput) modelOutput {
	return modelOutput{
		CaseID:              output.CaseID,
		Suite:               output.Suite,
		Output:              output.Output,
		Before:              output.Before,
		After:               output.After,
		DeterministicReview: output.DeterministicReview,
		Error:               output.Error,
	}
}

func loadCases(suite string, opt options) ([]benchCase, error) {
	path := opt.casesPath
	if path == "" {
		path = defaultCasesPath(opt.dataset, suite)
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	files, err := caseFiles(path, info.IsDir())
	if err != nil {
		return nil, err
	}

	var cases []benchCase
	for _, file := range files {
		loaded, err := loadCaseFile(file)
		if err != nil {
			return nil, err
		}
		cases = append(cases, loaded...)
	}
	return filterCases(cases, opt), nil
}

func caseFiles(path string, isDir bool) ([]string, error) {
	if !isDir {
		return []string{path}, nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			files = append(files, filepath.Join(path, entry.Name()))
		}
	}
	sort.Strings(files)
	return files, nil
}

func loadCaseFile(file string) ([]benchCase, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var one benchCase
	if err := json.Unmarshal(data, &one); err == nil && one.ID != "" {
		one.Raw = data
		return []benchCase{one}, nil
	}
	var many []benchCase
	if err := json.Unmarshal(data, &many); err != nil {
		return nil, fmt.Errorf("load cases %s: %w", file, err)
	}
	for i := range many {
		many[i].Raw, _ = json.MarshalIndent(many[i], "", "  ")
	}
	return many, nil
}

func filterCases(cases []benchCase, opt options) []benchCase {
	filtered := cases[:0]
	for _, c := range cases {
		if c.ID == "" {
			continue
		}
		if opt.caseID != "" && c.ID != opt.caseID {
			continue
		}
		filtered = append(filtered, c)
		if opt.limit > 0 && len(filtered) >= opt.limit {
			break
		}
	}
	return filtered
}
