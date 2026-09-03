package web

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/ppt-agent/pkg/assets/unsplash"
	"github.com/cloudwego/ppt-agent/pkg/tools/search"
)

func TestChatNeedsWebSearchForSupplementalMaterials(t *testing.T) {
	if !chatNeedsWebSearch("给我补充些材料") {
		t.Fatal("supplemental-material request should enable web search")
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
			PhotographerURL: "https://unsplash.com/@example?utm_medium=referral&utm_source=ppt_agent",
		}},
	})
	for _, want := range []string{"补充资料来源", "[官方攻略](https://example.com/guide)", "![摄影师](https://images.example/guide.jpg)", "摄影：[示例摄影师](https://unsplash.com/@example?utm_medium=referral&utm_source=ppt_agent)", "[Unsplash](https://unsplash.com/?utm_source=ppt_agent&utm_medium=referral)"} {
		if !strings.Contains(got, want) {
			t.Fatalf("supplement = %q, missing %q", got, want)
		}
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
