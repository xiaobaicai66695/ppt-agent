package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const bingAPIURL = "https://cn.bing.com/search"

var searchToolInfo = &schema.ToolInfo{
	Name: "search",
	Desc: "搜索互联网获取相关信息，用于PPT内容补充。输入搜索关键词，返回搜索结果列表。",
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"query": {
			Type:     "string",
			Desc:     "搜索关键词",
			Required: true,
		},
	}),
}

type searchTool struct{}

type searchInput struct {
	Query string `json:"query"`
}

type SearchRequest struct {
	Query string `json:"query"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Error   string        `json:"error,omitempty"`
}

type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

func NewSearchTool() tool.InvokableTool {
	return &searchTool{}
}

func (t *searchTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return searchToolInfo, nil
}

func (t *searchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	input := &searchInput{}
	if err := json.Unmarshal([]byte(argumentsInJSON), input); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	if input.Query == "" {
		return `{"error": "搜索关键词不能为空"}`, nil
	}

	results, err := searchBing(input.Query)
	if err != nil {
		return fmt.Sprintf(`{"error": "搜索失败: %v"}`, err), nil
	}

	resp := SearchResponse{Results: results}
	data, _ := json.Marshal(resp)
	return string(data), nil
}

func searchBing(query string) ([]SearchResult, error) {
	searchURL := bingAPIURL + "?q=" + url.QueryEscape(query) + "&first=0"

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseBingSearchResults(string(body)), nil
}

func parseBingSearchResults(html string) []SearchResult {
	var results []SearchResult

	// Bing 搜索结果匹配模式
	resultPattern := regexp.MustCompile(`<li class="b_algo"[^>]*>[\s\S]*?<h2><a href="([^"]+)"[^>]*>([^<]+)</a></h2>[\s\S]*?<p>([^<]+)</p>`)
	matches := resultPattern.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) >= 4 {
			result := SearchResult{
				URL:         match[1],
				Title:       cleanHTML(match[2]),
				Description: cleanHTML(match[3]),
			}
			results = append(results, result)
		}
		if len(results) >= 10 {
			break
		}
	}

	// 备用匹配模式：更简单的链接+标题匹配
	if len(results) == 0 {
		simplePattern := regexp.MustCompile(`<a href="(https?://[^"]+)"[^>]*>([^<]+)</a>[^<]*<p>([^<]+)</p>`)
		simpleMatches := simplePattern.FindAllStringSubmatch(html, -1)
		for _, match := range simpleMatches {
			if len(match) >= 3 && strings.Contains(match[1], "http") {
				result := SearchResult{
					URL:         match[1],
					Title:       cleanHTML(match[2]),
					Description: cleanHTML(match[3]),
				}
				results = append(results, result)
			}
			if len(results) >= 10 {
				break
			}
		}
	}

	return results
}

func cleanHTML(s string) string {
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	return strings.TrimSpace(s)
}

var imageSearchToolInfo = &schema.ToolInfo{
	Name: "search_image",
	Desc: "搜索图片素材，用于PPT配图。返回图片URL列表。",
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"query": {
			Type:     "string",
			Desc:     "图片搜索关键词",
			Required: true,
		},
	}),
}

type imageSearchTool struct{}

type imageSearchInput struct {
	Query string `json:"query"`
}

type ImageSearchResponse struct {
	Results []ImageResult `json:"results"`
	Error   string        `json:"error,omitempty"`
}

type ImageResult struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Source      string `json:"source"`
	LicenseType string `json:"license_type"`
}

func NewImageSearchTool() tool.InvokableTool {
	return &imageSearchTool{}
}

func (t *imageSearchTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return imageSearchToolInfo, nil
}

func (t *imageSearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	input := &imageSearchInput{}
	if err := json.Unmarshal([]byte(argumentsInJSON), input); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	if input.Query == "" {
		return `{"error": "图片搜索关键词不能为空"}`, nil
	}

	results := searchImages(input.Query)
	resp := ImageSearchResponse{Results: results}
	data, _ := json.Marshal(resp)
	return string(data), nil
}

func searchImages(query string) []ImageResult {
	searchURL := fmt.Sprintf("https://cn.bing.com/images/search?q=%s", url.QueryEscape(query))

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return getFallbackImages(query)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	// Bing 图片匹配模式
	imgPattern := regexp.MustCompile(`murl&quot;:&quot;(https?://[^&"]+)&quot;`)
	titlePattern := regexp.MustCompile(`t1&quot;:&quot;([^&quot;]+)&quot;`)

	imgMatches := imgPattern.FindAllStringSubmatch(html, -1)
	titleMatches := titlePattern.FindAllStringSubmatch(html, -1)

	var results []ImageResult
	minLen := len(imgMatches)
	if len(titleMatches) < minLen {
		minLen = len(titleMatches)
	}
	if minLen > 5 {
		minLen = 5
	}

	for i := 0; i < minLen; i++ {
		imgURL := imgMatches[i][1]
		imgURL = strings.ReplaceAll(imgURL, "&amp;", "&")
		results = append(results, ImageResult{
			URL:         imgURL,
			Title:       titleMatches[i][1],
			Width:       800,
			Height:      600,
			Source:      "Bing Images",
			LicenseType: "Unknown",
		})
	}

	if len(results) == 0 {
		return getFallbackImages(query)
	}

	return results
}

func getFallbackImages(query string) []ImageResult {
	fallbackURL := fmt.Sprintf("https://picsum.photos/seed/%s/800/600", url.QueryEscape(query))

	return []ImageResult{
		{
			URL:         fallbackURL,
			Title:       query + " (示例图片)",
			Width:       800,
			Height:      600,
			Source:      "Lorem Picsum",
			LicenseType: "public domain",
		},
		{
			URL:         fmt.Sprintf("https://source.unsplash.com/800x600/?%s", url.QueryEscape(query)),
			Title:       query + " (Unsplash)",
			Width:       800,
			Height:      600,
			Source:      "Unsplash",
			LicenseType: "Unsplash License",
		},
	}
}
