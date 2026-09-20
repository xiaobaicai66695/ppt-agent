// Package router owns HTTP route registration.  It deliberately knows only
// about Gin handlers and middleware; task orchestration and JSON models stay
// in the service and model packages.
package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handlers is the adapter from the application server to the route table.
// Keeping functions (instead of a concrete web.Server) makes the route table
// independently testable and prevents a router/service import cycle.
type Handlers struct {
	AuthMiddleware      gin.HandlerFunc
	OwnershipMiddleware gin.HandlerFunc
	AdminMiddleware     gin.HandlerFunc

	SendCode    gin.HandlerFunc
	Register    gin.HandlerFunc
	Login       gin.HandlerFunc
	GuestLogin  gin.HandlerFunc
	SetPassword gin.HandlerFunc
	Logout      gin.HandlerFunc
	Me          gin.HandlerFunc

	Message          gin.HandlerFunc
	CreatePlanDraft  gin.HandlerFunc
	ListPlanDrafts   gin.HandlerFunc
	GetPlanDraft     gin.HandlerFunc
	CreateTask       gin.HandlerFunc
	StartTask        gin.HandlerFunc
	GetTask          gin.HandlerFunc
	ListTasks        gin.HandlerFunc
	StreamTask       gin.HandlerFunc
	DownloadFile     gin.HandlerFunc
	Thumbnail        gin.HandlerFunc
	CancelTask       gin.HandlerFunc
	SaveTaskFeedback gin.HandlerFunc
	DeleteTask       gin.HandlerFunc
	ContinueTask     gin.HandlerFunc
	GetConversation  gin.HandlerFunc
	GetRuntimeEvent  gin.HandlerFunc

	GetUserAPIKey    gin.HandlerFunc
	UpdateUserAPIKey gin.HandlerFunc
	DeleteUserAPIKey gin.HandlerFunc
	ListLayouts      gin.HandlerFunc

	AdminStats                       gin.HandlerFunc
	AdminUsers                       gin.HandlerFunc
	AdminTasks                       gin.HandlerFunc
	AdminExecutionRecords            gin.HandlerFunc
	AdminFeedback                    gin.HandlerFunc
	AdminEvaluationCandidates        gin.HandlerFunc
	AdminEvaluationCandidateEvidence gin.HandlerFunc
	AdminApproveEvaluationCandidate  gin.HandlerFunc
	AdminWithdrawEvaluationCandidate gin.HandlerFunc
	AdminCreateEvaluationCase        gin.HandlerFunc
	AdminPublishEvaluationDataset    gin.HandlerFunc
	AdminCreateEvaluationExport      gin.HandlerFunc
	AdminDownloadEvaluationExport    gin.HandlerFunc

	HealthCheck gin.HandlerFunc
	Metrics     http.Handler
	NoRoute     gin.HandlerFunc
}

// Register adds the complete public HTTP surface to engine.  Route paths and
// middleware order intentionally mirror the historical web.Server setup.
func Register(engine *gin.Engine, h Handlers) {
	if engine == nil {
		return
	}

	auth := engine.Group("/api/auth")
	{
		auth.POST("/send-code", handler(h.SendCode))
		auth.POST("/register", handler(h.Register))
		auth.POST("/login", handler(h.Login))
		auth.POST("/guest", handler(h.GuestLogin))
		auth.POST("/set-password", use(h.AuthMiddleware), handler(h.SetPassword))
		auth.POST("/logout", handler(h.Logout))
		auth.GET("/me", use(h.AuthMiddleware), handler(h.Me))
	}

	messages := engine.Group("/api/messages")
	messages.Use(use(h.AuthMiddleware))
	messages.POST("", handler(h.Message))

	planDrafts := engine.Group("/api/plan-drafts")
	planDrafts.Use(use(h.AuthMiddleware))
	{
		planDrafts.POST("", handler(h.CreatePlanDraft))
		planDrafts.GET("", handler(h.ListPlanDrafts))
		planDrafts.GET("/:id", handler(h.GetPlanDraft))
	}

	tasks := engine.Group("/api/tasks")
	tasks.Use(use(h.AuthMiddleware))
	{
		tasks.POST("", handler(h.CreateTask))
		tasks.POST("/:id/start", use(h.OwnershipMiddleware), handler(h.StartTask))
		tasks.GET("", handler(h.ListTasks))
		tasks.GET("/:id", use(h.OwnershipMiddleware), handler(h.GetTask))
		tasks.GET("/:id/stream", use(h.OwnershipMiddleware), handler(h.StreamTask))
		tasks.GET("/:id/files/:filename", use(h.OwnershipMiddleware), handler(h.DownloadFile))
		tasks.GET("/:id/thumb/:filename", use(h.OwnershipMiddleware), handler(h.Thumbnail))
		tasks.POST("/:id/cancel", use(h.OwnershipMiddleware), handler(h.CancelTask))
		tasks.PUT("/:id/feedback", use(h.OwnershipMiddleware), handler(h.SaveTaskFeedback))
		tasks.DELETE("/:id", use(h.OwnershipMiddleware), handler(h.DeleteTask))
		tasks.POST("/:id/continue", use(h.OwnershipMiddleware), handler(h.ContinueTask))
		tasks.GET("/:id/conversation", use(h.OwnershipMiddleware), handler(h.GetConversation))
		tasks.GET("/:id/runtime-events/:event_id", use(h.OwnershipMiddleware), handler(h.GetRuntimeEvent))
	}

	users := engine.Group("/api/users")
	users.Use(use(h.AuthMiddleware))
	{
		users.GET("/me/api-key", handler(h.GetUserAPIKey))
		users.PUT("/me/api-key", handler(h.UpdateUserAPIKey))
		users.DELETE("/me/api-key", handler(h.DeleteUserAPIKey))
	}

	templates := engine.Group("/api/templates")
	templates.GET("/layouts", handler(h.ListLayouts))

	admin := engine.Group("/api/admin")
	admin.Use(use(h.AdminMiddleware))
	{
		admin.GET("/stats", handler(h.AdminStats))
		admin.GET("/users", handler(h.AdminUsers))
		admin.GET("/tasks", handler(h.AdminTasks))
		admin.GET("/execution-records", handler(h.AdminExecutionRecords))
		admin.GET("/feedback", handler(h.AdminFeedback))
		admin.GET("/evaluations/candidates", handler(h.AdminEvaluationCandidates))
		admin.GET("/evaluations/candidates/:id", handler(h.AdminEvaluationCandidateEvidence))
		admin.POST("/evaluations/candidates/:id/approve", handler(h.AdminApproveEvaluationCandidate))
		admin.POST("/evaluations/candidates/:id/withdraw", handler(h.AdminWithdrawEvaluationCandidate))
		admin.POST("/evaluations/candidates/:id/cases", handler(h.AdminCreateEvaluationCase))
		admin.POST("/evaluations/datasets", handler(h.AdminPublishEvaluationDataset))
		admin.POST("/evaluations/datasets/:id/exports", handler(h.AdminCreateEvaluationExport))
		admin.GET("/evaluations/exports/:id/download", handler(h.AdminDownloadEvaluationExport))
	}

	if h.Metrics != nil {
		engine.GET("/metrics", gin.WrapH(h.Metrics))
	} else {
		engine.GET("/metrics", func(c *gin.Context) { c.Status(http.StatusNotImplemented) })
	}
	engine.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	engine.GET("/health/ready", handler(h.HealthCheck))
	engine.NoRoute(noRoute(h.NoRoute))
}

func handler(fn gin.HandlerFunc) gin.HandlerFunc {
	if fn != nil {
		return fn
	}
	return func(c *gin.Context) { c.Status(http.StatusNotImplemented) }
}

func use(fn gin.HandlerFunc) gin.HandlerFunc {
	if fn != nil {
		return fn
	}
	return func(c *gin.Context) { c.Next() }
}

func noRoute(fn gin.HandlerFunc) gin.HandlerFunc {
	if fn != nil {
		return fn
	}
	return func(c *gin.Context) { c.Status(http.StatusNotFound) }
}
