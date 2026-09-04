package web

import (
	"net"
	"net/http"
	"strings"

	"github.com/cloudwego/ppt-agent/pkg/auth"
	"github.com/gin-gonic/gin"
)

func (s *Server) handleSendCode(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
		return
	}
	if err := auth.SendCode(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "验证码已发送"})
}

func (s *Server) handleLogin(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Code     string `json:"code"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
		return
	}
	if req.Code != "" {
		token, user, err := auth.LoginWithCode(req.Email, req.Code)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token, "id": user.ID, "email": user.Email, "is_admin": user.IsAdmin})
		return
	}
	if req.Password != "" {
		token, user, err := auth.LoginWithPassword(req.Email, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token, "id": user.ID, "email": user.Email, "is_admin": user.IsAdmin})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "请提供验证码或密码"})
}

func (s *Server) handleRegister(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Code     string `json:"code"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
		return
	}
	token, user, err := auth.Register(req.Email, req.Code, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"token": token, "id": user.ID, "email": user.Email, "is_admin": user.IsAdmin})
}

func (s *Server) handleGuestLogin(c *gin.Context) {
	token, user, err := auth.LoginAsGuest(guestRemoteIP(c))
	if err != nil {
		status := http.StatusInternalServerError
		if !auth.GuestLoginEnabled() {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "id": user.ID, "email": user.Email, "is_guest": true, "is_admin": false})
}

func guestRemoteIP(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(c.Request.RemoteAddr)
}

func (s *Server) handleSetPassword(c *gin.Context) {
	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
		return
	}
	if err := auth.SetPassword(userIDGin(c), req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) handleLogout(c *gin.Context) {
	if token := extractTokenFromGin(c); token != "" {
		auth.Logout(token)
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) handleMe(c *gin.Context) {
	uid := userIDGin(c)
	email, _ := auth.UsernameFromContext(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"id": uid, "email": email, "is_admin": isAdminGin(c), "is_guest": auth.IsGuestEmail(email)})
}
