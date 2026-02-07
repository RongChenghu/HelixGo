package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAdminRoleHandler_Create_EmptyNameReturns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAdminRoleHandler(nil, nil)
	engine := gin.New()
	engine.POST("/admin/admin-roles", h.Create)

	body := []byte(`{"name":"","description":"","perms":[]}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/admin-roles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for empty name, got %d", rec.Code)
	}
}

func TestAdminRoleHandler_Create_InvalidPermsReturns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAdminRoleHandler(nil, nil)
	engine := gin.New()
	engine.POST("/admin/admin-roles", h.Create)

	body := []byte(`{"name":"testrole","description":"","perms":["invalid.perm"]}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/admin-roles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for invalid perms, got %d", rec.Code)
	}
}
