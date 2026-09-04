package model

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// TextModel is the smallest model surface needed by chat/router benchmarks.
type TextModel interface {
	Generate(context.Context, []*schema.Message, ...interface{}) (*schema.Message, error)
}

type BenchmarkCreateRouteResult struct {
	Intent                string  `json:"intent"`
	TargetAgent           string  `json:"target_agent,omitempty"`
	NormalizedRequest     string  `json:"normalized_request,omitempty"`
	Reason                string  `json:"reason"`
	ClarificationQuestion string  `json:"clarification_question,omitempty"`
	Confidence            float64 `json:"confidence,omitempty"`
}

type BenchmarkMessageRouteResult struct {
	Intent            string  `json:"intent"`
	TargetAgent       string  `json:"target_agent,omitempty"`
	TaskID            string  `json:"task_id"`
	NormalizedRequest string  `json:"normalized_request,omitempty"`
	Action            string  `json:"action,omitempty"`
	Reason            string  `json:"reason"`
	Confidence        float64 `json:"confidence,omitempty"`
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

type ChatBenchmarkInput struct {
	Message             string                      `json:"message"`
	Fallback            string                      `json:"fallback,omitempty"`
	ConversationContext string                      `json:"conversation_context,omitempty"`
	WebResults          []ChatBenchmarkSearchResult `json:"web_results,omitempty"`
	Images              []ChatBenchmarkImageResult  `json:"images,omitempty"`
	WebSearchError      string                      `json:"web_search_error,omitempty"`
	ImageSearchError    string                      `json:"image_search_error,omitempty"`
}
