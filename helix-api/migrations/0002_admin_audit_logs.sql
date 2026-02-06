CREATE TABLE IF NOT EXISTS admin_audit_logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    action VARCHAR(64) NOT NULL,
    method VARCHAR(8) NOT NULL,
    path VARCHAR(255) NOT NULL,
    status INT NOT NULL,
    ip VARCHAR(64) NOT NULL,
    admin_user_id VARCHAR(64) NOT NULL,
    admin_username VARCHAR(64) NOT NULL,
    user_agent VARCHAR(255) NOT NULL,
    trace_id VARCHAR(64) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY idx_created_at (created_at),
    KEY idx_action (action),
    KEY idx_admin_user (admin_user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
