package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds bot runtime config (env only; no business rules).
type Config struct {
	// TELEGRAM_BOT_TOKEN is required for Telegram API.
	TelegramBotToken string
	// TELEGRAM_POLLING_TIMEOUT optional; seconds for long polling timeout.
	TelegramPollingTimeout time.Duration
	// TELEGRAM_POLL_OFFSET_FILE optional; path to persist offset so restart continues from last.
	PollOffsetFile string
	// Mode: "polling" or "webhook" (v0.1 focus on polling).
	Mode string
}

// sanitizeToken keeps only valid Telegram bot token chars (digits, ':', letters, '-', '_').
// Strips BOM, \\r, \\n and any other chars that cause 404 when read from .env files.
func sanitizeToken(s string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(s) {
		switch {
		case r >= '0' && r <= '9', r == ':', r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r == '-', r == '_':
			b.WriteRune(r)
		}
	}
	return b.String()
}

// LoadFromEnv reads config from environment.
func LoadFromEnv() Config {
	cfg := Config{
		TelegramBotToken:       sanitizeToken(os.Getenv("TELEGRAM_BOT_TOKEN")),
		TelegramPollingTimeout: 60 * time.Second,
		PollOffsetFile:         strings.TrimSpace(os.Getenv("TELEGRAM_POLL_OFFSET_FILE")),
		Mode:                   getEnv("TELEGRAM_MODE", "polling"),
	}
	if s := os.Getenv("TELEGRAM_POLLING_TIMEOUT"); s != "" {
		if sec, err := strconv.Atoi(s); err == nil && sec > 0 {
			cfg.TelegramPollingTimeout = time.Duration(sec) * time.Second
		}
	}
	return cfg
}

// Valid returns true if required fields are set for running.
func (c Config) Valid() bool {
	return c.TelegramBotToken != ""
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
