package web

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/cloudwego/ppt-agent/pkg/auth"
)

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
