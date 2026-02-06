package runtime

import (
	"time"

	"helix-bot/pkg/errors"
	"helix-bot/pkg/ports"
	"helix-bot/pkg/types"
)

type ctxImpl struct {
	update    types.BotUpdate
	requestID string
	logger    ports.Logger
	client    ports.TelegramClient
	bag       map[string]any
}

// NewCtx builds a Ctx for the given update (used by runtime to pass to handler).
func NewCtx(update types.BotUpdate, requestID string, logger ports.Logger, client ports.TelegramClient) ports.Ctx {
	return &ctxImpl{
		update:    update,
		requestID: requestID,
		logger:    logger,
		client:    client,
		bag:       make(map[string]any),
	}
}

func (c *ctxImpl) Update() any                { return c.update.Raw }
func (c *ctxImpl) UpdateID() int64            { return c.update.UpdateID }
func (c *ctxImpl) RequestID() string          { return c.requestID }
func (c *ctxImpl) Logger() ports.Logger       { return c.logger }
func (c *ctxImpl) Get(key string) (any, bool) { v, ok := c.bag[key]; return v, ok }
func (c *ctxImpl) Set(key string, val any)    { c.bag[key] = val }
func (c *ctxImpl) NowUnix() int64             { return time.Now().Unix() }

func (c *ctxImpl) ChatID() int64 {
	if c.update.Message != nil {
		return c.update.Message.ChatID
	}
	if c.update.CallbackQuery != nil {
		return c.update.CallbackQuery.ChatID
	}
	return 0
}

func (c *ctxImpl) UserID() int64 {
	if c.update.Message != nil {
		return c.update.Message.UserID
	}
	if c.update.CallbackQuery != nil {
		return c.update.CallbackQuery.UserID
	}
	return 0
}

func (c *ctxImpl) MessageID() int64 {
	if c.update.Message != nil {
		return c.update.Message.MessageID
	}
	if c.update.CallbackQuery != nil {
		return c.update.CallbackQuery.MessageID
	}
	return 0
}

func (c *ctxImpl) Text() string {
	if c.update.Message != nil {
		return c.update.Message.Text
	}
	return ""
}

func (c *ctxImpl) CallbackData() string {
	if c.update.CallbackQuery != nil {
		return c.update.CallbackQuery.Data
	}
	return ""
}

func (c *ctxImpl) ReplyText(text string, opts ...types.SendOption) (types.MessageRef, error) {
	chatID := c.ChatID()
	if chatID == 0 {
		return types.MessageRef{}, errors.ErrNoChatContext
	}
	var opt types.SendOption
	if len(opts) > 0 {
		opt = opts[0]
	}
	return c.client.SendMessage(chatID, text, opt)
}

func (c *ctxImpl) SendText(chatID int64, text string, opts ...types.SendOption) (types.MessageRef, error) {
	var opt types.SendOption
	if len(opts) > 0 {
		opt = opts[0]
	}
	return c.client.SendMessage(chatID, text, opt)
}

func (c *ctxImpl) EditText(ref types.MessageRef, text string, opts ...types.EditOption) error {
	var opt types.EditOption
	if len(opts) > 0 {
		opt = opts[0]
	}
	return c.client.EditMessageText(ref, text, opt)
}

func (c *ctxImpl) AnswerCallback(text string, opts ...types.AnswerOption) error {
	if c.update.CallbackQuery == nil {
		return nil
	}
	var opt types.AnswerOption
	if len(opts) > 0 {
		opt = opts[0]
	}
	return c.client.AnswerCallbackQuery(c.update.CallbackQuery.ID, text, opt)
}

func (c *ctxImpl) SendPhoto(chatID int64, photo types.InputFile, opts ...types.SendOption) (types.MessageRef, error) {
	var opt types.SendOption
	if len(opts) > 0 {
		opt = opts[0]
	}
	return c.client.SendPhoto(chatID, photo, opt)
}

func (c *ctxImpl) SendDocument(chatID int64, doc types.InputFile, opts ...types.SendOption) (types.MessageRef, error) {
	var opt types.SendOption
	if len(opts) > 0 {
		opt = opts[0]
	}
	return c.client.SendDocument(chatID, doc, opt)
}
