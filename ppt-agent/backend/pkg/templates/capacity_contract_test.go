package templates

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestComponentContractsSeparateTargetDensityFromRenderLimit(t *testing.T) {
	contractPath := filepath.Join("..", "..", "..", "skills", "ppt-deck-planner", "templates", "component_contracts.json")
	raw, err := os.ReadFile(contractPath)
	if err != nil {
		t.Fatal(err)
	}
	var document componentContractsFile
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("invalid component contract JSON: %v", err)
	}
	if len(document.ContentTypes) == 0 {
		t.Fatal("component contract has no content types")
	}
	for name, spec := range document.ContentTypes {
		min, minOK := spec.Capacity["target_components_min"].(float64)
		targetMax, targetOK := spec.Capacity["target_components_max"].(float64)
		renderLimit, limitOK := spec.Capacity["max_components"].(float64)
		if !minOK || !targetOK || !limitOK {
			t.Fatalf("%s capacity must define target_components_min, target_components_max and max_components", name)
		}
		if min < 0 || min > targetMax || targetMax > renderLimit {
			t.Fatalf("invalid capacity order for %s: min=%v targetMax=%v renderLimit=%v", name, min, targetMax, renderLimit)
		}
	}
}
