# Response Format (Legacy Admin)

## 现状
- 成功响应无统一包装，直接返回业务对象或数组
- 典型成功返回：
  - 登录：`{ token, admin: { id, username } }`
  - 当前管理员：`{ id, name, roles, permissions }`
  - 改密：`{ success: true }`
- 失败响应多数为 `res.status(x).json({ error, message })`
- 权限拒绝可能返回 `res.status(403).json({ error: "FORBIDDEN", permission: "<perm>" })`
- 前端请求层默认期望 `{ code, data }`（successCode = 0），与实际后端不一致

## 成功返回结构（真实）
```json
{
  "token": "<jwt>",
  "admin": {
    "id": 123,
    "username": "admin"
  }
}
```

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

```json
{
  "success": true
}
```

## 失败返回结构（真实）
```json
{
  "error": "INVALID_CREDENTIALS",
  "message": "用户名或密码错误"
}
```

```json
{
  "error": "FORBIDDEN",
  "permission": "admin.manage"
}
```

## TraceId 约定
- 现状：未发现统一注入的 traceId/traceID（无 middleware 统一设置）
- 建议：使用 `X-Request-Id` 请求头传入，响应体回传 `traceId` 字段

## TODO
- 若需统一返回结构，确定字段名与成功码（如 `code/message/data`）
