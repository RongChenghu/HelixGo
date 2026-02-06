package ports

import "context"

// Bot is the runtime (interfaces.md §6.1).
type Bot interface {
	Router() Router
	Use(mw ...Middleware)
	Run(ctx context.Context) error
	Shutdown(ctx context.Context) error
}
