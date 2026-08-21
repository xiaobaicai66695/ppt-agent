package image

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudwego/ppt-agent/pkg/assets/unsplash"
)

func TestImageSearchToolReturnsStructuredResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search/photos" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"total": 1,
			"total_pages": 1,
			"results": [{
				"id": "photo-1",
				"width": 1600,
				"height": 900,
				"alt_description": "drone over city",
				"urls": {
					"regular": "https://images.unsplash.com/photo-1",
					"small": "https://images.unsplash.com/photo-1-small"
				},
				"links": {
					"html": "https://unsplash.com/photos/photo-1"
				},
				"user": {
					"name": "Photographer",
					"links": {"html": "https://unsplash.com/@photographer"}
				}
			}]
		}`))
	}))
	defer server.Close()

	client, err := unsplash.NewClient("test-key", unsplash.WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	tool := NewImageSearchTool(client, t.TempDir())

	raw, err := tool.InvokableRun(context.Background(), `{
		"query":"aerial city skyline, wide landscape, clean negative space",
		"asset_purpose":"background",
		"asset_subject":"aerial city skyline at blue hour",
		"composition":"wide landscape, clean negative space on left",
		"orientation":"landscape",
		"per_page":3
	}`)
	if err != nil {
		t.Fatal(err)
	}

	var response imageSearchResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		t.Fatalf("decode tool result: %v; raw=%s", err, raw)
	}
	if response.Provider != "unsplash" || response.AssetPurpose != "background" ||
		response.AssetSubject != "aerial city skyline at blue hour" ||
		response.AssetQuery != "aerial city skyline, wide landscape, clean negative space" ||
		response.Composition != "wide landscape, clean negative space on left" ||
		len(response.Photos) != 1 {
		t.Fatalf("unexpected response: %#v", response)
	}
	if response.Photos[0].Attribution != "Photo by Photographer on Unsplash" {
		t.Fatalf("attribution = %q", response.Photos[0].Attribution)
	}
}

func TestImageSearchToolRejectsDownloadPathEscape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total":0,"total_pages":0,"results":[]}`))
	}))
	defer server.Close()

	client, err := unsplash.NewClient("test-key", unsplash.WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	workDir := t.TempDir()
	tool := NewImageSearchTool(client, workDir)

	raw, err := tool.InvokableRun(context.Background(), `{"query":"drone","download":true,"download_dir":"../../outside"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(raw, "不得逃逸") {
		t.Fatalf("raw = %s", raw)
	}
	if filepath.Dir(workDir) == "" {
		t.Fatal("unexpected empty work directory")
	}
}

func TestImageSearchToolMapsUnauthorizedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors":["OAuth error: The access token is invalid"]}`))
	}))
	defer server.Close()

	client, err := unsplash.NewClient("test-key", unsplash.WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	tool := NewImageSearchTool(client, t.TempDir())

	raw, err := tool.InvokableRun(context.Background(), `{"query":"drone"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(raw, "Access Key 无效或已被拒绝") {
		t.Fatalf("raw = %s", raw)
	}
	if strings.Contains(raw, "test-key") {
		t.Fatal("tool result leaked the access key")
	}
}
