package handler

import "github.com/gin-gonic/gin"

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Healthz(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}

func (h *HealthHandler) Version(c *gin.Context) {
	c.JSON(200, gin.H{"version": "v0.1"})
}
