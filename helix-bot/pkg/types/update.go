package types

// BotUpdate is the normalized update (interfaces.md §4.2).
type BotUpdate struct {
	Raw      any
	UpdateID int64

	Message       *BotMessage
	CallbackQuery *BotCallbackQuery
}

// BotMessage (interfaces.md §4.2).
type BotMessage struct {
	ChatID    int64
	UserID    int64
	MessageID int64
	Text      string
}

// BotCallbackQuery (interfaces.md §4.2).
type BotCallbackQuery struct {
	ID        string
	ChatID    int64
	UserID    int64
	MessageID int64
	Data      string
}
