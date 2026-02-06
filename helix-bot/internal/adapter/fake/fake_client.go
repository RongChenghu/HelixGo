package fake

import (
	"sync"

	"helix-bot/pkg/ports"
	"helix-bot/pkg/types"
)

// SendRecord captures a single TelegramClient call for assertions in tests.
type SendRecord struct {
	Method     string
	ChatID     int64
	Text       string
	MessageRef types.MessageRef
	Photo      types.InputFile
	Document   types.InputFile
	CallbackID string
	Action     string
}

// FakeTelegramClient is an in-memory implementation of ports.TelegramClient
// for tests. It only records calls and never touches the network.
type FakeTelegramClient struct {
	mu   sync.Mutex
	Sent []SendRecord
}

func (c *FakeTelegramClient) append(rec SendRecord) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Sent = append(c.Sent, rec)
}

func (c *FakeTelegramClient) SendMessage(chatID int64, text string, opts types.SendOption) (types.MessageRef, error) {
	ref := types.MessageRef{ChatID: chatID, MessageID: int64(len(c.Sent) + 1)}
	c.append(SendRecord{
		Method:     "SendMessage",
		ChatID:     chatID,
		Text:       text,
		MessageRef: ref,
	})
	return ref, nil
}

func (c *FakeTelegramClient) SendPhoto(chatID int64, photo types.InputFile, opts types.SendOption) (types.MessageRef, error) {
	ref := types.MessageRef{ChatID: chatID, MessageID: int64(len(c.Sent) + 1)}
	c.append(SendRecord{
		Method: "SendPhoto",
		ChatID: chatID,
		Photo:  photo,
	})
	return ref, nil
}

func (c *FakeTelegramClient) SendDocument(chatID int64, doc types.InputFile, opts types.SendOption) (types.MessageRef, error) {
	ref := types.MessageRef{ChatID: chatID, MessageID: int64(len(c.Sent) + 1)}
	c.append(SendRecord{
		Method:   "SendDocument",
		ChatID:   chatID,
		Document: doc,
	})
	return ref, nil
}

func (c *FakeTelegramClient) EditMessageText(ref types.MessageRef, text string, opts types.EditOption) error {
	c.append(SendRecord{
		Method:     "EditMessageText",
		MessageRef: ref,
		Text:       text,
	})
	return nil
}

func (c *FakeTelegramClient) DeleteMessage(ref types.MessageRef) error {
	c.append(SendRecord{
		Method:     "DeleteMessage",
		MessageRef: ref,
	})
	return nil
}

func (c *FakeTelegramClient) AnswerCallbackQuery(callbackID string, text string, opts types.AnswerOption) error {
	c.append(SendRecord{
		Method:     "AnswerCallbackQuery",
		CallbackID: callbackID,
		Text:       text,
	})
	return nil
}

func (c *FakeTelegramClient) SendChatAction(chatID int64, action string) error {
	c.append(SendRecord{
		Method: "SendChatAction",
		ChatID: chatID,
		Action: action,
	})
	return nil
}

var _ ports.TelegramClient = (*FakeTelegramClient)(nil)
