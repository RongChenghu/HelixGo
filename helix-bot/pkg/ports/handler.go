package ports

// Handler processes one update (interfaces.md §1.1).
type Handler func(ctx Ctx) error

// Middleware wraps a handler and returns a new handler (interfaces.md §1.1).
type Middleware func(next Handler) Handler
