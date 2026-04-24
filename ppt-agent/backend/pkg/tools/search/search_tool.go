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
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/go-shiori/dom"
	"github.com/markusmobius/go-trafilatura"
)

var (
	htmlTagRegex  = regexp.MustCompile(`<[^>]*>`)
	urlRegex      = regexp.MustCompile(`\(https?://[^\)]+\)|https?://\S+`)
	unicodeEscRe  = regexp.MustCompile(`\\u0026#34;|\\u0026`)
	htmlEntityRe  = regexp.MustCompile(`&#34;|&#x22;`)
	ampEntityRe   = regexp.MustCompile(`&amp;`)
	extraSpaceRe  = regexp.MustCompile(`\n{3,}`)
)

const bingAPIURL = "https://cn.bing.com/search"

var bingSearchURL string

func init() {
	bingSearchURL = bingAPIURL
}

const (
	maxConcurrentFetch = 5
)

var trustedDomains = []string{
	"cloud.tencent.com", // 腾讯云开发者社区
	"xinhuanet.com",     // 新华网
	"sohu.com",          // 搜狐
	"aliyun.com",        // 阿里云
	"baidu.com",         // 百度百科
	"cnblogs.com",       // 博客园
	"juejin.cn",         // 稀土掘金
}

var searchToolInfo = &schema.ToolInfo{
	Name: "search",
	Desc: "搜索互联网获取相关信息，用于PPT内容补充。输入搜索关键词，返回搜索结果列表。\n\n【使用原则】网络搜索开销很大，请遵循以下原则：\n1. 仅在以下情况使用搜索：\n   - 用户明确要求查找最新信息或数据\n   - PPT内容需要具体的数字、日期、统计数据\n   - 需要核实不确定的事实或概念\n   - 缺少必要的关键信息（如专业术语解释、事件时间线等）\n2. 优先使用已有知识：常见概念、通用知识、基础事实无需搜索\n3. 搜索前先思考：是否能从已有信息推断？是否必须联网查询？",
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"query": {
			Type:     "string",
			Desc:     "搜索关键词（必填）",
			Required: true,
		},
		"reason": {
			Type:     "string",
			Desc:     "搜索必要性说明（选填）：简述为什么需要搜索，如'需要2024年最新数据'、'核实某公司财报数据'等。用于帮助判断是否真正需要执行搜索。",
			Required: false,
		},
	}),
}

type searchTool struct{}

type searchInput struct {
	Query  string `json:"query"`
	Reason string `json:"reason,omitempty"`
}

type SearchRequest struct {
	Query string `json:"query"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Content string         `json:"content,omitempty"`
	Error   string         `json:"error,omitempty"`
}

type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

type fetchResult struct {
	URL     string
	Content string
	Err     error
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

	if input.Reason != "" {
		fmt.Printf("[搜索必要性] 关键词: %s | 原因: %s\n", input.Query, input.Reason)
	} else {
		fmt.Printf("[搜索必要性] 关键词: %s | 原因: 未说明（建议补充）\n", input.Query)
	}

	results, err := searchBing(input.Query)
	if err != nil {
		return fmt.Sprintf(`{"error": "搜索失败: %v"}`, err), nil
	}

	if len(results) == 0 {
		return `{"error": "未找到搜索结果，可能页面结构已更新"}`, nil
	}

	filteredResults := filterTrustedResults(results)
	if len(filteredResults) == 0 {
		return `{"error": "未找到来自可信网站的搜索结果"}`, nil
	}

	urls := make([]string, 0)
	for i, r := range filteredResults {
		if i >= maxConcurrentFetch {
			break
		}
		urls = append(urls, r.URL)
	}

	contents := fetchURLsConcurrently(ctx, urls)

	var combinedContent strings.Builder
	combinedContent.WriteString(fmt.Sprintf("关键词: %s\n\n", input.Query))
	combinedContent.WriteString("=== 搜索结果 ===\n\n")

	for i, r := range filteredResults {
		if i < maxConcurrentFetch {
			combinedContent.WriteString(fmt.Sprintf("[%d] %s\n", i+1, r.URL))
			if content, ok := contents[r.URL]; ok && content != "" {
				combinedContent.WriteString(fmt.Sprintf("正文:\n%s\n\n", content))
			} else {
				combinedContent.WriteString("（获取内容失败）\n\n")
			}
		} else {
			combinedContent.WriteString(fmt.Sprintf("[%d] %s\n", i+1, r.URL))
		}
	}

	resp := SearchResponse{
		Results: filteredResults,
		Content: combinedContent.String(),
	}
	data, _ := json.Marshal(resp)
	return string(data), nil
}

func filterTrustedResults(results []SearchResult) []SearchResult {
	var filtered []SearchResult
	for _, r := range results {
		if isTrustedURL(r.URL) {
			filtered = append(filtered, r)
			if len(filtered) >= 5 {
				break
			}
		}
	}
	return filtered
}

func isTrustedURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Host)
	for _, domain := range trustedDomains {
		if strings.Contains(host, domain) {
			return true
		}
	}
	return false
}

func fetchURLsConcurrently(ctx context.Context, urls []string) map[string]string {
	results := make(chan fetchResult, len(urls))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrentFetch)

	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			select {
			case <-ctx.Done():
				results <- fetchResult{URL: u, Err: ctx.Err()}
			default:
				content, err := fetchURL(u)
				results <- fetchResult{URL: u, Content: content, Err: err}
			}
		}(url)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	contentMap := make(map[string]string)
	for result := range results {
		if result.Err == nil && result.Content != "" {
			contentMap[result.URL] = result.Content
		}
	}
	return contentMap
}

func fetchURL(targetURL string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	parsedURL, err := url.ParseRequestURI(targetURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %v", err)
	}

	opts := trafilatura.Options{
		IncludeImages: true,
		OriginalURL:   parsedURL,
	}

	result, err := trafilatura.Extract(resp.Body, opts)
	if err != nil {
		return "", fmt.Errorf("failed to extract content: %v", err)
	}

	if result == nil || result.ContentText == "" {
		return "", fmt.Errorf("no content extracted from page")
	}

	doc := trafilatura.CreateReadableDocument(result)
	raw := dom.OuterHTML(doc)
	clean := htmlTagRegex.ReplaceAllString(raw, "")
	clean = urlRegex.ReplaceAllString(clean, "")
	clean = unicodeEscRe.ReplaceAllString(clean, `"`)
	clean = htmlEntityRe.ReplaceAllString(clean, `"`)
	clean = ampEntityRe.ReplaceAllString(clean, "&")
	clean = extraSpaceRe.ReplaceAllString(clean, "\n\n")
	return strings.TrimSpace(clean), nil
}

// buildSiteFilter 将信任域名列表拼接成 "site:xxx.com OR site:yyy.com OR site:zzz.com" 格式
func buildSiteFilter() string {
	if len(trustedDomains) == 0 {
		return ""
	}
	var parts []string
	for _, d := range trustedDomains {
		parts = append(parts, "site:"+d)
	}
	return strings.Join(parts, " OR ")
}

func searchBing(query string) ([]SearchResult, error) {
	siteFilter := buildSiteFilter()
	fullQuery := query + " " + siteFilter
	fmt.Printf("[DEBUG] 完整搜索词: %s\n", fullQuery)
	searchURL := bingSearchURL + "?q=" + url.QueryEscape(fullQuery) + "&first=0"
	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	htmlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	html := string(htmlBytes)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	var results []SearchResult
	seen := make(map[string]bool)

	selector := "li.b_algo a.tilk"
	count := doc.Find(selector).Length()
	fmt.Printf("[DEBUG searchBing] selector=%s 匹配到 %d 个元素\n", selector, count)

	doc.Find("li.b_algo a.tilk").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" {
			return
		}
		href = strings.TrimSpace(href)
		if seen[href] {
			return
		}
		if strings.Contains(href, "cn.bing.com") || strings.Contains(href, "bing.com") {
			return
		}
		seen[href] = true
		fmt.Printf("[DEBUG searchBing] 匹配到 URL: %s\n", href)
		results = append(results, SearchResult{
			Title:       "Bing搜索结果",
			URL:         href,
			Description: "来自Bing搜索",
		})
	})

	if len(results) == 0 {
		return nil, fmt.Errorf("no results found in Bing search page")
	}
	return results, nil
}
