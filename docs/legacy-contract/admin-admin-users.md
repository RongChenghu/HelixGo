# Users Admin Contracts

## GET /admin/users

- Auth: required
- Permission: TODO
- RateLimit: no
- Source: server/src/routes/adminUsers.js:10

### Request(JSON)
```json
{
  "query": {
    "page": "TODO",
    "pageSize": "TODO",
    "keyword": "TODO"
  }
}
```

### Response(JSON)
```json
TODO
```

### Errors
- Bad Request
- Internal Server Error

## POST /admin/users

- Auth: required
- Permission: TODO
- RateLimit: no
- Source: server/src/routes/adminUsers.js:120

### Request(JSON)
```json
{
  "body": {
    "telegramId": "TODO",
    "username": "TODO",
    "nickname": "TODO",
    "balance": "TODO"
  }
}
```

### Response(JSON)
```json
TODO
```

### Errors
- Bad Request
- Conflict
- Internal Server Error

## PUT /admin/users/:userId/wallet-address

- Auth: required
- Permission: TODO
- RateLimit: no
- Source: server/src/routes/adminUsers.js:51

### Request(JSON)
```json
{
  "body": {
    "walletAddress": "TODO"
  },
  "params": {
    "userId": "TODO"
  }
}
```

### Response(JSON)
```json
{
  "ok": "TODO",
  "walletAddress": "TODO"
}
```

### Errors
- Bad Request
- Not Found
- Internal Server Error

## GET /admin/admin-users

- Auth: required
- Permission: admin.manage
- RateLimit: no
- Source: server/src/routes/adminAdminUsers.js:10

### Request(JSON)
```json
{
  "query": {
    "page": "TODO",
    "pageSize": "TODO",
    "keyword": "TODO"
  }
}
```

### Response(JSON)
```json
TODO
```

### Errors
- Internal Server Error

## GET /admin/admin-users/:id/roles

- Auth: required
- Permission: admin.manage
- RateLimit: no
- Source: server/src/routes/adminAdminUsers.js:185

### Request(JSON)
```json
{
  "params": {
    "id": "TODO"
  }
}
```

### Response(JSON)
```json
TODO
```

### Errors
- Bad Request
- ADMIN_NOT_FOUND
- Internal Server Error

## POST /admin/admin-users

- Auth: required
- Permission: admin.manage
- RateLimit: no
- Source: server/src/routes/adminAdminUsers.js:36

### Request(JSON)
```json
{
  "body": {
    "username": "TODO",
    "password": "TODO",
    "roles": "TODO"
  }
}
```

### Response(JSON)
```json
TODO
```

### Errors
- Bad Request
- USERNAME_EXISTS
- Internal Server Error

## POST /admin/admin-users/:id/enable

- Auth: required
- Permission: admin.manage
- RateLimit: no
- Source: server/src/routes/adminAdminUsers.js:83

### Request(JSON)
```json
{
  "body": {
    "isEnabled": "TODO"
  },
  "params": {
    "id": "TODO"
  }
}
```

### Response(JSON)
```json
{
  "success": "TODO"
}
```

### Errors
- Bad Request
- ADMIN_NOT_FOUND
- Internal Server Error

## POST /admin/admin-users/:id/reset-password

- Auth: required
- Permission: admin.manage
- RateLimit: no
- Source: server/src/routes/adminAdminUsers.js:134

### Request(JSON)
```json
{
  "body": {
    "newPassword": "TODO"
  },
  "params": {
    "id": "TODO"
  }
}
```

### Response(JSON)
```json
{
  "success": "TODO"
}
```

### Errors
- Bad Request
- ADMIN_NOT_FOUND
- Internal Server Error

## POST /admin/admin-users/:id/roles

- Auth: required
- Permission: admin.manage
- RateLimit: no
- Source: server/src/routes/adminAdminUsers.js:218

### Request(JSON)
```json
{
  "body": {
    "roles": "TODO"
  },
  "params": {
    "id": "TODO"
  }
}
```

### Response(JSON)
```json
{
  "success": "TODO"
}
```

### Errors
- Bad Request
- ADMIN_NOT_FOUND
- Internal Server Error

---

## 待补齐清单
- Response: GET /admin/users
- Permission: GET /admin/users
- Response: POST /admin/users
- Permission: POST /admin/users
- Permission: PUT /admin/users/:userId/wallet-address
- Response: GET /admin/admin-users
- Response: GET /admin/admin-users/:id/roles
- Response: POST /admin/admin-users
