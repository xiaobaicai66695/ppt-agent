package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func finalizeSummary(summary runSummary) runSummary {
	for _, c := range summary.Cases {
		summary.Total++
		if c.Judged && c.Pass {
			summary.Passed++
		}
		if c.Judged {
			summary.Judged++
			summary.Average += float64(c.Score)
		}
	}
	if summary.Judged > 0 {
		summary.Average /= float64(summary.Judged)
	}
	summary.FinishedAt = time.Now().Format(time.RFC3339)
	return summary
}

func writeRunSummary(outPath string, summary runSummary) error {
	if err := writeJSON(filepath.Join(outPath, "summary.json"), summary); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outPath, "summary.md"), []byte(summaryMarkdown(summary)), 0o644)
}

func summaryMarkdown(summary runSummary) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "# PPT Agent Benchmark Summary\n\n")
	fmt.Fprintf(&b, "- Suite: `%s`\n", summary.Suite)
	fmt.Fprintf(&b, "- Total: %d\n", summary.Total)
	fmt.Fprintf(&b, "- Judged: %d\n", summary.Judged)
	fmt.Fprintf(&b, "- Passed: %d\n", summary.Passed)
	if summary.Judged > 0 {
		fmt.Fprintf(&b, "- Average score: %.2f\n", summary.Average)
	}
	fmt.Fprintf(&b, "\n| Case | Score | Pass | Error |\n")
	fmt.Fprintf(&b, "| --- | --- | --- | --- |\n")
	for _, c := range summary.Cases {
		score := "not judged"
		pass := "not judged"
		if c.Judged {
			score = strconv.Itoa(c.Score)
			pass = fmt.Sprintf("%v", c.Pass)
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", c.CaseID, score, pass, escapePipe(c.Error))
	}
	return b.String()
}

func escapePipe(value string) string {
	value = strings.ReplaceAll(value, "|", "\\|")
	value = strings.ReplaceAll(value, "\n", " ")
	return value
}
