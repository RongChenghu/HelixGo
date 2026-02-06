package telegram

import (
	"context"
	"strings"

	"helix-bot/pkg/errors"
	"helix-bot/pkg/ports"
	"helix-bot/pkg/types"
)

// Client implements ports.TelegramClient with real HTTP (net/http).
type Client struct {
	Token string
}

// NewClient creates a Telegram client (token required for real API calls).
// Token is trimmed to avoid 404 from trailing newline in .env.
func NewClient(token string) *Client {
	return &Client{Token: strings.TrimSpace(token)}
}

// SendMessage implements ports.TelegramClient (real sendMessage API).
func (c *Client) SendMessage(chatID int64, text string, opts types.SendOption) (types.MessageRef, error) {
	if c.Token == "" {
		return types.MessageRef{}, errors.ErrTelegramAPI
	}
	parseMode := opts.ParseMode
	messageID, err := SendMessage(context.Background(), c.Token, chatID, text, parseMode)
	if err != nil {
		return types.MessageRef{}, err
	}
	return types.MessageRef{ChatID: chatID, MessageID: messageID}, nil
}

// SendPhoto implements ports.TelegramClient (stub for v0.1).
func (c *Client) SendPhoto(chatID int64, photo types.InputFile, opts types.SendOption) (types.MessageRef, error) {
	if c.Token == "" {
		return types.MessageRef{}, errors.ErrTelegramAPI
	}
	return types.MessageRef{ChatID: chatID, MessageID: 0}, nil
}

// SendDocument implements ports.TelegramClient (stub for v0.1).
func (c *Client) SendDocument(chatID int64, doc types.InputFile, opts types.SendOption) (types.MessageRef, error) {
	if c.Token == "" {
		return types.MessageRef{}, errors.ErrTelegramAPI
	}
	return types.MessageRef{ChatID: chatID, MessageID: 0}, nil
}

// EditMessageText implements ports.TelegramClient (stub for v0.1).
func (c *Client) EditMessageText(ref types.MessageRef, text string, opts types.EditOption) error {
	if c.Token == "" {
		return errors.ErrTelegramAPI
	}
	return nil
}

// DeleteMessage implements ports.TelegramClient (stub for v0.1).
func (c *Client) DeleteMessage(ref types.MessageRef) error {
	if c.Token == "" {
		return errors.ErrTelegramAPI
	}
	return nil
}

// AnswerCallbackQuery implements ports.TelegramClient (stub for v0.1).
func (c *Client) AnswerCallbackQuery(callbackID string, text string, opts types.AnswerOption) error {
	if c.Token == "" {
		return errors.ErrTelegramAPI
	}
	return nil
}

// SendChatAction implements ports.TelegramClient (stub for v0.1).
func (c *Client) SendChatAction(chatID int64, action string) error {
	if c.Token == "" {
		return errors.ErrTelegramAPI
	}
	return nil
}

var _ ports.TelegramClient = (*Client)(nil)
