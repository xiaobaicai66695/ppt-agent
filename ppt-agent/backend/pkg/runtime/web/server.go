package web

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/cloudwego/ppt-agent/pkg/agent/ppt"
	"github.com/cloudwego/ppt-agent/pkg/chattrace"
	"github.com/cloudwego/ppt-agent/pkg/evaluation"
	"github.com/cloudwego/ppt-agent/pkg/runtime/task"
	webrouter "github.com/cloudwego/ppt-agent/pkg/runtime/web/router"
	webservice "github.com/cloudwego/ppt-agent/pkg/runtime/web/service"
	"github.com/cloudwego/ppt-agent/pkg/session"
	"github.com/cloudwego/ppt-agent/pkg/templates"
	"github.com/cloudwego/ppt-agent/pkg/utils/logger"
	"github.com/cloudwego/ppt-agent/pkg/utils/metrics"
)

// Server 提供 REST API + SSE 流式推送 + 静态前端服务（基于 Gin 框架）。
type Server struct {
	tasks           *task.TaskManager
	sessionManager  *session.SessionManager
	agentFactory    task.AgentFactory
	makeTaskConfig  func(taskID string) *ppt.PPTTaskConfig
	taskIDGen       func() string
	engine          *gin.Engine
	addr            string
	templateLoader  *templates.Loader
	skillDir        string
	operator        commandline.Operator
	httpServer      *http.Server
	runContext      context.Context
	chatTrace       chattrace.Store
	evaluation      *evaluation.Service
	taskService     *webservice.TaskService
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
	MakeTaskConfig func(taskID string) *ppt.PPTTaskConfig
	AIModelFactory func(ctx context.Context) (interface {
		Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
	}, error)
	// TextModelFactory 创建轻量级文本模型，用于继续请求分类等辅助任务。
	// 使用 ARK_TEXT_MODEL 环境变量（成本更低）。如果为 nil，则回退到 AIModelFactory。
	TextModelFactory func(ctx context.Context) (interface {
		Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
	}, error)
	// ChatTraceStore is a Redis-backed transient store for safe tool traces.
	// It must not be replaced with a MySQL implementation.
	ChatTraceStore chattrace.Store
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
	engine.Use(requestBodyLimitMiddleware)

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
		chatTrace:        cfg.ChatTraceStore,
		evaluation:       evaluation.NewService(""),
		runContext:       context.Background(),
	}

	// 创建任务管理器。风格要求由当前任务提示词显式携带。
	s.tasks = task.NewTaskManager(cfg.BaseDir,
		nil,
		nil,
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
	s.tasks.SetTimelineEventCallback(s.persistConversationTimelineEvent)

	// 页面能力只从 component_contracts.json 加载。
	s.templateLoader = templates.NewComponentLoader(filepath.Join(cfg.SkillsDir, "ppt-planner"))

	// TaskService owns task lifecycle orchestration; the router package below
	// only receives handler functions and middleware adapters.
	s.taskService = webservice.NewTaskService(webservice.TaskServiceConfig{
		Tasks:          s.tasks,
		Sessions:       s.sessionManager,
		AgentFactory:   s.agentFactory,
		MakeTaskConfig: s.makeTaskConfig,
		TemplateLoader: s.templateLoader,
	})
	webrouter.Register(engine, webrouter.Handlers{
		AuthMiddleware:                   s.authMiddleware(),
		OwnershipMiddleware:              s.taskOwnershipMiddleware(),
		AdminMiddleware:                  s.adminMiddleware(),
		SendCode:                         s.handleSendCode,
		Register:                         s.handleRegister,
		Login:                            s.handleLogin,
		GuestLogin:                       s.handleGuestLogin,
		SetPassword:                      s.handleSetPassword,
		Logout:                           s.handleLogout,
		Me:                               s.handleMe,
		Message:                          s.handleMessage,
		CreatePlanDraft:                  s.handleCreatePlanDraft,
		ListPlanDrafts:                   s.handleListPlanDrafts,
		GetPlanDraft:                     s.handleGetPlanDraft,
		CreateTask:                       s.handleCreateTask,
		StartTask:                        s.handleStartConversationTask,
		GetTask:                          s.handleGetTask,
		ListTasks:                        s.handleListTasks,
		StreamTask:                       s.handleStreamTask,
		DownloadFile:                     s.handleDownloadFile,
		Thumbnail:                        s.handleThumbnail,
		CancelTask:                       s.handleCancelTask,
		SaveTaskFeedback:                 s.handleSaveTaskFeedback,
		DeleteTask:                       s.handleDeleteTask,
		ContinueTask:                     s.handleContinueTask,
		GetConversation:                  s.handleGetConversation,
		GetRuntimeEvent:                  s.handleGetRuntimeEvent,
		GetUserAPIKey:                    s.handleGetUserAPIKey,
		UpdateUserAPIKey:                 s.handleUpdateUserAPIKey,
		DeleteUserAPIKey:                 s.handleDeleteUserAPIKey,
		ListLayouts:                      s.handleListLayouts,
		AdminStats:                       s.handleAdminStats,
		AdminUsers:                       s.handleAdminUsers,
		AdminTasks:                       s.handleAdminTasks,
		AdminFeedback:                    s.handleAdminFeedback,
		AdminEvaluationCandidates:        s.handleAdminEvaluationCandidates,
		AdminEvaluationCandidateEvidence: s.handleAdminEvaluationCandidateEvidence,
		AdminApproveEvaluationCandidate:  s.handleAdminApproveEvaluationCandidate,
		AdminWithdrawEvaluationCandidate: s.handleAdminWithdrawEvaluationCandidate,
		AdminCreateEvaluationCase:        s.handleAdminCreateEvaluationCase,
		AdminPublishEvaluationDataset:    s.handleAdminPublishEvaluationDataset,
		AdminCreateEvaluationExport:      s.handleAdminCreateEvaluationExport,
		AdminDownloadEvaluationExport:    s.handleAdminDownloadEvaluationExport,
		HealthCheck:                      s.handleHealthCheck,
		Metrics:                          promhttp.Handler(),
		NoRoute: func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(http.StatusNotFound, gin.H{"error": "api route not found"})
				return
			}
			serveStatic(c, frontendDir)
		},
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
	return s.StartContext(context.Background())
}

// StartContext serves HTTP until ctx is cancelled, then performs a bounded
// graceful shutdown. Explicit server timeouts prevent slow clients from
// holding connections forever (SSE handlers remain long-lived by design).
func (s *Server) StartContext(ctx context.Context) error {
	s.runContext = ctx
	s.tasks.SetBaseContext(ctx)
	logger.Info("server_starting", "addr", s.addr, "frontend", fmt.Sprintf("http://localhost%s", s.addr))
	readHeaderTimeout := durationEnv("HTTP_READ_HEADER_TIMEOUT", 10*time.Second)
	readTimeout := durationEnv("HTTP_READ_TIMEOUT", 30*time.Second)
	idleTimeout := durationEnv("HTTP_IDLE_TIMEOUT", 2*time.Minute)
	writeTimeout := durationEnv("HTTP_WRITE_TIMEOUT", 0)
	server := &http.Server{Handler: s.engine, ReadHeaderTimeout: readHeaderTimeout, ReadTimeout: readTimeout, IdleTimeout: idleTimeout, WriteTimeout: writeTimeout}
	s.httpServer = server
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	errCh := make(chan error, 1)
	go func() { errCh <- server.Serve(listener) }()
	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return s.Shutdown(shutdownCtx)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) runtimeContext() context.Context {
	if s.runContext != nil {
		return s.runContext
	}
	return context.Background()
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil && parsed >= 0 {
			return parsed
		}
		logger.Warn("invalid_http_timeout", "key", key, "value", value)
	}
	return fallback
}

const maxRequestBodyBytes = 4 << 20

func requestBodyLimitMiddleware(c *gin.Context) {
	if c.Request.Body != nil && c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodyBytes)
	}
	c.Next()
}
