package web

import (
	"strings"
	"testing"

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

func TestAppendChatSupplementRendersSourcesAndImages(t *testing.T) {
	got := appendChatSupplement("正文", chatAugmentations{
		webResults: []search.SearchResult{{Title: "官方攻略", URL: "https://example.com/guide"}},
		images:     []chatImageResult{{PreviewURL: "https://images.example/guide.jpg", Attribution: "摄影师"}},
	})
	for _, want := range []string{"补充资料来源", "[官方攻略](https://example.com/guide)", "![摄影师](https://images.example/guide.jpg)"} {
		if !strings.Contains(got, want) {
			t.Fatalf("supplement = %q, missing %q", got, want)
		}
	}
}
