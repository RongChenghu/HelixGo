package telegram

import (
	"context"
	"errors"

	bterrors "helix-bot/pkg/errors"
)

// reasonFromErr 将错误映射为固定字符串，禁止把 err.Error() 写入日志（可能含 URL/token）。
func reasonFromErr(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.Canceled) {
		return "context_canceled"
	}
	if errors.Is(err, bterrors.ErrRateLimited) {
		return "rate_limited"
	}
	if errors.Is(err, bterrors.ErrTelegramAPI) {
		var te *bterrors.TelegramError
		if errors.As(err, &te) && te != nil && (te.Code == 401 || te.Code == 403) {
			return "unauthorized"
		}
		return "telegram_api"
	}
	return "network"
}
