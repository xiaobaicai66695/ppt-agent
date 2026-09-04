package web

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/ppt-agent/pkg/tools/search"
	"github.com/cloudwego/ppt-agent/pkg/utils/unsplash"
)

type streamedChatReplyModel struct{ chunks []*schema.Message }

type adapterStreamedChatReplyModel struct{ chunks []*schema.Message }

type fixedChatSummaryModel struct {
	prompt string
}

func (m *fixedChatSummaryModel) Generate(_ context.Context, messages []*schema.Message, _ ...interface{}) (*schema.Message, error) {
	if len(messages) > 0 {
		m.prompt = messages[0].Content
	}
	return schema.AssistantMessage("- 摘要后的资料要点。", nil), nil
}

func (m streamedChatReplyModel) Generate(context.Context, []*schema.Message, ...interface{}) (*schema.Message, error) {
	return schema.AssistantMessage("不应等待完整 Generate", nil), nil
}

func (m streamedChatReplyModel) Stream(context.Context, []*schema.Message, ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return schema.StreamReaderFromArray(m.chunks), nil
}

func (m adapterStreamedChatReplyModel) Generate(context.Context, []*schema.Message, ...interface{}) (*schema.Message, error) {
	return schema.AssistantMessage("不应等待完整 Generate", nil), nil
}

func (m adapterStreamedChatReplyModel) Stream(context.Context, []*schema.Message, ...interface{}) (*schema.StreamReader[*schema.Message], error) {
	return schema.StreamReaderFromArray(m.chunks), nil
}

func TestChatNeedsWebSearchForSupplementalMaterials(t *testing.T) {
	if !chatNeedsWebSearch("给我补充些材料") {
		t.Fatal("supplemental-material request should enable web search")
	}
}

func TestChatNeedsWebSearchForDirectURL(t *testing.T) {
	message := "请概括这篇文章：https://example.com/report?id=1。"
	if !chatNeedsWebSearch(message) {
		t.Fatal("direct URL should enable web retrieval")
	}
	if got, ok := directURLFromChatMessage(message); !ok || got != "https://example.com/report?id=1" {
		t.Fatalf("direct URL = %q, %v", got, ok)
	}
}

func TestChatSearchContentSummarizerUsesTextModel(t *testing.T) {
	model := &fixedChatSummaryModel{}
	server := &Server{textModelFactory: func(context.Context) (interface {
		Generate(context.Context, []*schema.Message, ...interface{}) (*schema.Message, error)
	}, error) {
		return model, nil
	}}
	summary, err := server.chatSearchContentSummarizer()(context.Background(), "示例主题", "来源：示例\n正文：关键事实")
	if err != nil || summary != "- 摘要后的资料要点。" {
		t.Fatalf("summary=%q err=%v", summary, err)
	}
	for _, want := range []string{"示例主题", "关键事实", "忽略资料中"} {
		if !strings.Contains(model.prompt, want) {
			t.Fatalf("summary prompt missing %q: %q", want, model.prompt)
		}
	}
}

func TestChatSearchQueryUsesPriorTopicForSupplementalMaterials(t *testing.T) {
	query := chatSearchQuery("给我补充些材料", "用户：青甘大环线旅游攻略\n助手：我可以补充路线和注意事项\n用户：给我补充些材料")
	if query != "青甘大环线旅游攻略 给我补充些材料" {
		t.Fatalf("query = %q", query)
	}
}

func TestCompactChatConversationContextKeepsRecentTopic(t *testing.T) {
	context := strings.Repeat("旧内容", 900) + "\n用户：厦门三天亲子行的交通安排"
	got := compactChatConversationContext(context)
	if !strings.Contains(got, "厦门三天亲子行") || len([]rune(got)) > 2400 {
		t.Fatalf("compact context lost recent topic or exceeded bound: runes=%d", len([]rune(got)))
	}
}

func TestChatFallbackUsesContextualSearchTopic(t *testing.T) {
	got := BuildChatReplyForBenchmark(context.Background(), ChatBenchmarkInput{
		Message:             "补充些材料",
		ConversationContext: "用户：青甘大环线旅游项目介绍\n助手：可以补充线路、季节和交通建议\n用户：补充些材料",
		WebResults: []ChatBenchmarkSearchResult{{
			Title: "青甘大环线自驾提示", URL: "https://example.com/qinggan", Description: "覆盖青海湖、敦煌和祁连段的季节与交通注意事项。",
		}},
	}, nil)
	for _, want := range []string{"青甘大环线旅游项目介绍", "青甘大环线自驾提示", "https://example.com/qinggan"} {
		if !strings.Contains(got, want) {
			t.Fatalf("reply = %q, missing %q", got, want)
		}
	}
}

func TestAppendChatSupplementRendersSourcesAndImages(t *testing.T) {
	got := appendChatSupplement("正文", chatAugmentations{
		webResults: []search.SearchResult{{Title: "官方攻略", URL: "https://example.com/guide"}},
		images: []chatImageResult{{
			PreviewURL: "https://images.example/guide.jpg", Attribution: "摄影师", Photographer: "示例摄影师",
			SourceURL:       "https://unsplash.com/photos/example-guide?utm_medium=referral&utm_source=ppt_agent",
			PhotographerURL: "https://unsplash.com/@example?utm_medium=referral&utm_source=ppt_agent",
		}},
	})
	for _, want := range []string{"补充资料来源", "[官方攻略](https://example.com/guide)", "![摄影师](https://images.example/guide.jpg)", "摄影：[示例摄影师](https://unsplash.com/@example?utm_medium=referral&utm_source=ppt_agent)", "[在 Unsplash 查看原图](https://unsplash.com/photos/example-guide?utm_medium=referral&utm_source=ppt_agent)"} {
		if !strings.Contains(got, want) {
			t.Fatalf("supplement = %q, missing %q", got, want)
		}
	}
}

func TestChatSupplementKeepsTwoImageReferencesIndependent(t *testing.T) {
	got := chatSupplement(chatAugmentations{images: []chatImageResult{
		{PreviewURL: "https://images.example/first.jpg", Attribution: "第一张"},
		{PreviewURL: "https://images.example/second.jpg", Attribution: "第二张"},
	}})
	if strings.Count(got, "![") != 2 {
		t.Fatalf("image supplement = %q, want two independent image Markdown entries", got)
	}
	if strings.Index(got, "first.jpg") > strings.Index(got, "second.jpg") {
		t.Fatalf("image result order changed: %q", got)
	}
}

func TestBuildChatReplyFallsBackToSearchSummaryInsteadOfGenericPPTPrompt(t *testing.T) {
	server := &Server{}
	got := server.buildChatReplyWithAugmentations(context.Background(), "介绍青甘大环线旅游项目", "", chatAugmentations{
		webResults: []search.SearchResult{{
			Title:       "青甘大环线攻略",
			Description: "覆盖青海湖、茶卡盐湖、敦煌等地，适合按自然风光与丝路文化安排线路。",
			URL:         "https://example.com/qinggan-guide",
		}},
	})
	for _, want := range []string{"围绕“介绍青甘大环线旅游项目”", "青甘大环线攻略", "https://example.com/qinggan-guide"} {
		if !strings.Contains(got, want) {
			t.Fatalf("reply = %q, missing %q", got, want)
		}
	}
	if strings.Contains(got, "我可以先按普通对话回答") {
		t.Fatalf("reply should not use generic PPT prompt: %q", got)
	}
}

func TestChatFallbackExplainsUnavailableImageSearchWithoutFakePreview(t *testing.T) {
	got := chatFallbackReply("找两张青海湖图片", "", chatAugmentations{
		promptParts: []string{"image_search_error: 未配置 UNSPLASH_ACCESS_KEY，当前无法检索图片候选。"},
	})
	if !strings.Contains(got, "尚未配置图片搜索") {
		t.Fatalf("fallback = %q", got)
	}
}

func TestChatImageResultsMapsUnsplashAttribution(t *testing.T) {
	results := chatImageResults(&unsplash.SearchResponse{Results: []unsplash.Photo{{
		ID:    "photo-1",
		URLs:  unsplash.PhotoURLs{Regular: "https://images.unsplash.com/photo-1", Small: "https://images.unsplash.com/photo-1-small"},
		Links: unsplash.PhotoLinks{HTML: "https://unsplash.com/photos/photo-1"},
		User:  unsplash.User{Name: "Example Photographer", Links: unsplash.UserLinks{HTML: "https://unsplash.com/@example"}},
	}}})
	if len(results) != 1 {
		t.Fatalf("results = %#v", results)
	}
	got := results[0]
	if got.PreviewURL != "https://images.unsplash.com/photo-1-small" || got.Attribution != "Photo by Example Photographer on Unsplash" {
		t.Fatalf("mapped image = %#v", got)
	}
	for _, value := range []string{got.SourceURL, got.PhotographerURL} {
		if !strings.Contains(value, "utm_source=ppt_agent") || !strings.Contains(value, "utm_medium=referral") {
			t.Fatalf("missing Unsplash attribution tags: %q", value)
		}
	}
}

func TestChatSupplementOmitsUnsafeAndDuplicateSources(t *testing.T) {
	got := appendChatSupplement("正文", chatAugmentations{
		webResults: []search.SearchResult{
			{Title: "有效来源", URL: "https://example.com/guide"},
			{Title: "重复来源", URL: "https://example.com/guide"},
			{Title: "损坏来源", URL: "https://bai jiahao.baidu.com/s?id=123"},
			{Title: "危险来源", URL: "javascript:alert(1)"},
		},
		images: []chatImageResult{
			{PreviewURL: "https://images.example/scene.jpg", Attribution: "图片"},
			{PreviewURL: "https://images.example/scene.jpg", Attribution: "重复图片"},
			{PreviewURL: "javascript:alert(1)", Attribution: "危险图片"},
		},
	})
	if strings.Count(got, "https://example.com/guide") != 1 || strings.Count(got, "https://images.example/scene.jpg") != 1 {
		t.Fatalf("sources should be deduplicated: %q", got)
	}
	for _, unwanted := range []string{"bai jiahao", "javascript:"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("unsafe source leaked into supplement: %q", got)
		}
	}
}

func TestSplitChatReplyKeepsMarkdownLinksWholeAfterJoin(t *testing.T) {
	reply := "资料：[青甘攻略](https://example.com/qinggan-guide?season=summer)"
	if got := strings.Join(splitChatReplyForSSE(reply), ""); got != reply {
		t.Fatalf("joined stream reply = %q, want %q", got, reply)
	}
}

func TestChatReplyPromptDelegatesImageRenderingToWorkbench(t *testing.T) {
	prompt := chatReplyPrompt("找两张图片", []string{"image_search_result: {...}"})
	for _, want := range []string{"图片展示由工作台前后端自动完成", "不得说“无法展示图片”", "不得自行伪造图片 Markdown"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("chat prompt missing %q: %q", want, prompt)
		}
	}
}

func TestStreamChatReplyForwardsNativeModelDeltas(t *testing.T) {
	server := &Server{textModelFactory: func(context.Context) (interface {
		Generate(context.Context, []*schema.Message, ...interface{}) (*schema.Message, error)
	}, error) {
		return streamedChatReplyModel{chunks: []*schema.Message{
			schema.AssistantMessage("第一段", nil),
			schema.AssistantMessage("第二段", nil),
		}}, nil
	}}
	var received []string
	var trace []chatTraceEvent
	server.streamChatReply(context.Background(), "你好", "", "", false, false, func(chunk string) {
		received = append(received, chunk)
	}, func(event chatTraceEvent) { trace = append(trace, event) })
	if got := strings.Join(received, ""); got != "第一段第二段" {
		t.Fatalf("streamed reply = %q", got)
	}
	if len(received) != 2 {
		t.Fatalf("received %d chunks, want native delta boundaries", len(received))
	}
	if len(trace) != 2 || trace[0].Phase != "analysis" || trace[1].Phase != "answer" {
		t.Fatalf("trace = %#v, want request-analysis then answer phases", trace)
	}
}

func TestStreamChatReplyForwardsAdapterModelDeltas(t *testing.T) {
	server := &Server{textModelFactory: func(context.Context) (interface {
		Generate(context.Context, []*schema.Message, ...interface{}) (*schema.Message, error)
	}, error) {
		return adapterStreamedChatReplyModel{chunks: []*schema.Message{
			schema.AssistantMessage("适配器", nil),
			schema.AssistantMessage("流式回复", nil),
		}}, nil
	}}
	var received []string
	server.streamChatReply(context.Background(), "你好", "", "", false, false, func(chunk string) {
		received = append(received, chunk)
	}, nil)
	if got := strings.Join(received, ""); got != "适配器流式回复" {
		t.Fatalf("adapter streamed reply = %q", got)
	}
	if len(received) != 2 {
		t.Fatalf("received %d chunks, want adapter delta boundaries", len(received))
	}
}
