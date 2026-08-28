package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	opt, err := parseOptions()
	if err != nil {
		return err
	}
	loadEnv()
	if opt.outPath == "" {
		opt.outPath = defaultOutPath(opt.dataset, opt.suite)
	}
	if err := os.MkdirAll(opt.outPath, 0o755); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(opt.outPath, "config.json"), runConfig(opt)); err != nil {
		return err
	}

	summary := runSummary{Suite: opt.suite, StartedAt: time.Now().Format(time.RFC3339)}
	for _, suite := range selectedSuites(opt.suite) {
		cases, err := loadCases(suite, opt)
		if err != nil {
			return err
		}
		for _, c := range cases {
			caseSummary, err := runCase(suite, c, opt)
			if err != nil {
				return err
			}
			summary.Cases = append(summary.Cases, caseSummary)
		}
	}
	return writeRunSummary(opt.outPath, finalizeSummary(summary))
}

func parseOptions() (options, error) {
	var opt options
	var timeout string
	flag.StringVar(&opt.dataset, "dataset", "test", "test|validation")
	flag.StringVar(&opt.suite, "s", "all", "router|planner|reviewer|fixer|all")
	flag.StringVar(&opt.step, "p", "model", "model|judge|all")
	flag.StringVar(&opt.casesPath, "cases", "", "case file or directory; defaults to benchmark/cases/<suite> or benchmark/validation_cases/<suite>")
	flag.StringVar(&opt.outPath, "o", "", "run output directory")
	flag.IntVar(&opt.limit, "l", 0, "maximum cases per suite")
	flag.StringVar(&opt.caseID, "case", "", "single case id")
	flag.StringVar(&timeout, "timeout", "45m", "per case timeout")
	flag.Parse()
	if flag.NArg() > 0 {
		return opt, fmt.Errorf("unexpected positional argument %q; use -o <run-dir> to score an existing benchmark output", flag.Arg(0))
	}

	if !validSuite(opt.suite) {
		return opt, fmt.Errorf("unsupported suite %q", opt.suite)
	}
	if !validDataset(opt.dataset) {
		return opt, fmt.Errorf("unsupported dataset %q", opt.dataset)
	}
	if !validStep(opt.step) {
		return opt, fmt.Errorf("unsupported step %q", opt.step)
	}
	d, err := time.ParseDuration(timeout)
	if err != nil {
		return opt, err
	}
	opt.timeout = d
	return opt, nil
}

func runConfig(opt options) map[string]any {
	return map[string]any{
		"dataset": opt.dataset,
		"suite":   opt.suite,
		"step":    opt.step,
		"cases":   opt.casesPath,
		"case":    opt.caseID,
		"limit":   opt.limit,
		"timeout": opt.timeout.String(),
	}
}

func validDataset(dataset string) bool {
	return dataset == "test" || dataset == "validation"
}

func validStep(step string) bool {
	switch step {
	case "model", "judge", "all":
		return true
	default:
		return false
	}
}

func validSuite(suite string) bool {
	switch suite {
	case "all", "router", "planner", "reviewer", "fixer":
		return true
	default:
		return false
	}
}

func selectedSuites(suite string) []string {
	if suite == "all" {
		return []string{"router", "planner", "reviewer", "fixer"}
	}
	return []string{suite}
}
