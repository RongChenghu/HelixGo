package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"helix-api/internal/repo/memory"
	"helix-api/internal/service"
)

func TestSystemConfigHandlerListJSONKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := memory.NewSystemConfigRepoMemory()
	svc := service.NewSystemConfigService(repo)
	handler := NewSystemConfigHandler(svc, nil)

	svc.Upsert("feature.enabled", "true")

	engine := gin.New()
	engine.GET("/admin/system/configs", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/admin/system/configs", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var payload []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v", err)
	}

	if len(payload) != 1 {
		t.Fatalf("expected 1 item, got %d", len(payload))
	}

	expectedKeys := []string{"key", "value", "description", "updatedAt"}
	for _, key := range expectedKeys {
		if _, exists := payload[0][key]; !exists {
			t.Fatalf("missing key %s", key)
		}
	}
}
