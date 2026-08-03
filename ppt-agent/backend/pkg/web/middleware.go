package web

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/cloudwego/ppt-agent/pkg/auth"
)

const taskInfoContextKey = "ownedTaskInfo"

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractTokenFromGin(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		user, err := auth.ValidateSession(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		ctx := auth.WithUser(c.Request.Context(), user)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// taskOwnershipMiddleware keeps every task-specific route behind the same
// owner-or-admin authorization rule. Unauthorized task IDs intentionally look
// missing so the API does not expose whether another user's task exists.
func (s *Server) taskOwnershipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		info := s.tasks.GetTask(c.Param("id"))
		if info == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}

		if !canAccessTask(info.UserID, userIDGin(c), isAdminGin(c)) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}

		c.Set(taskInfoContextKey, info)
		c.Next()
	}
}

func canAccessTask(ownerID, userID int, isAdmin bool) bool {
	return isAdmin || (ownerID > 0 && ownerID == userID)
}

// adminMiddleware 要求用户已登录且具有管理员权限。
func (s *Server) adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractTokenFromGin(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		user, err := auth.ValidateSession(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if !user.IsAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			return
		}
		ctx := auth.WithUser(c.Request.Context(), user)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func extractTokenFromGin(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return h[7:]
	}
	if cookie, err := c.Cookie("session_token"); err == nil && cookie != "" {
		return cookie
	}
	return ""
}

func userIDGin(c *gin.Context) int {
	id, _ := auth.UserIDFromContext(c.Request.Context())
	return id
}

func isAdminGin(c *gin.Context) bool {
	id, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		return false
	}
	user, err := auth.ValidateUser(id)
	if err != nil || user == nil {
		return false
	}
	return user.IsAdmin
}
