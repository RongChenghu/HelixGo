package ports

import "helix-bot/pkg/types"

// IdempotencyPolicy (interfaces.md §8.1). Returns (key, ttlSeconds, ok).
type IdempotencyPolicy interface {
	KeyForUpdate(u types.BotUpdate) (string, int64, bool)
}

// RateLimitPolicy (interfaces.md §8.2). Returns (key, limit, windowSeconds, ok).
type RateLimitPolicy interface {
	KeyForSend(chatID int64, userID int64, method string) (string, int64, int64, bool)
}
