# Ping Bot Example (helix-bot v0.1 Acceptance)

本示例用于验证 **helix-bot v0.1 的最小闭环能力**：收到 `/ping` → 回复 `pong`。

> 本示例不包含任何业务逻辑，不接数据库，不依赖 helix-api。

------

## 快速验收（约 3 分钟）

### 1. 前置条件

- 已创建 Telegram Bot（@BotFather），并拿到 **Bot Token**
- 本地可访问 Telegram API（网络可用）

### 2. 设置环境变量

**方式 A：写进 `.env.development`（推荐）**

在 **helix-bot** 目录下创建或编辑 `helix-bot/.env.development`：

```bash
# 必填：从 @BotFather 获取
TELEGRAM_BOT_TOKEN=你的bot_token

# 可选：长轮询超时（秒），默认 60
# TELEGRAM_POLLING_TIMEOUT=30
```

**方式 B：临时导出**

```bash
export TELEGRAM_BOT_TOKEN=你的bot_token
```

### 3. 运行

**在仓库根目录（HelixGo）下：**

```bash
go run ./helix-bot/examples/ping-bot
```

**或在 helix-bot 目录下：**

```bash
cd helix-bot
make ping-bot
# 或
go run ./examples/ping-bot
```

看到类似输出即表示已启动：

```
[ping-bot] run; send /ping to get pong
[telegram] getUpdates duration=...
```

### 4. 验收步骤

1. 打开 Telegram，找到你的 Bot（或从 @BotFather 的链接进入）
2. 在聊天框输入：`/ping`
3. Bot 应回复：`pong`

✅ 收到 `pong` 即表示最小闭环验收通过。

------

## 可选：验收脚本

在 **helix-bot** 目录下执行：

```bash
./scripts/verify_ping.sh
```

脚本会检查是否已配置 `TELEGRAM_BOT_TOKEN`，然后启动 ping-bot，并提示你在 Telegram 中发送 `/ping` 验收（无交互，按 Ctrl+C 可停止 bot）。

------

## 目标行为

| 用户输入 | Bot 回复 |
|----------|----------|
| `/ping`  | `pong`   |

------

## 配置说明

| 变量 | 必填 | 说明 |
|------|------|------|
| `TELEGRAM_BOT_TOKEN` | 是 | 从 @BotFather 获取的 Bot Token |
| `TELEGRAM_POLLING_TIMEOUT` | 否 | 长轮询超时（秒），默认 60 |
| `TELEGRAM_POLL_OFFSET_FILE` | 否 | 持久化 offset 的文件路径，重启后从上次继续（每批写入 max(update_id)+1，原子写入） |

环境变量可从 `.env.development`、`.env` 或当前 shell 读取；从 **仓库根** 运行时也会自动尝试加载 `helix-bot/.env.development`。

------

## 重启不重复消费自测（可选）

用于验证 `TELEGRAM_POLL_OFFSET_FILE` 生效：重启后从持久化 offset 继续，不会重复处理旧消息。

1. **设置 offset 文件**（例如在 `helix-bot/.env.development` 中）：路径相对当前工作目录。从仓库根运行时用 `helix-bot/.poll_offset`，从 helix-bot 目录运行时用 `.poll_offset`。
   ```bash
   # 从仓库根运行
   TELEGRAM_POLL_OFFSET_FILE=helix-bot/.poll_offset
   # 从 helix-bot 目录运行
   # TELEGRAM_POLL_OFFSET_FILE=.poll_offset
   ```
2. **启动 bot**：`go run ./helix-bot/examples/ping-bot` 或 `make ping-bot`。
3. **在 Telegram 发一条** `/ping`，确认收到 `pong`。
4. **Ctrl+C 停止** bot。
5. **再次启动** 同一命令。
6. **再发一条** `/ping`，应再次收到 `pong`。

✅ **通过标准**：重启后不会把“停止前已处理过”的那条 `/ping` 再拉一遍（不会在同一会话里重复收到对同一条指令的回复）。可选：用 `cat helix-bot/.poll_offset` 查看持久化的 offset 数值，重启前后应递增。

------

## 验收清单（可选细查）

### 功能

- [ ] Bot 能启动并持续运行
- [ ] 发送 `/ping` 能收到 `pong`
- [ ] 非命令消息可走 NotFound 或忽略

### 可观测性

- [ ] 每次 update 有 requestId
- [ ] 日志中有 getUpdates / sendMessage 耗时

### 稳定性

- [ ] handler panic 不会导致进程退出（recover + stack 日志）
- [ ] 重复 update 不会重复执行（去重）

------

## 常见问题

**Q: 为什么只做 /ping？**  
因为 /ping 是最小可用闭环：**接收 → 路由 → 发送**，验证的是框架底座能力。

**Q: 从仓库根运行报 missing TELEGRAM_BOT_TOKEN？**  
确保 `helix-bot/.env.development` 存在且包含 `TELEGRAM_BOT_TOKEN=...`，且先加载 `.env.development` 再加载 `.env.example`（主入口已按此顺序加载）。

**Q: 示例不接 helix-api？**  
helix-bot 是通用 Adapter；业务可在 handler 内自行 HTTP 调用 helix-api。

------

## 下一步（可选扩展）

- `/whoami`：回复当前 userId/chatId（验证 ctx 字段）
- `callback confirm:`：按钮回调并 answerCallback（验证 callbackQuery）

> 注意：仍不引入任何业务规则。
