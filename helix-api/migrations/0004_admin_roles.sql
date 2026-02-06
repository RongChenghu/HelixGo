CREATE TABLE IF NOT EXISTS admin_roles (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL UNIQUE,
    description VARCHAR(255) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

-- default admin role
INSERT INTO admin_roles (name, description)
VALUES ('admin', 'System administrator')
ON DUPLICATE KEY UPDATE description = VALUES(description);

