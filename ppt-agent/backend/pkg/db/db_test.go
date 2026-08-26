package db

import (
	"testing"
	"time"

	drivermysql "github.com/go-sql-driver/mysql"
)

func TestDSNWithTimeoutsAddsBoundedDefaults(t *testing.T) {
	dsn, err := dsnWithTimeouts("user:pass@tcp(db.example:3306)/ppt?charset=utf8mb4&parseTime=true")
	if err != nil {
		t.Fatalf("dsnWithTimeouts() error = %v", err)
	}
	cfg, err := drivermysql.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("ParseDSN() error = %v", err)
	}
	if cfg.Timeout != 5*time.Second || cfg.ReadTimeout != 10*time.Second || cfg.WriteTimeout != 10*time.Second {
		t.Fatalf("timeouts = %s/%s/%s", cfg.Timeout, cfg.ReadTimeout, cfg.WriteTimeout)
	}
	if !cfg.InterpolateParams {
		t.Fatal("InterpolateParams = false, want true")
	}
}

func TestDSNWithTimeoutsPreservesExplicitValues(t *testing.T) {
	dsn, err := dsnWithTimeouts("user:pass@tcp(db.example:3306)/ppt?timeout=2s&readTimeout=3s&writeTimeout=4s")
	if err != nil {
		t.Fatalf("dsnWithTimeouts() error = %v", err)
	}
	cfg, err := drivermysql.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("ParseDSN() error = %v", err)
	}
	if cfg.Timeout != 2*time.Second || cfg.ReadTimeout != 3*time.Second || cfg.WriteTimeout != 4*time.Second {
		t.Fatalf("explicit timeouts changed to %s/%s/%s", cfg.Timeout, cfg.ReadTimeout, cfg.WriteTimeout)
	}
}

func TestIsBusinessDatabaseRejectsSystemSchemas(t *testing.T) {
	for _, name := range []string{"", "information_schema", "MYSQL", " performance_schema ", "sys"} {
		if isBusinessDatabase(name) {
			t.Fatalf("isBusinessDatabase(%q) = true", name)
		}
	}
	if !isBusinessDatabase("myapp") {
		t.Fatal("isBusinessDatabase(myapp) = false")
	}
}

func TestBuildUserAPIKeyUpsertSetsTimestampsWithoutOverwritingCreatedAt(t *testing.T) {
	now := time.Date(2026, 8, 26, 21, 45, 0, 0, time.UTC)
	record, updates := buildUserAPIKeyUpsert(7, " deepseek ", " key-value ", now)

	if record.UserID != 7 || record.Provider != "deepseek" || record.APIKey != "key-value" {
		t.Fatalf("record fields = %#v", record)
	}
	if record.CreatedAt.IsZero() || record.UpdatedAt.IsZero() {
		t.Fatalf("timestamps should be non-zero: %#v", record)
	}
	if _, ok := updates["created_at"]; ok {
		t.Fatalf("created_at should not be overwritten on conflict: %#v", updates)
	}
	if updates["provider"] != "deepseek" || updates["api_key"] != "key-value" || updates["updated_at"] != now {
		t.Fatalf("updates = %#v", updates)
	}
}
