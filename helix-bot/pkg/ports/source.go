package ports

import (
	"context"

	"helix-bot/pkg/types"
)

// UpdateSource delivers updates (interfaces.md §4.1).
type UpdateSource interface {
	Start(ctx context.Context, out chan<- types.BotUpdate) error
	Stop() error
}
