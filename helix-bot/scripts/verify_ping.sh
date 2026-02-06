#!/usr/bin/env bash
# verify_ping.sh — 检查 env 并启动 ping-bot，提示在 Telegram 发 /ping 验收（无交互）.
set -e

cd "$(dirname "$0")/.."
ROOT="$(pwd)"

# 检查 token：环境变量或 .env.development（脚本在 helix-bot 下运行时为 helix-bot/.env.development）
check_token() {
  if [ -n "$TELEGRAM_BOT_TOKEN" ]; then
    return 0
  fi
  if [ -f ".env.development" ] && grep -q '^TELEGRAM_BOT_TOKEN=' .env.development 2>/dev/null; then
    return 0
  fi
  if [ -f "helix-bot/.env.development" ] && grep -q '^TELEGRAM_BOT_TOKEN=' helix-bot/.env.development 2>/dev/null; then
    return 0
  fi
  return 1
}

if ! check_token; then
  echo "缺少 TELEGRAM_BOT_TOKEN。"
  echo "请任选其一："
  echo "  1) 在 helix-bot/.env.development 中设置 TELEGRAM_BOT_TOKEN=你的token"
  echo "  2) 或执行: export TELEGRAM_BOT_TOKEN=你的token"
  echo "Token 可从 @BotFather 创建 Bot 后获取。"
  exit 1
fi

echo "已检测到 TELEGRAM_BOT_TOKEN。"
echo "即将启动 ping-bot；请在 Telegram 中向你的 Bot 发送 /ping，应收到 pong。"
echo "按 Ctrl+C 停止 bot。"
echo ""

# 若当前目录是 helix-bot（含 examples/ping-bot），直接运行；否则从仓库根运行
if [ -d "examples/ping-bot" ]; then
  exec go run ./examples/ping-bot
else
  exec go run ./helix-bot/examples/ping-bot
fi
