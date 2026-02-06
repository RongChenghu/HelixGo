# Audit Admin Contracts

## GET /admin/audit/logs

- Auth: required
- Permission: admin.manage
- RateLimit: no
- Source: server/src/routes/adminAudit.js:9

### Request(JSON)
```json
{
  "query": {
    "page": "TODO",
    "pageSize": "TODO",
    "keyword": "TODO",
    "action": "TODO",
    "from": "TODO",
    "to": "TODO"
  }
}
```

### Response(JSON)
```json
TODO
```

### Errors
- Internal Server Error

---

## 待补齐清单
- Response: GET /admin/audit/logs
