package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

	"helix-api/internal/domain"
	"helix-api/internal/middleware"
	"helix-api/internal/pkg/resp"
	"helix-api/internal/pkg/token"
	"helix-api/internal/service"
)

// AdminUserHandler provides basic admin user management over MySQL.
type AdminUserHandler struct {
	db    *sqlx.DB
	audit *service.AuditService
}

func NewAdminUserHandler(db *sqlx.DB, audit *service.AuditService) *AdminUserHandler {
	return &AdminUserHandler{db: db, audit: audit}
}

type adminUserDTO struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	IsEnabled bool     `json:"isEnabled"`
	Roles     []string `json:"roles"`
	CreatedAt string   `json:"createdAt,omitempty"`
}

// List returns all admin users with roles.
// GET /admin/admin-users
func (h *AdminUserHandler) List(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	pageSize := parseIntDefault(c.Query("pageSize"), 10)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	keyword := strings.TrimSpace(c.Query("keyword"))

	// total
	var total int
	var countQuery string
	var args []interface{}
	if keyword != "" {
		countQuery = `SELECT COUNT(*) FROM admin_users WHERE username LIKE ?`
		args = append(args, "%"+keyword+"%")
	} else {
		countQuery = `SELECT COUNT(*) FROM admin_users`
	}
	if err := h.db.Get(&total, countQuery, args...); err != nil {
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "查询管理员列表失败")
		return
	}

	type userRow struct {
		ID        int64        `db:"id"`
		Username  string       `db:"username"`
		IsEnabled bool         `db:"is_enabled"`
		CreatedAt sql.NullTime `db:"created_at"`
	}
	users := make([]userRow, 0)
	var listQuery string
	args = args[:0]
	if keyword != "" {
		listQuery = `
SELECT id, username, is_enabled, created_at
FROM admin_users
WHERE username LIKE ?
ORDER BY id ASC
LIMIT ? OFFSET ?
`
		args = append(args, "%"+keyword+"%", pageSize, offset)
	} else {
		listQuery = `
SELECT id, username, is_enabled, created_at
FROM admin_users
ORDER BY id ASC
LIMIT ? OFFSET ?
`
		args = append(args, pageSize, offset)
	}

	if err := h.db.Select(&users, `
`+listQuery+``, args...); err != nil {
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "查询管理员列表失败")
		return
	}

	type roleRow struct {
		UserID int64  `db:"user_id"`
		Name   string `db:"name"`
	}
	roleRows := make([]roleRow, 0)
	if err := h.db.Select(&roleRows, `
SELECT ur.user_id, r.name
FROM admin_user_roles ur
JOIN admin_roles r ON r.id = ur.role_id
`); err != nil {
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "查询管理员角色失败")
		return
	}

	roleMap := make(map[int64][]string)
	for _, r := range roleRows {
		roleMap[r.UserID] = append(roleMap[r.UserID], r.Name)
	}

	result := make([]adminUserDTO, 0, len(users))
	for _, u := range users {
		roles := roleMap[u.ID]
		if roles == nil {
			roles = []string{}
		}
		dto := adminUserDTO{
			ID:        u.ID,
			Username:  u.Username,
			IsEnabled: u.IsEnabled,
			Roles:     roles,
		}
		if u.CreatedAt.Valid {
			dto.CreatedAt = u.CreatedAt.Time.Format(time.RFC3339)
		}
		result = append(result, dto)
	}

	resp.JSONOK(c, gin.H{
		"list":     result,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

type createAdminUserRequest struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

// Create creates a new admin user.
// POST /admin/admin-users
func (h *AdminUserHandler) Create(c *gin.Context) {
	var req createAdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "用户名和密码不能为空")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "创建管理员失败")
		return
	}

	tx, err := h.db.Beginx()
	if err != nil {
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "创建管理员失败")
		return
	}
	defer tx.Rollback() // nolint:errcheck

	res, err := tx.Exec(`
INSERT INTO admin_users (username, password_hash, is_enabled)
VALUES (?, ?, 1)
`, req.Username, string(hashed))
	if err != nil {
		// 唯一键冲突：用户名已存在
		if me, ok := err.(*mysql.MySQLError); ok && me.Number == 1062 {
			resp.JSONError(c, http.StatusBadRequest, "ADMIN_USER_EXISTS", "用户名已存在")
			return
		}
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "创建管理员失败")
		return
	}

	newID, err := res.LastInsertId()
	if err != nil || newID == 0 {
		// 回退方案：根据用户名查 ID
		if err := tx.Get(&newID, `SELECT id FROM admin_users WHERE username = ? LIMIT 1`, req.Username); err != nil {
			resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "创建管理员失败")
			return
		}
	}

	// 绑定角色（如果有）
	if len(req.Roles) > 0 {
		if err := bindRoles(tx, newID, req.Roles); err != nil {
			resp.JSONError(c, http.StatusBadRequest, "ROLE_BIND_FAILED", err.Error())
			return
		}
	}

	if err := tx.Commit(); err != nil {
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "创建管理员失败")
		return
	}

	resp.JSONOK(c, gin.H{
		"id":       newID,
		"username": req.Username,
	})

	h.appendAudit(c, "admin.user.create", http.StatusOK, strconv.FormatInt(newID, 10), req.Username)
}

type enableRequest struct {
	Enabled bool `json:"enabled"`
}

// Enable sets is_enabled flag.
// POST /admin/admin-users/:id/enable
func (h *AdminUserHandler) Enable(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "无效的管理员ID")
		return
	}
	var req enableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "参数错误")
		return
	}

	res, err := h.db.Exec(`UPDATE admin_users SET is_enabled = ? WHERE id = ?`, req.Enabled, id)
	if err != nil {
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "更新状态失败")
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		resp.JSONError(c, http.StatusNotFound, "ADMIN_NOT_FOUND", "管理员不存在")
		return
	}

	resp.JSONOK(c, gin.H{"ok": true})
	h.appendAudit(c, "admin.user.enable", http.StatusOK, strconv.FormatInt(id, 10), "")
}

type resetPasswordRequest struct {
	Password string `json:"password"`
}

// ResetPassword resets password hash.
// POST /admin/admin-users/:id/reset-password
func (h *AdminUserHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "无效的管理员ID")
		return
	}
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Password == "" {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "密码不能为空")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "重置密码失败")
		return
	}

	res, err := h.db.Exec(`UPDATE admin_users SET password_hash = ?, updated_at = NOW() WHERE id = ?`, string(hashed), id)
	if err != nil {
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "重置密码失败")
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		resp.JSONError(c, http.StatusNotFound, "ADMIN_NOT_FOUND", "管理员不存在")
		return
	}

	resp.JSONOK(c, gin.H{"ok": true})
	h.appendAudit(c, "admin.user.reset_password", http.StatusOK, strconv.FormatInt(id, 10), "")
}

// GetRoles returns roles for a given admin user.
// GET /admin/admin-users/:id/roles
func (h *AdminUserHandler) GetRoles(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "无效的管理员ID")
		return
	}
	var roles []string
	if err := h.db.Select(&roles, `
SELECT r.name
FROM admin_roles r
JOIN admin_user_roles ur ON ur.role_id = r.id
WHERE ur.user_id = ?
`, id); err != nil {
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "查询角色失败")
		return
	}

	resp.JSONOK(c, gin.H{"roles": roles})
}

type setRolesRequest struct {
	Roles []string `json:"roles"`
}

// SetRoles sets roles for a given admin user.
// POST /admin/admin-users/:id/roles
func (h *AdminUserHandler) SetRoles(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "无效的管理员ID")
		return
	}
	var req setRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "参数错误")
		return
	}

	tx, err := h.db.Beginx()
	if err != nil {
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "保存角色失败")
		return
	}
	defer tx.Rollback() // nolint:errcheck

	if _, err := tx.Exec(`DELETE FROM admin_user_roles WHERE user_id = ?`, id); err != nil {
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "保存角色失败")
		return
	}

	if len(req.Roles) > 0 {
		if err := bindRoles(tx, id, req.Roles); err != nil {
			resp.JSONError(c, http.StatusBadRequest, "ROLE_BIND_FAILED", err.Error())
			return
		}
	}

	if err := tx.Commit(); err != nil {
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "保存角色失败")
		return
	}

	resp.JSONOK(c, gin.H{"ok": true})
	h.appendAudit(c, "admin.user.assign_roles", http.StatusOK, strconv.FormatInt(id, 10), "")
}

// bindRoles binds role names to a user within an existing transaction.
func bindRoles(tx *sqlx.Tx, userID int64, roles []string) error {
	if len(roles) == 0 {
		return nil
	}
	// 查出需要的角色 ID
	type roleRow struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
	}
	var dbRoles []roleRow
	query, args, err := sqlx.In(`SELECT id, name FROM admin_roles WHERE name IN (?)`, roles)
	if err != nil {
		return err
	}
	query = tx.Rebind(query)
	if err := tx.Select(&dbRoles, query, args...); err != nil {
		return err
	}
	if len(dbRoles) == 0 {
		return sql.ErrNoRows
	}
	for _, r := range dbRoles {
		if _, err := tx.Exec(`
INSERT INTO admin_user_roles (user_id, role_id)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE user_id = user_id
`, userID, r.ID); err != nil {
			return err
		}
	}
	return nil
}

func (h *AdminUserHandler) appendAudit(c *gin.Context, action string, status int, targetID string, operatorName string) {
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

func parseIntDefault(raw string, def int) int {
	if raw == "" {
		return def
	}
	if v, err := strconv.Atoi(raw); err == nil {
		return v
	}
	return def
}
