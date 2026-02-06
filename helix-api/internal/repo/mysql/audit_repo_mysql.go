package mysql

import (
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"

	"helix-api/internal/domain"
	"helix-api/internal/repo"
)

type auditRepoMySQL struct {
	db *sqlx.DB
}

func NewAuditRepoMySQL(db *sqlx.DB) repo.AuditRepo {
	return &auditRepoMySQL{db: db}
}

func (r *auditRepoMySQL) Append(log domain.AuditLog) domain.AuditLog {
	_, err := r.db.Exec(`
		INSERT INTO admin_audit_logs (
			action, method, path, status, ip,
			admin_user_id, admin_username, user_agent, trace_id, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())
	`,
		log.Action,
		log.Method,
		log.Path,
		log.Status,
		log.IP,
		log.OperatorID,
		log.OperatorName,
		log.UserAgent,
		log.TraceID,
	)
	if err != nil {
		panic(fmt.Errorf("insert audit log: %w", err))
	}
	// 上层目前不依赖返回的 ID 等字段，直接返回入参即可
	return log
}

func (r *auditRepoMySQL) ListAll() []domain.AuditLog {
	type row struct {
		ID           int64     `db:"id"`
		Action       string    `db:"action"`
		Method       string    `db:"method"`
		Path         string    `db:"path"`
		Status       int       `db:"status"`
		IP           string    `db:"ip"`
		AdminUserID  string    `db:"admin_user_id"`
		AdminName    string    `db:"admin_username"`
		UserAgent    string    `db:"user_agent"`
		TraceID      string    `db:"trace_id"`
		CreatedAtRaw time.Time `db:"created_at"`
	}

	rows := []row{}
	err := r.db.Select(&rows, `
SELECT
  id,
  action,
  method,
  path,
  status,
  ip,
  admin_user_id,
  admin_username,
  user_agent,
  trace_id,
  created_at
FROM admin_audit_logs
ORDER BY created_at DESC, id DESC
`)
	if err != nil {
		// 读审计失败时不要影响主请求，记录日志并返回空列表
		log.Printf("[warn] list audit logs failed: %v", err)
		return []domain.AuditLog{}
	}

	items := make([]domain.AuditLog, 0, len(rows))
	for _, rrow := range rows {
		items = append(items, domain.AuditLog{
			ID:           rrow.ID,
			Action:       rrow.Action,
			Method:       rrow.Method,
			Path:         rrow.Path,
			Status:       rrow.Status,
			OperatorID:   rrow.AdminUserID,
			OperatorName: rrow.AdminName,
			IP:           rrow.IP,
			UserAgent:    rrow.UserAgent,
			TraceID:      rrow.TraceID,
			CreatedAt:    rrow.CreatedAtRaw,
		})
	}
	return items
}
