package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"helix-api/internal/domain"
	"helix-api/internal/middleware"
	"helix-api/internal/pkg/resp"
	"helix-api/internal/pkg/token"
	"helix-api/internal/service"
)

type AdminRoleHandler struct {
	db    *sqlx.DB
	audit *service.AuditService
}

func NewAdminRoleHandler(db *sqlx.DB, audit *service.AuditService) *AdminRoleHandler {
	return &AdminRoleHandler{db: db, audit: audit}
}

// List returns all admin roles.
// GET /admin/admin-roles
func (h *AdminRoleHandler) List(c *gin.Context) {
	type row struct {
		ID          int64    `db:"id" json:"id"`
		Name        string   `db:"name" json:"name"`
		Description string   `db:"description" json:"description"`
		Perms       []string `db:"-" json:"perms"`
		PermsJSON   string   `db:"perms_json" json:"-"`
	}
	roles := make([]row, 0)
	if err := h.db.Select(&roles, `
SELECT id,
       name,
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
			perms = []string{}
		}
		roles[i].Perms = perms
		roles[i].PermsJSON = ""
	}

	resp.JSONOK(c, roles)
}

var allowedPermCodes = func() map[string]struct{} {
	m := make(map[string]struct{})
	for _, p := range domain.AllPermissions() {
		m[p.Code] = struct{}{}
	}
	return m
}()

func validatePerms(perms []string) bool {
	for _, code := range perms {
		if code == "" {
			continue
		}
		if _, ok := allowedPermCodes[code]; !ok {
			return false
		}
	}
	return true
}

type createRoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Perms       []string `json:"perms"`
}

// Create creates a new admin role.
// POST /admin/admin-roles
func (h *AdminRoleHandler) Create(c *gin.Context) {
	var req createRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "请求体无效")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > 64 {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "角色名称不能为空且不超过64字符")
		return
	}
	if req.Perms != nil && !validatePerms(req.Perms) {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "包含无效的权限代码")
		return
	}
	permsJSON := "[]"
	if len(req.Perms) > 0 {
		b, _ := json.Marshal(req.Perms)
		permsJSON = string(b)
	}
	desc := strings.TrimSpace(req.Description)
	if len(desc) > 255 {
		desc = desc[:255]
	}

	res, err := h.db.Exec(`
INSERT INTO admin_roles (name, description, perms_json)
VALUES (?, ?, ?)
`, name, desc, permsJSON)
	if err != nil {
		if me, ok := err.(*mysql.MySQLError); ok && me.Number == 1062 {
			resp.JSONError(c, http.StatusBadRequest, "ROLE_NAME_EXISTS", "角色名称已存在")
			return
		}
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "创建角色失败")
		return
	}
	newID, err := res.LastInsertId()
	if err != nil || newID == 0 {
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "创建角色失败")
		return
	}

	resp.JSONOK(c, gin.H{"id": newID, "name": name})
	h.appendAudit(c, "admin.role.create", http.StatusOK, strconv.FormatInt(newID, 10))
}

type updateRoleRequest struct {
	Description *string  `json:"description"`
	Perms       []string `json:"perms"`
}

// Update updates an admin role (description and perms only).
// PUT /admin/admin-roles/:id
func (h *AdminRoleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "无效的角色ID")
		return
	}
	var req updateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "请求体无效")
		return
	}
	if req.Perms != nil && !validatePerms(req.Perms) {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "包含无效的权限代码")
		return
	}

	type currentRow struct {
		Name        string `db:"name"`
		Description string `db:"description"`
		PermsJSON   string `db:"perms_json"`
	}
	var cur currentRow
	if err := h.db.Get(&cur, `SELECT name, IFNULL(description,'') AS description, COALESCE(perms_json, JSON_ARRAY()) AS perms_json FROM admin_roles WHERE id = ?`, id); err != nil {
		resp.JSONError(c, http.StatusNotFound, "ROLE_NOT_FOUND", "角色不存在")
		return
	}

	desc := cur.Description
	if req.Description != nil {
		desc = strings.TrimSpace(*req.Description)
		if len(desc) > 255 {
			desc = desc[:255]
		}
	}
	permsJSON := cur.PermsJSON
	if req.Perms != nil {
		b, _ := json.Marshal(req.Perms)
		permsJSON = string(b)
	}

	result, err := h.db.Exec(`
UPDATE admin_roles SET description = ?, perms_json = ? WHERE id = ?
`, desc, permsJSON, id)
	if err != nil {
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "更新角色失败")
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		resp.JSONError(c, http.StatusNotFound, "ROLE_NOT_FOUND", "角色不存在")
		return
	}

	resp.JSONOK(c, gin.H{"id": id, "name": cur.Name})
	h.appendAudit(c, "admin.role.update", http.StatusOK, strconv.FormatInt(id, 10))
}

// Delete deletes an admin role. Built-in role "admin" cannot be deleted.
// DELETE /admin/admin-roles/:id
func (h *AdminRoleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "无效的角色ID")
		return
	}

	var name string
	if err := h.db.Get(&name, `SELECT name FROM admin_roles WHERE id = ?`, id); err != nil {
		resp.JSONError(c, http.StatusNotFound, "ROLE_NOT_FOUND", "角色不存在")
		return
	}
	if name == "admin" {
		resp.JSONError(c, http.StatusForbidden, "ROLE_PROTECTED", "不可删除系统内置角色 admin")
		return
	}

	_, err = h.db.Exec(`DELETE FROM admin_user_roles WHERE role_id = ?`, id)
	if err != nil {
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "删除角色失败")
		return
	}
	result, err := h.db.Exec(`DELETE FROM admin_roles WHERE id = ?`, id)
	if err != nil {
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "删除角色失败")
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		resp.JSONError(c, http.StatusNotFound, "ROLE_NOT_FOUND", "角色不存在")
		return
	}

	resp.JSONOK(c, gin.H{"id": id})
	h.appendAudit(c, "admin.role.delete", http.StatusOK, strconv.FormatInt(id, 10))
}

func (h *AdminRoleHandler) appendAudit(c *gin.Context, action string, status int, targetID string) {
	if h.audit == nil {
		return
	}
	traceID := ""
	if value, ok := c.Get(middleware.RequestIDKey); ok {
		if str, ok := value.(string); ok {
			traceID = str
		}
	}
	var operatorID, operator string
	if claimsAny, ok := c.Get(middleware.ClaimsKey); ok {
		if claims, ok := claimsAny.(*token.Claims); ok {
			operatorID = claims.Subject
			operator = claims.Name
		}
	}
	_ = targetID
	h.audit.Append(domain.AuditLog{
		Action:       action,
		Method:       c.Request.Method,
		Path:         c.Request.URL.Path,
		Status:       status,
		OperatorID:   operatorID,
		OperatorName: operator,
		IP:           c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
		TraceID:      traceID,
	})
}
