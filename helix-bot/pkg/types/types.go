package types

// MessageRef references a message for edit/delete.
type MessageRef struct {
	ChatID    int64
	MessageID int64
}

// InputFile represents file input (exactly one of FileID/URL/LocalPath set).
type InputFile struct {
	FileID    string // Telegram file_id
	URL       string // http(s) url
	LocalPath string // local file path
}

// SendOption for sendMessage/sendPhoto/sendDocument.
type SendOption struct {
	ParseMode         string // "MarkdownV2" / "HTML" / ""
	DisablePreview   bool
	ReplyToMessageID int64
	ReplyMarkup      any
}

// EditOption for editMessageText.
type EditOption struct {
	ParseMode   string
	ReplyMarkup any
}

// AnswerOption for answerCallbackQuery.
type AnswerOption struct {
	ShowAlert bool
}
