package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

var (
	extraSpaceRe = regexp.MustCompile(`\n{3,}`)
	userAgents   = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:131.0) Gecko/20100101 Firefox/131.0",
	}
)

const (
	qianfanBaseURL   = "https://qianfan.baidubce.com/v2/ai_search"
	maxSearchResults = 5
)

// getAPIKey 延迟获取 API Key，确保 main 函数中 godotenv.Load 已执行
func getAPIKey() string {
	key := os.Getenv("QIANFAN_API_KEY")
	if key == "" {
		key = os.Getenv("BAIDU_QIANFAN_API_KEY")
	}
	return key
}

var searchToolInfo = &schema.ToolInfo{
	Name: "search",
	Desc: "网络搜索工具，用于获取大模型不知道的真实信息和最新数据。【使用原则】搜索有成本，仅在必要时使用：\n1. 需要大模型不知道的具体数字、日期、统计数据（如某公司财报、特定年份数据）\n2. 需要核实大模型可能不确定的事实（如某产品发布时间、技术参数）\n3. 用户明确要求查找最新信息\n4. 缺少必要的关键信息（如专业术语解释、事件时间线等）\n【禁止】对常见概念、通用知识、基础事实进行搜索。\n\n输入参数：\n- query: 搜索关键词（必填，2-5个词，简洁精准）\n- reason: 搜索必要性说明（选填）",
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

// --- 百度千帆 API 请求/响应结构 ---

type qianfanRequest struct {
	Messages           []qianfanMessage     `json:"messages"`
	SearchSource       string               `json:"search_source"`
	SearchFilter       *qianfanSearchFilter `json:"search_filter,omitempty"`
	ResourceTypeFilter []qianfanResource    `json:"resource_type_filter"`
}

type qianfanMessage struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type qianfanResource struct {
	Type string `json:"type"`
	TopK int    `json:"top_k"`
}

type qianfanSearchFilter struct {
	Match *qianfanMatch `json:"match,omitempty"`
}

type qianfanMatch struct {
	Site []string `json:"site,omitempty"`
}

type qianfanResponse struct {
	RequestID  string       `json:"request_id"`
	Code       string       `json:"code"`
	Message    string       `json:"message"`
	References []qianfanRef `json:"references"`
}

type qianfanRef struct {
	ID             int     `json:"id"`
	URL            string  `json:"url"`
	Title          string  `json:"title"`
	Date           string  `json:"date"`
	Content        string  `json:"content"`
	Snippet        string  `json:"snippet"`
	Icon           string  `json:"icon"`
	WebAnchor      string  `json:"web_anchor"`
	Type           string  `json:"type"`
	Website        string  `json:"website"`
	RerankScore    float64 `json:"rerank_score"`
	AuthorityScore float64 `json:"authority_score"`
}

// --- 工具入口 ---

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

	if qianfanAPIKey := getAPIKey(); qianfanAPIKey == "" {
		return `{"error": "未配置百度千帆 API Key (Set QIANFAN_API_KEY or BAIDU_QIANFAN_API_KEY)"}`, nil
	}

	refs, err := callQianfanAPI(ctx, input.Query)
	if err != nil {
		return fmt.Sprintf(`{"error": "搜索失败: %v"}`, err), nil
	}

	if len(refs) == 0 {
		return `{"error": "未找到搜索结果"}`, nil
	}

	// 构造结果列表
	results := make([]SearchResult, 0, len(refs))
	var combinedContent strings.Builder
	combinedContent.WriteString(fmt.Sprintf("关键词: %s\n\n", input.Query))
	combinedContent.WriteString("=== 搜索结果 ===\n\n")

	for i, ref := range refs {
		results = append(results, SearchResult{
			Title:       ref.Title,
			URL:         ref.URL,
			Description: ref.Website,
		})

		text := ref.Content
		if text == "" {
			text = ref.Snippet
		}
		if text == "" {
			text = "（无正文内容）"
		}

		combinedContent.WriteString(fmt.Sprintf("[%d] %s (%s)\n", i+1, ref.Title, ref.URL))
		combinedContent.WriteString(fmt.Sprintf("来源: %s | 日期: %s\n", ref.Website, ref.Date))
		combinedContent.WriteString(fmt.Sprintf("正文:\n%s\n\n", text))
	}

	resp := SearchResponse{
		Results: results,
		Content: combinedContent.String(),
	}
	data, _ := json.Marshal(resp)
	return string(data), nil
}

// --- 百度千帆搜索（单次请求，直接解析 JSON 中的文字内容）---

func callQianfanAPI(ctx context.Context, query string) ([]qianfanRef, error) {
	reqBody := qianfanRequest{
		Messages: []qianfanMessage{
			{Content: query, Role: "user"},
		},
		SearchSource: "baidu_search_v2",
		SearchFilter: &qianfanSearchFilter{
			Match: &qianfanMatch{
				Site: []string{
					"cloud.tencent.com",
					"cloud.alibabacloud.com",
					"juejin.cn",
					"zhihu.com",
					"csdn.net",
					"baidu.com",
					"tencent.com",
					"aliyun.com",
					"cnblogs.com",
				},
			},
		},
		ResourceTypeFilter: []qianfanResource{
			{Type: "web", TopK: maxSearchResults},
			{Type: "video", TopK: 0},
			{Type: "image", TopK: 0},
			{Type: "aladdin", TopK: 0},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", qianfanBaseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+getAPIKey())
	req.Header.Set("User-Agent", userAgents[rand.IntN(len(userAgents))])

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("[DEBUG] 百度搜索响应: %s \n", string(respBytes))

	var qresp qianfanResponse
	if err := json.Unmarshal(respBytes, &qresp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %v | body: %s", err, string(respBytes))
	}

	if qresp.Code != "" && qresp.Code != "0" {
		return nil, fmt.Errorf("API error [%s]: %s", qresp.Code, qresp.Message)
	}

	var refs []qianfanRef
	for _, ref := range qresp.References {
		if ref.Type == "web" && ref.URL != "" {
			// 清洗正文文本
			text := cleanText(ref.Content)
			if text == "" {
				text = cleanText(ref.Snippet)
			}
			refs = append(refs, qianfanRef{
				ID:      ref.ID,
				URL:     ref.URL,
				Title:   ref.Title,
				Date:    ref.Date,
				Content: text,
				Snippet: cleanText(ref.Snippet),
				Website: ref.Website,
				Type:    ref.Type,
			})
		}
	}

	fmt.Printf("[DEBUG] 解析到 %d 条网页结果\n", len(refs))
	return refs, nil
}

// cleanText 对从 JSON 中提取的文本做基本清洗
func cleanText(text string) string {
	if text == "" {
		return ""
	}
	// 移除多余的连续换行
	text = extraSpaceRe.ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}
