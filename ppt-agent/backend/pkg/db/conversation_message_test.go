package db

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSplitConversationContentPreservesUTF8AndOrder(t *testing.T) {
	content := strings.Repeat("演示🙂", 20)
	chunks := splitConversationContent(content, 17)
	if len(chunks) < 2 {
		t.Fatal("expected content to be split")
	}
	if strings.Join(chunks, "") != content {
		t.Fatal("split content did not reassemble exactly")
	}
	for _, chunk := range chunks {
		if !utf8.ValidString(chunk) {
			t.Fatalf("invalid UTF-8 chunk: %q", chunk)
		}
		if len(chunk) > 17 {
			t.Fatalf("chunk is too large: %d", len(chunk))
		}
	}
}
