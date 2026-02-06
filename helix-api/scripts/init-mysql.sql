-- Optional helper script to create database and user for helix-api v0.3
-- Edit credentials as needed before running.

CREATE DATABASE IF NOT EXISTS `helixgo_admin` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE USER IF NOT EXISTS 'helixgo_admin'@'%' IDENTIFIED BY 'rHzpwS8Ajjh5awPs';
GRANT ALL PRIVILEGES ON `helixgo_admin`.* TO 'helixgo_admin'@'%';
FLUSH PRIVILEGES;

-- Optional seed for default admin role permissions.
-- NOTE: run this AFTER applying migrations (including 0004 and 0006), otherwise
-- the admin_roles table/column may not exist yet.
USE `helixgo_admin`;

INSERT INTO admin_roles (name, description, perms_json)
VALUES (
  'admin',
  'System administrator',
  JSON_ARRAY(
    'admin.manage',
    'system.config.read',
    'system.config.write',
    'audit.read',
    'admin.user.read',
    'admin.user.write',
    'admin.role.read'
  )
)
ON DUPLICATE KEY UPDATE
  description = VALUES(description),
  perms_json = VALUES(perms_json);

