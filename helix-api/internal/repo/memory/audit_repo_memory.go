package memory

import (
	"sort"
	"sync"
	"time"

	"helix-api/internal/domain"
)

type AuditRepoMemory struct {
	mu     sync.RWMutex
	nextID int64
	logs   []domain.AuditLog
}

func NewAuditRepoMemory() *AuditRepoMemory {
	return &AuditRepoMemory{
		nextID: 1,
		logs:   []domain.AuditLog{},
	}
}

func (r *AuditRepoMemory) Append(log domain.AuditLog) domain.AuditLog {
	r.mu.Lock()
	defer r.mu.Unlock()
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	log.ID = r.nextID
	r.nextID++
	r.logs = append(r.logs, log)
	return log
}

func (r *AuditRepoMemory) ListAll() []domain.AuditLog {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.AuditLog, len(r.logs))
	copy(list, r.logs)
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})
	return list
}
