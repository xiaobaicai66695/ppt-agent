package web

import (
	"context"
	"strings"

	"github.com/cloudwego/eino/schema"

	webmodel "github.com/cloudwego/ppt-agent/pkg/runtime/web/model"
	"github.com/cloudwego/ppt-agent/pkg/tools/search"
)

// Benchmark models are owned by web/model. Aliases retain the package-level
// API used by the existing evaluation command.
type ChatBenchmarkModel = webmodel.TextModel
type ChatBenchmarkSearchResult = webmodel.ChatBenchmarkSearchResult
type ChatBenchmarkImageResult = webmodel.ChatBenchmarkImageResult
type ChatBenchmarkInput = webmodel.ChatBenchmarkInput

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
