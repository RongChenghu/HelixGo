#!/usr/bin/env bash
set -e

: "${TELEGRAM_BOT_TOKEN:?missing TELEGRAM_BOT_TOKEN}"
: "${TELEGRAM_WEBHOOK_SECRET:?missing TELEGRAM_WEBHOOK_SECRET}"

DOMAIN=${DOMAIN:-bot.example.com}
URL="https://${DOMAIN}/tg/${TELEGRAM_WEBHOOK_SECRET}/webhook"

curl -s "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/setWebhook" \
  -d "url=${URL}" \
  -d 'allowed_updates=["message","callback_query"]' \
  -d "drop_pending_updates=true"

echo "Webhook set to: ${URL}"

