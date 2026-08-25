/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package plan_benchmark contains the opt-in PPT Planner benchmark workflow.
//
// Default tests are deterministic and low-cost. Tests that call LLM APIs or
// render PPTX files are gated by environment variables; see
// ../../../test/plan_benchmark/README.md.
package plan_benchmark_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"

	agentcommand "github.com/cloudwego/ppt-agent/pkg/agent/command"
	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
	"github.com/cloudwego/ppt-agent/pkg/tools/pythonutil"
)

type benchmarkCase struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Query        string   `json:"query"`
	GoldManifest string   `json:"gold_manifest"`
	MustHave     []string `json:"must_have"`
	MustNot      []string `json:"must_not"`
	casesRoot    string
}

type benchmarkSummary struct {
	CaseID       string                 `json:"case_id"`
	Query        string                 `json:"query"`
	StartedAt    string                 `json:"started_at"`
	DurationMS   int64                  `json:"duration_ms"`
	TotalSlides  int                    `json:"total_slides"`
	DoneSlides   int                    `json:"done_slides"`
	EventCounts  map[string]int         `json:"event_counts"`
	ReviewReport *deck.PlanReviewReport `json:"review_report,omitempty"`
	Error        string                 `json:"error,omitempty"`
}

type judgeResult struct {
	Pass       bool               `json:"pass"`
	Score      float64            `json:"score"`
	Dimensions map[string]float64 `json:"dimensions,omitempty"`
	Strengths  []string           `json:"strengths,omitempty"`
	Issues     []string           `json:"issues,omitempty"`
}

func TestGoldDeckSpecsPassReviewer(t *testing.T) {
	for _, c := range loadBenchmarkCases(t) {
		t.Run(c.ID, func(t *testing.T) {
			manifest := loadManifestFile(t, resolveCasePath(t, c.casesRoot, c.GoldManifest))
			report := deck.ReviewTasksManifest(manifest, "gold:"+c.ID, 1)
			if !report.Passed {
				t.Fatalf("gold DeckSpec should pass reviewer: %s issues=%s", report.Summary, mustJSON(report.Issues))
			}
			if report.TotalSlides != len(manifest.Tasks) {
				t.Fatalf("review total slides = %d, want %d", report.TotalSlides, len(manifest.Tasks))
			}
		})
	}
}

func TestPlannerWorkflowGeneratesReviewedDeckSpec(t *testing.T) {
	if !envBool("PPT_BENCH_RUN_PLANNER") {
		t.Skip("set PPT_BENCH_RUN_PLANNER=true to run the real Planner workflow")
	}
	loadEnv(t)
	cases := limitedCases(loadBenchmarkCases(t), envInt("PPT_BENCH_LIMIT", 0))
	runRoot := plannerRunRoot(t)
	t.Logf("planner benchmark artifacts: %s", runRoot)

	for _, c := range cases {
		t.Run(c.ID, func(t *testing.T) {
			caseDir := filepath.Join(runRoot, safePathName(c.ID))
			if err := os.MkdirAll(caseDir, 0o755); err != nil {
				t.Fatal(err)
			}
			writeFile(t, filepath.Join(caseDir, "query.txt"), []byte(c.Query))

			started := time.Now()
			summary := benchmarkSummary{
				CaseID:      c.ID,
				Query:       c.Query,
				StartedAt:   started.Format(time.RFC3339),
				EventCounts: map[string]int{},
			}
			var events []deck.AgentEvent
			defer func() {
				summary.DurationMS = time.Since(started).Milliseconds()
				writeJSON(t, filepath.Join(caseDir, "summary.json"), summary)
				writeJSON(t, filepath.Join(caseDir, "events.json"), events)
			}()

			ctx, cancel := context.WithTimeout(context.Background(), envDuration("PPT_BENCH_PLANNER_TIMEOUT", 15*time.Minute))
			defer cancel()

			operator := &agentcommand.LocalOperator{}
			ctx = operator.SetWorkDir(ctx, caseDir)
			cfg := &deck.PPTTaskConfig{
				WorkDir:     caseDir,
				TaskID:      "bench-" + c.ID,
				Query:       c.Query,
				Concurrency: 1,
				Operator:    operator,
				SkillsDir:   filepath.Join(projectRoot(t), "skills"),
			}
			agent, err := deck.NewPPTPlannerAgent(ctx, cfg)
			if err != nil {
				summary.Error = err.Error()
				t.Fatal(err)
			}
			result, err := deck.RunPPTPlannerWithCallback(ctx, agent, cfg, c.Query, func(event deck.AgentEvent) {
				events = append(events, event)
				summary.EventCounts[event.Type]++
			})
			if err != nil {
				summary.Error = err.Error()
				t.Fatal(err)
			}

			manifest, err := deck.ReadTasksManifest(caseDir)
			if err != nil {
				summary.Error = err.Error()
				t.Fatal(err)
			}
			report := deck.ReviewTasksManifest(manifest, "generated:"+c.ID, 1)
			summary.ReviewReport = report
			summary.TotalSlides = result.TotalSlides
			summary.DoneSlides = result.DoneSlides
			if !report.Passed {
				t.Fatalf("generated DeckSpec did not pass reviewer: %s issues=%s", report.Summary, mustJSON(report.Issues))
			}
			if _, err := os.Stat(filepath.Join(caseDir, "tasks.review.json")); err != nil {
				t.Fatalf("planner workflow should persist tasks.review.json: %v", err)
			}
		})
	}
}

func TestPlanJudgeAPIScoresDeckSpecs(t *testing.T) {
	if !envBool("PPT_BENCH_RUN_JUDGE") {
		t.Skip("set PPT_BENCH_RUN_JUDGE=true to call the Judge API")
	}
	loadEnv(t)
	apiKey := firstEnv("PLAN_JUDGE_API_KEY", "OPENAI_API_KEY", "ARK_API_KEY")
	model := firstEnv("PLAN_JUDGE_MODEL", "ARK_QA_MODEL", "ARK_TEXT_MODEL", "ARK_MODEL")
	if apiKey == "" {
		t.Skip("set PLAN_JUDGE_API_KEY, OPENAI_API_KEY, or ARK_API_KEY")
	}
	if model == "" {
		t.Skip("set PLAN_JUDGE_MODEL, ARK_QA_MODEL, ARK_TEXT_MODEL, or ARK_MODEL")
	}

	cases := loadBenchmarkCases(t)
	targets := judgeTargets(t, cases)
	if limit := envInt("PPT_BENCH_LIMIT", 0); limit > 0 && len(targets) > limit {
		targets = targets[:limit]
	}
	rubric := string(readFile(t, filepath.Join(projectRoot(t), "test", "plan_benchmark", "rubric.md")))
	minScore := envFloat("PLAN_JUDGE_MIN_SCORE", 7.0)

	for _, target := range targets {
		t.Run(target.caseID, func(t *testing.T) {
			manifest := loadManifestFile(t, target.manifestPath)
			report := deck.ReviewTasksManifest(manifest, target.manifestPath, 1)
			if !report.Passed {
				t.Fatalf("manifest must pass deterministic reviewer before Judge: %s issues=%s", report.Summary, mustJSON(report.Issues))
			}
			result := callJudge(t, apiKey, model, target, manifest, rubric)
			writeJudgeResult(t, target, result)
			if !result.Pass || result.Score < minScore {
				t.Fatalf("judge failed: pass=%v score=%.2f min=%.2f issues=%v", result.Pass, result.Score, minScore, result.Issues)
			}
		})
	}
}

func TestGoldDecksRenderWithSkillScripts(t *testing.T) {
	if !envBool("PPT_BENCH_RUN_GOLD_RENDER") {
		t.Skip("set PPT_BENCH_RUN_GOLD_RENDER=true to render gold manifests")
	}
	pythonBin := pythonutil.GetPythonBinary()
	if info, err := os.Stat(pythonBin); err != nil || info.IsDir() {
		t.Fatalf("python binary %q is unavailable; set PYTHON_BIN to the PPT generator environment", pythonBin)
	}

	for _, c := range loadBenchmarkCases(t) {
		t.Run(c.ID, func(t *testing.T) {
			workDir := t.TempDir()
			src := resolveCasePath(t, c.casesRoot, c.GoldManifest)
			copyFile(t, src, filepath.Join(workDir, "tasks.json"))

			ctx, cancel := context.WithTimeout(context.Background(), envDuration("PPT_BENCH_RENDER_TIMEOUT", 5*time.Minute))
			defer cancel()
			cfg := &deck.PPTTaskConfig{
				WorkDir:     workDir,
				TaskID:      "gold-render-" + c.ID,
				Concurrency: 2,
				SkillsDir:   filepath.Join(projectRoot(t), "skills"),
			}
			result, err := deck.RenderDeckByTaskIDWorkflow(ctx, cfg, nil)
			if err != nil {
				t.Fatal(err)
			}
			if result.TotalSlides == 0 || result.DoneSlides != result.TotalSlides {
				t.Fatalf("render done slides = %d/%d", result.DoneSlides, result.TotalSlides)
			}
		})
	}
}

type judgeTarget struct {
	caseID       string
	name         string
	query        string
	mustHave     []string
	mustNot      []string
	manifestPath string
}

func judgeTargets(t *testing.T, cases []benchmarkCase) []judgeTarget {
	t.Helper()
	if manifest := strings.TrimSpace(os.Getenv("PPT_BENCH_JUDGE_MANIFEST")); manifest != "" {
		query := strings.TrimSpace(os.Getenv("PPT_BENCH_JUDGE_QUERY"))
		return []judgeTarget{{
			caseID:       firstEnv("PPT_BENCH_JUDGE_CASE_ID", "ad_hoc_manifest"),
			name:         "Ad hoc manifest",
			query:        query,
			manifestPath: manifest,
		}}
	}
	if resultsDir := strings.TrimSpace(os.Getenv("PPT_BENCH_JUDGE_RESULTS_DIR")); resultsDir != "" {
		return generatedJudgeTargets(t, resultsDir, cases)
	}
	targets := make([]judgeTarget, 0, len(cases))
	for _, c := range cases {
		targets = append(targets, judgeTarget{
			caseID:       c.ID,
			name:         c.Name,
			query:        c.Query,
			mustHave:     c.MustHave,
			mustNot:      c.MustNot,
			manifestPath: resolveCasePath(t, c.casesRoot, c.GoldManifest),
		})
	}
	return targets
}

func generatedJudgeTargets(t *testing.T, resultsDir string, cases []benchmarkCase) []judgeTarget {
	t.Helper()
	caseByID := map[string]benchmarkCase{}
	for _, c := range cases {
		caseByID[c.ID] = c
	}
	var manifests []string
	if err := filepath.WalkDir(resultsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Base(path) != "tasks.json" {
			return nil
		}
		manifests = append(manifests, path)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	sort.Strings(manifests)
	targets := make([]judgeTarget, 0, len(manifests))
	for _, manifest := range manifests {
		caseID := filepath.Base(filepath.Dir(manifest))
		c := caseByID[caseID]
		targets = append(targets, judgeTarget{
			caseID:       caseID,
			name:         c.Name,
			query:        c.Query,
			mustHave:     c.MustHave,
			mustNot:      c.MustNot,
			manifestPath: manifest,
		})
	}
	return targets
}

func callJudge(t *testing.T, apiKey, model string, target judgeTarget, manifest *deck.TasksManifest, rubric string) judgeResult {
	t.Helper()
	payload := map[string]any{
		"model":       model,
		"temperature": 0,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a strict PPT planning quality judge. Return JSON only."},
			{"role": "user", "content": judgePrompt(target, manifest, rubric)},
		},
	}
	if os.Getenv("PLAN_JUDGE_JSON_MODE") != "0" {
		payload["response_format"] = map[string]string{"type": "json_object"}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, judgeURL(), bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: envDuration("PLAN_JUDGE_TIMEOUT", 90*time.Second)}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("judge API HTTP %d: %s", resp.StatusCode, truncate(string(data), 1000))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("decode judge response: %v body=%s", err, truncate(string(data), 1000))
	}
	if len(parsed.Choices) == 0 {
		t.Fatalf("judge response has no choices: %s", truncate(string(data), 1000))
	}
	var result judgeResult
	if err := unmarshalJSONObject(parsed.Choices[0].Message.Content, &result); err != nil {
		t.Fatalf("decode judge content: %v content=%s", err, parsed.Choices[0].Message.Content)
	}
	return result
}

func judgePrompt(target judgeTarget, manifest *deck.TasksManifest, rubric string) string {
	payload := map[string]any{
		"case": map[string]any{
			"id":        target.caseID,
			"name":      target.name,
			"query":     target.query,
			"must_have": target.mustHave,
			"must_not":  target.mustNot,
		},
		"contract_summary": map[string]any{
			"judge_scope": "Hard schema validity and deterministic DeckSpec review already passed in Go. Judge semantic plan quality only.",
		},
		"rubric":         rubric,
		"tasks_manifest": manifest,
		"required_output_schema": map[string]any{
			"pass":  "boolean",
			"score": "number from 0 to 10",
			"dimensions": map[string]string{
				"intent_coverage":     "0-10",
				"narrative":           "0-10",
				"content_specificity": "0-10",
				"layout_fit":          "0-10",
				"data_and_sources":    "0-10",
				"visual_planning":     "0-10",
				"capacity_control":    "0-10",
			},
			"strengths": []string{"short strings"},
			"issues":    []string{"short strings"},
		},
	}
	data, _ := json.MarshalIndent(payload, "", "  ")
	return string(data)
}

func judgeURL() string {
	base := firstEnv("PLAN_JUDGE_BASE_URL", "OPENAI_BASE_URL", "ARK_BASE_URL")
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	base = strings.TrimRight(base, "/")
	if strings.HasSuffix(base, "/chat/completions") {
		return base
	}
	return base + "/chat/completions"
}

func writeJudgeResult(t *testing.T, target judgeTarget, result judgeResult) {
	t.Helper()
	resultsDir := filepath.Join(projectRoot(t), "test", "plan_benchmark", "results", "judge-"+time.Now().Format("20060102"))
	if err := os.MkdirAll(resultsDir, 0o755); err != nil {
		t.Logf("warning: create judge results dir: %v", err)
		return
	}
	writeJSON(t, filepath.Join(resultsDir, safePathName(target.caseID)+".json"), result)
}

func loadBenchmarkCases(t *testing.T) []benchmarkCase {
	t.Helper()
	casesPath := strings.TrimSpace(os.Getenv("PPT_BENCH_CASES"))
	if casesPath == "" {
		casesPath = filepath.Join(projectRoot(t), "test", "plan_benchmark", "testdata", "planner_cases.json")
	}
	data := readFile(t, casesPath)
	var cases []benchmarkCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatalf("load benchmark cases %s: %v", casesPath, err)
	}
	if len(cases) == 0 {
		t.Fatalf("benchmark cases empty: %s", casesPath)
	}
	root := filepath.Dir(casesPath)
	for i := range cases {
		cases[i].casesRoot = root
	}
	return cases
}

func limitedCases(cases []benchmarkCase, limit int) []benchmarkCase {
	if limit > 0 && len(cases) > limit {
		return cases[:limit]
	}
	return cases
}

func loadManifestFile(t *testing.T, path string) *deck.TasksManifest {
	t.Helper()
	var manifest deck.TasksManifest
	if err := json.Unmarshal(readFile(t, path), &manifest); err != nil {
		t.Fatalf("load manifest %s: %v", path, err)
	}
	return &manifest
}

func resolveCasePath(t *testing.T, root, value string) string {
	t.Helper()
	if filepath.IsAbs(value) {
		return value
	}
	path, err := filepath.Abs(filepath.Join(root, value))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func plannerRunRoot(t *testing.T) string {
	t.Helper()
	if root := strings.TrimSpace(os.Getenv("PPT_BENCH_RESULTS_DIR")); root != "" {
		if err := os.MkdirAll(root, 0o755); err != nil {
			t.Fatal(err)
		}
		return root
	}
	root := filepath.Join(projectRoot(t), "test", "plan_benchmark", "results", "planner-"+time.Now().Format("20060102-150405"))
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func projectRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func loadEnv(t *testing.T) {
	t.Helper()
	root := projectRoot(t)
	_ = godotenv.Load(filepath.Join(root, ".env"))
	_ = godotenv.Load(filepath.Join(root, "backend", ".env"))
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, path, data)
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	data := readFile(t, src)
	writeFile(t, dst, data)
}

func mustJSON(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func unmarshalJSONObject(value string, target any) error {
	value = strings.TrimSpace(value)
	if err := json.Unmarshal([]byte(value), target); err == nil {
		return nil
	}
	re := regexp.MustCompile(`(?s)\{.*\}`)
	match := re.FindString(value)
	if match == "" {
		return fmt.Errorf("no JSON object in content")
	}
	return json.Unmarshal([]byte(match), target)
}

func firstEnv(names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}
	return ""
}

func envBool(name string) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	return value == "1" || value == "true" || value == "yes"
}

func envInt(name string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}

func envFloat(name string, fallback float64) float64 {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(name string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func safePathName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "case"
	}
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "...(truncated)"
}
