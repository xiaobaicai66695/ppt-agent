package web

import (
	"context"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/cloudwego/ppt-agent/pkg/auth"
	"github.com/cloudwego/ppt-agent/pkg/db"
	"github.com/gin-gonic/gin"
)

func TestCanAccessTask(t *testing.T) {
	tests := []struct {
		name    string
		ownerID int
		userID  int
		admin   bool
		want    bool
	}{
		{name: "owner", ownerID: 7, userID: 7, want: true},
		{name: "other user", ownerID: 7, userID: 8, want: false},
		{name: "administrator", ownerID: 7, userID: 8, admin: true, want: true},
		{name: "legacy task without owner", ownerID: 0, userID: 0, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canAccessTask(tt.ownerID, tt.userID, tt.admin); got != tt.want {
				t.Fatalf("canAccessTask(%d, %d, %v) = %v, want %v", tt.ownerID, tt.userID, tt.admin, got, tt.want)
			}
		})
	}
}

func TestUserIDGinReadsAuthenticatedRequestContext(t *testing.T) {
	ctx := auth.WithUser(context.Background(), &db.User{ID: 42, Email: "owner@example.com"})
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("GET", "/api/tasks/example", nil).WithContext(ctx)

	if got := userIDGin(c); got != 42 {
		t.Fatalf("userIDGin() = %d, want 42", got)
	}
}

func TestMaskAPIKeyDoesNotExposeFullSecret(t *testing.T) {
	got := maskAPIKey("abcd1234567890wxyz")
	if got != "abcd********wxyz" {
		t.Fatalf("maskAPIKey() = %q", got)
	}
	if got == "abcd1234567890wxyz" {
		t.Fatal("maskAPIKey exposed the full secret")
	}
}

func TestUserAPIKeyRoutesAreRegistered(t *testing.T) {
	server := NewServer(&ServerConfig{BaseDir: t.TempDir()})
	var routes []string
	for _, route := range server.engine.Routes() {
		routes = append(routes, route.Method+" "+route.Path)
	}
	for _, want := range []string{
		"POST /api/auth/register",
		"GET /api/users/me/api-key",
		"PUT /api/users/me/api-key",
		"DELETE /api/users/me/api-key",
	} {
		if !slices.Contains(routes, want) {
			t.Fatalf("route %q not registered; routes=%v", want, routes)
		}
	}
}
