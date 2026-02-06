package memory

import (
	"sort"
	"sync"
	"time"

	"helix-api/internal/domain"
)

type SystemConfigRepoMemory struct {
	mu      sync.RWMutex
	records map[string]domain.SystemConfig
}

func NewSystemConfigRepoMemory() *SystemConfigRepoMemory {
	return &SystemConfigRepoMemory{
		records: make(map[string]domain.SystemConfig),
	}
}

func (r *SystemConfigRepoMemory) List() []domain.SystemConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.SystemConfig, 0, len(r.records))
	for _, item := range r.records {
		list = append(list, item)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Key < list[j].Key
	})
	return list
}

func (r *SystemConfigRepoMemory) Upsert(key, value string) domain.SystemConfig {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing := r.records[key]
	record := domain.SystemConfig{
		Key:         key,
		Value:       value,
		Description: existing.Description,
		UpdatedAt:   time.Now(),
	}
	r.records[key] = record
	return record
}
