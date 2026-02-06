package mysql

import (
	"encoding/json"
	"log"
	"sort"

	"github.com/jmoiron/sqlx"

	"helix-api/internal/domain"
	"helix-api/internal/repo"
)

type userRepoMySQL struct {
	db *sqlx.DB
}

func NewUserRepoMySQL(db *sqlx.DB) repo.UserRepo {
	return &userRepoMySQL{db: db}
}

func (r *userRepoMySQL) GetByUsername(username string) (*domain.User, bool) {
	var row struct {
		ID           int64  `db:"id"`
		Username     string `db:"username"`
		PasswordHash string `db:"password_hash"`
		IsEnabled    bool   `db:"is_enabled"`
	}
	if err := r.db.Get(&row, `
SELECT id, username, password_hash, is_enabled
FROM admin_users
WHERE username = ?
LIMIT 1
`, username); err != nil {
		return nil, false
	}

	// Load roles and their perms_json
	type roleRow struct {
		Name      string `db:"name"`
		PermsJSON string `db:"perms_json"`
	}

	roleRows := make([]roleRow, 0)
	if err := r.db.Select(&roleRows, `
SELECT r.name, COALESCE(r.perms_json, JSON_ARRAY()) AS perms_json
FROM admin_roles r
JOIN admin_user_roles ur ON ur.role_id = r.id
WHERE ur.user_id = ?
`, row.ID); err != nil {
		// If role query fails, still allow login but with empty roles/perms.
		log.Printf("[warn] load roles for user %s failed: %v", username, err)
		roleRows = []roleRow{}
	}

	roles := make([]string, 0, len(roleRows))
	permSet := map[string]struct{}{}

	for _, rr := range roleRows {
		roles = append(roles, rr.Name)

		var codes []string
		if err := json.Unmarshal([]byte(rr.PermsJSON), &codes); err != nil {
			log.Printf("[warn] invalid perms_json for role %s: %v", rr.Name, err)
			continue
		}
		for _, code := range codes {
			if code == "" {
				continue
			}
			permSet[code] = struct{}{}
		}
	}

	// Fallback: if no perms from JSON, derive minimal perms from roles.
	if len(permSet) == 0 {
		for _, p := range rolesToPerms(roles) {
			permSet[p] = struct{}{}
		}
	}

	perms := make([]string, 0, len(permSet))
	for p := range permSet {
		perms = append(perms, p)
	}
	sort.Strings(perms)

	return &domain.User{
		ID:           row.ID,
		Username:     row.Username,
		PasswordHash: row.PasswordHash,
		IsEnabled:    row.IsEnabled,
		Roles:        roles,
		Permissions:  perms,
	}, true
}

func (r *userRepoMySQL) UpdatePassword(username, password string) error {
	_, err := r.db.Exec(`
UPDATE admin_users
SET password_hash = ?, updated_at = NOW()
WHERE username = ?
`, password, username)
	return err
}

// rolesToPerms provides a minimal legacy mapping from role names to permissions.
// It is kept as a fallback when perms_json is empty or not yet configured.
func rolesToPerms(roles []string) []string {
	m := map[string]struct{}{}
	for _, role := range roles {
		if role == "admin" {
			m[domain.PermAdminManage] = struct{}{}
		}
	}
	out := make([]string, 0, len(m))
	for p := range m {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}
