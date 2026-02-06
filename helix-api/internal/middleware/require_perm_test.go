package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"helix-api/internal/domain"
	"helix-api/internal/pkg/token"
)

func withClaims(perms []string) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(ClaimsKey, &token.Claims{
		Perms: perms,
	})
	return c
}

func TestRequireAny_AllowsWithPerm(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c := withClaims([]string{"foo.read"})

	called := false
	h := RequireAny("foo.read")
	h(c)
	if c.IsAborted() {
		t.Fatalf("expected context not aborted")
	}
	if !called {
		// Wrap handler to ensure Next is called; here we just assert not aborted.
	}
}

func TestRequireAny_AllowsWithAdminManage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c := withClaims([]string{domain.PermAdminManage})

	h := RequireAny("non.existing")
	h(c)
	if c.IsAborted() {
		t.Fatalf("expected context not aborted when admin.manage present")
	}
}

func TestRequireAll_DeniesMissingPerm(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(ClaimsKey, &token.Claims{
		Perms: []string{"a", "b"},
	})

	h := RequireAll("a", "c")
	h(c)

	if !c.IsAborted() {
		t.Fatalf("expected context aborted when missing required perm")
	}
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}
