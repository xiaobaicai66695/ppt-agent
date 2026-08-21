package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudwego/ppt-agent/pkg/assets/unsplash"
	imagetool "github.com/cloudwego/ppt-agent/pkg/tools/image"
)

type toolOutput struct {
	Provider  string `json:"provider"`
	Query     string `json:"query"`
	Total     int    `json:"total"`
	Returned  int    `json:"returned"`
	LocalPath string `json:"local_path,omitempty"`
	SourceURL string `json:"source_url,omitempty"`
}

func main() {
	query := flag.String("query", "drone", "Unsplash search query")
	download := flag.Bool("download", false, "download the first result")
	workDir := flag.String("work-dir", filepath.Join(".", "tmp", "unsplash-tool-probe"), "tool work directory")
	flag.Parse()

	client, err := unsplash.NewClientFromEnv()
	if err != nil {
		exitWithError(err)
	}
	if err := os.MkdirAll(*workDir, 0o755); err != nil {
		exitWithError(err)
	}

	searchTool := imagetool.NewImageSearchTool(client, *workDir)
	input := map[string]interface{}{
		"query":          *query,
		"orientation":    "landscape",
		"content_filter": "high",
		"per_page":       3,
		"download":       *download,
	}
	rawInput, err := json.Marshal(input)
	if err != nil {
		exitWithError(err)
	}

	rawOutput, err := searchTool.InvokableRun(context.Background(), string(rawInput))
	if err != nil {
		exitWithError(err)
	}

	var response struct {
		Error    string `json:"error"`
		Provider string `json:"provider"`
		Query    string `json:"query"`
		Total    int    `json:"total"`
		Photos   []struct {
			LocalPath string `json:"local_path"`
			SourceURL string `json:"source_url"`
		} `json:"photos"`
	}
	if err := json.Unmarshal([]byte(rawOutput), &response); err != nil {
		exitWithError(err)
	}
	if response.Error != "" {
		exitWithError(fmt.Errorf("%s", response.Error))
	}

	output := toolOutput{
		Provider: response.Provider,
		Query:    response.Query,
		Total:    response.Total,
		Returned: len(response.Photos),
	}
	if len(response.Photos) > 0 {
		output.LocalPath = response.Photos[0].LocalPath
		output.SourceURL = response.Photos[0].SourceURL
	}
	data, err := json.Marshal(output)
	if err != nil {
		exitWithError(err)
	}
	fmt.Println(string(data))
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}
