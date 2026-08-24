/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package image exposes image search to the Planner without leaking provider
// authentication details into prompts or tool results.
package image

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/assets/unsplash"
	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/params"
)

const (
	defaultPlannerImagesPerCall = 2
	maxPlannerImagesPerCall     = 3
)

var imageSearchToolInfo = &schema.ToolInfo{
	Name: "search_images",
	Desc: `搜索适合 PPT 专题页使用的图片，当前 provider 为 Unsplash。
调用前先区分 asset_purpose：background 搜索宽幅、低细节、留白充足的氛围图；scene/evidence 搜索具体真实对象、动作和场景。query 应是经过视觉转换的最终检索词，不直接复制用户标题。
只用于寻找与页面视觉意图匹配的图片，不替代事实搜索。结果包含图片直链、来源页、摄影师署名和可选本地文件路径。
需要本地图片时将 download=true；下载结果会保留 Unsplash 来源和署名信息。`,
	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"query": {
			Type:     "string",
			Desc:     "经过视觉转换的最终图片检索词；背景图使用视觉代理和构图词，实景图使用具体对象、动作和环境",
			Required: true,
		},
		"asset_purpose": {
			Type: "string",
			Desc: "图片用途：background、scene、evidence 或 decorative",
		},
		"asset_subject": {
			Type: "string",
			Desc: "本次搜索对应的视觉主体或语义代理，例如 urban drone logistics、sunlit rural fields",
		},
		"composition": {
			Type: "string",
			Desc: "构图约束，例如 wide landscape, clean negative space on left",
		},
		"orientation": {
			Type: "string",
			Desc: "landscape、portrait 或 squarish；PPT 横向页面通常使用 landscape",
		},
		"content_filter": {
			Type: "string",
			Desc: "low 或 high；默认 high",
		},
		"color": {
			Type: "string",
			Desc: "颜色筛选，如 blue、green、black_and_white",
		},
		"order_by": {
			Type: "string",
			Desc: "relevant 或 latest；默认 relevant",
		},
		"page": {
			Type: "integer",
			Desc: "页码，从 1 开始",
		},
		"per_page": {
			Type: "integer",
			Desc: "返回候选数量，1-3，默认 2；整轮任务优先复用候选图，避免重复批量下载",
		},
		"download": {
			Type: "boolean",
			Desc: "是否下载到当前任务工作目录的 assets/images 下，默认 false",
		},
		"download_dir": {
			Type: "string",
			Desc: "下载目录，相对于当前任务工作目录，默认 assets/images",
		},
		"reason": {
			Type: "string",
			Desc: "检索原因，简述该图片如何服务于页面叙事",
		},
	}),
}

type imageSearchTool struct {
	client  *unsplash.Client
	workDir string
}

type imageSearchInput struct {
	Query         string `json:"query"`
	AssetPurpose  string `json:"asset_purpose,omitempty"`
	AssetSubject  string `json:"asset_subject,omitempty"`
	Composition   string `json:"composition,omitempty"`
	Orientation   string `json:"orientation,omitempty"`
	ContentFilter string `json:"content_filter,omitempty"`
	Color         string `json:"color,omitempty"`
	OrderBy       string `json:"order_by,omitempty"`
	Page          int    `json:"page,omitempty"`
	PerPage       int    `json:"per_page,omitempty"`
	Download      bool   `json:"download,omitempty"`
	DownloadDir   string `json:"download_dir,omitempty"`
	Reason        string `json:"reason,omitempty"`
}

type imageSearchResponse struct {
	Provider     string              `json:"provider"`
	AssetPurpose string              `json:"asset_purpose"`
	AssetSubject string              `json:"asset_subject,omitempty"`
	AssetQuery   string              `json:"asset_query"`
	Composition  string              `json:"composition,omitempty"`
	Total        int                 `json:"total"`
	TotalPages   int                 `json:"total_pages"`
	Photos       []imageSearchResult `json:"photos"`
}

type imageSearchResult struct {
	ID              string `json:"id"`
	Width           int    `json:"width"`
	Height          int    `json:"height"`
	Description     string `json:"description,omitempty"`
	AltDescription  string `json:"alt_description,omitempty"`
	ImageURL        string `json:"image_url"`
	PreviewURL      string `json:"preview_url,omitempty"`
	SourceURL       string `json:"source_url"`
	Photographer    string `json:"photographer,omitempty"`
	PhotographerURL string `json:"photographer_url,omitempty"`
	Attribution     string `json:"attribution"`
	LocalPath       string `json:"local_path,omitempty"`
	DownloadError   string `json:"download_error,omitempty"`
}

// NewImageSearchTool creates the Planner-facing Unsplash image search tool.
func NewImageSearchTool(client *unsplash.Client, workDir string) tool.InvokableTool {
	return &imageSearchTool{client: client, workDir: strings.TrimSpace(workDir)}
}

func (t *imageSearchTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return imageSearchToolInfo, nil
}

func (t *imageSearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	input := &imageSearchInput{}
	if err := json.Unmarshal([]byte(argumentsInJSON), input); err != nil {
		return "", fmt.Errorf("图片搜索参数解析失败: %w", err)
	}
	if strings.TrimSpace(input.Query) == "" {
		return marshalImageSearchError("图片搜索关键词不能为空")
	}
	if t.client == nil {
		return marshalImageSearchError("未配置 UNSPLASH_ACCESS_KEY，当前不可用图片搜索")
	}

	purpose, err := normalizeAssetPurpose(input.AssetPurpose)
	if err != nil {
		return marshalImageSearchError(err.Error())
	}
	if input.PerPage == 0 {
		input.PerPage = defaultPlannerImagesPerCall
	}
	if input.PerPage < 1 {
		input.PerPage = 1
	}
	if input.PerPage > maxPlannerImagesPerCall {
		input.PerPage = maxPlannerImagesPerCall
	}
	if purpose == "background" && input.Orientation == "" {
		input.Orientation = "landscape"
	}
	if input.ContentFilter == "" {
		input.ContentFilter = "high"
	}
	if input.OrderBy == "" {
		input.OrderBy = "relevant"
	}
	if input.Reason != "" {
		logger.Info("image_search_request", "query", input.Query, "asset_purpose", purpose, "asset_subject", input.AssetSubject, "reason", input.Reason)
	} else {
		logger.Info("image_search_request", "query", input.Query, "asset_purpose", purpose, "asset_subject", input.AssetSubject, "reason", "unspecified")
	}

	result, err := t.client.Search(ctx, unsplash.SearchOptions{
		Query:         input.Query,
		Orientation:   input.Orientation,
		ContentFilter: input.ContentFilter,
		Color:         input.Color,
		OrderBy:       input.OrderBy,
		Page:          input.Page,
		PerPage:       input.PerPage,
	})
	if err != nil {
		return marshalImageSearchError(publicImageSearchError(err))
	}

	response := imageSearchResponse{
		Provider:     "unsplash",
		AssetPurpose: purpose,
		AssetSubject: strings.TrimSpace(input.AssetSubject),
		AssetQuery:   input.Query,
		Composition:  strings.TrimSpace(input.Composition),
		Total:        result.Total,
		TotalPages:   result.TotalPages,
		Photos:       make([]imageSearchResult, 0, len(result.Results)),
	}

	var downloadDir string
	if input.Download {
		downloadDir, err = t.resolveDownloadDir(ctx, input.DownloadDir)
		if err != nil {
			return marshalImageSearchError(err.Error())
		}
	}

	for _, photo := range result.Results {
		item := imageSearchResult{
			ID:              photo.ID,
			Width:           photo.Width,
			Height:          photo.Height,
			Description:     photo.Description,
			AltDescription:  photo.AltDescription,
			ImageURL:        firstNonEmpty(photo.URLs.Regular, photo.URLs.Full, photo.URLs.Small),
			PreviewURL:      firstNonEmpty(photo.URLs.Small, photo.URLs.Thumb),
			SourceURL:       photo.Links.HTML,
			Photographer:    firstNonEmpty(photo.User.Name, photo.User.Username),
			PhotographerURL: photo.User.Links.HTML,
			Attribution:     attributionFor(photo),
		}
		if input.Download {
			asset, downloadErr := t.client.Download(ctx, photo, downloadDir)
			if downloadErr != nil {
				item.DownloadError = publicImageSearchError(downloadErr)
			} else {
				item.LocalPath = asset.LocalPath
			}
		}
		response.Photos = append(response.Photos, item)
	}

	data, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("图片搜索结果序列化失败: %w", err)
	}
	return string(data), nil
}

func (t *imageSearchTool) resolveDownloadDir(ctx context.Context, requested string) (string, error) {
	workDir := t.workDir
	if workDir == "" {
		if contextWorkDir, ok := params.GetTypedContextParams[string](ctx, params.WorkDirSessionKey); ok {
			workDir = contextWorkDir
		}
	}
	if workDir == "" {
		workDir, _ = os.Getwd()
	}
	workDir, err := filepath.Abs(workDir)
	if err != nil {
		return "", fmt.Errorf("解析图片工作目录失败: %w", err)
	}

	relativeDir := strings.TrimSpace(requested)
	if relativeDir == "" {
		relativeDir = filepath.Join("assets", "images")
	}
	if filepath.IsAbs(relativeDir) {
		return "", errors.New("download_dir 必须是相对于当前任务工作目录的路径")
	}

	target := filepath.Clean(filepath.Join(workDir, relativeDir))
	relativeToWorkDir, err := filepath.Rel(workDir, target)
	if err != nil || relativeToWorkDir == ".." || strings.HasPrefix(relativeToWorkDir, ".."+string(filepath.Separator)) {
		return "", errors.New("download_dir 不得逃逸当前任务工作目录")
	}
	return target, nil
}

func marshalImageSearchError(message string) (string, error) {
	data, err := json.Marshal(map[string]string{"error": message})
	return string(data), err
}

func publicImageSearchError(err error) string {
	if errors.Is(err, unsplash.ErrUnauthorized) {
		return "Unsplash Access Key 无效或已被拒绝"
	}
	return err.Error()
}

func normalizeAssetPurpose(value string) (string, error) {
	purpose := strings.ToLower(strings.TrimSpace(value))
	if purpose == "" {
		return "scene", nil
	}
	switch purpose {
	case "background", "scene", "evidence", "decorative":
		return purpose, nil
	default:
		return "", fmt.Errorf("asset_purpose 必须是 background、scene、evidence 或 decorative")
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func attributionFor(photo unsplash.Photo) string {
	name := firstNonEmpty(photo.User.Name, photo.User.Username)
	if name == "" {
		return "Photo on Unsplash"
	}
	return "Photo by " + name + " on Unsplash"
}
