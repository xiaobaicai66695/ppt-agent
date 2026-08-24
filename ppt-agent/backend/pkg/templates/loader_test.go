package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoaderPreservesLayoutContractMetadata(t *testing.T) {
	root := t.TempDir()
	presetsDir := filepath.Join(root, "presets")
	layoutsDir := filepath.Join(root, "layouts")
	backgroundsDir := filepath.Join(root, "backgrounds")
	for _, dir := range []string{presetsDir, layoutsDir, backgroundsDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	layoutJSON := `{
		"name":"content_slide",
		"display_name":"内容页",
		"description":"结构化内容页",
		"fields":[{"name":"title","label":"标题","type":"text","required":true}],
		"contract":{"capacity":{"max_items":6,"density":"normal"},"required_fields":["title"]}
	}`
	if err := os.WriteFile(filepath.Join(layoutsDir, "content_slide.json"), []byte(layoutJSON), 0600); err != nil {
		t.Fatal(err)
	}

	layouts := NewLoader(presetsDir, layoutsDir, backgroundsDir).ListLayouts()
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

func TestLoaderFallsBackToBuiltInCatalogWithoutLegacyTemplateDirs(t *testing.T) {
	root := t.TempDir()
	loader := NewLoader(
		filepath.Join(root, "missing-full-decks"),
		filepath.Join(root, "missing-single-page"),
		filepath.Join(root, "missing-backgrounds"),
	)
	if loader.GetPreset("generic") == nil {
		t.Fatal("expected built-in generic preset when legacy full-deck files are absent")
	}
	if loader.GetLayout("content_slide") == nil {
		t.Fatal("expected built-in component layout when legacy single-page files are absent")
	}
	if got := loader.ListBackgrounds(); len(got) != 0 {
		t.Fatalf("backgrounds = %#v, want no legacy local backgrounds", got)
	}
}
