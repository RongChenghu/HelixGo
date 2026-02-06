package router

import (
	"helix-bot/pkg/ports"
	"helix-bot/pkg/types"
	"strings"
	"sync"
)

// Router implements ports.Router with minimal command/text/callback matching.
type Router struct {
	mu          sync.RWMutex
	middlewares []ports.Middleware
	commands    map[string]ports.Handler // e.g. "/ping" -> handler
	textMatchers []struct {
		m ports.TextMatcher
		h ports.Handler
	}
	callbackMatchers []struct {
		m ports.CallbackMatcher
		h ports.Handler
	}
	notFound ports.Handler
}

// New returns a new Router.
func New() *Router {
	return &Router{
		commands: make(map[string]ports.Handler),
	}
}

// Use adds middlewares.
func (r *Router) Use(mw ...ports.Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middlewares = append(r.middlewares, mw...)
}

// OnCommand registers handler for command (e.g. "/ping").
func (r *Router) OnCommand(cmd string, h ports.Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cmd = strings.TrimPrefix(cmd, "/")
	if cmd != "" {
		r.commands["/"+cmd] = h
	}
}

// OnText registers handler for text matcher.
func (r *Router) OnText(m ports.TextMatcher, h ports.Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.textMatchers = append(r.textMatchers, struct {
		m ports.TextMatcher
		h ports.Handler
	}{m, h})
}

// OnCallback registers handler for callback matcher.
func (r *Router) OnCallback(m ports.CallbackMatcher, h ports.Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.callbackMatchers = append(r.callbackMatchers, struct {
		m ports.CallbackMatcher
		h ports.Handler
	}{m, h})
}

// SetNotFound sets the not-found handler.
func (r *Router) SetNotFound(h ports.Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notFound = h
}

// Route returns the handler for the update (and a ctx builder is needed by runtime; handler only).
// Used internally by runtime to dispatch. Not part of ports.Router.
func (r *Router) Route(update types.BotUpdate) ports.Handler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Command: message with text starting with /command
	if update.Message != nil && update.Message.Text != "" {
		text := strings.TrimSpace(update.Message.Text)
		if idx := strings.Index(text, " "); idx > 0 {
			text = text[:idx]
		} else if idx := strings.Index(text, "\n"); idx > 0 {
			text = text[:idx]
		}
		if h, ok := r.commands[text]; ok {
			return r.chain(h)
		}
		for _, tm := range r.textMatchers {
			if tm.m.Match(update.Message.Text) {
				return r.chain(tm.h)
			}
		}
	}

	// Callback query
	if update.CallbackQuery != nil {
		for _, cm := range r.callbackMatchers {
			if cm.m.Match(update.CallbackQuery.Data) {
				return r.chain(cm.h)
			}
		}
	}

	if r.notFound != nil {
		return r.chain(r.notFound)
	}
	return nil
}

func (r *Router) chain(h ports.Handler) ports.Handler {
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}
	return h
}

// Ensure Router implements ports.Router.
var _ ports.Router = (*Router)(nil)
