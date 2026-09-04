package web

import (
	"testing"
)

// TestRouteContract guards the public method/path pairs. Handler internals may
// move between files, but clients must continue to see the same HTTP surface.
func TestRouteContract(t *testing.T) {
	server := NewServer(&ServerConfig{BaseDir: t.TempDir()})
	routes := map[string]bool{}
	for _, route := range server.engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, want := range []string{
		"POST /api/auth/login",
		"GET /api/auth/me",
		"POST /api/tasks",
		"GET /api/tasks",
		"GET /api/tasks/:id",
		"GET /api/tasks/:id/stream",
		"POST /api/tasks/:id/continue",
		"GET /api/tasks/:id/files/:filename",
		"GET /api/tasks/:id/conversation",
		"GET /api/tasks/:id/runtime-events/:event_id",
		"GET /api/health",
	} {
		if !routes[want] {
			t.Errorf("missing route %s", want)
		}
	}
}
