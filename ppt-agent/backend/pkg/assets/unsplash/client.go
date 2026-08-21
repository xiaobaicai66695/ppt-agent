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

// Package unsplash provides a small, provider-specific client for the
// public Unsplash API used by the PPT asset workflow.
package unsplash

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL       = "https://api.unsplash.com/"
	defaultPerPage       = 10
	maxPerPage           = 30
	defaultMaxImageBytes = 20 << 20
)

var (
	ErrMissingAccessKey = errors.New("unsplash access key is not configured")
	ErrUnauthorized     = errors.New("unsplash access key was rejected")
)

// ClientOption customizes a Client. Options are primarily useful for tests
// and for deployments that need an explicit HTTP timeout.
type ClientOption func(*Client) error

// WithBaseURL replaces the API base URL. It is intended for httptest servers
// and must not be used to bypass the download host allowlist.
func WithBaseURL(rawURL string) ClientOption {
	return func(c *Client) error {
		parsed, err := url.Parse(strings.TrimSpace(rawURL))
		if err != nil {
			return fmt.Errorf("parse unsplash base URL: %w", err)
		}
		if parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("unsplash base URL must include scheme and host")
		}
		parsed.Path = strings.TrimRight(parsed.Path, "/") + "/"
		c.baseURL = parsed
		return nil
	}
}

// WithHTTPClient injects the HTTP client used for API and image requests.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) error {
		if httpClient == nil {
			return errors.New("unsplash HTTP client cannot be nil")
		}
		c.httpClient = httpClient
		return nil
	}
}

// WithMaxDownloadBytes limits the size of a downloaded image.
func WithMaxDownloadBytes(maxBytes int64) ClientOption {
	return func(c *Client) error {
		if maxBytes <= 0 {
			return errors.New("unsplash max download size must be positive")
		}
		c.maxDownloadBytes = maxBytes
		return nil
	}
}

// WithAllowedDownloadHosts adds exact hosts or URL hosts to the image
// download allowlist. It is useful for httptest-based download tests.
func WithAllowedDownloadHosts(hosts ...string) ClientOption {
	return func(c *Client) error {
		if c.allowedDownloadHosts == nil {
			c.allowedDownloadHosts = make(map[string]struct{})
		}
		for _, rawHost := range hosts {
			host := strings.TrimSpace(rawHost)
			if host == "" {
				continue
			}
			if parsed, err := url.Parse(host); err == nil && parsed.Host != "" {
				host = parsed.Host
			}
			c.allowedDownloadHosts[strings.ToLower(host)] = struct{}{}
		}
		return nil
	}
}

// NewClient creates an Unsplash API client from an Access Key.
func NewClient(accessKey string, options ...ClientOption) (*Client, error) {
	accessKey = strings.TrimSpace(accessKey)
	if accessKey == "" {
		return nil, ErrMissingAccessKey
	}

	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		return nil, err
	}

	client := &Client{
		accessKey:        accessKey,
		baseURL:          baseURL,
		httpClient:       &http.Client{Timeout: 30 * time.Second},
		maxDownloadBytes: defaultMaxImageBytes,
		allowedDownloadHosts: map[string]struct{}{
			"unsplash.com": {},
		},
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(client); err != nil {
			return nil, err
		}
	}
	return client, nil
}

// NewClientFromEnv creates a client using UNSPLASH_ACCESS_KEY.
func NewClientFromEnv(options ...ClientOption) (*Client, error) {
	return NewClient(os.Getenv("UNSPLASH_ACCESS_KEY"), options...)
}

// IsConfigured reports whether the public Access Key is available.
func IsConfigured() bool {
	return strings.TrimSpace(os.Getenv("UNSPLASH_ACCESS_KEY")) != ""
}

// Client is a provider-specific Unsplash API client.
type Client struct {
	accessKey            string
	baseURL              *url.URL
	httpClient           *http.Client
	maxDownloadBytes     int64
	allowedDownloadHosts map[string]struct{}
}

// SearchOptions contains the public /search/photos parameters.
type SearchOptions struct {
	Query         string
	Orientation   string
	ContentFilter string
	Color         string
	OrderBy       string
	Page          int
	PerPage       int
}

// SearchResponse is the abbreviated photo search response returned by
// Unsplash.
type SearchResponse struct {
	Total      int     `json:"total"`
	TotalPages int     `json:"total_pages"`
	Results    []Photo `json:"results"`
}

// Photo is the subset of an Unsplash photo object needed by the asset
// workflow.
type Photo struct {
	ID             string     `json:"id"`
	Width          int        `json:"width"`
	Height         int        `json:"height"`
	Color          string     `json:"color"`
	Description    string     `json:"description"`
	AltDescription string     `json:"alt_description"`
	BlurHash       string     `json:"blur_hash"`
	URLs           PhotoURLs  `json:"urls"`
	Links          PhotoLinks `json:"links"`
	User           User       `json:"user"`
}

type PhotoURLs struct {
	Raw     string `json:"raw"`
	Full    string `json:"full"`
	Regular string `json:"regular"`
	Small   string `json:"small"`
	Thumb   string `json:"thumb"`
}

type PhotoLinks struct {
	HTML             string `json:"html"`
	Download         string `json:"download"`
	DownloadLocation string `json:"download_location"`
}

type User struct {
	Name     string    `json:"name"`
	Username string    `json:"username"`
	Links    UserLinks `json:"links"`
}

type UserLinks struct {
	HTML string `json:"html"`
}

// Search executes GET /search/photos using public authentication.
func (c *Client) Search(ctx context.Context, options SearchOptions) (*SearchResponse, error) {
	if c == nil {
		return nil, errors.New("unsplash client is nil")
	}
	query := strings.TrimSpace(options.Query)
	if query == "" {
		return nil, errors.New("unsplash search query cannot be empty")
	}
	if options.Page < 0 {
		return nil, errors.New("unsplash page must be zero or greater")
	}
	if options.PerPage < 0 || options.PerPage > maxPerPage {
		return nil, fmt.Errorf("unsplash per_page must be between 1 and %d", maxPerPage)
	}

	perPage := options.PerPage
	if perPage == 0 {
		perPage = defaultPerPage
	}
	page := options.Page
	if page == 0 {
		page = 1
	}

	endpoint := c.baseURL.ResolveReference(&url.URL{Path: "search/photos"})
	params := endpoint.Query()
	params.Set("query", query)
	params.Set("page", strconv.Itoa(page))
	params.Set("per_page", strconv.Itoa(perPage))
	if value := strings.TrimSpace(options.Orientation); value != "" {
		params.Set("orientation", value)
	}
	if value := strings.TrimSpace(options.ContentFilter); value != "" {
		params.Set("content_filter", value)
	}
	if value := strings.TrimSpace(options.Color); value != "" {
		params.Set("color", value)
	}
	if value := strings.TrimSpace(options.OrderBy); value != "" {
		params.Set("order_by", value)
	}
	endpoint.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create unsplash search request: %w", err)
	}
	c.setPublicHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call unsplash search API: %w", err)
	}
	defer resp.Body.Close()

	body, err := readLimited(resp.Body, 8<<20)
	if err != nil {
		return nil, fmt.Errorf("read unsplash search response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, c.apiError(resp.StatusCode, body)
	}

	var result SearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode unsplash search response: %w", err)
	}
	return &result, nil
}

// DownloadedAsset contains the local file and attribution metadata for a
// downloaded photo.
type DownloadedAsset struct {
	PhotoID         string `json:"photo_id"`
	LocalPath       string `json:"local_path"`
	ImageURL        string `json:"image_url"`
	SourceURL       string `json:"source_url"`
	Photographer    string `json:"photographer"`
	PhotographerURL string `json:"photographer_url"`
	Attribution     string `json:"attribution"`
	Width           int    `json:"width"`
	Height          int    `json:"height"`
}

// Download tracks the Unsplash download event and saves the best available
// image URL into dir. The source URL and attribution are retained for later
// PPT metadata or notes.
func (c *Client) Download(ctx context.Context, photo Photo, dir string) (*DownloadedAsset, error) {
	if c == nil {
		return nil, errors.New("unsplash client is nil")
	}
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, errors.New("unsplash download directory cannot be empty")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create unsplash download directory: %w", err)
	}

	// PPT 页面通常不需要原始分辨率。优先使用 Unsplash 的 regular
	// 变体，避免 download_location 返回原图后占用过大带宽和磁盘。
	imageURL := firstNonEmpty(photo.URLs.Regular, photo.URLs.Full, photo.URLs.Small)
	if imageURL == "" {
		return nil, errors.New("unsplash photo has no downloadable image URL")
	}

	downloadURL := imageURL
	if strings.TrimSpace(photo.Links.DownloadLocation) != "" {
		if _, err := c.trackDownload(ctx, photo.Links.DownloadLocation); err != nil {
			return nil, err
		}
	}

	parsedURL, err := url.Parse(downloadURL)
	if err != nil || !c.allowedDownloadURL(parsedURL) {
		return nil, errors.New("unsplash image URL is not an allowed HTTPS host")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create unsplash image request: %w", err)
	}
	c.setPublicHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download unsplash image: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("download unsplash image: HTTP %d", resp.StatusCode)
	}
	if resp.Request != nil && !c.allowedDownloadURL(resp.Request.URL) {
		return nil, errors.New("unsplash image redirected to an untrusted host")
	}

	contentType := strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Type")))
	if contentType != "" && !strings.HasPrefix(contentType, "image/") && contentType != "application/octet-stream" {
		return nil, fmt.Errorf("unsplash image returned unsupported content type %q", contentType)
	}

	extension := extensionForContentType(contentType)
	tmp, err := os.CreateTemp(dir, ".unsplash-*")
	if err != nil {
		return nil, fmt.Errorf("create temporary unsplash image: %w", err)
	}
	tmpPath := tmp.Name()
	keepTemp := false
	defer func() {
		_ = tmp.Close()
		if !keepTemp {
			_ = os.Remove(tmpPath)
		}
	}()

	written, err := copyLimited(tmp, resp.Body, c.maxDownloadBytes)
	if err != nil {
		return nil, fmt.Errorf("write unsplash image: %w", err)
	}
	if written == 0 {
		return nil, errors.New("unsplash image response is empty")
	}
	if err := tmp.Close(); err != nil {
		return nil, fmt.Errorf("close temporary unsplash image: %w", err)
	}
	if extension == "" {
		extension = extensionForURL(parsedURL)
	}
	if extension == "" {
		extension = ".jpg"
	}

	filename := "unsplash_" + safePhotoID(photo.ID, imageURL) + extension
	targetPath := filepath.Join(dir, filename)
	if existing, err := os.Stat(targetPath); err == nil && existing.Size() > 0 {
		keepTemp = true
		_ = os.Remove(tmpPath)
	} else if err := os.Rename(tmpPath, targetPath); err != nil {
		return nil, fmt.Errorf("commit unsplash image: %w", err)
	} else {
		keepTemp = true
	}

	return &DownloadedAsset{
		PhotoID:         photo.ID,
		LocalPath:       targetPath,
		ImageURL:        imageURL,
		SourceURL:       photo.Links.HTML,
		Photographer:    photo.User.Name,
		PhotographerURL: photo.User.Links.HTML,
		Attribution:     attributionFor(photo),
		Width:           photo.Width,
		Height:          photo.Height,
	}, nil
}

func (c *Client) trackDownload(ctx context.Context, rawURL string) (string, error) {
	parsedURL, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || !c.allowedAPIURL(parsedURL) {
		return "", errors.New("unsplash download tracking URL is not an allowed HTTPS host")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create unsplash download tracking request: %w", err)
	}
	c.setPublicHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("track unsplash download: %w", err)
	}
	defer resp.Body.Close()
	body, err := readLimited(resp.Body, 64<<10)
	if err != nil {
		return "", fmt.Errorf("read unsplash download tracking response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("track unsplash download: HTTP %d", resp.StatusCode)
	}

	var payload struct {
		URL string `json:"url"`
	}
	if len(body) == 0 {
		return "", nil
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("decode unsplash download tracking response: %w", err)
	}
	return strings.TrimSpace(payload.URL), nil
}

func (c *Client) setPublicHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Client-ID "+c.accessKey)
	req.Header.Set("Accept-Version", "v1")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "ppt-agent/unsplash")
}

func (c *Client) apiError(status int, body []byte) error {
	if status == http.StatusUnauthorized {
		return fmt.Errorf("%w: HTTP %d", ErrUnauthorized, status)
	}
	var payload struct {
		Errors []string `json:"errors"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && len(payload.Errors) > 0 {
		return fmt.Errorf("unsplash API HTTP %d: %s", status, strings.Join(payload.Errors, "; "))
	}
	return fmt.Errorf("unsplash API HTTP %d", status)
}

func (c *Client) allowedAPIURL(parsedURL *url.URL) bool {
	if parsedURL == nil || parsedURL.Scheme != "https" {
		if parsedURL == nil || parsedURL.Scheme != "http" {
			return false
		}
	}
	if parsedURL.Hostname() == "" {
		return false
	}
	// The base URL is injectable for tests; production URLs are constrained to
	// the Unsplash API host by default.
	if c.baseURL != nil && parsedURL.Host == c.baseURL.Host {
		return true
	}
	return isUnsplashHost(parsedURL.Hostname())
}

func (c *Client) allowedDownloadURL(parsedURL *url.URL) bool {
	if parsedURL == nil {
		return false
	}
	host := strings.ToLower(parsedURL.Host)
	if _, ok := c.allowedDownloadHosts[host]; ok {
		// Exact hosts added by WithAllowedDownloadHosts are test/deployment
		// overrides; production Unsplash hosts below remain HTTPS-only.
		return true
	}
	if parsedURL.Scheme != "https" {
		return false
	}
	return isUnsplashHost(parsedURL.Hostname())
}

func isUnsplashHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	return host == "unsplash.com" || strings.HasSuffix(host, ".unsplash.com")
}

func readLimited(reader io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		maxBytes = 1
	}
	data, err := io.ReadAll(io.LimitReader(reader, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("response exceeds %d bytes", maxBytes)
	}
	return data, nil
}

func copyLimited(dst io.Writer, src io.Reader, maxBytes int64) (int64, error) {
	if maxBytes <= 0 {
		return 0, errors.New("download size limit must be positive")
	}
	written, err := io.Copy(dst, io.LimitReader(src, maxBytes+1))
	if err != nil {
		return written, err
	}
	if written > maxBytes {
		return written, fmt.Errorf("image exceeds %d bytes", maxBytes)
	}
	return written, nil
}

func extensionForContentType(contentType string) string {
	if contentType == "" || contentType == "application/octet-stream" {
		return ""
	}
	if contentType == "image/jpeg" {
		return ".jpg"
	}
	if contentType == "image/png" {
		return ".png"
	}
	if contentType == "image/webp" {
		return ".webp"
	}
	extensions, _ := mime.ExtensionsByType(contentType)
	if len(extensions) == 0 {
		return ""
	}
	for _, extension := range extensions {
		if extension == ".jpg" || extension == ".jpeg" || extension == ".png" || extension == ".webp" {
			return extension
		}
	}
	return extensions[0]
}

func extensionForURL(parsedURL *url.URL) string {
	if parsedURL == nil {
		return ""
	}
	extension := strings.ToLower(filepath.Ext(parsedURL.Path))
	switch extension {
	case ".jpg", ".jpeg", ".png", ".webp":
		return extension
	default:
		return ""
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func safePhotoID(id, fallback string) string {
	id = strings.TrimSpace(id)
	var builder strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			builder.WriteRune(r)
		}
	}
	if result := builder.String(); result != "" {
		return result
	}
	sum := sha256.Sum256([]byte(fallback))
	return hex.EncodeToString(sum[:8])
}

func attributionFor(photo Photo) string {
	name := strings.TrimSpace(photo.User.Name)
	if name == "" {
		name = strings.TrimSpace(photo.User.Username)
	}
	if name == "" {
		return "Photo on Unsplash"
	}
	return "Photo by " + name + " on Unsplash"
}
