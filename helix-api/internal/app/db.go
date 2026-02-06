package app

import (
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"helix-api/internal/config"
)

// NewDB initializes a sqlx DB with sane pool defaults and validates connectivity.
func NewDB(cfg config.Config) (*sqlx.DB, error) {
	dsn := cfg.DSN()
	if dsn == "" {
		return nil, fmt.Errorf("mysql DSN is empty (set DB_NAME to enable mysql)")
	}

	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	// Pool settings: conservative defaults for admin backend
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return db, nil
}

// EnsureTables checks required tables and guides migration execution.
func EnsureTables(db *sqlx.DB) error {
	required := []string{
		"admin_system_configs",
		"admin_audit_logs",
		"admin_users",
		"admin_roles",
		"admin_user_roles",
	}
	for _, tbl := range required {
		var exists int
		query := `
			SELECT COUNT(*)
			FROM information_schema.tables
			WHERE table_schema = DATABASE()
			  AND table_name = ?
		`
		if err := db.Get(&exists, query, tbl); err != nil {
			return fmt.Errorf("check table %s: %w", tbl, err)
		}
		if exists == 0 {
			return fmt.Errorf("missing table %s. Please run migrations in ./migrations", tbl)
		}
	}
	log.Printf("[mysql] all required tables exist")
	return nil
}
