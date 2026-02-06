-- MySQL 5.7 兼容：先增加可为 NULL 的 JSON 列，再补数据，最后改为 NOT NULL。

ALTER TABLE admin_roles
  ADD COLUMN perms_json JSON NULL;

-- Seed default admin role permissions (idempotent for fresh DB)
UPDATE admin_roles
SET perms_json = JSON_ARRAY(
  'admin.manage',
  'system.config.read',
  'system.config.write',
  'audit.read',
  'admin.user.read',
  'admin.user.write',
  'admin.role.read'
)
WHERE name = 'admin' AND perms_json IS NULL;

-- 现在把列收紧为 NOT NULL
ALTER TABLE admin_roles
  MODIFY COLUMN perms_json JSON NOT NULL;

