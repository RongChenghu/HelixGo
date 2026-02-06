package router

import (
	"helix-bot/pkg/ports"
	"helix-bot/pkg/types"
	"strings"
	"sync"
)

// routeEntry holds a handler and its route-specific middlewares.
type routeEntry struct {
	h  ports.Handler
	mw []ports.Middleware
}

// Router implements ports.Router with minimal command/text/callback matching and
// a predictable middleware execution model:
// Incoming Update -> Global Middleware (Use) -> Route Middleware (per route) -> Handler.
type Router struct {
	mu           sync.RWMutex
	middlewares  []ports.Middleware
	commands     map[string]routeEntry // e.g. "/ping" -> handler + route mw
	textMatchers []struct {
		m  ports.TextMatcher
		re routeEntry
	}
	callbackMatchers []struct {
		m  ports.CallbackMatcher
		re routeEntry
	}
	notFound routeEntry
}

// New returns a new Router.
func New() *Router {
	return &Router{
		commands: make(map[string]routeEntry),
	}
}

// Use adds middlewares.
func (r *Router) Use(mw ...ports.Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middlewares = append(r.middlewares, mw...)
}

// OnCommand registers handler for command (e.g. "/ping").
// Route-specific middlewares run after global middlewares but before the handler.
func (r *Router) OnCommand(cmd string, h ports.Handler, mw ...ports.Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cmd = strings.TrimPrefix(cmd, "/")
	if cmd != "" {
		r.commands["/"+cmd] = routeEntry{
			h:  h,
			mw: cloneMiddlewareSlice(mw),
		}
	}
}

// OnText registers handler for text matcher.
func (r *Router) OnText(m ports.TextMatcher, h ports.Handler, mw ...ports.Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.textMatchers = append(r.textMatchers, struct {
		m  ports.TextMatcher
		re routeEntry
	}{
		m: m,
		re: routeEntry{
			h:  h,
			mw: cloneMiddlewareSlice(mw),
		},
	})
}

// OnCallback registers handler for callback matcher.
func (r *Router) OnCallback(m ports.CallbackMatcher, h ports.Handler, mw ...ports.Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.callbackMatchers = append(r.callbackMatchers, struct {
		m  ports.CallbackMatcher
		re routeEntry
	}{
		m: m,
		re: routeEntry{
			h:  h,
			mw: cloneMiddlewareSlice(mw),
		},
	})
}

// SetNotFound sets the not-found handler.
func (r *Router) SetNotFound(h ports.Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notFound = routeEntry{h: h}
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
		if re, ok := r.commands[text]; ok {
			return r.chain(re)
		}
		for _, tm := range r.textMatchers {
			if tm.m.Match(update.Message.Text) {
				return r.chain(tm.re)
			}
		}
	}

	// Callback query
	if update.CallbackQuery != nil {
		for _, cm := range r.callbackMatchers {
			if cm.m.Match(update.CallbackQuery.Data) {
				return r.chain(cm.re)
			}
		}
	}

	if r.notFound.h != nil {
		return r.chain(r.notFound)
	}
	return nil
}

// chain applies global middlewares (Use) then route-specific middlewares, preserving
// registration order. For mw1,mw2 registered in order, and route mwA,mwB:
//
//	incoming -> mw1 -> mw2 -> mwA -> mwB -> handler.
func (r *Router) chain(re routeEntry) ports.Handler {
	h := re.h

	// Route-specific middleware: registered order, so apply in reverse to wrap correctly.
	for i := len(re.mw) - 1; i >= 0; i-- {
		h = re.mw[i](h)
	}
	// Global middleware: registered order, apply in reverse as well.
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}
	return h
}

// cloneMiddlewareSlice makes a copy to avoid external mutation affecting router internals.
func cloneMiddlewareSlice(in []ports.Middleware) []ports.Middleware {
	if len(in) == 0 {
		return nil
	}
	out := make([]ports.Middleware, len(in))
	copy(out, in)
	return out
}

// Ensure Router implements ports.Router.
var _ ports.Router = (*Router)(nil)
