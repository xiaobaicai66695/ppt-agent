package model

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestEvaluationSessionSourceTaskIDHasSingleUniqueIndex(t *testing.T) {
	parsed, err := schema.Parse(&EvaluationSession{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse evaluation session schema: %v", err)
	}

	matching := 0
	for _, index := range parsed.ParseIndexes() {
		if len(index.Fields) != 1 || index.Fields[0].DBName != "source_task_id" {
			continue
		}
		matching++
		if index.Class != "UNIQUE" {
			t.Fatalf("source_task_id index class = %q, want UNIQUE", index.Class)
		}
	}
	if matching != 1 {
		t.Fatalf("source_task_id index count = %d, want 1", matching)
	}
}
