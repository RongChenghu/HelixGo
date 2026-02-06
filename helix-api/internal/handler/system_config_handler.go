package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"helix-api/internal/domain"
	"helix-api/internal/middleware"
	"helix-api/internal/pkg/resp"
	"helix-api/internal/pkg/token"
	"helix-api/internal/service"
)

type SystemConfigHandler struct {
	service *service.SystemConfigService
	audit   *service.AuditService
}

func NewSystemConfigHandler(service *service.SystemConfigService, audit *service.AuditService) *SystemConfigHandler {
	return &SystemConfigHandler{service: service, audit: audit}
}

func (h *SystemConfigHandler) List(c *gin.Context) {
	configs := h.service.List()
	resp.JSONOK(c, configs)
}

type configUpdateRequest struct {
	Value interface{} `json:"value"`
}

func (h *SystemConfigHandler) Upsert(c *gin.Context) {
	key := c.Param("key")
	var req configUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Value == nil {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "配置值不能为空")
		return
	}

	value := fmt.Sprintf("%v", req.Value)
	h.service.Upsert(key, value)
	resp.JSONOK(c, gin.H{"ok": true})

	if h.audit == nil {
		return
	}
	claimsAny, ok := c.Get(middleware.ClaimsKey)
	if !ok {
		return
	}
	claims := claimsAny.(*token.Claims)
	traceID := ""
	if value, ok := c.Get(middleware.RequestIDKey); ok {
		if str, ok := value.(string); ok {
			traceID = str
		}
	}
	h.audit.Append(domain.AuditLog{
		Action:       "system.config.update",
		Method:       c.Request.Method,
		Path:         c.Request.URL.Path,
		Status:       http.StatusOK,
		OperatorID:   claims.Subject,
		OperatorName: claims.Name,
		IP:           c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
		TraceID:      traceID,
	})
}
