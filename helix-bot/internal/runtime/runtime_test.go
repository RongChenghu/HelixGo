package runtime

import (
	"context"
	"testing"

	"helix-bot/internal/adapter/fake"
	"helix-bot/internal/router"
	"helix-bot/pkg/ports"
	"helix-bot/pkg/types"
)

// testLogger is a minimal Logger implementation for tests.
type testLogger struct {
	lastErrorMsg string
}

func (l *testLogger) Debug(msg string, kv ...any) {}
func (l *testLogger) Info(msg string, kv ...any)  {}
func (l *testLogger) Warn(msg string, kv ...any)  {}
func (l *testLogger) Error(msg string, kv ...any) {
	l.lastErrorMsg = msg
}

// TestRuntimePingPong verifies /ping -> pong and that dedup in FakeUpdateSource
// prevents duplicate updates from being processed twice.
func TestRuntimePingPong(t *testing.T) {
	src := &fake.FakeUpdateSource{
		Updates: []types.BotUpdate{
			{
				UpdateID: 1,
				Message: &types.BotMessage{
					ChatID: 123,
					UserID: 42,
					Text:   "/ping",
				},
			},
			// Duplicate update_id should be skipped by FakeUpdateSource's deduper.
			{
				UpdateID: 1,
				Message: &types.BotMessage{
					ChatID: 123,
					UserID: 42,
					Text:   "/ping",
				},
			},
		},
	}
	client := &fake.FakeTelegramClient{}
	log := &testLogger{}

	r := router.New()
	rt := NewBotRuntime(r, src, log, client)

	r.OnCommand("/ping", func(ctx ports.Ctx) error {
		_, err := ctx.ReplyText("pong")
		return err
	})

	ctx := context.Background()
	if err := rt.Run(ctx); err != nil {
		t.Fatalf("runtime.Run error: %v", err)
	}

	if len(client.Sent) != 1 {
		t.Fatalf("expected 1 message sent (dedup), got %d", len(client.Sent))
	}
	rec := client.Sent[0]
	if rec.Method != "SendMessage" {
		t.Fatalf("expected method SendMessage, got %q", rec.Method)
	}
	if rec.ChatID != 123 {
		t.Fatalf("expected ChatID=123, got %d", rec.ChatID)
	}
	if rec.Text != "pong" {
		t.Fatalf("expected text 'pong', got %q", rec.Text)
	}
}

// TestRuntimeHandlerPanicRecovered verifies that a panicking handler is recovered
// by Runtime and does not crash the loop.
func TestRuntimeHandlerPanicRecovered(t *testing.T) {
	src := &fake.FakeUpdateSource{
		Updates: []types.BotUpdate{
			{
				UpdateID: 2,
				Message: &types.BotMessage{
					ChatID: 100,
					UserID: 1,
					Text:   "/panic",
				},
			},
		},
	}
	client := &fake.FakeTelegramClient{}
	log := &testLogger{}

	r := router.New()
	rt := NewBotRuntime(r, src, log, client)

	r.OnCommand("/panic", func(ctx ports.Ctx) error {
		panic("boom")
	})

	ctx := context.Background()
	if err := rt.Run(ctx); err != nil {
		t.Fatalf("runtime.Run error: %v", err)
	}

	if log.lastErrorMsg != "handler panic" {
		t.Fatalf("expected handler panic log, got %q", log.lastErrorMsg)
	}
}
