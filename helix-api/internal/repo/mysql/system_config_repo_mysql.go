package mysql

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"helix-api/internal/domain"
	"helix-api/internal/repo"
)

type systemConfigRepoMySQL struct {
	db *sqlx.DB
}

func NewSystemConfigRepoMySQL(db *sqlx.DB) repo.SystemConfigRepo {
	return &systemConfigRepoMySQL{db: db}
}

func (r *systemConfigRepoMySQL) List() []domain.SystemConfig {
	items := []domain.SystemConfig{}
	err := r.db.Select(&items, `
SELECT
  `+"`key`"+`,
  `+"`value`"+`,
  IFNULL(description, '') AS description,
  updated_at
FROM admin_system_configs
ORDER BY `+"`key`"+` ASC
`)
	if err != nil {
		// fail fast, caller will handle
		panic(fmt.Errorf("list system configs: %w", err))
	}
	return items
}

func (r *systemConfigRepoMySQL) Upsert(key, value, description string) domain.SystemConfig {
	_, err := r.db.Exec(`
INSERT INTO admin_system_configs (`+"`key`"+`, `+"`value`"+`, description)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE
  `+"`value`"+` = VALUES(`+"`value`"+`),
  description = VALUES(description),
  updated_at = CURRENT_TIMESTAMP
`, key, value, description)
	if err != nil {
		panic(fmt.Errorf("upsert system config: %w", err))
	}

	var item domain.SystemConfig
	if err := r.db.Get(&item, `
SELECT
  `+"`key`"+`,
  `+"`value`"+`,
  IFNULL(description, '') AS description,
  updated_at
FROM admin_system_configs
WHERE `+"`key`"+` = ?
LIMIT 1
`, key); err != nil {
		panic(fmt.Errorf("fetch system config after upsert: %w", err))
	}
	return item
}
