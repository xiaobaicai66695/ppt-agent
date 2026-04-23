package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const bingAPIURL = "https://cn.bing.com/search"

const (
	maxConcurrentFetch = 5
	maxContentLength   = 0
)

var trustedDomains = []string{
	"baidu.com",              // 百度
	"baiduusercontent.com",   // 百度用户内容
	"zhihu.com",              // 知乎
	"zhihuusercontent.com",    // 知乎用户内容
	"csdn.net",               // CSDN
	"juejin.cn",              // 稀土掘金
	"jianshu.com",            // 简书
	"toutiao.com",            // 今日头条
	"weibo.com",              // 微博
	"36kr.com",               // 36氪
	"ithome.com",             // IT之家
	"oschina.net",            // 开源中国
	"cloud.tencent.com",      // 腾讯云开发者社区
	"segmentfault.com",       // 思否
	"bilibili.com",           // 哔哩哔哩
	"douban.com",             // 豆瓣
	"cnblogs.com",            // 博客园
	"infoq.cn",               // InfoQ
}

var searchToolInfo = &schema.ToolInfo{
	Name: "search",
	Desc: "搜索互联网获取相关信息，用于PPT内容补充。输入搜索关键词，返回搜索结果列表。\n\n【使用原则】网络搜索开销很大，请遵循以下原则：\n1. 仅在以下情况使用搜索：\n   - 用户明确要求查找最新信息或数据\n   - PPT内容需要具体的数字、日期、统计数据\n   - 需要核实不确定的事实或概念\n   - 缺少必要的关键信息（如专业术语解释、事件时间线等）\n2. 优先使用已有知识：常见概念、通用知识、基础事实无需搜索\n3. 搜索前先思考：是否能从已有信息推断？是否必须联网查询？",
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"query": {
			Type:        "string",
			Desc:        "搜索关键词（必填）",
			Required:    true,
		},
		"reason": {
			Type:        "string",
			Desc:        "搜索必要性说明（选填）：简述为什么需要搜索，如'需要2024年最新数据'、'核实某公司财报数据'等。用于帮助判断是否真正需要执行搜索。",
			Required:    false,
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
	Results  []SearchResult `json:"results"`
	Content  string         `json:"content,omitempty"`
	Error    string         `json:"error,omitempty"`
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
				// 提取正文主体（最长的一段）
				mainContent := extractMainBody(content)
				if mainContent != "" {
					combinedContent.WriteString(fmt.Sprintf("正文:\n%s\n\n", mainContent))
				} else {
					combinedContent.WriteString(fmt.Sprintf("内容摘要: %s\n\n", truncateContent(content, 500)))
				}
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	html := string(body)
	content := extractTextContent(html)

	if maxContentLength > 0 && len(content) > maxContentLength {
		content = content[:maxContentLength] + "..."
	}
	return content, nil
}

func extractTextContent(html string) string {
	content := html

	// 尝试从 script 标签中提取 JSON 数据（处理知乎等动态渲染页面）
	jsonText := extractJSONFromScript(html)
	if jsonText != "" {
		// 从 JSON 中提取文本内容
		textFromJSON := extractTextFromJSON(jsonText)
		if textFromJSON != "" {
			return textFromJSON
		}
	}

	// 标准 HTML 提取流程
	content = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`).ReplaceAllString(content, "")
	content = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`).ReplaceAllString(content, "")
	content = regexp.MustCompile(`<!--[\s\S]*?-->`).ReplaceAllString(content, "")
	content = regexp.MustCompile(`<br\s*/?>`).ReplaceAllString(content, "\n")
	content = regexp.MustCompile(`</p>`).ReplaceAllString(content, "\n")
	content = regexp.MustCompile(`</div>`).ReplaceAllString(content, "\n")
	content = regexp.MustCompile(`</li>`).ReplaceAllString(content, "\n")
	content = regexp.MustCompile(`</h[1-6]>`).ReplaceAllString(content, "\n")
	content = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(content, "")
	content = strings.ReplaceAll(content, "&nbsp;", " ")
	content = strings.ReplaceAll(content, "&amp;", "&")
	content = strings.ReplaceAll(content, "&lt;", "<")
	content = strings.ReplaceAll(content, "&gt;", ">")
	content = strings.ReplaceAll(content, "&quot;", "\"")
	content = strings.ReplaceAll(content, "&#39;", "'")
	content = strings.ReplaceAll(content, "&#x27;", "'")
	content = strings.ReplaceAll(content, "&mdash;", "—")
	content = strings.ReplaceAll(content, "&ndash;", "–")
	content = regexp.MustCompile(`\n{3,}`).ReplaceAllString(content, "\n\n")
	content = strings.TrimSpace(content)

	lines := strings.Split(content, "\n")
	var filteredLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 10 {
			filteredLines = append(filteredLines, line)
		}
	}
	return strings.Join(filteredLines, "\n")
}

// extractJSONFromScript 从 HTML 中提取 __INITIAL_STATE__ 等 JSON 数据
func extractJSONFromScript(html string) string {
	// 尝试多种模式匹配
	patterns := []string{
		`window\.__INITIAL_STATE__\s*=\s*({.+})\s*;?\s*</script>`,
		`window\.__PRELOADED_STATE__\s*=\s*({.+})\s*;?\s*</script>`,
		`window\.__NUXT__\s*=\s*({.+})\s*;?\s*</script>`,
	}
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			// 用括号匹配来正确处理嵌套的 { }
			jsonStr := extractMatchingBraces(matches[1])
			if jsonStr != "" {
				return jsonStr
			}
		}
	}
	return ""
}

// extractMatchingBraces 从第一个 { 开始，找到匹配的最后一个 }
func extractMatchingBraces(s string) string {
	start := strings.Index(s, "{")
	if start == -1 {
		return ""
	}
	braceCount := 0
	for i := start; i < len(s); i++ {
		if s[i] == '{' {
			braceCount++
		} else if s[i] == '}' {
			braceCount--
			if braceCount == 0 {
				return s[start : i+1]
			}
		}
	}
	return ""
}

// extractTextFromJSON 从 JSON 中提取纯文本内容
func extractTextFromJSON(jsonStr string) string {
	// 递归提取 JSON 中的字符串值，优先找长文本
	var result strings.Builder
	var extractJSONStrings func(v interface{})
	extractJSONStrings = func(v interface{}) {
		switch val := v.(type) {
		case string:
			// 跳过短文本
			if len(val) < 30 {
				return
			}
			// 跳过明显的噪音
			lower := strings.ToLower(val)
			noise := []string{"http://", "https://", "function(", "undefined", "null", "window.", "document.", "typeof"}
			for _, n := range noise {
				if strings.Contains(lower, n) {
					return
				}
			}
			// 检查是否是有效的中文正文（包含常见中文字符）
			chineseCount := 0
			for _, r := range val {
				if r >= 0x4e00 && r <= 0x9fff {
					chineseCount++
				}
			}
			// 至少30%的中文字符
			if chineseCount > len(val)/3 {
				result.WriteString(val)
				result.WriteString("\n")
			}
		case map[string]interface{}:
			for _, child := range val {
				extractJSONStrings(child)
			}
		case []interface{}:
			for _, child := range val {
				extractJSONStrings(child)
			}
		}
	}

	// 尝试解析 JSON
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err == nil {
		extractJSONStrings(data)
	}
	return strings.TrimSpace(result.String())
}

// extractMainBody 返回正文主体：找出所有段落中内容最丰富的若干段
// 这能有效避免侧边栏、广告等噪音内容
func extractMainBody(text string) string {
	if text == "" {
		return ""
	}

	// 按换行分割成段落
	lines := strings.Split(text, "\n")

	// 收集所有有效的正文段落
	type paragraph struct {
		text string
		len  int
	}
	var validParagraphs []paragraph

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 跳过太短的行
		if len(line) < 50 {
			continue
		}
		// 跳过可能是导航、版权、评论区等噪音
		lower := strings.ToLower(line)
		noiseKeywords := []string{"copyright", "版权所有", "沪ICP备", "京ICP备", "声明", "未经授权", "右侧", "评论区", "相关推荐", "热门文章", "广告", "sponsored", "share to", "欢迎来到", "发现问题的"}
		isNoise := false
		for _, kw := range noiseKeywords {
			if strings.Contains(lower, kw) {
				isNoise = true
				break
			}
		}
		if isNoise {
			continue
		}

		// 优先选择中文内容多的段落
		chineseCount := 0
		for _, r := range line {
			if r >= 0x4e00 && r <= 0x9fff {
				chineseCount++
			}
		}
		if chineseCount < len(line)/3 && chineseCount < 20 {
			continue
		}

		validParagraphs = append(validParagraphs, paragraph{text: line, len: len(line)})
	}

	// 按长度排序，取最长的几个段落拼接
	sort.Slice(validParagraphs, func(i, j int) bool {
		return validParagraphs[i].len > validParagraphs[j].len
	})

	// 取最长的3个段落
	var result strings.Builder
	count := 0
	for _, p := range validParagraphs {
		if count >= 3 {
			break
		}
		result.WriteString(p.text)
		result.WriteString("\n\n")
		count++
	}

	return strings.TrimSpace(result.String())
}

// truncateContent 截断超长内容到指定长度
func truncateContent(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	// 在句子边界处截断
	truncated := text[:maxLen]
	lastSep := strings.LastIndexAny(truncated, "。！？；\n")
	if lastSep > maxLen/2 {
		return text[:lastSep+1]
	}
	return truncated + "..."
}

func searchBing(query string) ([]SearchResult, error) {
	searchURL := bingAPIURL + "?q=" + url.QueryEscape(query) + "&first=0"
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	html := string(body)
	results := parseBingSearchResults(html)
	if len(results) == 0 {
		results = parseBingSearchResultsFallback(html)
	}
	return results, nil
}

func parseBingSearchResults(html string) []SearchResult {
	var results []SearchResult
	olContent := extractOLContent(html)
	if olContent == "" {
		return results
	}

	liMatches := findAllLIWithClass(olContent, "b_algo")
	for _, liContent := range liMatches {
		links := extractLinksFromLI(liContent)
		for _, href := range links {
			if !isValidURL(href) {
				continue
			}
			results = append(results, SearchResult{
				Title:       "Bing搜索结果",
				URL:         href,
				Description: "来自Bing搜索",
			})
			if len(results) >= 10 {
				return results
			}
		}
	}
	return results
}

func extractOLContent(html string) string {
	pattern := `div[^>]*\sid\s*=\s*["']b_content["'][^>]*>([\s\S]*?)(?:<div[^>]*>[\s\S]*?){0,3}<main[^>]*aria-label\s*=\s*["'][^"']*搜索结果[^"']*["'][^>]*>([\s\S]*?)(?:<div[^>]*>[\s\S]*?){0,3}<ol[^>]*\sid\s*=\s*["']b_results["'][^>]*>`
	mainMatch := regexp.MustCompile(pattern).FindStringSubmatch(html)
	if len(mainMatch) >= 2 && mainMatch[1] != "" {
		olMatch := regexp.MustCompile(`<ol[^>]*\sid\s*=\s*["']b_results["'][^>]*>([\s\S]*?)</ol>`).FindStringSubmatch(mainMatch[1])
		if len(olMatch) >= 2 {
			return olMatch[1]
		}
		olMatch2 := regexp.MustCompile(`<ol[^>]*id=["']b_results["'][^>]*>([\s\S]*?)</ol>`).FindStringSubmatch(mainMatch[1])
		if len(olMatch2) >= 2 {
			return olMatch2[1]
		}
	}
	if len(mainMatch) >= 2 && mainMatch[2] != "" {
		olMatch := regexp.MustCompile(`<ol[^>]*\sid\s*=\s*["']b_results["'][^>]*>([\s\S]*?)</ol>`).FindStringSubmatch(mainMatch[2])
		if len(olMatch) >= 2 {
			return olMatch[1]
		}
		olMatch2 := regexp.MustCompile(`<ol[^>]*id=["']b_results["'][^>]*>([\s\S]*?)</ol>`).FindStringSubmatch(mainMatch[2])
		if len(olMatch2) >= 2 {
			return olMatch2[1]
		}
	}
	olMatch := regexp.MustCompile(`<ol[^>]*\s*id\s*=\s*["']b_results["'][^>]*>([\s\S]*?)</ol>`).FindStringSubmatch(html)
	if len(olMatch) >= 2 {
		return olMatch[1]
	}
	return ""
}

func findAllLIWithClass(content string, className string) []string {
	var results []string
	pattern := fmt.Sprintf(`<li[^>]*class\s*=\s*["'][^"']*%s[^"']*["'][^>]*>([\s\S]*?)</li>`, regexp.QuoteMeta(className))
	matches := regexp.MustCompile(pattern).FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) >= 2 {
			results = append(results, m[1])
		}
	}
	return results
}

func extractLinksFromLI(liContent string) []string {
	var urls []string
	links := regexp.MustCompile(`<a[^>]*\s*class\s*=\s*["'][^"']*\btilk\b[^"']*["'][^>]*>`).FindAllString(liContent, -1)
	for _, link := range links {
		hrefMatch := regexp.MustCompile(`href\s*=\s*["']([^"']+)["']`).FindStringSubmatch(link)
		if len(hrefMatch) >= 2 {
			urls = append(urls, hrefMatch[1])
		}
	}
	return urls
}

func parseBingSearchResultsFallback(html string) []SearchResult {
	var results []SearchResult
	linkPatterns := []string{
		`<a[^>]*class\s*=\s*["'][^"']*\bh\s+[tl]\b[^"']*["'][^>]*href\s*=\s*["']([^"']+)["'][^>]*>`,
		`<a[^>]*href\s*=\s*["'](https?://[^"']+)["'][^>]*class\s*=\s*["'][^"']*\btilk\b[^"']*["']`,
		`<a[^>]*class\s*=\s*["'][^"']*tilt[^"']*["'][^>]*href\s*=\s*["']([^"']+)["']`,
		`<h[23][^>]*>[\s\S]*?<a[^>]*href\s*=\s*["']([^"']+)["'][^>]*>`,
		`<a[^>]*href\s*=\s*["'](https?://[^"']{20,})["'][^>]*>`,
	}

	seen := make(map[string]bool)
	for _, pattern := range linkPatterns {
		matches := regexp.MustCompile(pattern).FindAllStringSubmatch(html, -1)
		for _, m := range matches {
			if len(m) < 2 {
				continue
			}
			href := m[1]
			if !isValidURL(href) || seen[href] {
				continue
			}
			if strings.Contains(href, "microsoft.com") && strings.Contains(href, "translator") {
				continue
			}
			if strings.Contains(href, "cn.bing.com") {
				continue
			}
			if strings.HasPrefix(href, "javascript:") {
				continue
			}
			seen[href] = true
			results = append(results, SearchResult{
				Title:       "Bing搜索结果",
				URL:         href,
				Description: "来自Bing搜索",
			})
			if len(results) >= 10 {
				return results
			}
		}
	}
	return results
}

func isValidURL(href string) bool {
	if href == "" {
		return false
	}
	if strings.HasPrefix(href, "/") && !strings.HasPrefix(href, "//") {
		return false
	}
	if strings.Contains(href, "cn.bing.com") || strings.Contains(href, "bing.com") {
		return false
	}
	parsed, err := url.Parse(href)
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" && parsed.Scheme != "" {
		return false
	}
	if parsed.Host == "" && parsed.Scheme == "" {
		return false
	}
	return true
}
