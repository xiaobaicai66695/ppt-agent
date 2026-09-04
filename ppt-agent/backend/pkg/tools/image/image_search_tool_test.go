package image

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudwego/ppt-agent/pkg/utils/unsplash"
)

func TestImageSearchToolReturnsAttributableCandidates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search/photos" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("per_page"); got != "2" {
			t.Fatalf("per_page = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"id":"photo-1","urls":{"regular":"https://images.unsplash.com/photo-1","small":"https://images.unsplash.com/photo-1-small"},"links":{"html":"https://unsplash.com/photos/photo-1"},"user":{"name":"Photographer","links":{"html":"https://unsplash.com/@photographer"}}}]}`))
	}))
	defer server.Close()

	client, err := unsplash.NewClient("test-key", unsplash.WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := NewImageSearchTool(client).InvokableRun(context.Background(), `{"query":"northwest china landscape"}`)
	if err != nil {
		t.Fatal(err)
	}
	var response ImageSearchResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		t.Fatalf("unmarshal response: %v; raw=%s", err, raw)
	}
	if response.Provider != "unsplash" || len(response.Photos) != 1 {
		t.Fatalf("response = %#v", response)
	}
	photo := response.Photos[0]
	if photo.PreviewURL != "https://images.unsplash.com/photo-1-small" || photo.Attribution != "Photo by Photographer on Unsplash" {
		t.Fatalf("photo = %#v", photo)
	}
	if photo.SourceURL == "https://unsplash.com/photos/photo-1" || photo.PhotographerURL == "https://unsplash.com/@photographer" {
		t.Fatalf("Unsplash referral tags are missing: %#v", photo)
	}
}

func TestImageSearchToolRejectsEmptyQuery(t *testing.T) {
	raw, err := NewImageSearchTool(nil).InvokableRun(context.Background(), `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if raw != `{"error":"图片搜索关键词不能为空"}` {
		t.Fatalf("raw = %s", raw)
	}
}
