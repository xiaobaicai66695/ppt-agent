package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandleGuestLoginReportsDisabledFeature(t *testing.T) {
	t.Setenv("GUEST_LOGIN_ENABLED", "false")
	gin.SetMode(gin.TestMode)
	server := &Server{}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/auth/guest", nil)

	server.handleGuestLogin(ctx)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}
