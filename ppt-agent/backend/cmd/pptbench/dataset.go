package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/ppt-agent/pkg/evaluation"
)

type localDatasetRegistry struct {
	Datasets map[string]localDatasetRecord `json:"datasets"`
}

type localDatasetRecord struct {
	DatasetID string `json:"dataset_id"`
	Role      string `json:"role"`
	SHA256    string `json:"sha256"`
}

func runDatasetCommand(args []string) error {
	if len(args) == 0 || args[0] != "pull" {
		return fmt.Errorf("usage: pptbench dataset pull --url <download-url> [--token <bearer-token>] [--name <local-name>]")
	}
	return pullDataset(args[1:])
}

func pullDataset(args []string) error {
	flags := newFlagSet("pptbench dataset pull")
	url := flags.String("url", "", "admin evaluation export download URL")
	token := flags.String("token", "", "admin bearer token; optional when the URL is otherwise authenticated")
	name := flags.String("name", "", "local dataset name; defaults to bundle manifest name")
	root := flags.String("root", importRoot(), "private benchmark import root")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*url) == "" {
		return fmt.Errorf("--url is required")
	}
	req, err := http.NewRequest(http.MethodGet, *url, nil)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*token) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(*token))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dataset download returned %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	manifest, files, err := evaluation.VerifyBundle(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("verify downloaded dataset: %w", err)
	}
	// Reuse the server's SHA header as a second integrity boundary when present.
	if expected := strings.TrimSpace(resp.Header.Get("X-Evaluation-Export-SHA256")); expected != "" {
		if actual := sha256Hex(data); actual != expected {
			return fmt.Errorf("export checksum mismatch")
		}
	}
	localName := strings.TrimSpace(*name)
	if localName == "" {
		localName = manifest.Name
	}
	if !validImportedDatasetName(localName) {
		return fmt.Errorf("invalid local dataset name %q", localName)
	}
	destination := filepath.Join(*root, localName)
	if _, err := os.Stat(destination); err == nil {
		return fmt.Errorf("local dataset %q already exists", localName)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Join(destination, "cases"), 0o700); err != nil {
		return err
	}
	for _, item := range manifest.Cases {
		content := files[item.Path]
		casePath := filepath.Join(destination, filepath.FromSlash(item.Path))
		if err := os.MkdirAll(filepath.Dir(casePath), 0o700); err != nil {
			return err
		}
		if err := os.WriteFile(casePath, content, 0o600); err != nil {
			return err
		}
	}
	manifestBytes, _ := json.MarshalIndent(manifest, "", "  ")
	if err := os.WriteFile(filepath.Join(destination, "manifest.json"), append(manifestBytes, '\n'), 0o600); err != nil {
		return err
	}
	return updateLocalRegistry(*root, localName, localDatasetRecord{DatasetID: manifest.DatasetID, Role: manifest.Role, SHA256: sha256Hex(data)})
}

func updateLocalRegistry(root, name string, record localDatasetRecord) error {
	path := filepath.Join(root, "registry.json")
	registry := localDatasetRegistry{Datasets: map[string]localDatasetRecord{}}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &registry); err != nil {
			return err
		}
	}
	if registry.Datasets == nil {
		registry.Datasets = map[string]localDatasetRecord{}
	}
	registry.Datasets[name] = record
	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

func newFlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	return flags
}

func sha256Hex(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
