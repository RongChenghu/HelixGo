package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"helix-api/internal/config"
	"helix-api/internal/domain"
	"helix-api/internal/middleware"
	"helix-api/internal/pkg/token"
)

// TestSystemConfigWritePermission verifies that the route-level middleware enforces permissions
// for PUT /admin/system/configs/:key.
func TestSystemConfigWritePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		JwtSecret:  "secret",
		JwtExpires: time.Minute,
	}
	tokens := token.NewManager(cfg.JwtSecret, cfg.JwtExpires)

	engine := gin.New()
	engine.PUT("/admin/system/configs/:key",
		middleware.JWTAuth(tokens),
		middleware.RequireAny(domain.PermAdminManage, domain.PermSystemConfigWrite),
		func(c *gin.Context) {
			// simple OK handler to differentiate from middleware 403
			c.Status(http.StatusOK)
		},
	)

	// helper to issue token with given perms
	issue := func(perms []string) string {
		t.Helper()
		jwtToken, err := tokens.Issue("1", "test", []string{"admin"}, perms)
		if err != nil {
			t.Fatalf("issue token: %v", err)
		}
		return jwtToken
	}

	// Token without system.config.write or admin.manage -> 403
	req1, _ := http.NewRequest(http.MethodPut, "/admin/system/configs/foo", nil)
	req1.Header.Set("Authorization", "Bearer "+issue([]string{}))
	w1 := httptest.NewRecorder()
	engine.ServeHTTP(w1, req1)
	if w1.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for missing write perm, got %d", w1.Code)
	}

	// Token with system.config.write -> 200
	req2, _ := http.NewRequest(http.MethodPut, "/admin/system/configs/foo", nil)
	req2.Header.Set("Authorization", "Bearer "+issue([]string{domain.PermSystemConfigWrite}))
	w2 := httptest.NewRecorder()
	engine.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 when system.config.write present, got %d", w2.Code)
	}

	// Token with admin.manage -> 200
	req3, _ := http.NewRequest(http.MethodPut, "/admin/system/configs/foo", nil)
	req3.Header.Set("Authorization", "Bearer "+issue([]string{domain.PermAdminManage}))
	w3 := httptest.NewRecorder()
	engine.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("expected 200 when admin.manage present, got %d", w3.Code)
	}
}
