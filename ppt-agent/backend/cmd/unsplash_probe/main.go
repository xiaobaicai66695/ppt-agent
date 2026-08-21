package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudwego/ppt-agent/pkg/assets/unsplash"
)

type probeOutput struct {
	Provider       string `json:"provider"`
	Query          string `json:"query"`
	Total          int    `json:"total"`
	Returned       int    `json:"returned"`
	FirstID        string `json:"first_id,omitempty"`
	FirstWidth     int    `json:"first_width,omitempty"`
	FirstHeight    int    `json:"first_height,omitempty"`
	FirstSourceURL string `json:"first_source_url,omitempty"`
	DownloadedPath string `json:"downloaded_path,omitempty"`
}

func main() {
	query := flag.String("query", "drone", "Unsplash search query")
	orientation := flag.String("orientation", "landscape", "landscape, portrait, or squarish")
	perPage := flag.Int("per-page", 3, "number of results, from 1 to 30")
	download := flag.Bool("download", false, "download the first result")
	downloadDir := flag.String("download-dir", "", "directory for the optional first download")
	flag.Parse()

	client, err := unsplash.NewClientFromEnv()
	if err != nil {
		exitWithError(err)
	}

	result, err := client.Search(context.Background(), unsplash.SearchOptions{
		Query:         *query,
		Orientation:   *orientation,
		ContentFilter: "high",
		OrderBy:       "relevant",
		PerPage:       *perPage,
	})
	if err != nil {
		exitWithError(err)
	}

	output := probeOutput{
		Provider: "unsplash",
		Query:    *query,
		Total:    result.Total,
		Returned: len(result.Results),
	}
	if len(result.Results) > 0 {
		first := result.Results[0]
		output.FirstID = first.ID
		output.FirstWidth = first.Width
		output.FirstHeight = first.Height
		output.FirstSourceURL = first.Links.HTML

		if *download {
			dir := *downloadDir
			if dir == "" {
				dir = filepath.Join(".", "tmp", "unsplash-probe")
			}
			asset, err := client.Download(context.Background(), first, dir)
			if err != nil {
				exitWithError(err)
			}
			output.DownloadedPath, _ = filepath.Abs(asset.LocalPath)
		}
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
