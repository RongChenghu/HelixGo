# **Helix Bot v0.1 规范文档**
## **1. 定位与设计目标**
### **1.1 模块定位**
**helix-bot 是 HelixGo 框架中的 Telegram Bot 通用能力层（Adapter / Gateway）。**
它的职责是：
- 对接 Telegram Bot API
- 接收与解析 Update
- 提供统一的 Router / Middleware / Context
- 封装发送消息等基础能力
- 提供幂等、限流、日志等“必需但枯燥”的基础设施
**helix-bot 不负责任何业务规则。**
------
### **1.2 明确不做的事情（非常重要）**
helix-bot **永远不做**：
- ❌ 游戏规则 / 积分结算 / 逻辑
- ❌ 钱包 / 资金 / 交易 / Web3
- ❌ 业务状态机
- ❌ 用户体系 / 鉴权（除 Telegram 自身身份）
- ❌ 项目级数据库表结构
这些全部属于 **业务项目或 helix-event / service 层**。

------

## **2. 支持的接入模式**

### **2.1 Long Polling（调试 / 本地）**

- 用于本地开发、无公网环境
- 简单可靠
- 不要求 Webhook 证书与公网域名
### **2.2 Webhook（生产）**
- 支持 Telegram Webhook
- 支持 Secret Token 校验
- 支持重复投递（由幂等机制处理）
> v0.1 要求：**两种模式必须共存，但共享同一套 Update → Router → Handler 流程**
------
## **3. 核心抽象**
### **3.1 Update（统一输入）**
所有来自 Telegram 的事件，都会被解析为统一的 BotUpdate：
包含但不限于：
- message
- edited_message
- callback_query
- inline_query
- chat_join_request
业务侧不直接处理原始 JSON。
------
### **3.2 Context（Ctx）**
Ctx 是 handler 唯一感知的对象。
**Ctx 必须包含：**
- 原始 update
- chatId / userId / messageId（可为空）
- requestId / traceId
- logger（带 requestId）
- helper 方法：

  - Reply(text)
  - Send(text)
  - Edit(text)
  - AnswerCallback(text)

- context bag（供 middleware 读写）
业务 handler **禁止直接调用 Telegram HTTP API**。
------
## **4. Router 设计（轻量、通用）**
### **4.1 Router 能力**
helix-bot 提供类似 HTTP Router 的机制，但只针对 Bot 场景：
- Command Router
``` bash
  /start
  /help
  /ping
```
- Text Router

  - 正则
  - predicate（函数判断）

- Callback Router

  - 按 prefix / pattern 匹配 callback data

###  **4.2 示例（概念）**

``` bash
bot.OnCommand("/ping", handler)
bot.OnText(regex, handler)
bot.OnCallback("confirm:", handler)
bot.Use(middleware...)
```

Router **只负责匹配，不负责业务语义**。

## **5. Middleware（框架级能力）**
### **5.1 内置 Middleware（v0.1）**
#### **1) RequestId / Trace**
- 每个 update 生成唯一 requestId
- 贯穿日志与 handler
#### **2) Logging**
- 记录：

   - update_id
  - chatId / userId
  - update 类型
  - handler 耗时

- 不记录业务 payload（避免泄露）
#### **3) Idempotency（必须）**
- 防止 webhook 重复投递
- 默认 key：update_id
- 默认 TTL：5–10 分钟
- v0.1 提供 MemoryStore
#### **4) Rate Limit（基础）**
- 按 chatId / userId
- 防止 Telegram API 频率限制
- v0.1 只需简单 token bucket
#### **5) Recover / Error Boundary**
- handler panic 不影响 bot 主循环
- 记录错误日志
- 不向 Telegram 返回异常信息
------
## **6. Storage 抽象（只为基础设施）**
### **6.1 Store 接口（v0.1）**
用于：
- 幂等
- 限流
- 简单状态缓存
``` Go
Get(key)
Set(key, value, ttl)
Delete(key)
```

### **6.2 默认实现**
- MemoryStore（进程内）
- v0.1 不强制 Redis
- Redis / KV 存储作为后续扩展
------
## **7. 发送能力（Telegram API 封装）**
helix-bot 必须封装以下基础方法：
- sendMessage
- sendPhoto
- sendDocument
- editMessageText
- deleteMessage
- answerCallbackQuery
- sendChatAction
要求：
- 自动处理 JSON / multipart
- 自动重试（幂等方法）
- 统一错误返回
业务侧不直接感知 HTTP。
------
## **8. 配置规范**

### **8.1 必需配置**

- BOT_TOKEN
- MODE（polling / webhook）
- WEBHOOK_URL（webhook 模式）
- WEBHOOK_SECRET（可选）
### **8.2 不允许的配置**
- 业务规则开关
- 游戏/积分/结算配置
这些应进入 helix-api / 项目 service 层。
------
## **9. v0.1 验收标准（不涉及业务）**
helix-bot v0.1 必须满足：
1. 可启动 polling 模式
2. 能接收并解析 update
3. /ping → 回复 pong
4. callbackQuery 能正确 answer
5. webhook 重复 update 不会重复执行 handler
6. 发送消息有基础限流
7. panic 不会导致 bot 停止

------
## **10. 与 HelixGo 其他模块的关系**

| **模块**    | **关系**                 |
| ----------- | ------------------------ |
| helix-api   | ❌ 不直接依赖             |
| helix-event | ✅ 未来可作为事件源       |
| helix-core  | ✅ 可抽公共 logger / util |
| 业务项目    | ✅ 引用 helix-bot SDK     |

------
## **11. 设计原则（长期有效）**
- Adapter 优先
- 业务零感知
- 默认安全（幂等、限流、recover）
- 小而稳，拒绝“万能 Bot 框架”

------
## **12. 版本策略**
- v0.1：基础能力（本规范）
- v0.2：Redis Store / 多 Bot 支持
- v0.3：与 helix-event 集成（事件驱动）
------
### **结束语**
> **helix-bot 是“所有项目都会用到，但没人愿意反复写”的那一层。**
> 它的价值在于稳定、克制和边界清晰。
