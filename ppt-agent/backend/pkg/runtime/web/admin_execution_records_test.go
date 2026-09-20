package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cloudwego/ppt-agent/pkg/db"
)

func TestParseAdminExecutionRecordQuery(t *testing.T) {
	tests := []struct {
		name                   string
		query                  string
		wantPage, wantPageSize int
		wantUserID             uint
		wantErr                bool
	}{
		{name: "defaults", wantPage: 1, wantPageSize: 50},
		{name: "explicit pagination and user", query: "?page=2&page_size=100&user_id=42", wantPage: 2, wantPageSize: 100, wantUserID: 42},
		{name: "invalid page", query: "?page=0", wantErr: true},
		{name: "oversized page size", query: "?page_size=201", wantErr: true},
		{name: "invalid user", query: "?user_id=abc", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest(http.MethodGet, "/api/admin/execution-records"+tt.query, nil)
			page, pageSize, userID, err := parseAdminExecutionRecordQuery(ctx)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, want error=%v", err, tt.wantErr)
			}
			if !tt.wantErr && (page != tt.wantPage || pageSize != tt.wantPageSize || userID != tt.wantUserID) {
				t.Fatalf("query = (%d, %d, %d), want (%d, %d, %d)", page, pageSize, userID, tt.wantPage, tt.wantPageSize, tt.wantUserID)
			}
		})
	}
}

func TestAdminExecutionRecordFromDBDoesNotExposeInternalPaths(t *testing.T) {
	now := time.Date(2026, 9, 20, 8, 0, 0, 0, time.UTC)
	response := adminExecutionRecordFromDB(db.AdminTaskRecord{
		ID:          "task-1",
		UserID:      7,
		UserEmail:   "member@example.com",
		Query:       "制作季度复盘",
		Status:      "completed",
		TotalTokens: 128,
		Intent:      "create",
		CreatedAt:   now,
	})
	if response.ID != "task-1" || response.UserEmail != "member@example.com" || response.TotalTokens != 128 || response.Intent != "create" {
		t.Fatalf("execution response fields = %#v", response)
	}
}
