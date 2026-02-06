package ports

import "helix-bot/pkg/types"

// TelegramClient abstracts Telegram API (interfaces.md §3).
type TelegramClient interface {
	SendMessage(chatID int64, text string, opts types.SendOption) (types.MessageRef, error)
	SendPhoto(chatID int64, photo types.InputFile, opts types.SendOption) (types.MessageRef, error)
	SendDocument(chatID int64, doc types.InputFile, opts types.SendOption) (types.MessageRef, error)
	EditMessageText(ref types.MessageRef, text string, opts types.EditOption) error
	DeleteMessage(ref types.MessageRef) error
	AnswerCallbackQuery(callbackID string, text string, opts types.AnswerOption) error
	SendChatAction(chatID int64, action string) error
}
