package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoaderPreservesLayoutContractMetadata(t *testing.T) {
	root := t.TempDir()
	templatesDir := filepath.Join(root, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatal(err)
	}

	contractJSON := `{
		"content_types": {
			"content_slide": {
				"best_for":["结构化内容页"],
				"recommended_components":["headline","list"],
				"capacity":{"max_items":6,"density":"normal"},
				"variants":["balanced"],
				"ppt_rule":""
			}
		}
	}`
	if err := os.WriteFile(filepath.Join(templatesDir, "component_contracts.json"), []byte(contractJSON), 0600); err != nil {
		t.Fatal(err)
	}

	layouts := NewComponentLoader(root).ListLayouts()
	if len(layouts) != 1 || layouts[0].Contract == nil {
		t.Fatalf("layouts = %#v, want one layout with contract metadata", layouts)
	}
	if got := layouts[0].Contract.Capacity["max_items"]; got != float64(6) {
		t.Fatalf("max_items = %#v, want 6", got)
	}
	if len(layouts[0].Contract.RequiredFields) != 1 || layouts[0].Contract.RequiredFields[0] != "title" {
		t.Fatalf("required_fields = %#v", layouts[0].Contract.RequiredFields)
	}
}

func TestLoaderFallsBackToBuiltInLayoutsWhenContractMissing(t *testing.T) {
	loader := NewComponentLoader(t.TempDir())
	if loader.GetLayout("content_slide") == nil {
		t.Fatal("expected built-in component layout when legacy single-page files are absent")
	}
}

func TestLoadComponentContractProjectsCapacityAndHash(t *testing.T) {
	root := t.TempDir()
	templatesDir := filepath.Join(root, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatal(err)
	}
	data := []byte(`{
		"version": 7,
		"content_types": {
			"content_slide": {
				"recommended_components":["headline","insight"],
				"capacity":{"target_components_min":2,"target_components_max":4,"max_components":5}
			}
		}
	}`)
	if err := os.WriteFile(filepath.Join(templatesDir, "component_contracts.json"), data, 0600); err != nil {
		t.Fatal(err)
	}

	contract, err := LoadComponentContract(root)
	if err != nil {
		t.Fatal(err)
	}
	capacity := contract.CapacityFor("content_slide")
	if contract.Version != 7 || contract.SHA256 == "" || capacity.RecommendedMin != 2 || capacity.RecommendedMax != 4 || capacity.MaxComponents != 5 {
		t.Fatalf("unexpected contract projection: %#v / %#v", contract, capacity)
	}
	if !strings.Contains(contract.PromptText(), "content_slide: recommended=2-4, max=5") {
		t.Fatalf("prompt summary omitted capacity: %s", contract.PromptText())
	}
}
