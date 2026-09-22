package types

// BotUpdate is the normalized update (interfaces.md §4.2).
type BotUpdate struct {
	Raw      any
	UpdateID int64

	Message       *BotMessage
	CallbackQuery *BotCallbackQuery
}

// BotDice is the normalized dice from message.dice (Telegram 🎲 etc.).
type BotDice struct {
	Emoji string // e.g. "🎲"
	Value int    // 1–6 for 🎲
}

// BotMessage (interfaces.md §4.2).
type BotMessage struct {
	ChatID    int64
	UserID    int64
	MessageID int64
	Text      string
	Dice      *BotDice // set when message contains a dice (e.g. 🎲)
}

// BotCallbackQuery (interfaces.md §4.2).
type BotCallbackQuery struct {
	ID        string
	ChatID    int64
	UserID    int64
	MessageID int64
	Data      string
}
