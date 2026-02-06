package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()
		traceID, _ := c.Get(RequestIDKey)

		log.Printf("traceId=%v method=%s path=%s status=%d latency=%s",
			traceID, c.Request.Method, c.Request.URL.Path, status, latency)
	}
}
