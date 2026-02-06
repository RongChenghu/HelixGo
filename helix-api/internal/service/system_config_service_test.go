package service

import (
	"testing"

	"helix-api/internal/repo/memory"
)

func TestSystemConfigServiceListAndUpsert(t *testing.T) {
	repo := memory.NewSystemConfigRepoMemory()
	svc := NewSystemConfigService(repo)

	if len(svc.List()) != 0 {
		t.Fatalf("expected empty list")
	}

	svc.Upsert("feature.enabled", "true")
	svc.Upsert("max.count", "10")

	list := svc.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 configs, got %d", len(list))
	}
}
