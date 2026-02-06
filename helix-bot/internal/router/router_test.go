package router

import (
	"testing"

	"helix-bot/pkg/ports"
	"helix-bot/pkg/types"
)

// TestMiddlewareOrder verifies that global and route-specific middlewares run
// in the expected Before/After order:
// incoming -> mwG1 -> mwG2 -> mwR1 -> handler -> mwR1(after) -> mwG2(after) -> mwG1(after)
func TestMiddlewareOrder(t *testing.T) {
	r := New()

	var calls []string

	mwG1 := func(next ports.Handler) ports.Handler {
		return func(ctx ports.Ctx) error {
			calls = append(calls, "G1-before")
			err := next(ctx)
			calls = append(calls, "G1-after")
			return err
		}
	}
	mwG2 := func(next ports.Handler) ports.Handler {
		return func(ctx ports.Ctx) error {
			calls = append(calls, "G2-before")
			err := next(ctx)
			calls = append(calls, "G2-after")
			return err
		}
	}
	mwR1 := func(next ports.Handler) ports.Handler {
		return func(ctx ports.Ctx) error {
			calls = append(calls, "R1-before")
			err := next(ctx)
			calls = append(calls, "R1-after")
			return err
		}
	}

	handler := func(ctx ports.Ctx) error {
		calls = append(calls, "H")
		return nil
	}

	r.Use(mwG1, mwG2)
	r.OnCommand("/ping", handler, mwR1)

	u := types.BotUpdate{
		UpdateID: 1,
		Message: &types.BotMessage{
			ChatID: 123,
			Text:   "/ping",
		},
	}

	h := r.Route(u)
	if h == nil {
		t.Fatalf("expected handler, got nil")
	}
	if err := h(nil); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	want := []string{
		"G1-before",
		"G2-before",
		"R1-before",
		"H",
		"R1-after",
		"G2-after",
		"G1-after",
	}
	if len(calls) != len(want) {
		t.Fatalf("calls len=%d, want=%d, calls=%v", len(calls), len(want), calls)
	}
	for i, s := range want {
		if calls[i] != s {
			t.Fatalf("calls[%d]=%q, want %q; full=%v", i, calls[i], s, calls)
		}
	}
}
