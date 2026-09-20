package evaluation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPersistJSONUsesContentAddressedStorageWithoutDatabase(t *testing.T) {
	root := t.TempDir()
	capturer := NewCapturer(root)
	artifact, err := capturer.persistJSON(nil, "stage_input", map[string]any{"request": "测试"})
	if err != nil {
		t.Fatal(err)
	}
	if artifact.SHA256 == "" || artifact.StorageKey == "" {
		t.Fatalf("artifact = %#v", artifact)
	}
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(artifact.StorageKey)))
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]string
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["request"] != "测试" {
		t.Fatalf("stored value = %#v", decoded)
	}
}

func TestEvaluationStageIdentifiersAreStable(t *testing.T) {
	if got, want := evaluationSessionID("task-1"), evaluationSessionID("task-1"); got != want {
		t.Fatalf("session ids differ: %s vs %s", got, want)
	}
	if evaluationStageID("task-1", StagePlanner, 1) == evaluationStageID("task-1", StagePlanner, 2) {
		t.Fatal("stage attempts must have distinct identifiers")
	}
}
