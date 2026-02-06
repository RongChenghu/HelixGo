package ports

import "context"

// Store for idempotency / rate limit (interfaces.md §7.1).
type Store interface {
	Get(ctx context.Context, key string) (val string, ok bool, err error)
	Set(ctx context.Context, key string, val string, ttlSeconds int64) error
	Delete(ctx context.Context, key string) error
	Incr(ctx context.Context, key string, ttlSeconds int64) (newVal int64, err error)
}
