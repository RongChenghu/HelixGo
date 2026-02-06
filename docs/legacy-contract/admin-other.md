# Other Admin Contracts

## POST /admin/wallet/manual-credit

- Auth: required
- Permission: TODO
- RateLimit: no
- Source: server/src/routes/adminWallet.js:11

### Request(JSON)
```json
{
  "body": {
    "userId": "TODO",
    "amount": "TODO",
    "reason": "TODO",
    "idempotencyKey": "TODO"
  }
}
```

### Response(JSON)
```json
TODO
```

### Errors
- Bad Request
- Not Found
- Internal Server Error

## POST /admin/wallet/manual-debit

- Auth: required
- Permission: TODO
- RateLimit: no
- Source: server/src/routes/adminWallet.js:108

### Request(JSON)
```json
{
  "body": {
    "userId": "TODO",
    "amount": "TODO",
    "reason": "TODO",
    "idempotencyKey": "TODO"
  }
}
```

### Response(JSON)
```json
TODO
```

### Errors
- Bad Request
- Not Found
- INSUFFICIENT_BALANCE
- Internal Server Error

## GET /admin/withdrawals

- Auth: required
- Permission: TODO
- RateLimit: no
- Source: server/src/routes/adminWithdrawals.js:142

### Request(JSON)
```json
{
  "query": {
    "page": "TODO",
    "pageSize": "TODO",
    "status": "TODO"
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

## POST /admin/withdrawals

- Auth: required
- Permission: TODO
- RateLimit: no
- Source: server/src/routes/adminWithdrawals.js:17

### Request(JSON)
```json
{
  "body": {
    "userId": "TODO",
    "amount": "TODO",
    "remark": "TODO",
    "idempotencyKey": "TODO"
  }
}
```

### Response(JSON)
```json
TODO
```

### Errors
- Bad Request
- Not Found
- Wallet Address Missing
- Insufficient Balance
- WITHDRAW_DISABLED
- Internal Server Error

## POST /admin/withdrawals/:id/approve

- Auth: required
- Permission: TODO
- RateLimit: no
- Source: server/src/routes/adminWithdrawals.js:183

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
- Not Found
- Internal Server Error

## POST /admin/withdrawals/:id/reject

- Auth: required
- Permission: TODO
- RateLimit: no
- Source: server/src/routes/adminWithdrawals.js:256

### Request(JSON)
```json
{
  "body": {
    "reason": "TODO"
  },
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
- Not Found
- Internal Server Error

## POST /admin/withdrawals/:id/pay

- Auth: required
- Permission: TODO
- RateLimit: no
- Source: server/src/routes/adminWithdrawals.js:342

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
- Wallet Address Missing
- Not Found
- Internal Server Error

## GET /admin/transactions

- Auth: required
- Permission: TODO
- RateLimit: no
- Source: server/src/routes/adminTransactions.js:10

### Request(JSON)
```json
{
  "query": {
    "page": "TODO",
    "pageSize": "TODO",
    "userId": "TODO",
    "type": "TODO",
    "interactionId": "TODO",
    "bizStatus": "TODO",
    "dateFrom": "TODO",
    "dateTo": "TODO",
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

## GET /admin/game/interactions

- Auth: required
- Permission: TODO
- RateLimit: no
- Source: server/src/routes/adminGame.js:14

### Request(JSON)
```json
{
  "query": {
    "page": "TODO",
    "pageSize": "TODO",
    "status": "TODO",
    "interactionType": "TODO",
    "interactionId": "TODO",
    "interactionCode": "TODO",
    "creatorId": "TODO",
    "chatId": "TODO",
    "dateFrom": "TODO",
    "dateTo": "TODO"
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

## GET /admin/game/participations

- Auth: required
- Permission: TODO
- RateLimit: no
- Source: server/src/routes/adminGame.js:336

### Request(JSON)
```json
{
  "query": {
    "page": "TODO",
    "pageSize": "TODO",
    "interactionId": "TODO",
    "userId": "TODO",
    "isTriggered": "TODO",
    "dateFrom": "TODO",
    "dateTo": "TODO"
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

## GET /admin/game/interactions/:id

- Auth: required
- Permission: TODO
- RateLimit: no
- Source: server/src/routes/adminGame.js:393

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
{
  "id": "TODO",
  "interactionCode": "TODO",
  "chatId": "TODO",
  "creatorId": "TODO",
  "baseValue": "TODO",
  "ruleFlag": "TODO",
  "maxParticipants": "TODO",
  "currentParticipants": "TODO",
  "status": "TODO",
  "resultData": "TODO",
  "participations": "TODO",
  "createdAt": "TODO",
  "updatedAt": "TODO",
  "expiresAt": "TODO"
}
```

### Errors
- Bad Request
- Not Found
- Internal Server Error

## POST /admin/game/welfare-interactions

- Auth: required
- Permission: TODO
- RateLimit: no
- Source: server/src/routes/adminGame.js:85

### Request(JSON)
```json
{
  "body": {
    "data": "TODO"
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

## POST /admin/game/interactions/:id/assign-hit-slots

- Auth: required
- Permission: TODO
- RateLimit: no
- Source: server/src/routes/adminGame.js:475

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
{
  "success": "TODO",
  "message": "TODO",
  "data": "TODO"
}
```

### Errors
- Bad Request
- Internal Server Error

## DELETE /admin/game/welfare-interactions/:id

- Auth: required
- Permission: game.interaction.delete
- RateLimit: no
- Source: server/src/routes/adminGame.js:272

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
{
  "ok": "TODO",
  "id": "TODO"
}
```

### Errors
- Bad Request
- Not Found
- Internal Server Error

## GET /admin/promotion/records

- Auth: required
- Permission: promotion.view
- RateLimit: no
- Source: server/src/routes/adminPromotion.js:10

### Request(JSON)
```json
{
  "query": {
    "page": "TODO",
    "pageSize": "TODO",
    "keyword": "TODO",
    "level": "TODO",
    "from": "TODO",
    "to": "TODO",
    "sourceTxnId": "TODO",
    "sourceUserId": "TODO"
  }
}
```

### Response(JSON)
```json
TODO
```

### Errors
- Internal Server Error

## GET /admin/promotion/summary

- Auth: required
- Permission: promotion.view
- RateLimit: no
- Source: server/src/routes/adminPromotion.js:46

### Request(JSON)
```json
{
  "query": {
    "page": "TODO",
    "pageSize": "TODO",
    "keyword": "TODO",
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

## GET /admin/promotion/trace

- Auth: required
- Permission: TODO
- RateLimit: no
- Source: server/src/routes/adminPromotion.js:76

### Request(JSON)
```json
{
  "query": {
    "sourceTxnId": "TODO"
  }
}
```

### Response(JSON)
```json
TODO
```

### Errors
- Internal Server Error

## GET /admin/referral/logs

- Auth: required
- Permission: TODO
- RateLimit: no
- Source: server/src/routes/adminReferral.js:9

### Request(JSON)
```json
{
  "query": {
    "page": "TODO",
    "pageSize": "TODO",
    "keyword": "TODO",
    "status": "TODO",
    "source": "TODO",
    "chatId": "TODO",
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
- Bad Request
- Internal Server Error

## GET /admin/reports/user-daily

- Auth: required
- Permission: report.view
- RateLimit: no
- Source: server/src/routes/adminReports.js:19

### Request(JSON)
```json
{
  "query": {
    "page": "TODO",
    "pageSize": "TODO",
    "keyword": "TODO",
    "from": "TODO",
    "to": "TODO",
    "realtime": "TODO"
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

## GET /admin/reports/platform-daily

- Auth: required
- Permission: report.view
- RateLimit: no
- Source: server/src/routes/adminReports.js:115

### Request(JSON)
```json
{
  "query": {
    "page": "TODO",
    "pageSize": "TODO",
    "from": "TODO",
    "to": "TODO",
    "realtime": "TODO"
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

---

## 待补齐清单
- Response: POST /admin/wallet/manual-credit
- Permission: POST /admin/wallet/manual-credit
- Response: POST /admin/wallet/manual-debit
- Permission: POST /admin/wallet/manual-debit
- Response: GET /admin/withdrawals
- Permission: GET /admin/withdrawals
- Response: POST /admin/withdrawals
- Permission: POST /admin/withdrawals
- Response: POST /admin/withdrawals/:id/approve
- Permission: POST /admin/withdrawals/:id/approve
- Response: POST /admin/withdrawals/:id/reject
- Permission: POST /admin/withdrawals/:id/reject
- Response: POST /admin/withdrawals/:id/pay
- Permission: POST /admin/withdrawals/:id/pay
- Response: GET /admin/transactions
- Permission: GET /admin/transactions
- Response: GET /admin/game/interactions
- Permission: GET /admin/game/interactions
- Response: GET /admin/game/participations
- Permission: GET /admin/game/participations
- Permission: GET /admin/game/interactions/:id
- Response: POST /admin/game/welfare-interactions
- Permission: POST /admin/game/welfare-interactions
- Permission: POST /admin/game/interactions/:id/assign-hit-slots
- Response: GET /admin/promotion/records
- Response: GET /admin/promotion/summary
- Response: GET /admin/promotion/trace
- Permission: GET /admin/promotion/trace
- Response: GET /admin/referral/logs
- Permission: GET /admin/referral/logs
- Response: GET /admin/reports/user-daily
- Response: GET /admin/reports/platform-daily
