// Package chat_benchmark_test provides deterministic contract checks and an
// opt-in real-model benchmark for the production workbench chat reply path.
package chat_benchmark_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"

	agentutils "github.com/cloudwego/ppt-agent/pkg/runtime/model"
	"github.com/cloudwego/ppt-agent/pkg/runtime/web"
)

type chatCase struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Input          web.ChatBenchmarkInput `json:"input"`
	ModelReply     string                 `json:"model_reply,omitempty"`
	MustHave       []string               `json:"must_have"`
	MustMentionAny []string               `json:"must_mention_any,omitempty"`
	MustNot        []string               `json:"must_not"`
	MinLinks       int                    `json:"min_links,omitempty"`
}

type fixedChatModel struct{ reply string }

func (m fixedChatModel) Generate(_ context.Context, _ []*schema.Message, _ ...interface{}) (*schema.Message, error) {
	return schema.AssistantMessage(m.reply, nil), nil
}

// liveChatModelAdapter keeps the benchmark adapter deliberately narrow while
// delegating actual inference to the production text-model fallback chain.
type liveChatModelAdapter struct {
	inner   einomodel.ToolCallingChatModel
	lastErr error
}

func (m *liveChatModelAdapter) Generate(ctx context.Context, messages []*schema.Message, _ ...interface{}) (*schema.Message, error) {
	message, err := m.inner.Generate(ctx, messages)
	m.lastErr = err
	return message, err
}

func TestChatBenchmarkDatasetCoverageAndIsolation(t *testing.T) {
	testCases := loadCases(t, "cases.json")
	validationCases := loadCases(t, "validation_cases.json")
	if len(testCases) < 10 || len(validationCases) < 10 {
		t.Fatalf("chat benchmark requires >=10 cases per dataset, test=%d validation=%d", len(testCases), len(validationCases))
	}
	ids := map[string]struct{}{}
	requests := map[string]struct{}{}
	for _, item := range testCases {
		assertCaseIdentity(t, item, ids)
		requests[normalizedRequest(item.Input.Message)] = struct{}{}
	}
	for _, item := range validationCases {
		assertCaseIdentity(t, item, ids)
		if _, overlaps := requests[normalizedRequest(item.Input.Message)]; overlaps {
			t.Fatalf("validation request overlaps test request: %q", item.Input.Message)
		}
	}
}

// TestChatBenchmarkDeterministicContracts does not call a model. It protects
// source safety, retrieval fallback and degradation behavior cheaply on every
// local/CI run; it is intentionally not reported as a model benchmark.
func TestChatBenchmarkDeterministicContracts(t *testing.T) {
	for _, dataset := range []string{"cases.json", "validation_cases.json"} {
		for _, item := range loadCases(t, dataset) {
			t.Run(dataset+"/"+item.ID, func(t *testing.T) {
				var model web.ChatBenchmarkModel
				if strings.TrimSpace(item.ModelReply) != "" {
					model = fixedChatModel{reply: item.ModelReply}
				}
				reply := web.BuildChatReplyForBenchmark(context.Background(), item.Input, model)
				assertReplyContract(t, item, reply, false)
			})
		}
	}
}

// TestChatBenchmarkLiveModelResponses evaluates every chat case through the
// configured production text-model chain. It is opt-in so ordinary go test
// remains deterministic and does not spend model quota.
func TestChatBenchmarkLiveModelResponses(t *testing.T) {
	if !envBool("PPT_CHAT_BENCH_RUN_LIVE") {
		t.Skip("set PPT_CHAT_BENCH_RUN_LIVE=true to call the configured real text model")
	}
	loadEnv(t)
	ctx, cancel := context.WithTimeout(context.Background(), envDuration("PPT_CHAT_BENCH_MODEL_INIT_TIMEOUT", 45*time.Second))
	defer cancel()
	model, err := agentutils.NewFallbackToolCallingChatModel(ctx, agentutils.WithTextModel())
	if err != nil {
		t.Fatalf("initialize real chat benchmark model: %v", err)
	}

	cases := allCases(t)
	if limit := envInt("PPT_CHAT_BENCH_LIMIT", 0); limit > 0 && len(cases) > limit {
		cases = cases[:limit]
	}
	runDir := filepath.Join(projectRoot(t), "test", "chat_benchmark", "results", "live-"+time.Now().Format("20060102-150405"))
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Logf("real-model chat benchmark artifacts: %s", runDir)

	for _, item := range cases {
		item := item
		t.Run(item.dataset+"/"+item.ID, func(t *testing.T) {
			started := time.Now()
			caseCtx, caseCancel := context.WithTimeout(context.Background(), envDuration("PPT_CHAT_BENCH_CASE_TIMEOUT", 45*time.Second))
			defer caseCancel()
			adapter := &liveChatModelAdapter{inner: model}
			reply := web.BuildChatReplyForBenchmark(caseCtx, item.chatCase.Input, adapter)
			result := liveResult{Dataset: item.dataset, CaseID: item.ID, StartedAt: started.Format(time.RFC3339), DurationMS: time.Since(started).Milliseconds(), Reply: reply}
			if adapter.lastErr != nil {
				result.ModelError = adapter.lastErr.Error()
			}
			writeJSON(t, filepath.Join(runDir, safePathName(item.dataset+"-"+item.ID)+".json"), result)
			if adapter.lastErr != nil {
				t.Fatalf("real model generate failed: %v", adapter.lastErr)
			}
			assertReplyContract(t, item.chatCase, reply, true)
		})
	}
}

type namedChatCase struct {
	dataset string
	chatCase
}

type liveResult struct {
	Dataset    string `json:"dataset"`
	CaseID     string `json:"case_id"`
	StartedAt  string `json:"started_at"`
	DurationMS int64  `json:"duration_ms"`
	Reply      string `json:"reply"`
	ModelError string `json:"model_error,omitempty"`
}

func allCases(t *testing.T) []namedChatCase {
	t.Helper()
	items := make([]namedChatCase, 0, 20)
	for _, dataset := range []string{"cases.json", "validation_cases.json"} {
		for _, item := range loadCases(t, dataset) {
			items = append(items, namedChatCase{dataset: strings.TrimSuffix(dataset, ".json"), chatCase: item})
		}
	}
	return items
}

func assertReplyContract(t *testing.T, item chatCase, reply string, requireSemanticAnswer bool) {
	t.Helper()
	if strings.TrimSpace(reply) == "" {
		t.Fatal("reply is empty")
	}
	for _, want := range item.MustHave {
		if !strings.Contains(reply, want) {
			t.Fatalf("reply missing %q:\n%s", want, reply)
		}
	}
	for _, unwanted := range item.MustNot {
		if strings.Contains(reply, unwanted) {
			t.Fatalf("reply contains forbidden %q:\n%s", unwanted, reply)
		}
	}
	if links := len(chatLinkPattern.FindAllStringSubmatch(reply, -1)); links < item.MinLinks {
		t.Fatalf("links=%d, want >=%d:\n%s", links, item.MinLinks, reply)
	}
	if requireSemanticAnswer && len(item.MustMentionAny) > 0 && !containsAny(reply, item.MustMentionAny) {
		t.Fatalf("real model reply misses every topic signal %q:\n%s", item.MustMentionAny, reply)
	}
	if requireSemanticAnswer && strings.Contains(reply, "暂时无法调用普通对话模型") {
		t.Fatalf("real model did not produce an answer:\n%s", reply)
	}
}

func containsAny(value string, choices []string) bool {
	for _, choice := range choices {
		if strings.Contains(value, choice) {
			return true
		}
	}
	return false
}

var chatLinkPattern = regexp.MustCompile(`\]\(https?://[^\s)]+\)`)

func assertCaseIdentity(t *testing.T, item chatCase, ids map[string]struct{}) {
	t.Helper()
	if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.Input.Message) == "" {
		t.Fatalf("case must have id and message: %#v", item)
	}
	if _, exists := ids[item.ID]; exists {
		t.Fatalf("duplicate case id: %s", item.ID)
	}
	ids[item.ID] = struct{}{}
}

func normalizedRequest(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), ""))
}

func loadCases(t *testing.T, name string) []chatCase {
	t.Helper()
	path := filepath.Join(projectRoot(t), "test", "chat_benchmark", "testdata", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var cases []chatCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatalf("load %s: %v", path, err)
	}
	return cases
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
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(fmt.Errorf("write %s: %w", path, err))
	}
}
