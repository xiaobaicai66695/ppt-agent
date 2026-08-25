package templates

import (
	"os"
	"path/filepath"
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
				"deck_rule":""
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
