package telegram

import (
	"encoding/json"
	"testing"

	"helix-bot/pkg/types"
)

// TestNormalizeUpdate_Dice ensures message.dice is mapped to BotMessage.Dice (Emoji, Value).
func TestNormalizeUpdate_Dice(t *testing.T) {
	// Telegram update with message.dice (🎲 value 4)
	raw := []byte(`{"update_id":1,"message":{"message_id":10,"chat":{"id":123},"from":{"id":42},"dice":{"emoji":"🎲","value":4}}}`)
	var u tgUpdate
	if err := json.Unmarshal(raw, &u); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	nu := normalizeUpdate(u)
	if nu.Message == nil {
		t.Fatal("expected Message")
	}
	if nu.Message.Dice == nil {
		t.Fatal("expected Message.Dice")
	}
	if nu.Message.Dice.Emoji != "🎲" {
		t.Errorf("Dice.Emoji = %q, want 🎲", nu.Message.Dice.Emoji)
	}
	if nu.Message.Dice.Value != 4 {
		t.Errorf("Dice.Value = %d, want 4", nu.Message.Dice.Value)
	}
	// Sanity: chat/user preserved
	if nu.Message.ChatID != 123 || nu.Message.UserID != 42 || nu.Message.MessageID != 10 {
		t.Errorf("ChatID=%d UserID=%d MessageID=%d", nu.Message.ChatID, nu.Message.UserID, nu.Message.MessageID)
	}
}

// TestNormalizeUpdate_NoDice ensures updates without dice leave Dice nil.
func TestNormalizeUpdate_NoDice(t *testing.T) {
	raw := []byte(`{"update_id":2,"message":{"message_id":1,"chat":{"id":1},"from":{"id":1},"text":"/ping"}}`)
	var u tgUpdate
	if err := json.Unmarshal(raw, &u); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	nu := normalizeUpdate(u)
	if nu.Message == nil {
		t.Fatal("expected Message")
	}
	if nu.Message.Dice != nil {
		t.Errorf("expected Dice nil for text message, got %+v", nu.Message.Dice)
	}
}

// Ensure BotMessage.Dice is present in types for NiuNiuGame / callers.
var _ = types.BotMessage{}.Dice
