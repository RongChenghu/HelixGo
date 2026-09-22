package helixbot

import (
	"testing"
)

func TestLoadConfigFromEnv(t *testing.T) {
	cfg := LoadConfigFromEnv()
	// Just ensure it doesn't panic; env may be empty in CI.
	_ = cfg.Mode
	_ = cfg.HTTPListenAddr
	if cfg.TelegramPollingTimeout == 0 {
		t.Log("default polling timeout is set by internal config")
	}
}

func TestNew_InvalidConfig(t *testing.T) {
	cfg := Config{} // no token
	bot, err := New(cfg)
	if err == nil {
		t.Fatal("expected error when token is empty")
	}
	if err != ErrMissingToken {
		t.Fatalf("expected ErrMissingToken, got %v", err)
	}
	if bot != nil {
		t.Fatal("expected nil bot")
	}
}

func TestConfig_Valid(t *testing.T) {
	if (Config{}).Valid() {
		t.Error("empty config should not be valid")
	}
	if !(Config{TelegramBotToken: "x"}).Valid() {
		t.Error("config with token should be valid")
	}
}
