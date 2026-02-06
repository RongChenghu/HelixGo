# SystemConfigs Admin Contracts

## GET /admin/system/configs

- Auth: required
- Permission: TODO
- RateLimit: no
- Source: server/src/routes/adminSystemConfigs.js:11

### Request(JSON)
```json
TODO
```

### Response(JSON)
```json
TODO
```

### Errors
- Internal Server Error

## PUT /admin/system/configs/:key

- Auth: required
- Permission: TODO
- RateLimit: no
- Source: server/src/routes/adminSystemConfigs.js:28

### Request(JSON)
```json
{
  "body": {
    "value": "TODO"
  },
  "params": {
    "key": "TODO"
  }
}
```

### Response(JSON)
```json
{
  "ok": "TODO"
}
```

### Errors
- Bad Request
- Internal Server Error

---

## 待补齐清单
- Request: GET /admin/system/configs
- Response: GET /admin/system/configs
- Permission: GET /admin/system/configs
- Permission: PUT /admin/system/configs/:key
