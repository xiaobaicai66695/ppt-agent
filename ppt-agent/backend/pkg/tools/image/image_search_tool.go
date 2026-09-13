// Package image exposes an Unsplash-backed image search tool for project
// agents. It deliberately returns remote candidates and attribution only; the
// PPT planner's deterministic materialization stage remains responsible for
// ppt asset downloads.
package image

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/utils/unsplash"
)

const (
	defaultImagesPerCall = 2
	maxImagesPerCall     = 3
)

var searchToolInfo = &schema.ToolInfo{
	Name: "search_images",
	Desc: "在 Unsplash 搜索图片候选。返回图片预览、来源页、摄影师署名和可被 PPTSpec/确定性素材物化阶段继续使用的候选字段。PPT Planner 与 Task Reviewer 可将真实返回字段写入其授权页面的 visual_intent 或 image 组件，但不负责下载本地素材。",
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"query": {
			Type:     "string",
			Desc:     "明确的可见主体、场景或风格检索词",
			Required: true,
		},
		"orientation":    {Type: "string", Desc: "landscape、portrait 或 squarish"},
		"content_filter": {Type: "string", Desc: "low 或 high；默认 high"},
		"order_by":       {Type: "string", Desc: "relevant 或 latest；默认 relevant"},
		"per_page":       {Type: "integer", Desc: "候选数量，1-3，默认 2"},
		"reason":         {Type: "string", Desc: "简述为何需要图片候选"},
	}),
}

type imageSearchTool struct{ client *unsplash.Client }

type imageSearchInput struct {
	Query         string `json:"query"`
	Orientation   string `json:"orientation,omitempty"`
	ContentFilter string `json:"content_filter,omitempty"`
	OrderBy       string `json:"order_by,omitempty"`
	PerPage       int    `json:"per_page,omitempty"`
	Reason        string `json:"reason,omitempty"`
}

type ImageSearchResponse struct {
	Provider string       `json:"provider"`
	Photos   []ImagePhoto `json:"photos"`
}

type ImagePhoto struct {
	ID               string `json:"id"`
	AssetID          string `json:"asset_id,omitempty"`
	AssetQuery       string `json:"asset_query,omitempty"`
	ImageURL         string `json:"image_url"`
	PreviewURL       string `json:"preview_url,omitempty"`
	SourceURL        string `json:"source_url"`
	DownloadLocation string `json:"download_location,omitempty"`
	Photographer     string `json:"photographer,omitempty"`
	PhotographerURL  string `json:"photographer_url,omitempty"`
	Attribution      string `json:"attribution"`
}

// NewImageSearchTool creates a project-agent tool. The caller decides whether
// the provider is configured before registering or invoking it.
func NewImageSearchTool(client *unsplash.Client) tool.InvokableTool {
	return &imageSearchTool{client: client}
}

func (t *imageSearchTool) Info(context.Context) (*schema.ToolInfo, error) { return searchToolInfo, nil }

func (t *imageSearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	var input imageSearchInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("图片搜索参数解析失败: %w", err)
	}
	input.Query = strings.TrimSpace(input.Query)
	if input.Query == "" {
		return marshalSearchError("图片搜索关键词不能为空")
	}
	if t.client == nil {
		return marshalSearchError("未配置 UNSPLASH_ACCESS_KEY，当前不可用图片搜索")
	}
	if input.PerPage <= 0 {
		input.PerPage = defaultImagesPerCall
	}
	if input.PerPage > maxImagesPerCall {
		input.PerPage = maxImagesPerCall
	}
	if input.ContentFilter == "" {
		input.ContentFilter = "high"
	}
	if input.OrderBy == "" {
		input.OrderBy = "relevant"
	}

	result, err := t.client.Search(ctx, unsplash.SearchOptions{
		Query: input.Query, Orientation: strings.TrimSpace(input.Orientation),
		ContentFilter: input.ContentFilter, OrderBy: input.OrderBy, PerPage: input.PerPage,
	})
	if err != nil {
		return marshalSearchError(publicSearchError(err))
	}
	response := ImageSearchResponse{Provider: "unsplash", Photos: make([]ImagePhoto, 0, len(result.Results))}
	for _, photo := range result.Results {
		photographer := firstNonEmpty(photo.User.Name, photo.User.Username)
		response.Photos = append(response.Photos, ImagePhoto{
			ID: photo.ID, AssetID: photo.ID, AssetQuery: input.Query,
			ImageURL: firstNonEmpty(photo.URLs.Regular, photo.URLs.Full, photo.URLs.Small),
			PreviewURL: firstNonEmpty(photo.URLs.Small, photo.URLs.Thumb),
			SourceURL:  unsplash.AttributionURL(photo.Links.HTML), DownloadLocation: photo.Links.DownloadLocation,
			Photographer: photographer, PhotographerURL: unsplash.AttributionURL(photo.User.Links.HTML),
			Attribution: attributionFor(photographer),
		})
	}
	data, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("图片搜索结果序列化失败: %w", err)
	}
	return string(data), nil
}

func marshalSearchError(message string) (string, error) {
	data, err := json.Marshal(map[string]string{"error": message})
	return string(data), err
}

func publicSearchError(err error) string {
	if strings.Contains(err.Error(), unsplash.ErrUnauthorized.Error()) {
		return "Unsplash Access Key 无效或已被拒绝"
	}
	return err.Error()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func attributionFor(photographer string) string {
	if photographer == "" {
		return "Photo on Unsplash"
	}
	return "Photo by " + photographer + " on Unsplash"
}
