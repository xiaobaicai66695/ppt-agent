package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/cloudwego/ppt-agent/pkg/agent/deep"
	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/metrics"
	"github.com/cloudwego/ppt-agent/pkg/session"
	"github.com/cloudwego/ppt-agent/pkg/style"
	"github.com/cloudwego/ppt-agent/pkg/task"
	"github.com/cloudwego/ppt-agent/pkg/templates"
)

// Server provides REST API + SSE streaming + static frontend via Gin.
type Server struct {
	tasks          *task.TaskManager
	sessionManager *session.SessionManager
	styleStore     *style.ProfileStore
	styleExtractor *style.Extractor
	agentFactory   task.AgentFactory
	makeTaskConfig func(taskID string) *deep.PPTTaskConfig
	taskIDGen      func() string
	engine         *gin.Engine
	addr           string
	templateLoader *templates.Loader
	aiModelFactory func(ctx context.Context) (interface {
		Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
	}, error)
	// textModelFactory creates a lightweight model for side-tasks like intent classification.
	// Uses ARK_TEXT_MODEL env var to keep costs low. Falls back to AIModelFactory if nil.
	textModelFactory func(ctx context.Context) (interface {
		Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
	}, error)
}

// ServerConfig holds configuration for creating a Server.
type ServerConfig struct {
	Addr           string
	BaseDir        string
	FrontendDir    string
	SkillsDir      string
	AgentFactory   task.AgentFactory
	MakeTaskConfig func(taskID string) *deep.PPTTaskConfig
	AIModelFactory func(ctx context.Context) (interface {
		Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
	}, error)
	// StyleModelFactory creates a model for LLM-driven style extraction.
	// If nil, falls back to rule-based extraction.
	StyleModelFactory func(ctx context.Context) (model.ToolCallingChatModel, error)
	// TextModelFactory creates a lightweight text model for side-tasks like intent classification.
	// Uses ARK_TEXT_MODEL env var (cheaper). Falls back to AIModelFactory if nil.
	TextModelFactory func(ctx context.Context) (interface {
		Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
	}, error)
}

// NewServer creates a new Server with Gin routes.
func NewServer(cfg *ServerConfig) *Server {
	frontendDir := cfg.FrontendDir
	if frontendDir == "" {
		frontendDir = filepath.Join("..", "frontend", "dist")
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	// Prometheus HTTP metrics middleware
	engine.Use(metricsMiddleware)

	engine.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	s := &Server{
		sessionManager:   session.NewSessionManager(),
		styleStore:       style.NewProfileStore(cfg.BaseDir),
		agentFactory:     cfg.AgentFactory,
		makeTaskConfig:   cfg.MakeTaskConfig,
		taskIDGen:        func() string { return uuid.New().String() },
		engine:           engine,
		addr:             cfg.Addr,
		aiModelFactory:   cfg.AIModelFactory,
		textModelFactory: cfg.TextModelFactory,
	}

	// LLM 驱动的风格提取器（从 PPTX 文本内容分析用户偏好）
	if cfg.StyleModelFactory != nil {
		s.styleExtractor = style.NewExtractor(cfg.StyleModelFactory)
	}

	// Create task manager with style update callback
	s.tasks = task.NewTaskManager(cfg.BaseDir, func(userID int, workDir, query string) {
		s.updateUserStyleFromTask(userID, workDir, query)
	})

	// Initialize template loader
	presetsDir := filepath.Join(cfg.SkillsDir, "visual_designer", "templates", "full-decks")
	layoutsDir := filepath.Join(cfg.SkillsDir, "visual_designer", "templates", "single-page")
	s.templateLoader = templates.NewLoader(presetsDir, layoutsDir)

	// Auth routes (public)
	auth := engine.Group("/api/auth")
	{
		auth.POST("/send-code", s.handleSendCode)
		auth.POST("/login", s.handleLogin)
		auth.POST("/set-password", s.authMiddleware(), s.handleSetPassword)
		auth.POST("/logout", s.handleLogout)
		auth.GET("/me", s.authMiddleware(), s.handleMe)
	}

	// Task routes (protected)
	tasks := engine.Group("/api/tasks")
	tasks.Use(s.authMiddleware())
	{
		tasks.POST("", s.handleCreateTask)
		tasks.GET("", s.handleListTasks)
		tasks.GET("/:id", s.handleGetTask)
		tasks.GET("/:id/stream", s.handleStreamTask)
		tasks.GET("/:id/files/:filename", s.handleDownloadFile)
		tasks.GET("/:id/thumb/:filename", s.handleThumbnail)
		tasks.POST("/:id/cancel", s.handleCancelTask)
		tasks.DELETE("/:id", s.handleDeleteTask)
		// Session/continue routes
		tasks.POST("/:id/continue", s.handleContinueTask)
		tasks.GET("/:id/conversation", s.handleGetConversation)
		tasks.GET("/:id/events", s.handleListTaskEvents)
	}

	// User profile routes (protected)
	users := engine.Group("/api/users")
	users.Use(s.authMiddleware())
	{
		users.GET("/me/profile", s.handleGetUserProfile)
		users.PUT("/me/profile", s.handleUpdateUserProfile)
		users.POST("/me/profile/reset", s.handleResetUserProfile)
	}

	// Template routes (public)
	tpls := engine.Group("/api/templates")
	{
		tpls.GET("", s.handleListTemplates)
		tpls.GET("/:name", s.handleGetTemplate)
		tpls.GET("/layouts", s.handleListLayouts)
	}

	// Theme routes (public)
	engine.GET("/api/themes", s.handleListThemes)

	// AI routes (public)
	ai := engine.Group("/api/ai")
	{
		ai.POST("/expand", s.handleAIExpand)
		ai.POST("/generate-outline", s.handleAIGenerateOutline)
	}

	// Metrics
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health (basic)
	engine.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Detailed health check (auth optional for K8s probes)
	engine.GET("/health/ready", s.handleHealthCheck)

	// Static frontend
	engine.NoRoute(func(c *gin.Context) {
		serveStatic(c, frontendDir)
	})

	return s
}

func serveStatic(c *gin.Context, frontendDir string) {
	path := c.Request.URL.Path
	fullPath := filepath.Join(frontendDir, path)

	if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
		c.File(fullPath)
		return
	}
	indexPath := filepath.Join(frontendDir, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		c.File(indexPath)
		return
	}
	c.Status(http.StatusNotFound)
}

// metricsMiddleware records HTTP request metrics for Prometheus.
func metricsMiddleware(c *gin.Context) {
	start := time.Now()
	c.Next()
	duration := time.Since(start).Seconds()

	// Skip metrics endpoint itself to avoid noise
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	metrics.HTTPRequestsTotal.WithLabelValues(c.Request.Method, path, fmt.Sprintf("%d", c.Writer.Status())).Inc()
	metrics.HTTPRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	logger.Info("server_starting", "addr", s.addr, "frontend", fmt.Sprintf("http://localhost%s", s.addr))
	return s.engine.Run(s.addr)
}
