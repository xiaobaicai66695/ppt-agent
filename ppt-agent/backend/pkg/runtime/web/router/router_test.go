package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterKeepsPublicRouteContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	Register(engine, Handlers{})
	routes := make(map[string]bool)
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, want := range []string{
		"POST /api/auth/login",
		"GET /api/auth/me",
		"POST /api/messages",
		"POST /api/plan-drafts",
		"POST /api/tasks",
		"GET /api/tasks/:id/stream",
		"POST /api/tasks/:id/continue",
		"GET /api/tasks/:id/files/:filename",
		"GET /api/tasks/:id/runtime-events/:event_id",
		"GET /api/users/me/api-key",
		"GET /api/templates/layouts",
		"GET /api/admin/stats",
		"GET /api/health",
		"GET /health/ready",
		"GET /metrics",
	} {
		if !routes[want] {
			t.Errorf("missing route %s", want)
		}
	}
}

func TestRegisterUsesMiddlewareBeforeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	called := false
	Register(engine, Handlers{
		AuthMiddleware: func(c *gin.Context) { called = true; c.AbortWithStatus(http.StatusUnauthorized) },
		Message:        func(c *gin.Context) { c.Status(http.StatusOK) },
	})
	req, _ := http.NewRequest(http.MethodPost, "/api/messages", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if !called || rec.Code != http.StatusUnauthorized {
		t.Fatalf("middleware/handler order mismatch: called=%v status=%d", called, rec.Code)
	}
}

func TestRegisterFallsBackToNotFoundForUnknownRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	Register(engine, Handlers{})
	req, _ := http.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown route status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
