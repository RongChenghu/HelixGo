package ports

import "helix-bot/pkg/types"

// Ctx is the only object handlers see (interfaces.md §2).
type Ctx interface {
	Update() any
	UpdateID() int64
	ChatID() int64
	UserID() int64
	MessageID() int64
	Text() string
	CallbackData() string
	RequestID() string
	Logger() Logger
	Get(key string) (any, bool)
	Set(key string, val any)

	ReplyText(text string, opts ...types.SendOption) (types.MessageRef, error)
	SendText(chatID int64, text string, opts ...types.SendOption) (types.MessageRef, error)
	EditText(ref types.MessageRef, text string, opts ...types.EditOption) error
	AnswerCallback(text string, opts ...types.AnswerOption) error
	SendPhoto(chatID int64, photo types.InputFile, opts ...types.SendOption) (types.MessageRef, error)
	SendDocument(chatID int64, doc types.InputFile, opts ...types.SendOption) (types.MessageRef, error)
	NowUnix() int64
}
