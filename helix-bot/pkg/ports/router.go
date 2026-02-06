package ports

// TextMatcher matches message text (interfaces.md §5.2).
type TextMatcher interface {
	Match(text string) bool
}

// CallbackMatcher matches callback data (interfaces.md §5.2).
type CallbackMatcher interface {
	Match(data string) bool
}

// Router routes updates to handlers (interfaces.md §5.1).
type Router interface {
	Use(mw ...Middleware)
	// OnCommand registers a command handler. Optional middlewares are route-specific and
	// run after global middlewares but before the handler.
	OnCommand(cmd string, h Handler, mw ...Middleware)
	// OnText registers a text matcher handler. Optional middlewares are route-specific.
	OnText(m TextMatcher, h Handler, mw ...Middleware)
	// OnCallback registers a callback matcher handler. Optional middlewares are route-specific.
	OnCallback(m CallbackMatcher, h Handler, mw ...Middleware)
	SetNotFound(h Handler)
}
