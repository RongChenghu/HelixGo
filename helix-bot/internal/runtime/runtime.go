package runtime

import (
	"context"
	"runtime/debug"
	"strconv"

	"helix-bot/internal/router"
	"helix-bot/pkg/ports"
	"helix-bot/pkg/types"
)

// BotRuntime implements ports.Bot.
type BotRuntime struct {
	router *router.Router
	source ports.UpdateSource
	logger ports.Logger
	client ports.TelegramClient
}

// NewBotRuntime builds a Bot runtime.
func NewBotRuntime(r *router.Router, source ports.UpdateSource, logger ports.Logger, client ports.TelegramClient) *BotRuntime {
	return &BotRuntime{router: r, source: source, logger: logger, client: client}
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
			func() {
				defer func() {
					if e := recover(); e != nil {
						stack := debug.Stack()
						b.logger.Error("handler panic", "err", e, "updateID", u.UpdateID, "stack", string(stack))
					}
				}()
				_ = handler(c)
			}()
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
