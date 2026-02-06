package service

import (
	"testing"
	"time"

	"helix-api/internal/domain"
	"helix-api/internal/repo/memory"
)

func TestAuditServiceAppendAndList(t *testing.T) {
	repo := memory.NewAuditRepoMemory()
	svc := NewAuditService(repo)

	svc.Append(domain.AuditLog{
		Action:       "admin.login",
		Method:       "POST",
		Path:         "/admin/auth/login",
		Status:       200,
		OperatorID:   "1",
		OperatorName: "admin",
		CreatedAt:    time.Now().Add(-time.Minute),
	})
	svc.Append(domain.AuditLog{
		Action:       "system.config.update",
		Method:       "PUT",
		Path:         "/admin/system/configs/foo",
		Status:       200,
		OperatorID:   "1",
		OperatorName: "admin",
		CreatedAt:    time.Now(),
	})

	result := svc.ListPage(AuditListParams{Page: 1, PageSize: 10})
	if result.Total != 2 {
		t.Fatalf("expected total 2, got %d", result.Total)
	}
	if len(result.List) != 2 {
		t.Fatalf("expected list length 2, got %d", len(result.List))
	}
}
