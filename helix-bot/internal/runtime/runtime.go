package runtime

import (
	"context"
	"strconv"
	"strings"
	"sync/atomic"

	"helix-bot/internal/router"
	"helix-bot/pkg/ports"
	"helix-bot/pkg/types"
)

// routeKey 从 update 推导可打印的路由标识，禁止输出用户正文（仅命令或 "text"/"callback"）。
func routeKey(u types.BotUpdate) string {
	if u.Message != nil && u.Message.Text != "" {
		text := strings.TrimSpace(u.Message.Text)
		if idx := strings.Index(text, " "); idx > 0 {
			text = text[:idx]
		} else if idx := strings.Index(text, "\n"); idx > 0 {
			text = text[:idx]
		}
		if strings.HasPrefix(text, "/") {
			return text
		}
		return "text"
	}
	if u.CallbackQuery != nil {
		return "callback"
	}
	return "unknown"
}

// RuntimeState is a read-only snapshot of runtime state for health checks.
type RuntimeState struct {
	Started bool
}

// BotRuntime implements ports.Bot.
type BotRuntime struct {
	router *router.Router
	source ports.UpdateSource
	logger ports.Logger
	client ports.TelegramClient

	started atomic.Bool
}

// NewBotRuntime builds a Bot runtime.
func NewBotRuntime(r *router.Router, source ports.UpdateSource, logger ports.Logger, client ports.TelegramClient) *BotRuntime {
	return &BotRuntime{router: r, source: source, logger: logger, client: client}
}

// State returns a read-only snapshot of the runtime state.
func (b *BotRuntime) State() RuntimeState {
	return RuntimeState{
		Started: b.started.Load(),
	}
}

// Router implements ports.Bot.
func (b *BotRuntime) Router() ports.Router {
	return b.router
}

// Use implements ports.Bot.
func (b *BotRuntime) Use(mw ...ports.Middleware) {
	b.router.Use(mw...)
}

// Run starts the update source and dispatches updates to handlers (interfaces.md §6.1).
func (b *BotRuntime) Run(ctx context.Context) error {
	updates := make(chan types.BotUpdate, 64)
	if err := b.source.Start(ctx, updates); err != nil {
		return err
	}
	defer func() { _ = b.source.Stop() }()

	// Mark runtime as started once the source has been successfully started.
	b.started.Store(true)

	for {
		select {
		case u, ok := <-updates:
			if !ok {
				return nil
			}
			requestID := "req-" + strconv.FormatInt(u.UpdateID, 10)
			handler := b.router.Route(u)
			if handler == nil {
				continue
			}
			c := NewCtx(u, requestID, b.logger, b.client)
			var handlerErr error
			panicked := false
			func() {
				defer func() {
					if recover() != nil {
						panicked = true
						// 不输出 panic 值/stack，避免泄露用户内容或 token。
						b.logger.Error("handler panic", "update_id", u.UpdateID, "reason", "panic_recovered")
					}
				}()
				handlerErr = handler(c)
			}()
			handled := !panicked && handlerErr == nil
			chatID := int64(0)
			if u.Message != nil {
				chatID = u.Message.ChatID
			} else if u.CallbackQuery != nil {
				chatID = u.CallbackQuery.ChatID
			}
			textLen := 0
			if u.Message != nil && u.Message.Text != "" {
				textLen = len(u.Message.Text)
			}
			b.logger.Info("update handled", "update_id", u.UpdateID, "chat_id", chatID, "route_key", routeKey(u), "text_len", textLen, "handled", handled)
			if !handled && !panicked {
				b.logger.Warn("handler error", "update_id", u.UpdateID, "reason", "handler_error")
			}
		case <-ctx.Done():
			return b.source.Stop()
		}
	}
}

// Shutdown stops the runtime (interfaces.md §6.1).
func (b *BotRuntime) Shutdown(ctx context.Context) error {
	return b.source.Stop()
}

// Ensure BotRuntime implements ports.Bot.
var _ ports.Bot = (*BotRuntime)(nil)
