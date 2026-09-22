package helixbot

import (
	"context"
	"errors"
	"log"
	"net/http"

	"helix-bot/internal/app"
	"helix-bot/internal/config"
	"helix-bot/internal/runtime"
	"helix-bot/pkg/ports"
)

// Bot is the public bot instance: config + runtime. Register routes then Run(ctx).
type Bot struct {
	inner ports.Bot
	cfg   Config
}

// ErrMissingToken is returned by New when TELEGRAM_BOT_TOKEN is not set.
var ErrMissingToken = errors.New("helix-bot: missing required config: TELEGRAM_BOT_TOKEN")

// New builds a Bot from config. Returns error if config is invalid.
func New(cfg Config) (*Bot, error) {
	if !cfg.Valid() {
		return nil, ErrMissingToken
	}
	internalCfg := config.Config{
		TelegramBotToken:       cfg.TelegramBotToken,
		TelegramPollingTimeout:  cfg.TelegramPollingTimeout,
		PollOffsetFile:         cfg.PollOffsetFile,
		Mode:                   cfg.Mode,
		WebhookSecret:          cfg.WebhookSecret,
		HTTPListenAddr:         cfg.HTTPListenAddr,
	}
	inner := app.New(internalCfg)
	return &Bot{inner: inner, cfg: cfg}, nil
}

// MustNew is like New but panics on error (e.g. missing token).
func MustNew(cfg Config) *Bot {
	bot, err := New(cfg)
	if err != nil {
		panic(err)
	}
	return bot
}

// OnCommand registers a command handler (e.g. "/ping"). Optional mw are route-specific middlewares.
func (b *Bot) OnCommand(cmd string, h ports.Handler, mw ...ports.Middleware) {
	b.inner.Router().OnCommand(cmd, h, mw...)
}

// OnText registers a text matcher handler.
func (b *Bot) OnText(m ports.TextMatcher, h ports.Handler, mw ...ports.Middleware) {
	b.inner.Router().OnText(m, h, mw...)
}

// OnCallback registers a callback matcher handler.
func (b *Bot) OnCallback(m ports.CallbackMatcher, h ports.Handler, mw ...ports.Middleware) {
	b.inner.Router().OnCallback(m, h, mw...)
}

// Use adds global middlewares.
func (b *Bot) Use(mw ...ports.Middleware) {
	b.inner.Use(mw...)
}

// SetNotFound sets the handler for unmatched updates.
func (b *Bot) SetNotFound(h ports.Handler) {
	b.inner.Router().SetNotFound(h)
}

// Run starts the update source and dispatches updates until ctx is done.
// In polling mode, also starts a health server on cfg.HTTPListenAddr for /healthz and /readyz.
func (b *Bot) Run(ctx context.Context) error {
	if b.cfg.Mode != "webhook" {
		if rt, ok := b.inner.(*runtime.BotRuntime); ok {
			startHealthServer(ctx, b.cfg.HTTPListenAddr, rt)
		}
	}
	return b.inner.Run(ctx)
}

// startHealthServer runs /healthz and /readyz on addr until ctx is done.
func startHealthServer(ctx context.Context, addr string, rt *runtime.BotRuntime) {
	if addr == "" {
		addr = ":4000"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if !rt.State().Started {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = w.Write([]byte("not ready"))
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})
	srv := &http.Server{Addr: addr, Handler: mux}
	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[helix-bot] health server error: %v", err)
		}
	}()
}
