package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"helix-api/internal/pkg/resp"
)

type AdminRoleHandler struct {
	db *sqlx.DB
}

func NewAdminRoleHandler(db *sqlx.DB) *AdminRoleHandler {
	return &AdminRoleHandler{db: db}
}

// List returns all admin roles.
// GET /admin/admin-roles
func (h *AdminRoleHandler) List(c *gin.Context) {
	type row struct {
		Name        string   `db:"name" json:"name"`
		Description string   `db:"description" json:"description"`
		Perms       []string `db:"-" json:"perms"`
		PermsJSON   string   `db:"perms_json" json:"-"`
	}
	roles := make([]row, 0)
	if err := h.db.Select(&roles, `
SELECT name,
       IFNULL(description, '') AS description,
       COALESCE(perms_json, JSON_ARRAY()) AS perms_json
FROM admin_roles
ORDER BY id ASC
`); err != nil {
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "查询角色失败")
		return
	}

	// Map perms_json -> perms slice, but keep output camelCase.
	for i := range roles {
		if roles[i].PermsJSON == "" {
			roles[i].Perms = []string{}
			continue
		}
		var perms []string
		if err := json.Unmarshal([]byte(roles[i].PermsJSON), &perms); err != nil {
			// On invalid JSON, fall back to empty slice to avoid breaking UI.
			perms = []string{}
		}
		roles[i].Perms = perms
		roles[i].PermsJSON = ""
	}

	resp.JSONOK(c, roles)
}
