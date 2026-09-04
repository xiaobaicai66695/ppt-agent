package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/cloudwego/ppt-agent/pkg/db"
)

func signedTestSession(t *testing.T, userID uint, email string, isAdmin bool) string {
	t.Helper()
	claims := jwt.MapClaims{
		"user_id":  userID,
		"email":    email,
		"is_admin": isAdmin,
		"exp":      time.Now().Add(time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("pptagent"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	return token
}

func TestAuthMiddlewareAcceptsValidJWTWhenDatabaseIsUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousDB := db.DB
	db.DB = nil
	t.Cleanup(func() { db.DB = previousDB })

	server := &Server{}
	router := gin.New()
	router.GET("/protected", server.authMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user_id": userIDGin(c), "is_admin": isAdminGin(c)})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signedTestSession(t, 42, "user@example.com", false))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestAdminMiddlewareReturnsServiceUnavailableForDatabaseFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousDB := db.DB
	db.DB = nil
	t.Cleanup(func() { db.DB = previousDB })

	server := &Server{}
	router := gin.New()
	router.GET("/admin", server.adminMiddleware(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+signedTestSession(t, 1, "admin@example.com", true))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Retry-After") != "5" {
		t.Fatalf("Retry-After = %q", recorder.Header().Get("Retry-After"))
	}
}
