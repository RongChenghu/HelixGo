// Package main is the ping-bot example: /ping -> pong (minimal acceptance).
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

func main() {
	if os.Getenv("GO_ENV") == "" || os.Getenv("GO_ENV") == "development" {
		for _, path := range []string{
			".env.development", "helix-bot/.env.development",
			".env.example", ".env", "helix-bot/.env.example", "helix-bot/.env",
		} {
			_ = godotenv.Load(path)
		}
	}
	cfg := helixbot.LoadConfigFromEnv()
	if !cfg.Valid() {
		log.Fatal("missing TELEGRAM_BOT_TOKEN")
	}
	bot := helixbot.MustNew(cfg)
	bot.OnCommand("/ping", func(ctx ports.Ctx) error {
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
