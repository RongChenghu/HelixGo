package repo

import "helix-api/internal/domain"

type AuditRepo interface {
	Append(log domain.AuditLog) domain.AuditLog
	ListAll() []domain.AuditLog
}
