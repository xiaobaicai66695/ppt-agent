package search

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestSearchToolReadsDirectURLAndReturnsModelSummary(t *testing.T) {
	var gotURL string
	var gotEvidence string
	tool := NewSearchTool(
		WithURLContentReader(func(_ context.Context, rawURL string) (URLContent, error) {
			gotURL = rawURL
			return URLContent{Title: "示例报告", Text: strings.Repeat("这是用于测试的网页正文。", 2400), Source: "example.com"}, nil
		}),
		WithContentSummarizer(func(_ context.Context, query, evidence string) (string, error) {
			if query == "" {
				t.Fatal("summary query should not be empty")
			}
			gotEvidence = evidence
			return "- 已提取网页中的关键事实。", nil
		}),
	)

	raw, err := tool.InvokableRun(context.Background(), `{"query":"请阅读 https://example.com/report?id=1 并整理重点"}`)
	if err != nil {
		t.Fatalf("InvokableRun error = %v", err)
	}
	if gotURL != "https://example.com/report?id=1" {
		t.Fatalf("URL reader received %q", gotURL)
	}
	if len([]rune(gotEvidence)) > maxEvidenceRunes+1 {
		t.Fatalf("evidence was not bounded: %d runes", len([]rune(gotEvidence)))
	}
	var response SearchResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		t.Fatal(err)
	}
	if response.Mode != "url" || response.Content != "- 已提取网页中的关键事实。" || len(response.Results) != 1 {
		t.Fatalf("unexpected response: %#v", response)
	}
	if strings.Contains(response.Content, "网页正文") {
		t.Fatalf("raw webpage content leaked into response: %q", response.Content)
	}
}

func TestDirectURLFromQueryTrimsChinesePunctuation(t *testing.T) {
	got, ok := directURLFromQuery("资料在：https://example.com/a?x=1。请阅读")
	if !ok || got != "https://example.com/a?x=1" {
		t.Fatalf("directURLFromQuery = %q, %v", got, ok)
	}
}

func TestValidatePublicURLRejectsLoopback(t *testing.T) {
	if _, err := validatePublicURL(context.Background(), "http://127.0.0.1:8080/private"); err == nil {
		t.Fatal("loopback URL should be rejected")
	}
}

func TestExtractPageTextRemovesNonContentTags(t *testing.T) {
	title, text := extractPageText(`<html><head><title>报告 &amp; 数据</title><style>hidden</style></head><body><script>ignore()</script><h1>重点</h1><p>公开数据</p></body></html>`)
	if title != "报告 & 数据" || !strings.Contains(text, "重点 公开数据") || strings.Contains(text, "ignore") || strings.Contains(text, "hidden") {
		t.Fatalf("title=%q text=%q", title, text)
	}
}
