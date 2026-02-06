package service

import (
	"testing"
	"time"

	"helix-api/internal/domain"
	"helix-api/internal/pkg/token"
	"helix-api/internal/repo/memory"
)

type stubUserRepo struct {
	user *domain.User
	ok   bool
}

func (s *stubUserRepo) GetByUsername(username string) (*domain.User, bool) {
	return s.user, s.ok
}

func (s *stubUserRepo) UpdatePassword(username, password string) error {
	return nil
}

func TestLogin_IncludesPermissionsFromRepo(t *testing.T) {
	mgr := token.NewManager("secret", time.Minute)
	expectedPerms := []string{"system.config.read", "audit.read"}
	svc := NewAuthService(&stubUserRepo{
		user: &domain.User{
			ID:           1,
			Username:     "admin",
			PasswordHash: "admin123", // plaintext for memory-style comparison
			IsEnabled:    true,
			Roles:        []string{"admin"},
			Permissions:  expectedPerms,
		},
		ok: true,
	}, mgr)

	out, _, err := svc.Login("admin", "admin123")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}

	claims, err := mgr.Parse(out)
	if err != nil {
		t.Fatalf("Parse token error: %v", err)
	}

	if len(claims.Perms) != len(expectedPerms) {
		t.Fatalf("expected %d perms, got %d", len(expectedPerms), len(claims.Perms))
	}
	for i, p := range expectedPerms {
		if claims.Perms[i] != p {
			t.Fatalf("expected perm[%d]=%s, got %s", i, p, claims.Perms[i])
		}
	}
}

func TestAuthServiceLoginSuccess(t *testing.T) {
	repo := memory.NewUserRepoMemory("admin", "admin123")
	tokens := token.NewManager("test-secret", time.Hour)
	svc := NewAuthService(repo, tokens)

	jwtToken, user, err := svc.Login("admin", "admin123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if jwtToken == "" {
		t.Fatalf("expected token, got empty")
	}
	if user.Username != "admin" {
		t.Fatalf("expected username admin, got %s", user.Username)
	}
}

func TestAuthServiceLoginFailure(t *testing.T) {
	repo := memory.NewUserRepoMemory("admin", "admin123")
	tokens := token.NewManager("test-secret", time.Hour)
	svc := NewAuthService(repo, tokens)

	_, _, err := svc.Login("admin", "bad")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthServiceLoginClaimsContainRolesAndPerms(t *testing.T) {
	repo := memory.NewUserRepoMemory("admin", "admin123")
	tokens := token.NewManager("test-secret", time.Hour)
	svc := NewAuthService(repo, tokens)

	jwtToken, _, err := svc.Login("admin", "admin123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	claims, err := tokens.Parse(jwtToken)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if len(claims.Roles) == 0 || claims.Roles[0] != "admin" {
		t.Fatalf("expected role admin, got %+v", claims.Roles)
	}
	found := false
	for _, p := range claims.Perms {
		if p == "admin.manage" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected perm admin.manage in %+v", claims.Perms)
	}
}
