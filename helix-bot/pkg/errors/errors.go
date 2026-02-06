package errors

import "errors"

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

// TelegramError carries structured Telegram API error details.
// It is always wrapped by ErrTelegramAPI or ErrRateLimited.
type TelegramError struct {
	Code        int    // Telegram error_code or HTTP status code
	Description string // Telegram description or HTTP status text
	RetryAfter  int    // seconds, only meaningful for 429
}

func (e *TelegramError) Error() string {
	// Minimal string; callers should prefer accessing fields via errors.As.
	if e == nil {
		return ""
	}
	if e.RetryAfter > 0 {
		return "telegram api error"
	}
	return "telegram api error"
}

// TelegramAPIErr wraps ErrTelegramAPI with TelegramError (no retry_after).
func TelegramAPIErr(code int, description string) error {
	return errors.Join(ErrTelegramAPI, &TelegramError{
		Code:        code,
		Description: description,
	})
}

// RateLimitedErr wraps ErrRateLimited with TelegramError including retry_after.
func RateLimitedErr(code int, description string, retryAfterSec int) error {
	return errors.Join(ErrRateLimited, &TelegramError{
		Code:        code,
		Description: description,
		RetryAfter:  retryAfterSec,
	})
}
