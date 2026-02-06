# Auth Admin Contracts

## Token Header

- Header: `Authorization: Bearer <token>`
- Source: `server/src/middleware/adminAuth.js`, `admin/apps/web-antd/src/api/request.ts`

## GET /admin/auth/me

- Auth: required
- Permission: none (no explicit permission check)
- RateLimit: no
- Source: server/src/routes/adminAuth.js:64

### Request(JSON)
```json
{
  "headers": {
    "Authorization": "Bearer <token>"
  }
}
```

### Response(JSON)
```json
{
  "id": 123,
  "name": "admin",
  "roles": [
    "super_admin"
  ],
  "permissions": [
    "admin.manage",
    "report.view"
  ]
}
```

### Errors
- ADMIN_NOT_FOUND
- Internal Server Error

## POST /admin/auth/login

- Auth: optional
- Permission: none (no explicit permission check)
- RateLimit: yes
- Source: server/src/routes/adminAuth.js:13

### Request(JSON)
```json
{
  "body": {
    "username": "TODO",
    "password": "TODO"
  }
}
```

### Response(JSON)
```json
{
  "token": "<jwt>",
  "admin": {
    "id": 123,
    "username": "admin"
  }
}
```

### Errors
- Bad Request
- ADMIN_LOGIN_DISABLED
- INVALID_CREDENTIALS
- ADMIN_DISABLED
- Internal Server Error

## POST /admin/auth/change-password

- Auth: required
- Permission: none (no explicit permission check)
- RateLimit: no
- Source: server/src/routes/adminAuth.js:98

### Request(JSON)
```json
{
  "body": {
    "oldPassword": "TODO",
    "newPassword": "TODO"
  }
}
```

### Response(JSON)
```json
{
  "success": true
}
```

### Errors
- Bad Request
- ADMIN_NOT_FOUND
- INVALID_OLD_PASSWORD
- Internal Server Error

---

## 待补齐清单
- 无
