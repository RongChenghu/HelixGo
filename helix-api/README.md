# helix-api v0.5

Minimal admin-auth + system-configs + audit + admin RBAC (Go + Gin) aligned with legacy admin contracts.
v0.3 switched system-configs & audit to MySQL; v0.4 added MySQL-based admin users/roles management with JWT-based RBAC; v0.5 enables JSON-based permissions on roles (`admin_roles.perms_json`) and exposes a permissions dictionary.

## 1) Configure MySQL

Edit `configs/.env.example` (or your `.env`) and set:
```
DB_HOST=mysql.cn-hangzhou.rds.aliyuncs.com
DB_PORT=3306
DB_USER=helixgo_admin
DB_PASS=Helixgo123!
DB_NAME=helixgo
DB_PARAMS=charset=utf8mb4&parseTime=true&loc=Local
```

If `DB_NAME` is set, the service *must* connect to MySQL (no silent memory fallback). Startup fails fast if tables are missing.

Optional helper to create DB/user:
```
mysql -h <host> -P <port> -u root -p < scripts/init-mysql.sql
```

## 2) Apply migrations

Run the SQL files in order:
```bash
mysql -h <host> -P <port> -u helixgo_admin -p helixgo_admin < migrations/0001_admin_system_configs.sql
mysql -h <host> -P <port> -u helixgo_admin -p helixgo_admin < migrations/0002_admin_audit_logs.sql
mysql -h <host> -P <port> -u helixgo_admin -p helixgo_admin < migrations/0003_admin_users.sql
mysql -h <host> -P <port> -u helixgo_admin -p helixgo_admin < migrations/0004_admin_roles.sql
mysql -h <host> -P <port> -u helixgo_admin -p helixgo_admin < migrations/0005_admin_user_roles.sql
```

On startup the service checks:

- `admin_system_configs`
- `admin_audit_logs`
- `admin_users`
- `admin_roles`
- `admin_user_roles`

If any is missing, it prints an error asking you to run the migrations.

### Default admin user & role

The seed in `0005_admin_user_roles.sql` creates:

- Username: `admin`
- Password: `admin123` (bcrypt hash stored in `password_hash`)
- Role: `admin`
- Permissions (derived in code): `["admin.manage"]`

## 3) Run

```bash
cd helix-api
make dev
```

## 4) Curl (happy path)

### Login (copy token from response)
```bash
curl -s -X POST http://localhost:8080/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```
### Configs list (auth required)
```bash
curl -s http://localhost:8080/admin/system/configs \
  -H "Authorization: Bearer <token>"
```
### Upsert config
```bash
curl -s -X PUT http://localhost:8080/admin/system/configs/feature.enabled \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"value":"true"}'
```
### Audit list
```bash
curl -s "http://localhost:8080/admin/audit/logs?page=1&pageSize=10" \
  -H "Authorization: Bearer <token>"
```

Example audit list response (legacy camelCase):
```json
{
  "list": [
    {
      "id": 1,
      "action": "admin.login",
      "method": "POST",
      "path": "/admin/auth/login",
      "status": 200,
      "ip": "127.0.0.1",
      "adminUserId": "1",
      "adminUsername": "admin",
      "userAgent": "curl/8.0.0",
      "traceId": "req-123",
      "createdAt": "2026-02-03T12:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "pageSize": 10
}
```

### Admin users & roles (v0.4)

List admin users:
```bash
curl -s http://localhost:8080/admin/admin-users \
  -H "Authorization: Bearer <token>"
```

Create admin user:
```bash
curl -s -X POST http://localhost:8080/admin/admin-users \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"username":"ops","password":"ops123456","roles":["admin"]}'
```

Enable/disable:
```bash
curl -s -X POST http://localhost:8080/admin/admin-users/2/enable \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"enabled":false}'
```

Reset password:
```bash
curl -s -X POST http://localhost:8080/admin/admin-users/2/reset-password \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"password":"newPass123"}'
```

List admin roles (now including perms, v0.5):
```bash
curl -s http://localhost:8080/admin/admin-roles \
  -H "Authorization: Bearer <token>"
```

Example response:
```json
[
  {
    "name": "admin",
    "description": "System administrator",
    "perms": [
      "admin.manage",
      "system.config.read",
      "system.config.write",
      "audit.read",
      "admin.user.read",
      "admin.user.write",
      "admin.role.read"
    ]
  }
]
```

### Permissions dictionary (v0.5)

List all known permission codes and descriptions:

```bash
curl -s http://localhost:8080/admin/permissions \
  -H "Authorization: Bearer <token>"
```

Example response:
```json
[
  { "code": "admin.manage", "description": "Super admin permission (all access)" },
  { "code": "admin.user.read", "description": "View admin users" },
  { "code": "admin.user.write", "description": "Manage admin users (create/enable/reset roles)" },
  { "code": "admin.role.read", "description": "View admin roles and their permissions" },
  { "code": "system.config.read", "description": "Read system configs" },
  { "code": "system.config.write", "description": "Update system configs" },
  { "code": "audit.read", "description": "Read audit logs" }
]
```
