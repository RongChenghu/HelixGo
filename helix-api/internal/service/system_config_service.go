package service

import (
	"helix-api/internal/domain"
	"helix-api/internal/repo"
)

type SystemConfigService struct {
	repo repo.SystemConfigRepo
}

func NewSystemConfigService(repo repo.SystemConfigRepo) *SystemConfigService {
	return &SystemConfigService{repo: repo}
}

func (s *SystemConfigService) List() []domain.SystemConfig {
	return s.repo.List()
}

func (s *SystemConfigService) Upsert(key, value string) domain.SystemConfig {
	return s.repo.Upsert(key, value)
}
