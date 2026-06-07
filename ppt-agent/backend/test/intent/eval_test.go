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

// Package intent_eval 提供意图分类器的离线回归评估。
//
// 运行方式：
//
//	cd backend
//	go test ./test/intent/ -v -run TestIntentEval -count 1     # 纯规则匹配（无需 API key）
//	EVAL_USE_LLM=true go test ./test/intent/ -v -run TestIntentEval -count 1 -timeout 10m   # 大模型增强
//
// 环境变量：
//
//	EVAL_DATAFILE     测试数据文件路径（默认 ../../../test/intent/testdata/intent_eval.jsonl）
//	EVAL_USE_LLM      设为 "true" 启用大模型增强分类（规则置信度 < 0.85 时走 LLM）
//	EVAL_LLM_LIMIT    最多走 LLM 的用例数（默认 20，设为 0 表示不限制）
//	EVAL_LLM_CONCUR    LLM 并发数（默认 5）
//	EVAL_MODEL        覆盖评估用的 LLM 模型名（默认读取 ARK_MODEL 环境变量）
//	EVAL_TEXT_MODEL   覆盖评估用的轻量文本模型（默认读取 ARK_TEXT_MODEL，回退 ARK_MODEL）
package intent_eval

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"

	"github.com/cloudwego/ppt-agent/pkg/agent/intent"
	"github.com/cloudwego/ppt-agent/pkg/agent/utils"
)

// EvalCase 表示单条人工标注的评估用例。
type EvalCase struct {
	ID                    int    `json:"id"`
	Query                 string `json:"query"`
	ExpectedIntent        string `json:"expected_intent"`
	ExpectedDomain        string `json:"expected_domain"`
	ExpectedComplexityMin int    `json:"expected_complexity_min"`
	ExpectedComplexityMax int    `json:"expected_complexity_max"`
}

// EvalResult 包含单条用例的预测结果和对比信息。
type EvalResult struct {
	Case           EvalCase
	PredIntent     string
	PredDomain     string
	PredComplexity int
	Confidence     float64
	Source         string // "rule" 或 "llm"
	IntentCorrect  bool
	DomainCorrect  bool
	ComplexityOK   bool // true when predicted complexity ∈ [min, max]
}

func TestIntentEval(t *testing.T) {
	// 加载 .env，确保 LLM 模式下 ARK_MODEL / ARK_API_KEY 等环境变量可用
	_ = godotenv.Load(filepath.Join("..", "..", ".env"))

	datafile := os.Getenv("EVAL_DATAFILE")
	if datafile == "" {
		datafile = filepath.Join("..", "..", "..", "test", "intent", "testdata", "intent_eval.jsonl")
	}

	cases, err := loadEvalCases(datafile)
	if err != nil {
		t.Fatalf("加载测试数据失败 %q: %v", datafile, err)
	}
	if len(cases) == 0 {
		t.Fatal("测试数据集为空，请检查 JSONL 文件")
	}

	useLLM := strings.ToLower(os.Getenv("EVAL_USE_LLM")) == "true"
	llmLimit := envInt("EVAL_LLM_LIMIT", 20) // 默认最多 20 条走 LLM
	llmConcur := envInt("EVAL_LLM_CONCUR", 5)
	if llmConcur < 1 {
		llmConcur = 1
	}

	t.Logf("加载 %d 条评估用例 (LLM增强: %v, limit: %d, concur: %d)",
		len(cases), useLLM, llmLimit, llmConcur)

	classifier := buildClassifier(t, useLLM)
	ctx := context.Background()

	results := evaluateCases(t, ctx, classifier, cases, useLLM, llmLimit, llmConcur)

	printReport(t, results, useLLM)
	saveResults(t, results, useLLM)

	// 回归门禁：LLM 模式使用更高基线
	intentAcc := accuracy(results, func(r EvalResult) bool { return r.IntentCorrect })
	domainAcc := accuracy(results, func(r EvalResult) bool { return r.DomainCorrect })
	complexityRate := accuracy(results, func(r EvalResult) bool { return r.ComplexityOK })

	minIntent, minDomain, minComplexity := 0.75, 0.50, 0.60
	if useLLM {
		minIntent, minDomain, minComplexity = 0.85, 0.65, 0.75
	}

	if intentAcc < minIntent {
		t.Errorf("意图准确率 %.1f%% 低于基线 %.1f%%", intentAcc*100, minIntent*100)
	}
	if domainAcc < minDomain {
		t.Errorf("领域准确率 %.1f%% 低于基线 %.1f%%", domainAcc*100, minDomain*100)
	}
	if complexityRate < minComplexity {
		t.Errorf("复杂度区间命中率 %.1f%% 低于基线 %.1f%%", complexityRate*100, minComplexity*100)
	}
}

// evaluateCases 对全部用例求值。
// 当 useLLM 为 true 时，先做一次规则匹配筛选出低置信度用例，再对这些用例并发调 LLM。
func evaluateCases(t *testing.T, ctx context.Context, classifier *intent.Classifier,
	cases []EvalCase, useLLM bool, llmLimit, llmConcur int) []EvalResult {
	t.Helper()

	if !useLLM {
		// 纯规则模式：直接串行跑，速度很快
		results := make([]EvalResult, 0, len(cases))
		for _, c := range cases {
			cr, err := classifier.Classify(ctx, c.Query, 0)
			if err != nil {
				t.Errorf("用例 %d Classify 失败: %v", c.ID, err)
				continue
			}
			results = append(results, makeResult(c, cr, "rule"))
		}
		return results
	}

	// ── LLM 增强模式 ──
	// Step 1: 先创建纯规则分类器，快速筛选出需要 LLM 的用例
	ruleClassifier := intent.NewClassifier(nil, nil)
	type workItem struct {
		ec      EvalCase
		needsLLM bool
	}
	var items []workItem
	for _, c := range cases {
		cr, _ := ruleClassifier.Classify(ctx, c.Query, 0)
		needsLLM := cr.Confidence < 0.85
		items = append(items, workItem{ec: c, needsLLM: needsLLM})
	}

	// Step 2: 统计需要走 LLM 的用例数，限制在 llmLimit 内
	llmIndices := make(map[int]bool)
	llmCount := 0
	for i, item := range items {
		if item.needsLLM && (llmLimit == 0 || llmCount < llmLimit) {
			llmIndices[i] = true
			llmCount++
		}
	}
	needsLLMTotal := 0
	for _, item := range items {
		if item.needsLLM {
			needsLLMTotal++
		}
	}
	t.Logf("  规则筛选: %d 条低置信度, 实际走LLM: %d 条 (limit=%d)", needsLLMTotal, llmCount, llmLimit)

	// Step 3: 并发执行全部用例
	results := make([]EvalResult, len(items))
	var wg sync.WaitGroup
	sem := make(chan struct{}, llmConcur) // 并发控制信号量

	for i, item := range items {
		wg.Add(1)
		go func(idx int, wi workItem) {
			defer wg.Done()

			useLLMForThis := llmIndices[idx]
			source := "rule"
			var cr *intent.ClassificationResult
			var err error

			if useLLMForThis {
				sem <- struct{}{}        // 获取并发槽位
				cr, err = classifier.Classify(ctx, wi.ec.Query, 0)
				<-sem                    // 释放槽位
				source = "llm"
			} else {
				cr, err = ruleClassifier.Classify(ctx, wi.ec.Query, 0)
			}

			if err != nil {
				t.Errorf("用例 %d Classify 失败: %v", wi.ec.ID, err)
				cr = &intent.ClassificationResult{Intent: intent.IntentUnknown}
			}

			results[idx] = makeResult(wi.ec, cr, source)
		}(i, item)
	}
	wg.Wait()

	// 过滤掉失败的（Intent == unknown）
	var final []EvalResult
	for _, r := range results {
		if r.PredIntent != "unknown" || r.Case.Query != "" {
			final = append(final, r)
		}
	}
	if len(final) == 0 {
		return results
	}
	return final
}

func makeResult(c EvalCase, cr *intent.ClassificationResult, source string) EvalResult {
	return EvalResult{
		Case:           c,
		PredIntent:     cr.Intent.String(),
		PredDomain:     cr.Domain.String(),
		PredComplexity: cr.Complexity.Level,
		Confidence:     cr.Confidence,
		Source:         source,
		IntentCorrect:  cr.Intent.String() == c.ExpectedIntent,
		DomainCorrect:  cr.Domain.String() == c.ExpectedDomain,
		ComplexityOK: cr.Complexity.Level >= c.ExpectedComplexityMin &&
			cr.Complexity.Level <= c.ExpectedComplexityMax,
	}
}

// makeTextModelFactory 创建一个轻量级文本模型工厂。
func makeTextModelFactory(modelName string) func(ctx context.Context) (interface {
	Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
}, error) {
	return func(ctx context.Context) (interface {
		Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
	}, error) {
		m, err := utils.NewFallbackToolCallingChatModel(ctx, utils.WithModel(modelName))
		if err != nil {
			return nil, err
		}
		return &textModelAdapter{m}, nil
	}
}

// buildClassifier 根据模式构建分类器。
func buildClassifier(t *testing.T, useLLM bool) *intent.Classifier {
	t.Helper()
	if !useLLM {
		return intent.NewClassifier(nil, nil)
	}

	ctx := context.Background()

	evalModel := os.Getenv("EVAL_MODEL")
	if evalModel == "" {
		evalModel = os.Getenv("ARK_MODEL")
	}
	if evalModel == "" {
		t.Fatal("LLM 模式需要设置 ARK_MODEL 或 EVAL_MODEL 环境变量")
	}

	// 验证模型连通性
	t.Logf("  main model: %s", evalModel)
	testModel, err := utils.NewFallbackToolCallingChatModel(ctx, utils.WithModel(evalModel))
	if err != nil {
		t.Fatalf("创建评估模型失败: %v", err)
	}
	primaryName := safePrimaryModelName(testModel)
	t.Logf("  连通性验证: %s ✓", primaryName)

	// 主模型工厂
	modelFactory := func(ctx context.Context) (model.ToolCallingChatModel, error) {
		return utils.NewFallbackToolCallingChatModel(ctx, utils.WithModel(evalModel))
	}

	// 可选：轻量文本模型
	var textModelFactory func(ctx context.Context) (interface {
		Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
	}, error)

	textModel := os.Getenv("EVAL_TEXT_MODEL")
	if textModel == "" {
		textModel = os.Getenv("ARK_TEXT_MODEL")
	}
	if textModel != "" {
		t.Logf("  text model: %s", textModel)
		textModelFactory = makeTextModelFactory(textModel)
	}

	return intent.NewClassifier(modelFactory, textModelFactory)
}

// textModelAdapter 将 ToolCallingChatModel 适配为纯 Generate 接口。
// 桥接 ...model.Option → ...interface{}，满足 textModelFactory 的返回类型。
type textModelAdapter struct {
	inner model.ToolCallingChatModel
}

func (a *textModelAdapter) Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error) {
	modelOpts := make([]model.Option, 0, len(opts))
	for _, o := range opts {
		if opt, ok := o.(model.Option); ok {
			modelOpts = append(modelOpts, opt)
		}
	}
	return a.inner.Generate(ctx, messages, modelOpts...)
}

// safePrimaryModelName 安全获取主模型名。
func safePrimaryModelName(m model.ToolCallingChatModel) string {
	type nameGetter interface {
		PrimaryModelName() string
	}
	if ng, ok := m.(nameGetter); ok {
		return ng.PrimaryModelName()
	}
	return "unknown"
}

// loadEvalCases 从 JSONL 文件加载评估用例。
func loadEvalCases(path string) ([]EvalCase, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cases []EvalCase
	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		var c EvalCase
		if err := json.Unmarshal([]byte(line), &c); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}
		cases = append(cases, c)
	}
	return cases, scanner.Err()
}

// envInt 从环境变量读取整数，失败时返回默认值。
func envInt(key string, defaultVal int) int {
	s := os.Getenv(key)
	if s == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return defaultVal
	}
	return n
}

// accuracy 计算满足条件的比例。
func accuracy(results []EvalResult, pred func(EvalResult) bool) float64 {
	if len(results) == 0 {
		return 0
	}
	n := 0
	for _, r := range results {
		if pred(r) {
			n++
		}
	}
	return float64(n) / float64(len(results))
}

// printReport 输出详细的评估报告。
func printReport(t *testing.T, results []EvalResult, useLLM bool) {
	total := len(results)
	intentOK := count(results, func(r EvalResult) bool { return r.IntentCorrect })
	domainOK := count(results, func(r EvalResult) bool { return r.DomainCorrect })
	complexityOK := count(results, func(r EvalResult) bool { return r.ComplexityOK })

	// 统计 source 分布
	ruleCount, llmCount := 0, 0
	for _, r := range results {
		switch r.Source {
		case "rule":
			ruleCount++
		case "llm":
			llmCount++
		}
	}

	mode := "纯规则匹配"
	if useLLM {
		mode = fmt.Sprintf("规则 + LLM增强 (llm=%d rule=%d)", llmCount, ruleCount)
	}

	t.Logf("")
	t.Logf("═══════════════════════════════════════════")
	t.Logf("        意图分类器 离线评估报告")
	t.Logf("───────────────────────────────────────────")
	t.Logf("  模式          : %s", mode)
	t.Logf("  总用例数      : %d", total)
	t.Logf("")
	t.Logf("  意图准确率    : %d/%d (%.1f%%)", intentOK, total, float64(intentOK)/float64(total)*100)
	t.Logf("  领域准确率    : %d/%d (%.1f%%)", domainOK, total, float64(domainOK)/float64(total)*100)
	t.Logf("  复杂度区间命中: %d/%d (%.1f%%)", complexityOK, total, float64(complexityOK)/float64(total)*100)
	t.Logf("═══════════════════════════════════════════")

	// 按意图类型分组统计
	t.Logf("")
	t.Logf("── 按意图类型分组 ──")
	for _, intent := range []string{"create", "edit", "extend", "regenerate", "customize", "query", "continue"} {
		s := intentStatsFor(results, intent)
		if s.total > 0 {
			t.Logf("  %-12s : %d/%d (%.0f%%)", intent, s.ok, s.total, float64(s.ok)/float64(s.total)*100)
		}
	}

	// 对走 LLM 的用例单独统计
	if llmCount > 0 {
		t.Logf("")
		t.Logf("── LLM 增强用例独立统计 (%d 条) ──", llmCount)
		var llmResults []EvalResult
		for _, r := range results {
			if r.Source == "llm" {
				llmResults = append(llmResults, r)
			}
		}
		llmIntentOK := count(llmResults, func(r EvalResult) bool { return r.IntentCorrect })
		t.Logf("  意图准确率    : %d/%d (%.1f%%)",
			llmIntentOK, llmCount, float64(llmIntentOK)/float64(llmCount)*100)
	}

	// 打印错误用例（按 ID 排序，前 15 条）
	t.Logf("")
	t.Logf("── 错误用例 (意图/复杂度) ──")
	sort.Slice(results, func(i, j int) bool { return results[i].Case.ID < results[j].Case.ID })
	shown := 0
	for _, r := range results {
		if r.IntentCorrect && r.ComplexityOK {
			continue
		}
		if shown >= 15 {
			t.Logf("  ... 还有 %d 条错误用例未显示",
				len(results)-shown-countCorrectInRemaining(results, shown))
			break
		}
		t.Logf("  [%d] %q  src=%s", r.Case.ID, truncateForReport(r.Case.Query, 60), r.Source)
		if !r.IntentCorrect {
			t.Logf("      意图: pred=%s exp=%s conf=%.2f",
				r.PredIntent, r.Case.ExpectedIntent, r.Confidence)
		}
		if !r.ComplexityOK {
			t.Logf("      复杂度: pred=%d exp=[%d,%d]",
				r.PredComplexity, r.Case.ExpectedComplexityMin, r.Case.ExpectedComplexityMax)
		}
		shown++
	}
}

type intentStat struct{ ok, total int }

func intentStatsFor(results []EvalResult, intent string) intentStat {
	var s intentStat
	for _, r := range results {
		if r.Case.ExpectedIntent == intent {
			s.total++
			if r.IntentCorrect {
				s.ok++
			}
		}
	}
	return s
}

func count(results []EvalResult, pred func(EvalResult) bool) int {
	n := 0
	for _, r := range results {
		if pred(r) {
			n++
		}
	}
	return n
}

func countCorrectInRemaining(results []EvalResult, from int) int {
	n := 0
	for i := from; i < len(results); i++ {
		if results[i].IntentCorrect && results[i].ComplexityOK {
			n++
		}
	}
	return n
}

func truncateForReport(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "…"
}

// ── 结果持久化 ─────────────────────────────────────────────────────────────

// EvalReport 是写入 JSON 文件的最终评估报告结构。
type EvalReport struct {
	GeneratedAt  string            `json:"generated_at"`
	Mode         string            `json:"mode"`
	TotalCases   int               `json:"total_cases"`
	LLMCount     int               `json:"llm_count"`
	RuleCount    int               `json:"rule_count"`
	IntentAcc    float64           `json:"intent_accuracy"`
	DomainAcc    float64           `json:"domain_accuracy"`
	ComplexityAcc float64          `json:"complexity_accuracy"`
	ByIntent     map[string]IntentStat `json:"by_intent"`
	Errors       []ErrorEntry      `json:"errors"`
}

// IntentStat 单个意图的准确率统计。
type IntentStat struct {
	OK    int     `json:"ok"`
	Total int     `json:"total"`
	Rate  float64 `json:"rate"`
}

// ErrorEntry 单条错误用例。
type ErrorEntry struct {
	ID         int     `json:"id"`
	Query      string  `json:"query"`
	Source     string  `json:"source"`
	PredIntent string  `json:"pred_intent"`
	ExpIntent  string  `json:"exp_intent"`
	PredCmplx  int     `json:"pred_complexity"`
	ExpCmplx   string  `json:"exp_complexity"`
	Confidence float64 `json:"confidence"`
}

// saveResults 将本轮评估结果序列化到 test/intent/results/eval-{ts}.json。
func saveResults(t *testing.T, results []EvalResult, useLLM bool) {
	t.Helper()

	// 统计
	llmCount := 0
	for _, r := range results {
		if r.Source == "llm" {
			llmCount++
		}
	}
	mode := "rule"
	if useLLM {
		mode = "rule+llm"
	}

	byIntent := make(map[string]IntentStat)
	for _, intent := range []string{"create", "edit", "extend", "regenerate", "customize", "query", "continue"} {
		s := intentStatsFor(results, intent)
		if s.total > 0 {
			byIntent[intent] = IntentStat{
				OK: s.ok, Total: s.total,
				Rate: float64(s.ok) / float64(s.total),
			}
		}
	}

	var errors []ErrorEntry
	for _, r := range results {
		if r.IntentCorrect && r.ComplexityOK {
			continue
		}
		errors = append(errors, ErrorEntry{
			ID: r.Case.ID, Query: r.Case.Query, Source: r.Source,
			PredIntent: r.PredIntent, ExpIntent: r.Case.ExpectedIntent,
			PredCmplx: r.PredComplexity,
			ExpCmplx: fmt.Sprintf("[%d,%d]", r.Case.ExpectedComplexityMin, r.Case.ExpectedComplexityMax),
			Confidence: r.Confidence,
		})
	}

	report := EvalReport{
		GeneratedAt:  time.Now().Format("2006-01-02T15:04:05"),
		Mode:         mode,
		TotalCases:   len(results),
		LLMCount:     llmCount,
		RuleCount:    len(results) - llmCount,
		IntentAcc:    accuracy(results, func(r EvalResult) bool { return r.IntentCorrect }),
		DomainAcc:    accuracy(results, func(r EvalResult) bool { return r.DomainCorrect }),
		ComplexityAcc: accuracy(results, func(r EvalResult) bool { return r.ComplexityOK }),
		ByIntent:     byIntent,
		Errors:       errors,
	}

	// 输出路径：backend/test/intent/ 向上三级 → test/intent/results/
	resultsDir := filepath.Join("..", "..", "..", "test", "intent", "results")
	if err := os.MkdirAll(resultsDir, 0o755); err != nil {
		t.Logf("  [warn] 创建 results 目录失败: %v", err)
		return
	}

	filename := filepath.Join(resultsDir, fmt.Sprintf("eval-%s.json", time.Now().Format("20060102-150405")))
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Logf("  [warn] 序列化报告失败: %v", err)
		return
	}
	if err := os.WriteFile(filename, data, 0o644); err != nil {
		t.Logf("  [warn] 写入报告文件失败: %v", err)
		return
	}
	t.Logf("  结果已保存: %s", filename)
}
