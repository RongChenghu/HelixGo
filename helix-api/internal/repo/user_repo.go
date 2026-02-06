package repo

import "helix-api/internal/domain"

type UserRepo interface {
	GetByUsername(username string) (*domain.User, bool)
	UpdatePassword(username, password string) error
}
