package helixbot

import (
	"time"

	"helix-bot/internal/config"
)

// Config is the stable public config for the bot (env-based).
// Use LoadConfigFromEnv() to populate from environment.
type Config struct {
	// TelegramBotToken is required (TELEGRAM_BOT_TOKEN).
	TelegramBotToken string
	// TelegramPollingTimeout is the long-poll timeout (TELEGRAM_POLLING_TIMEOUT, seconds).
	TelegramPollingTimeout time.Duration
	// PollOffsetFile is the path to persist offset for restarts (TELEGRAM_POLL_OFFSET_FILE).
	PollOffsetFile string
	// Mode is "polling" or "webhook" (TELEGRAM_MODE).
	Mode string
	// WebhookSecret is used in /tg/{secret}/webhook (TELEGRAM_WEBHOOK_SECRET).
	WebhookSecret string
	// HTTPListenAddr is the HTTP server listen address, e.g. ":4000" (HTTP_LISTEN_ADDR).
	HTTPListenAddr string
}

// LoadConfigFromEnv reads config from the environment.
// Call this after loading .env (e.g. godotenv) if you use it.
func LoadConfigFromEnv() Config {
	c := config.LoadFromEnv()
	return Config{
		TelegramBotToken:       c.TelegramBotToken,
		TelegramPollingTimeout:  c.TelegramPollingTimeout,
		PollOffsetFile:         c.PollOffsetFile,
		Mode:                   c.Mode,
		WebhookSecret:          c.WebhookSecret,
		HTTPListenAddr:         c.HTTPListenAddr,
	}
}

// Valid returns true if required fields are set for running.
func (c Config) Valid() bool {
	return c.TelegramBotToken != ""
}
