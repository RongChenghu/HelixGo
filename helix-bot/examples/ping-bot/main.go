// Package main is the ping-bot example: /ping -> pong (minimal acceptance).
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"helix-bot/internal/app"
	"helix-bot/internal/config"
	"helix-bot/pkg/ports"
)

func main() {
	// Same env load order as cmd/helix-bot: .env.development first, then helix-bot/ paths for repo-root run.
	if os.Getenv("GO_ENV") == "" || os.Getenv("GO_ENV") == "development" {
		for _, path := range []string{
			".env.development", "helix-bot/.env.development",
			".env.example", ".env", "helix-bot/.env.example", "helix-bot/.env",
		} {
			_ = godotenv.Load(path)
		}
	}
	cfg := config.LoadFromEnv()
	if !cfg.Valid() {
		log.Fatal("missing TELEGRAM_BOT_TOKEN")
	}
	bot := app.New(cfg)
	bot.Router().OnCommand("/ping", func(ctx ports.Ctx) error {
		_, err := ctx.ReplyText("pong")
		return err
	})
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		log.Println("[ping-bot] shutting down")
	}()
	log.Println("[ping-bot] run; send /ping to get pong")
	_ = bot.Run(ctx)
	log.Println("[ping-bot] stopped")
}
