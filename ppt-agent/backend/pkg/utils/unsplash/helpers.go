package unsplash

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/url"
	"path/filepath"
	"strings"
)

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
