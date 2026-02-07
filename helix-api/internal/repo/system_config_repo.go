package repo

import "helix-api/internal/domain"

type SystemConfigRepo interface {
	List() []domain.SystemConfig
	Upsert(key, value, description string) domain.SystemConfig
}
