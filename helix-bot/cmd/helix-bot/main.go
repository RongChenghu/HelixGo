package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"helix-bot/pkg/helixbot"
	"helix-bot/pkg/ports"
)

const (
	version = "v0.1.0"
	commit  = "dev"
)

func main() {
	log.Printf("[helix-bot] %s (commit %s)", version, commit)

	if os.Getenv("GO_ENV") == "" || os.Getenv("GO_ENV") == "development" {
		for _, path := range []string{
			".env.development", "helix-bot/.env.development",
			".env.example", ".env", "helix-bot/.env.example", "helix-bot/.env",
		} {
			if err := godotenv.Load(path); err == nil && (path == ".env.development" || path == "helix-bot/.env.development") {
				log.Println("[info] loaded", path)
			}
		}
	}

	cfg := helixbot.LoadConfigFromEnv()
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

	bot := helixbot.MustNew(cfg)
	bot.OnCommand("/ping", func(ctx ports.Ctx) error {
		_, err := ctx.ReplyText("pong")
		return err
	})
	bot.SetNotFound(func(ctx ports.Ctx) error {
		if ctx.ChatID() != 0 {
			_, _ = ctx.ReplyText("Unknown command. Try /ping")
		}
		return nil
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
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
