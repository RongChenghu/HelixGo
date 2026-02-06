package memory

import (
	"errors"
	"sync"

	"helix-api/internal/domain"
)

type UserRepoMemory struct {
	mu    sync.RWMutex
	users map[string]*domain.User
}

func NewUserRepoMemory(adminUser, adminPass string) *UserRepoMemory {
	return &UserRepoMemory{
		users: map[string]*domain.User{
			adminUser: {
				ID:           1,
				Username:     adminUser,
				PasswordHash: adminPass,
				IsEnabled:    true,
				Roles:        []string{"admin"},
				Permissions:  []string{"admin.manage"},
			},
		},
	}
}

func (r *UserRepoMemory) GetByUsername(username string) (*domain.User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[username]
	if !ok {
		return nil, false
	}
	copied := *user
	return &copied, true
}

func (r *UserRepoMemory) UpdatePassword(username, password string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.users[username]
	if !ok {
		return errors.New("user not found")
	}
	user.PasswordHash = password
	return nil
}
