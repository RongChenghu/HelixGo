package service

import (
	"errors"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"helix-api/internal/domain"
	"helix-api/internal/pkg/token"
	"helix-api/internal/repo"
)

var (
	ErrInvalidCredentials = errors.New("INVALID_CREDENTIALS")
	ErrAdminNotFound      = errors.New("ADMIN_NOT_FOUND")
	ErrInvalidOldPassword = errors.New("INVALID_OLD_PASSWORD")
	ErrUserDisabled       = errors.New("USER_DISABLED")
)

type AuthService struct {
	repo   repo.UserRepo
	tokens *token.Manager
}

func NewAuthService(repo repo.UserRepo, tokens *token.Manager) *AuthService {
	return &AuthService{repo: repo, tokens: tokens}
}

func (s *AuthService) Login(username, password string) (string, *domain.User, error) {
	user, ok := s.repo.GetByUsername(username)
	if !ok {
		return "", nil, ErrInvalidCredentials
	}
	if !user.IsEnabled {
		return "", nil, ErrUserDisabled
	}

	// bcrypt first, fallback to plain text (memory repo/testing)
	if strings.HasPrefix(user.PasswordHash, "$2") {
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			return "", nil, ErrInvalidCredentials
		}
	} else if user.PasswordHash != password {
		return "", nil, ErrInvalidCredentials
	}
	jwtToken, err := s.tokens.Issue(strconv.FormatInt(user.ID, 10), user.Username, user.Roles, user.Permissions)
	if err != nil {
		return "", nil, err
	}
	return jwtToken, user, nil
}

func (s *AuthService) GetMe(username string) (*domain.User, error) {
	user, ok := s.repo.GetByUsername(username)
	if !ok {
		return nil, ErrAdminNotFound
	}
	return user, nil
}

func (s *AuthService) ChangePassword(username, oldPassword, newPassword string) error {
	user, ok := s.repo.GetByUsername(username)
	if !ok {
		return ErrAdminNotFound
	}

	// verify old password
	if strings.HasPrefix(user.PasswordHash, "$2") {
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
			return ErrInvalidOldPassword
		}
	} else if user.PasswordHash != oldPassword {
		return ErrInvalidOldPassword
	}

	// always store bcrypt when changing password
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(username, string(hashed))
}
