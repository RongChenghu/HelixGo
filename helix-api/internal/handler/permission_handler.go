package handler

import (
	"github.com/gin-gonic/gin"

	"helix-api/internal/domain"
	"helix-api/internal/pkg/resp"
)

type PermissionHandler struct{}

func NewPermissionHandler() *PermissionHandler {
	return &PermissionHandler{}
}

// List returns all known permissions.
// GET /admin/permissions
func (h *PermissionHandler) List(c *gin.Context) {
	perms := domain.AllPermissions()
	if perms == nil {
		resp.JSONOK(c, []interface{}{})
		return
	}
	resp.JSONOK(c, perms)
}
