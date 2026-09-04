package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cloudwego/eino/schema"

	agentutils "github.com/cloudwego/ppt-agent/pkg/runtime/model"
)

func buildJudgeInput(suite string, c benchCase, output modelOutput) (judgeInput, error) {
	rubric, err := os.ReadFile(filepath.Join(projectRoot(), "benchmark", "rubrics", suite+".md"))
	if err != nil {
		return judgeInput{}, err
	}
	var rawCase any
	if err := json.Unmarshal(c.Raw, &rawCase); err != nil {
		return judgeInput{}, err
	}
	return judgeInput{
		Case:        rawCase,
		Suite:       suite,
		Rubric:      string(rubric),
		ModelOutput: output,
		ScoringScale: []string{
			"1 = 完全失败或违反硬性约束",
			"2 = 主要目标失败，只完成少量要求",
			"3 = 基本可用，但存在明显质量或约束问题",
			"4 = 良好，满足主要要求，仅有小问题",
			"5 = 优秀，完整满足要求且质量高",
		},
		RequiredOutputSchema: map[string]any{
			"case_id":           "string",
			"suite":             suite,
			"score":             "integer 1-5",
			"pass":              "boolean, true only when score >= 4 and no critical_failures",
			"dimension_scores":  "object of integer 1-5 scores",
			"strengths":         []string{"short strings"},
			"weaknesses":        []string{"short strings"},
			"critical_failures": []string{"short strings"},
			"recommended_fix":   "short string",
		},
	}, nil
}

func callJudge(ctx context.Context, input judgeInput, opt options) judgeResult {
	data, _ := json.MarshalIndent(input, "", "  ")
	prompt := "你是 PPT Agent benchmark judge。你不会执行代码，只根据输入 case、Agent 输出、期望约束和评分 rubric 评分。必须严格输出 JSON，不要输出 Markdown。\n\n" + string(data)
	ctx, cancel := context.WithTimeout(ctx, opt.timeout)
	defer cancel()
	model, err := agentutils.NewFallbackToolCallingChatModel(ctx,
		agentutils.WithTextModel(),
		agentutils.WithAPIKeyForProvider(benchmarkModelProvider, benchmarkAPIKey()),
		agentutils.WithMaxTokens(4096),
		agentutils.WithTemperature(0),
		agentutils.WithTopP(0),
	)
	if err != nil {
		return judgeResult{Error: err.Error()}
	}
	resp, err := model.Generate(ctx, []*schema.Message{schema.UserMessage(prompt)})
	if err != nil {
		return judgeResult{Error: err.Error()}
	}
	content := ""
	if resp != nil {
		content = strings.TrimSpace(resp.Content)
	}
	var result judgeResult
	if err := unmarshalJSONObject(content, &result); err != nil {
		return judgeResult{RawContent: content, Error: err.Error()}
	}
	return normalizeJudgeResult(result)
}

const benchmarkModelProvider = "deepseek"

func benchmarkAPIKey() string {
	for _, name := range []string{"PPT_BENCH_JUDGE_API_KEY"} {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}
	return ""
}

func normalizeJudgeResult(result judgeResult) judgeResult {
	if result.Score < 1 || result.Score > 5 {
		result.Error = "judge score must be 1-5"
	}
	if len(result.CriticalFailures) > 0 && result.Score > 2 {
		result.Score = 2
		result.Pass = false
	}
	if result.Score >= 4 && len(result.CriticalFailures) == 0 {
		result.Pass = true
	}
	return result
}

func unmarshalJSONObject(value string, target any) error {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "```json")
	value = strings.TrimPrefix(value, "```")
	value = strings.TrimSuffix(value, "```")
	value = strings.TrimSpace(value)
	if err := json.Unmarshal([]byte(value), target); err == nil {
		return nil
	}
	re := regexp.MustCompile(`(?s)\{.*\}`)
	match := re.FindString(value)
	if match == "" {
		return fmt.Errorf("no JSON object in judge content")
	}
	return json.Unmarshal([]byte(match), target)
}
