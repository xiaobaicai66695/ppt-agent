package templates

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInformationTemplatesSeparateTargetDensityFromRenderLimit(t *testing.T) {
	templateDir := filepath.Join("..", "..", "..", "skills", "visual_designer", "templates", "single-page")
	templateNames := []string{
		"content_slide.json",
		"summary_slide.json",
		"two_column.json",
		"three_column.json",
		"card_grid.json",
		"quote_slide.json",
	}

	for _, name := range templateNames {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(templateDir, name))
			if err != nil {
				t.Fatal(err)
			}
			var document struct {
				Contract struct {
					Capacity map[string]any `json:"capacity"`
				} `json:"contract"`
			}
			if err := json.Unmarshal(raw, &document); err != nil {
				t.Fatalf("invalid template JSON: %v", err)
			}

			pairs := 0
			for key, value := range document.Contract.Capacity {
				if !strings.HasPrefix(key, "target_") || !strings.HasSuffix(key, "_max") {
					continue
				}
				base := strings.TrimSuffix(strings.TrimPrefix(key, "target_"), "_max")
				minKey := "target_" + base + "_min"
				limitKey := "max_" + base
				targetMax, targetOK := value.(float64)
				targetMin, minOK := document.Contract.Capacity[minKey].(float64)
				renderLimit, limitOK := document.Contract.Capacity[limitKey].(float64)
				if !targetOK || !minOK || !limitOK {
					t.Fatalf("capacity %q must define numeric %s, %s and %s", base, minKey, key, limitKey)
				}
				if targetMin <= 0 || targetMin > targetMax || targetMax > renderLimit {
					t.Fatalf("invalid capacity order for %q: min=%v targetMax=%v renderLimit=%v", base, targetMin, targetMax, renderLimit)
				}
				pairs++
			}
			if pairs == 0 {
				t.Fatal("template has no target density and render limit pair")
			}
		})
	}
}
