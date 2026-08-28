package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
	loganalysis "github.com/cloudwego/ppt-agent/pkg/log_analysis"
	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/metrics"
	"github.com/cloudwego/ppt-agent/pkg/session"
	"github.com/cloudwego/ppt-agent/pkg/task"
	"github.com/cloudwego/ppt-agent/pkg/templates"
)

// Server 提供 REST API + SSE 流式推送 + 静态前端服务（基于 Gin 框架）。
type Server struct {
	tasks           *task.TaskManager
	sessionManager  *session.SessionManager
	agentFactory    task.AgentFactory
	makeTaskConfig  func(taskID string) *deck.PPTTaskConfig
	taskIDGen       func() string
	engine          *gin.Engine
	addr            string
	templateLoader  *templates.Loader
	skillDir        string
	operator        commandline.Operator
	logAnalysis     *loganalysis.Service
	continueStarter func(taskID string, ts *task.TaskState, message string, uid int, sess *session.ConversationSession)
	aiModelFactory  func(ctx context.Context) (interface {
		Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
	}, error)
	// textModelFactory 创建轻量级模型，用于继续请求分类等辅助任务。
	// 使用 ARK_TEXT_MODEL 环境变量以降低成本。如果为 nil，则回退到 AIModelFactory。
	textModelFactory func(ctx context.Context) (interface {
		Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
	}, error)
}

// ServerConfig 用于创建 Server 的配置结构。
type ServerConfig struct {
	Addr           string
	BaseDir        string
	FrontendDir    string
	SkillsDir      string
	Operator       commandline.Operator
	AgentFactory   task.AgentFactory
	MakeTaskConfig func(taskID string) *deck.PPTTaskConfig
	AIModelFactory func(ctx context.Context) (interface {
		Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
	}, error)
	// TextModelFactory 创建轻量级文本模型，用于继续请求分类等辅助任务。
	// 使用 ARK_TEXT_MODEL 环境变量（成本更低）。如果为 nil，则回退到 AIModelFactory。
	TextModelFactory func(ctx context.Context) (interface {
		Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
	}, error)
	// LogAnalysisModelFactory 创建用于后台日志分析的模型。
	// 如果为 nil，则禁用日志分析功能。
	LogAnalysisModelFactory func(ctx context.Context) (model.ToolCallingChatModel, error)
	// LogAnalysisIdleInterval 控制空闲日志分析的运行频率。
	// 默认为 5 分钟。设置为 0 可禁用空闲分析。
	LogAnalysisIdleInterval time.Duration
}

// NewServer 创建并初始化一个新的 Gin Server。
func NewServer(cfg *ServerConfig) *Server {
	frontendDir := cfg.FrontendDir
	if frontendDir == "" {
		frontendDir = filepath.Join("..", "frontend", "dist")
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		errMsg := "panic recovered"
		if err, ok := recovered.(error); ok {
			errMsg = err.Error()
		} else if s, ok := recovered.(string); ok {
			errMsg = s
		}
		stack := string(debug.Stack())
		// 只取前 20 行堆栈，避免日志过长
		lines := strings.Split(stack, "\n")
		if len(lines) > 40 {
			stack = strings.Join(lines[:40], "\n") + "\n... (truncated)"
		}
		logger.Error("http_panic", "error", errMsg, "stack", stack)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误，请稍后重试"})
	}))

	// Prometheus HTTP 指标中间件
	engine.Use(metricsMiddleware)

	engine.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	s := &Server{
		sessionManager:   session.NewSessionManager(),
		agentFactory:     cfg.AgentFactory,
		makeTaskConfig:   cfg.MakeTaskConfig,
		taskIDGen:        func() string { return uuid.New().String() },
		engine:           engine,
		addr:             cfg.Addr,
		aiModelFactory:   cfg.AIModelFactory,
		textModelFactory: cfg.TextModelFactory,
		skillDir:         cfg.SkillsDir,
		operator:         cfg.Operator,
	}

	// 创建任务管理器。风格要求由当前任务提示词显式携带。
	s.tasks = task.NewTaskManager(cfg.BaseDir,
		nil,
		func(taskID string) {
			if s.logAnalysis != nil {
				s.logAnalysis.Trigger(taskID, "failed")
			}
		},
		func(taskID string) {
			s.onTaskContinue(taskID)
		},
	)
	s.tasks.SetFileReadyCallback(func(taskID, workDir, filename string) {
		s.prepareThumbnail(taskID, workDir, filename)
	})
	s.tasks.SetAssistantTurnCallback(func(taskID, workDir, content string) {
		if err := s.sessionManager.GetOrCreate(taskID, workDir).AddAssistantMessage(content); err != nil {
			logger.Error("assistant_turn_persist_failed", "task_id", taskID, "error", err.Error())
		}
	})

	// 初始化日志分析后台服务
	if cfg.LogAnalysisModelFactory != nil {
		s.logAnalysis = loganalysis.NewService(&loganalysis.ServiceConfig{
			ModelFactory: cfg.LogAnalysisModelFactory,
			IdleInterval: cfg.LogAnalysisIdleInterval,
			LogLines:     300,
			SkillsDir:    cfg.SkillsDir,
		})
		loganalysis.HasRunningTasksFunc = s.tasks.HasRunningTasks
		s.logAnalysis.Start()
	}

	// 页面能力只从 component_contracts.json 加载。
	s.templateLoader = templates.NewComponentLoader(filepath.Join(cfg.SkillsDir, "ppt-deck-planner"))

	// 认证路由（公开）
	auth := engine.Group("/api/auth")
	{
		auth.POST("/send-code", s.handleSendCode)
		auth.POST("/login", s.handleLogin)
		auth.POST("/set-password", s.authMiddleware(), s.handleSetPassword)
		auth.POST("/logout", s.handleLogout)
		auth.GET("/me", s.authMiddleware(), s.handleMe)
	}

	// 任务路由（需要认证）
	tasks := engine.Group("/api/tasks")
	tasks.Use(s.authMiddleware())
	{
		tasks.POST("", s.handleCreateTask)
		tasks.GET("", s.handleListTasks)
		tasks.GET("/:id", s.taskOwnershipMiddleware(), s.handleGetTask)
		tasks.GET("/:id/stream", s.taskOwnershipMiddleware(), s.handleStreamTask)
		tasks.GET("/:id/files/:filename", s.taskOwnershipMiddleware(), s.handleDownloadFile)
		tasks.GET("/:id/thumb/:filename", s.taskOwnershipMiddleware(), s.handleThumbnail)
		tasks.POST("/:id/cancel", s.taskOwnershipMiddleware(), s.handleCancelTask)
		tasks.DELETE("/:id", s.taskOwnershipMiddleware(), s.handleDeleteTask)
		// 会话/继续路由
		tasks.POST("/:id/continue", s.taskOwnershipMiddleware(), s.handleContinueTask)
		tasks.GET("/:id/conversation", s.taskOwnershipMiddleware(), s.handleGetConversation)
		tasks.GET("/:id/runtime-events/:event_id", s.taskOwnershipMiddleware(), s.handleGetRuntimeEvent)
	}

	// 用户资料路由（需要认证）
	users := engine.Group("/api/users")
	users.Use(s.authMiddleware())
	{
		users.GET("/me/api-key", s.handleGetUserAPIKey)
		users.PUT("/me/api-key", s.handleUpdateUserAPIKey)
		users.DELETE("/me/api-key", s.handleDeleteUserAPIKey)
	}

	// 组件布局路由（公开）
	tpls := engine.Group("/api/templates")
	{
		tpls.GET("/layouts", s.handleListLayouts)
	}

	// 日志分析路由（需要认证）
	logs := engine.Group("/api/log-analyses")
	logs.Use(s.authMiddleware())
	{
		logs.GET("", s.handleListLogAnalyses)
		logs.GET("/task/:task_id", s.handleGetTaskLogAnalyses)
	}

	// 管理员路由（需要管理员权限）
	admin := engine.Group("/api/admin")
	admin.Use(s.adminMiddleware())
	{
		admin.GET("/stats", s.handleAdminStats)
		admin.GET("/users", s.handleAdminUsers)
		admin.GET("/tasks", s.handleAdminTasks)
		admin.GET("/log-analyses", s.handleAdminLogAnalyses)
		admin.DELETE("/log-analyses/:id", s.handleAdminDeleteLogAnalysis)
	}

	// 指标
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 健康检查（基础）
	engine.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 详细健康检查（K8s 探针，可选认证）
	engine.GET("/health/ready", s.handleHealthCheck)

	// 静态前端
	engine.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "api route not found"})
			return
		}
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

// metricsMiddleware 记录 HTTP 请求指标供 Prometheus 使用。
func metricsMiddleware(c *gin.Context) {
	start := time.Now()
	c.Next()
	duration := time.Since(start).Seconds()

	// 跳过指标端点本身以避免干扰
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	metrics.HTTPRequestsTotal.WithLabelValues(c.Request.Method, path, fmt.Sprintf("%d", c.Writer.Status())).Inc()
	metrics.HTTPRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
}

// Start 启动 HTTP 服务器。
func (s *Server) Start() error {
	logger.Info("server_starting", "addr", s.addr, "frontend", fmt.Sprintf("http://localhost%s", s.addr))
	return s.engine.Run(s.addr)
}
