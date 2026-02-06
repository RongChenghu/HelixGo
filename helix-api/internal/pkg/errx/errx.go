package errx

type Error struct {
	Code    string
	Message string
	Status  int
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Code + ": " + e.Message
}
