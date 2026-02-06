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
	OnCommand(cmd string, h Handler)
	OnText(m TextMatcher, h Handler)
	OnCallback(m CallbackMatcher, h Handler)
	SetNotFound(h Handler)
}
