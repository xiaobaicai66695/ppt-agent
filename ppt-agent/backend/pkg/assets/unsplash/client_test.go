package unsplash

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSearchBuildsPublicAuthenticationRequest(t *testing.T) {
	var gotRequest *http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequest = r
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(SearchResponse{
			Total:      1,
			TotalPages: 1,
			Results:    []Photo{{ID: "photo-1", Width: 1600, Height: 900}},
		})
	}))
	defer server.Close()

	client, err := NewClient("test-access-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.Search(context.Background(), SearchOptions{
		Query:         "drone city",
		Orientation:   "landscape",
		ContentFilter: "high",
		PerPage:       3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 1 || len(result.Results) != 1 {
		t.Fatalf("unexpected search result: %#v", result)
	}
	if gotRequest == nil {
		t.Fatal("server did not receive a request")
	}
	if gotRequest.URL.Path != "/search/photos" {
		t.Fatalf("path = %q", gotRequest.URL.Path)
	}
	if gotRequest.URL.Query().Get("query") != "drone city" {
		t.Fatalf("query = %q", gotRequest.URL.Query().Get("query"))
	}
	if gotRequest.URL.Query().Get("orientation") != "landscape" {
		t.Fatalf("orientation = %q", gotRequest.URL.Query().Get("orientation"))
	}
	if gotRequest.Header.Get("Authorization") != "Client-ID test-access-key" {
		t.Fatalf("authorization header = %q", gotRequest.Header.Get("Authorization"))
	}
	if gotRequest.URL.Query().Get("client_id") != "" {
		t.Fatal("access key should not be placed in the query string")
	}
}

func TestSearchMapsUnauthorizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors":["OAuth error: The access token is invalid"]}`))
	}))
	defer server.Close()

	client, err := NewClient("rejected-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Search(context.Background(), SearchOptions{Query: "drone"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("error = %v, want ErrUnauthorized", err)
	}
	if strings.Contains(err.Error(), "rejected-key") {
		t.Fatal("error leaked the access key")
	}
}

func TestSearchRejectsInvalidPerPage(t *testing.T) {
	client, err := NewClient("test-access-key")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Search(context.Background(), SearchOptions{Query: "drone", PerPage: 31})
	if err == nil || !strings.Contains(err.Error(), "per_page") {
		t.Fatalf("error = %v, want per_page validation error", err)
	}
}

func TestDownloadTracksAndWritesImage(t *testing.T) {
	var tracked bool
	serverURL := ""
	imageBody := []byte("fake-jpeg")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/download":
			tracked = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"url":"` + serverURL + `/image.jpg"}`))
		case "/image.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write(imageBody)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	serverURL = server.URL

	client, err := NewClient(
		"test-access-key",
		WithBaseURL(server.URL),
		WithAllowedDownloadHosts(server.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	asset, err := client.Download(context.Background(), Photo{
		ID:     "abc123",
		Width:  1200,
		Height: 800,
		URLs:   PhotoURLs{Regular: server.URL + "/image.jpg"},
		Links:  PhotoLinks{DownloadLocation: server.URL + "/download", HTML: "https://unsplash.com/photos/abc123"},
		User:   User{Name: "摄影师", Links: UserLinks{HTML: "https://unsplash.com/@photographer"}},
	}, dir)
	if err != nil {
		t.Fatal(err)
	}
	if !tracked {
		t.Fatal("download tracking endpoint was not called")
	}
	if asset.LocalPath == "" || asset.Attribution != "Photo by 摄影师 on Unsplash" {
		t.Fatalf("unexpected asset metadata: %#v", asset)
	}
	if asset.PhotographerURL != "https://unsplash.com/@photographer?utm_medium=referral&utm_source=ppt_agent" {
		t.Fatalf("photographer URL = %q", asset.PhotographerURL)
	}
	content, err := os.ReadFile(asset.LocalPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != string(imageBody) {
		t.Fatalf("downloaded content = %q", string(content))
	}
	if filepath.Ext(asset.LocalPath) != ".jpg" {
		t.Fatalf("downloaded extension = %q", filepath.Ext(asset.LocalPath))
	}
}

func TestAttributionURLOnlyTagsUnsplashLinks(t *testing.T) {
	if got := AttributionURL("https://unsplash.com/@alice?foo=bar"); got != "https://unsplash.com/@alice?foo=bar&utm_medium=referral&utm_source=ppt_agent" {
		t.Fatalf("Unsplash attribution URL = %q", got)
	}
	if got := AttributionURL("https://example.com/profile"); got != "https://example.com/profile" {
		t.Fatalf("non-Unsplash URL = %q", got)
	}
}
