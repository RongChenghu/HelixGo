package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"helix-api/internal/pkg/resp"
	"helix-api/internal/pkg/token"
)

const ClaimsKey = "jwtClaims"

func JWTAuth(tokens *token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			resp.JSONError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			resp.JSONError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization format")
			c.Abort()
			return
		}

		claims, err := tokens.Parse(parts[1])
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				resp.JSONError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Token expired")
				c.Abort()
				return
			}
			resp.JSONError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token")
			c.Abort()
			return
		}

		if claims.Subject == "" || claims.Name == "" {
			resp.JSONError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token payload")
			c.Abort()
			return
		}

		c.Set(ClaimsKey, claims)
		c.Next()
	}
}
