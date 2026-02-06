package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"helix-bot/internal/app"
	"helix-bot/internal/config"
	"helix-bot/internal/runtime"
	"helix-bot/pkg/ports"
)

const (
	version = "v0.1.0"
	commit  = "dev"
)

func main() {
	log.Printf("[helix-bot] %s (commit %s)", version, commit)

	// Load .env in development. Load .env.development FIRST so real token is set before
	// .env.example (godotenv Load() does not overwrite existing vars, so order matters).
	if os.Getenv("GO_ENV") == "" || os.Getenv("GO_ENV") == "development" {
		for _, path := range []string{
			".env.development", "helix-bot/.env.development", // token first
			".env.example", ".env", "helix-bot/.env.example", "helix-bot/.env",
		} {
			if err := godotenv.Load(path); err == nil && (path == ".env.development" || path == "helix-bot/.env.development") {
				log.Println("[info] loaded", path)
			}
		}
	}

	cfg := config.LoadFromEnv()
	if !cfg.Valid() {
		log.Println("[helix-bot] missing required config: TELEGRAM_BOT_TOKEN")
		os.Exit(1)
	}

	switch cfg.Mode {
	case "webhook":
		log.Println("[helix-bot] mode=webhook (HTTP webhook)")
	default:
		log.Println("[helix-bot] mode=polling (long polling)")
	}

	bot := app.New(cfg)
	// Minimal loop: /ping -> pong; any other text -> NotFound reply (for B3 acceptance).
	bot.Router().OnCommand("/ping", func(ctx ports.Ctx) error {
		_, err := ctx.ReplyText("pong")
		return err
	})
	bot.Router().SetNotFound(func(ctx ports.Ctx) error {
		if ctx.ChatID() != 0 {
			_, _ = ctx.ReplyText("Unknown command. Try /ping")
		}
		return nil
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// In polling mode, expose /healthz and /readyz via a lightweight HTTP server
	// on the same HTTP_LISTEN_ADDR as webhook mode. Webhook mode reuses the
	// webhook HTTP server (handlers registered in WebhookSource).
	if cfg.Mode != "webhook" {
		if rt, ok := bot.(*runtime.BotRuntime); ok {
			startHealthServer(ctx, cfg.HTTPListenAddr, rt)
		}
	}

	go func() {
		<-ctx.Done()
		log.Println("[helix-bot] shutting down")
	}()

	if cfg.Mode == "webhook" {
		log.Println("[helix-bot] webhook started, waiting for updates...")
	} else {
		log.Println("[helix-bot] polling started, waiting for updates...")
	}
	if err := bot.Run(ctx); err != nil && ctx.Err() == nil {
		log.Printf("[helix-bot] run error: %v", err)
		os.Exit(1)
	}
	log.Println("[helix-bot] stopped")
}

// startHealthServer starts an HTTP server that serves /healthz and /readyz.
// /healthz: process alive -> 200
// /readyz: runtime.State.Started == true -> 200, otherwise 503.
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
		state := rt.State()
		if !state.Started {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = w.Write([]byte("not ready"))
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

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
