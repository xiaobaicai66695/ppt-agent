package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/cloudwego/ppt-agent/pkg/agent/deep"
)

// Server is the web server that provides REST API + SSE streaming + static frontend.
type Server struct {
	tasks          *TaskManager
	agentFactory   AgentFactory
	makeTaskConfig func(taskID string) *deep.PPTTaskConfig
	taskIDGen      func() string
	httpServer     *http.Server
}

// ServerConfig holds configuration for creating a Server.
type ServerConfig struct {
	Addr           string
	BaseDir        string
	FrontendDir    string
	AgentFactory   AgentFactory
	MakeTaskConfig func(taskID string) *deep.PPTTaskConfig
}

// NewServer creates a new Server.
func NewServer(cfg *ServerConfig) *Server {
	frontendDir := cfg.FrontendDir
	if frontendDir == "" {
		frontendDir = filepath.Join("..", "frontend", "dist")
	}

	s := &Server{
		tasks:          NewTaskManager(cfg.BaseDir),
		agentFactory:   cfg.AgentFactory,
		makeTaskConfig: cfg.MakeTaskConfig,
		taskIDGen:      func() string { return uuid.New().String() },
	}

	mux := http.NewServeMux()

	// API routes.
	mux.HandleFunc("POST /api/tasks", s.handleCreateTask)
	mux.HandleFunc("GET /api/tasks", s.handleListTasks)
	mux.HandleFunc("GET /api/tasks/{id}", s.handleGetTask)
	mux.HandleFunc("GET /api/tasks/{id}/stream", s.handleStreamTask)
	mux.HandleFunc("GET /api/tasks/{id}/files/{filename}", s.handleDownloadFile)
		mux.HandleFunc("GET /api/tasks/{id}/thumb/{filename}", s.handleThumbnail)
	mux.HandleFunc("POST /api/tasks/{id}/cancel", s.handleCancelTask)

	// Health check.
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Serve frontend static files or fallback to index.html for SPA routing.
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		fullPath := filepath.Join(frontendDir, path)

		// If the file exists, serve it.
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, fullPath)
			return
		}

		// If it's a directory with index.html, serve that.
		indexPath := filepath.Join(fullPath, "index.html")
		if info, err := os.Stat(indexPath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, indexPath)
			return
		}

		// SPA fallback: serve index.html for all unknown paths.
		spaIndex := filepath.Join(frontendDir, "index.html")
		if _, err := os.Stat(spaIndex); err == nil {
			http.ServeFile(w, r, spaIndex)
		} else {
			http.NotFound(w, r)
		}
	})

	s.httpServer = &http.Server{
		Addr:    cfg.Addr,
		Handler: withCORS(mux),
	}

	return s
}

// Start starts the HTTP server. Blocks until the server stops.
func (s *Server) Start() error {
	fmt.Printf("[Web] 服务器启动于 http://localhost%s\n", s.httpServer.Addr)
	fmt.Printf("[Web] 前端界面: http://localhost%s\n", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Addr returns the server's listen address.
func (s *Server) Addr() string {
	return s.httpServer.Addr
}

// CORS middleware for development (Vite dev server on different port).
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
