package app

import (
	"helix-bot/internal/adapter/telegram"
	"helix-bot/internal/config"
	"helix-bot/internal/router"
	"helix-bot/internal/runtime"
	"helix-bot/pkg/ports"
)

// New builds a Bot from config (config + logger + runtime; no business rules).
func New(cfg config.Config) ports.Bot {
	log := NewStdLogger("[helix-bot]")
	client := telegram.NewClient(cfg.TelegramBotToken)
	source := telegram.NewSource(cfg.TelegramBotToken, cfg.TelegramPollingTimeout, cfg.PollOffsetFile)
	r := router.New()
	rt := runtime.NewBotRuntime(r, source, log, client)
	return rt
}
