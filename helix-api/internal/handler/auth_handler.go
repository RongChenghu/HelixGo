package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"helix-api/internal/domain"
	"helix-api/internal/middleware"
	"helix-api/internal/pkg/resp"
	"helix-api/internal/pkg/token"
	"helix-api/internal/service"
)

type AuthHandler struct {
	service *service.AuthService
	audit   *service.AuditService
}

func NewAuthHandler(service *service.AuthService, audit *service.AuditService) *AuthHandler {
	return &AuthHandler{service: service, audit: audit}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type changePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "用户名和密码不能为空")
		return
	}

	jwtToken, user, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			resp.JSONError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "用户名或密码错误")
			return
		}
		if err == service.ErrUserDisabled {
			resp.JSONError(c, http.StatusForbidden, "USER_DISABLED", "账号已禁用")
			return
		}
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "登录失败")
		return
	}

	resp.JSONOK(c, gin.H{
		"token": jwtToken,
		"admin": gin.H{
			"id":       user.ID,
			"username": user.Username,
		},
	})

	h.appendAudit(c, "admin.login", http.StatusOK, strconv.FormatInt(user.ID, 10), user.Username)
}

func (h *AuthHandler) Me(c *gin.Context) {
	claimsAny, ok := c.Get(middleware.ClaimsKey)
	if !ok {
		resp.JSONError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token payload")
		return
	}
	claims := claimsAny.(*token.Claims)
	user, err := h.service.GetMe(claims.Name)
	if err != nil {
		if err == service.ErrAdminNotFound {
			resp.JSONError(c, http.StatusNotFound, "ADMIN_NOT_FOUND", "管理员不存在")
			return
		}
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "获取用户信息失败")
		return
	}

	resp.JSONOK(c, gin.H{
		"id":          user.ID,
		"name":        user.Username,
		"roles":       user.Roles,
		"permissions": user.Permissions,
	})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	claimsAny, ok := c.Get(middleware.ClaimsKey)
	if !ok {
		resp.JSONError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token payload")
		return
	}
	claims := claimsAny.(*token.Claims)

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "旧密码不能为空")
		return
	}
	if req.OldPassword == "" {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "旧密码不能为空")
		return
	}
	if req.NewPassword == "" {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "新密码不能为空")
		return
	}
	if len(req.NewPassword) < 6 {
		resp.JSONError(c, http.StatusBadRequest, "Bad Request", "新密码长度至少为6位")
		return
	}

	err := h.service.ChangePassword(claims.Name, req.OldPassword, req.NewPassword)
	if err != nil {
		if err == service.ErrAdminNotFound {
			resp.JSONError(c, http.StatusNotFound, "ADMIN_NOT_FOUND", "管理员不存在")
			return
		}
		if err == service.ErrInvalidOldPassword {
			resp.JSONError(c, http.StatusBadRequest, "INVALID_OLD_PASSWORD", "旧密码错误")
			return
		}
		resp.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "修改密码失败")
		return
	}

	resp.JSONOK(c, gin.H{"success": true})

	operatorID := ""
	if claims.Subject != "" {
		operatorID = claims.Subject
	}
	h.appendAudit(c, "admin.change_password", http.StatusOK, operatorID, claims.Name)
}

func (h *AuthHandler) appendAudit(c *gin.Context, action string, status int, operatorID string, operatorName string) {
	if h.audit == nil {
		return
	}
	traceID := ""
	if value, ok := c.Get(middleware.RequestIDKey); ok {
		if str, ok := value.(string); ok {
			traceID = str
		}
	}
	h.audit.Append(domain.AuditLog{
		Action:       action,
		Method:       c.Request.Method,
		Path:         c.Request.URL.Path,
		Status:       status,
		OperatorID:   operatorID,
		OperatorName: operatorName,
		IP:           c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
		TraceID:      traceID,
	})
}
