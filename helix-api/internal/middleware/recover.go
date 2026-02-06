package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"helix-api/internal/pkg/resp"
)

func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recoverErr := recover(); recoverErr != nil {
				resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "Internal Server Error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
