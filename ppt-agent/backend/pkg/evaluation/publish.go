package evaluation

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/cloudwego/ppt-agent/pkg/db"
)

type Service struct{ root string }

func NewService(root string) *Service { return &Service{root: NewCapturer(root).Root()} }

type BundleManifest struct {
	Format      int          `json:"format"`
	DatasetID   string       `json:"dataset_id"`
	Name        string       `json:"name"`
	Role        string       `json:"role"`
	PublishedAt string       `json:"published_at"`
	Cases       []BundleCase `json:"cases"`
}

type BundleCase struct {
	ID     string `json:"id"`
	Suite  string `json:"suite"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

func (s *Service) PublishDataset(name, role, parentID string, caseRevisionIDs []string) (*db.EvaluationDatasetVersion, error) {
	if s == nil {
		return nil, fmt.Errorf("evaluation service unavailable")
	}
	record := &db.EvaluationDatasetVersion{ID: "dataset-" + uuid.NewString(), Name: strings.TrimSpace(name), Role: strings.TrimSpace(role), ParentID: strings.TrimSpace(parentID)}
	if err := db.CreateEvaluationDatasetVersion(record, caseRevisionIDs); err != nil {
		return nil, err
	}
	return record, nil
}

func (s *Service) BuildExport(datasetID string) (*db.EvaluationExport, error) {
	if s == nil || s.root == "" {
		return nil, fmt.Errorf("evaluation artifact root is not configured")
	}
	dataset, err := db.GetEvaluationDatasetVersion(datasetID)
	if err != nil {
		return nil, err
	}
	if dataset.State != "frozen" || dataset.FrozenAt == nil {
		return nil, fmt.Errorf("only frozen datasets can be exported")
	}
	cases, err := db.ListEvaluationDatasetCases(datasetID)
	if err != nil {
		return nil, err
	}
	if len(cases) == 0 {
		return nil, fmt.Errorf("dataset has no cases")
	}
	manifest := BundleManifest{Format: 1, DatasetID: dataset.ID, Name: dataset.Name, Role: dataset.Role, PublishedAt: dataset.FrozenAt.UTC().Format(time.RFC3339)}
	exportID := "export-" + uuid.NewString()
	storageKey := filepath.ToSlash(filepath.Join("exports", exportID+".zip"))
	path := filepath.Join(s.root, filepath.FromSlash(storageKey))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".dataset-*.zip")
	if err != nil {
		return nil, err
	}
	tmpName := tmp.Name()
	zipWriter := zip.NewWriter(tmp)
	for _, revision := range cases {
		if !json.Valid([]byte(revision.CaseJSON)) {
			_ = zipWriter.Close()
			_ = tmp.Close()
			_ = os.Remove(tmpName)
			return nil, fmt.Errorf("case revision %s has invalid JSON", revision.ID)
		}
		casePath := filepath.ToSlash(filepath.Join("cases", safeFileName(revision.ID)+".json"))
		data := []byte(revision.CaseJSON)
		digest := sha256.Sum256(data)
		writer, err := zipWriter.Create(casePath)
		if err == nil {
			_, err = writer.Write(data)
		}
		if err != nil {
			_ = zipWriter.Close()
			_ = tmp.Close()
			_ = os.Remove(tmpName)
			return nil, err
		}
		manifest.Cases = append(manifest.Cases, BundleCase{ID: revision.ID, Suite: revision.Suite, Path: casePath, SHA256: hex.EncodeToString(digest[:])})
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err == nil {
		writer, createErr := zipWriter.Create("manifest.json")
		if createErr == nil {
			_, err = writer.Write(append(manifestBytes, '\n'))
		} else {
			err = createErr
		}
	}
	if closeErr := zipWriter.Close(); err == nil {
		err = closeErr
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(tmpName)
		return nil, err
	}
	data, err := os.ReadFile(tmpName)
	if err != nil {
		_ = os.Remove(tmpName)
		return nil, err
	}
	digest := sha256.Sum256(data)
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return nil, err
	}
	record := &db.EvaluationExport{ID: exportID, DatasetVersionID: dataset.ID, StorageKey: storageKey, SHA256: hex.EncodeToString(digest[:]), CreatedAt: time.Now().UTC()}
	if err := db.CreateEvaluationExport(record); err != nil {
		return nil, err
	}
	return record, nil
}

func (s *Service) ExportPath(id string) (string, *db.EvaluationExport, error) {
	if s == nil {
		return "", nil, fmt.Errorf("evaluation service unavailable")
	}
	record, err := db.GetEvaluationExport(id)
	if err != nil {
		return "", nil, err
	}
	if withdrawn, err := db.EvaluationDatasetHasWithdrawnEvidence(record.DatasetVersionID); err != nil {
		return "", nil, err
	} else if withdrawn {
		return "", nil, fmt.Errorf("evaluation export contains withdrawn evidence")
	}
	path := filepath.Join(s.root, filepath.FromSlash(record.StorageKey))
	if filepath.Clean(path) != path || !strings.HasPrefix(path, filepath.Clean(s.root)+string(os.PathSeparator)) {
		return "", nil, fmt.Errorf("invalid export storage key")
	}
	if _, err := os.Stat(path); err != nil {
		return "", nil, err
	}
	return path, record, nil
}

func (s *Service) CandidateEvidence(id string) (map[string]any, error) {
	if s == nil {
		return nil, fmt.Errorf("evaluation service unavailable")
	}
	candidate, err := db.GetEvaluationCandidate(id)
	if err != nil {
		return nil, err
	}
	stage, err := db.GetEvaluationStageRun(candidate.StageRunID)
	if err != nil {
		return nil, err
	}
	input, err := s.readArtifact(stage.InputArtifactID)
	if err != nil {
		return nil, err
	}
	output, err := s.readArtifact(stage.OutputArtifactID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"candidate": candidate, "stage_run": stage, "input": json.RawMessage(input), "output": json.RawMessage(output)}, nil
}

func (s *Service) readArtifact(id string) ([]byte, error) {
	artifact, err := db.GetEvaluationArtifact(id)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(s.root, filepath.FromSlash(artifact.StorageKey))
	if filepath.Clean(path) != path || !strings.HasPrefix(path, filepath.Clean(s.root)+string(os.PathSeparator)) {
		return nil, fmt.Errorf("invalid evaluation artifact storage key")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(data)
	if hex.EncodeToString(digest[:]) != artifact.SHA256 {
		return nil, fmt.Errorf("evaluation artifact checksum mismatch")
	}
	return data, nil
}

func safeFileName(value string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, value)
}

// VerifyBundle is shared by local import and tests. It verifies every declared
// case before any caller activates the downloaded dataset.
func VerifyBundle(reader io.Reader) (BundleManifest, map[string][]byte, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return BundleManifest{}, nil, err
	}
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return BundleManifest{}, nil, err
	}
	files := make(map[string][]byte, len(zipReader.File))
	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() || strings.Contains(filepath.ToSlash(file.Name), "../") || filepath.IsAbs(file.Name) {
			return BundleManifest{}, nil, fmt.Errorf("invalid bundle path %q", file.Name)
		}
		entry, err := file.Open()
		if err != nil {
			return BundleManifest{}, nil, err
		}
		content, readErr := io.ReadAll(entry)
		closeErr := entry.Close()
		if readErr != nil {
			return BundleManifest{}, nil, readErr
		}
		if closeErr != nil {
			return BundleManifest{}, nil, closeErr
		}
		files[filepath.ToSlash(file.Name)] = content
	}
	manifestBytes, ok := files["manifest.json"]
	if !ok {
		return BundleManifest{}, nil, fmt.Errorf("bundle is missing manifest.json")
	}
	var manifest BundleManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return BundleManifest{}, nil, err
	}
	if manifest.Format != 1 || manifest.Name == "" || len(manifest.Cases) == 0 {
		return BundleManifest{}, nil, fmt.Errorf("invalid bundle manifest")
	}
	for _, item := range manifest.Cases {
		content, ok := files[item.Path]
		if !ok {
			return BundleManifest{}, nil, fmt.Errorf("bundle is missing %s", item.Path)
		}
		digest := sha256.Sum256(content)
		if hex.EncodeToString(digest[:]) != item.SHA256 {
			return BundleManifest{}, nil, fmt.Errorf("checksum mismatch for %s", item.Path)
		}
		if !json.Valid(content) {
			return BundleManifest{}, nil, fmt.Errorf("case %s is not valid JSON", item.ID)
		}
	}
	return manifest, files, nil
}
