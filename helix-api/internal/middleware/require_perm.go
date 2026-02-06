package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"helix-api/internal/domain"
	"helix-api/internal/pkg/resp"
	"helix-api/internal/pkg/token"
)

// RequirePerm checks that current JWT claims contain the given permission.
// It is a shorthand for RequireAny with a single permission.
func RequirePerm(perm string) gin.HandlerFunc {
	return RequireAny(perm)
}

// RequireAny checks that JWT claims contain at least one of the given permissions.
// If claims.Perms contains admin.manage, it always allows (super permission).
// On failure it returns legacy-style JSON error {error,message} with 403.
func RequireAny(perms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := getClaims(c)
		if !ok {
			return
		}

		if hasPerm(claims.Perms, domain.PermAdminManage) {
			c.Next()
			return
		}

		for _, want := range perms {
			if hasPerm(claims.Perms, want) {
				c.Next()
				return
			}
		}

		resp.JSONError(c, http.StatusForbidden, "FORBIDDEN", "Permission denied")
		c.Abort()
	}
}

// RequireAll checks that JWT claims contain all of the given permissions.
// If claims.Perms contains admin.manage, it always allows (super permission).
func RequireAll(perms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := getClaims(c)
		if !ok {
			return
		}

		if hasPerm(claims.Perms, domain.PermAdminManage) {
			c.Next()
			return
		}

		for _, want := range perms {
			if !hasPerm(claims.Perms, want) {
				resp.JSONError(c, http.StatusForbidden, "FORBIDDEN", "Permission denied")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func getClaims(c *gin.Context) (*token.Claims, bool) {
	claimsAny, ok := c.Get(ClaimsKey)
	if !ok {
		resp.JSONError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token payload")
		c.Abort()
		return nil, false
	}
	claims, ok := claimsAny.(*token.Claims)
	if !ok {
		resp.JSONError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token payload")
		c.Abort()
		return nil, false
	}
	return claims, true
}

func hasPerm(perms []string, want string) bool {
	for _, p := range perms {
		if p == want {
			return true
		}
	}
	return false
}
