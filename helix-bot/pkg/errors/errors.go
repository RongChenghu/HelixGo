package errors

import (
	"errors"
	"fmt"
)

// Contract-level errors (stable; implementations use err wrapping).
var (
	// ErrNoChatContext: current update has no chat/message, cannot reply.
	ErrNoChatContext = errors.New("helix-bot: no chat context for reply")
	// ErrRateLimited: send rate limit hit (may include retryAfter).
	ErrRateLimited = errors.New("helix-bot: rate limited")
	// ErrDuplicateUpdate: idempotency dedup hit (should be silent, not surfaced to handler).
	ErrDuplicateUpdate = errors.New("helix-bot: duplicate update")
	// ErrTelegramAPI: Telegram returned non-2xx (method/status/description in wrapped error).
	ErrTelegramAPI = errors.New("helix-bot: telegram api error")
)

// TelegramAPIErr wraps ErrTelegramAPI with error_code and description (no token/body).
func TelegramAPIErr(code int, description string) error {
	return fmt.Errorf("%w: code=%d desc=%s", ErrTelegramAPI, code, description)
}

// RateLimitedErr wraps ErrRateLimited with optional retry_after seconds.
func RateLimitedErr(retryAfterSec int) error {
	if retryAfterSec <= 0 {
		return ErrRateLimited
	}
	return fmt.Errorf("%w: retry_after=%d", ErrRateLimited, retryAfterSec)
}
