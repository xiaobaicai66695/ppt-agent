package web

import (
	"context"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/ppt-agent/pkg/tools/search"
)

// ChatBenchmarkModel is the small text-model surface used by the production
// chat reply builder. Benchmark callers can pass nil for deterministic fixture
// checks, or a real model adapter for opt-in semantic evaluation.
type ChatBenchmarkModel interface {
	Generate(context.Context, []*schema.Message, ...interface{}) (*schema.Message, error)
}

type ChatBenchmarkSearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Source      string `json:"source,omitempty"`
}

type ChatBenchmarkImageResult struct {
	PreviewURL      string `json:"preview_url"`
	ImageURL        string `json:"image_url"`
	SourceURL       string `json:"source_url"`
	Photographer    string `json:"photographer"`
	PhotographerURL string `json:"photographer_url"`
	Attribution     string `json:"attribution"`
}

// ChatBenchmarkInput controls every non-model input to the production chat
// assembly. It deliberately accepts fixture evidence instead of invoking
// external search services, so default benchmark runs are repeatable.
type ChatBenchmarkInput struct {
	Message             string                      `json:"message"`
	Fallback            string                      `json:"fallback,omitempty"`
	ConversationContext string                      `json:"conversation_context,omitempty"`
	WebResults          []ChatBenchmarkSearchResult `json:"web_results,omitempty"`
	Images              []ChatBenchmarkImageResult  `json:"images,omitempty"`
	WebSearchError      string                      `json:"web_search_error,omitempty"`
	ImageSearchError    string                      `json:"image_search_error,omitempty"`
}

// BuildChatReplyForBenchmark invokes the same reply construction and Markdown
// supplement functions used by workbench conversations. Passing nil as model
// evaluates deterministic fallback behavior; a non-nil model exercises the
// production prompt with the exact same fixture evidence.
func BuildChatReplyForBenchmark(ctx context.Context, input ChatBenchmarkInput, model ChatBenchmarkModel) string {
	augmentations := chatAugmentations{
		query: chatSearchQuery(input.Message, input.ConversationContext),
	}
	if conversation := compactChatConversationContext(input.ConversationContext); conversation != "" {
		augmentations.promptParts = append(augmentations.promptParts, "conversation_context:\n"+conversation)
	}
	for _, result := range input.WebResults {
		augmentations.webResults = append(augmentations.webResults, search.SearchResult{
			Title: result.Title, URL: result.URL, Description: result.Description, Source: result.Source,
		})
	}
	for _, image := range input.Images {
		augmentations.images = append(augmentations.images, chatImageResult{
			PreviewURL: image.PreviewURL, ImageURL: image.ImageURL, SourceURL: image.SourceURL,
			Photographer: image.Photographer, PhotographerURL: image.PhotographerURL, Attribution: image.Attribution,
		})
	}
	if value := strings.TrimSpace(input.WebSearchError); value != "" {
		augmentations.promptParts = append(augmentations.promptParts, "web_search_error: "+value)
	}
	if value := strings.TrimSpace(input.ImageSearchError); value != "" {
		augmentations.promptParts = append(augmentations.promptParts, "image_search_error: "+value)
	}

	server := &Server{}
	if model != nil {
		server.textModelFactory = func(context.Context) (interface {
			Generate(context.Context, []*schema.Message, ...interface{}) (*schema.Message, error)
		}, error) {
			return model, nil
		}
	}
	return server.buildChatReplyWithAugmentations(ctx, input.Message, input.Fallback, augmentations)
}
