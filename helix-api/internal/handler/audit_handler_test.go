package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"helix-api/internal/domain"
	"helix-api/internal/repo/memory"
	"helix-api/internal/service"
)

func TestAuditHandlerListLogsJSONKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := memory.NewAuditRepoMemory()
	svc := service.NewAuditService(repo)
	handler := NewAuditHandler(svc)

	svc.Append(domain.AuditLog{
		Action:       "admin.login",
		Method:       "POST",
		Path:         "/admin/auth/login",
		Status:       200,
		OperatorID:   "1",
		OperatorName: "admin",
		IP:           "127.0.0.1",
		UserAgent:    "test",
		TraceID:      "trace-1",
		CreatedAt:    time.Date(2026, 2, 3, 12, 0, 0, 0, time.UTC),
	})

	engine := gin.New()
	engine.GET("/admin/audit/logs", handler.ListLogs)

	req := httptest.NewRequest(http.MethodGet, "/admin/audit/logs?page=1&pageSize=10", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v", err)
	}

	list, ok := payload["list"].([]interface{})
	if !ok || len(list) != 1 {
		t.Fatalf("expected list with 1 item")
	}

	item, ok := list[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected list item map")
	}

	expectedKeys := []string{
		"action", "method", "path", "status", "ip",
		"adminUserId", "adminUsername", "userAgent", "traceId", "createdAt",
	}
	for _, key := range expectedKeys {
		if _, exists := item[key]; !exists {
			t.Fatalf("missing key %s", key)
		}
	}
}
