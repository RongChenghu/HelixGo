package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
	Perms []string `json:"perms"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret  []byte
	expires time.Duration
}

func NewManager(secret string, expires time.Duration) *Manager {
	return &Manager{
		secret:  []byte(secret),
		expires: expires,
	}
}

func (m *Manager) Issue(subject, name string, roles, perms []string) (string, error) {
	now := time.Now()
	claims := Claims{
		Name:  name,
		Roles: roles,
		Perms: perms,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expires)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) Parse(raw string) (*Claims, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	parsed, err := parser.ParseWithClaims(raw, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
