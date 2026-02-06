CREATE TABLE IF NOT EXISTS admin_user_roles (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_role (user_id, role_id),
    KEY idx_user_id (user_id),
    KEY idx_role_id (role_id)
) ENGINE=InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

-- seed default admin user and its role binding
-- password_hash is bcrypt("admin123") generated with bcrypt.DefaultCost
INSERT INTO admin_users (username, password_hash, is_enabled)
VALUES (
  'admin',
  '$2a$10$u1y4iG3I8X5o6JmF7T1sPuwf7b9l2w36q2Dkmv2fHKc1HQTVagj9e',
  1
)
ON DUPLICATE KEY UPDATE
  password_hash = VALUES(password_hash),
  is_enabled = VALUES(is_enabled);

INSERT INTO admin_user_roles (user_id, role_id)
SELECT u.id, r.id
FROM admin_users u, admin_roles r
WHERE u.username = 'admin' AND r.name = 'admin'
ON DUPLICATE KEY UPDATE user_id = user_id;

