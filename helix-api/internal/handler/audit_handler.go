package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"helix-api/internal/pkg/resp"
	"helix-api/internal/service"
)

type AuditHandler struct {
	service *service.AuditService
}

func NewAuditHandler(service *service.AuditService) *AuditHandler {
	return &AuditHandler{service: service}
}

func (h *AuditHandler) ListLogs(c *gin.Context) {
	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("pageSize"), 10)
	limit := parseInt(c.Query("limit"), 0)
	offset := parseInt(c.Query("offset"), 0)
	keyword := c.Query("keyword")
	action := c.Query("action")
	from := c.Query("from")
	to := c.Query("to")

	result := h.service.ListPage(service.AuditListParams{
		Page:     page,
		PageSize: pageSize,
		Limit:    limit,
		Offset:   offset,
		Keyword:  keyword,
		Action:   action,
		From:     from,
		To:       to,
	})

	list := make([]gin.H, 0, len(result.List))
	for _, item := range result.List {
		list = append(list, gin.H{
			"id":            item.ID,
			"action":        item.Action,
			"method":        item.Method,
			"path":          item.Path,
			"status":        item.Status,
			"ip":            item.IP,
			"adminUserId":   item.OperatorID,
			"adminUsername": item.OperatorName,
			"userAgent":     item.UserAgent,
			"traceId":       item.TraceID,
			"createdAt":     item.CreatedAt.Format(time.RFC3339),
		})
	}

	resp.JSONOK(c, gin.H{
		"list":     list,
		"total":    result.Total,
		"page":     result.Page,
		"pageSize": result.PageSize,
	})
}

func parseInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	if val, err := strconv.Atoi(raw); err == nil {
		return val
	}
	return fallback
}
