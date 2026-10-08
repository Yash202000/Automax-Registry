package services

const (
	KindInvalid  = "invalid"
	KindNotFound = "not_found"
	KindConflict = "conflict"
)

type Error struct {
	Kind    string
	Message string
	Fields  map[string]string
}

func (e *Error) Error() string { return e.Message }

func Invalid(message string, fields map[string]string) *Error {
	return &Error{Kind: KindInvalid, Message: message, Fields: fields}
}

func NotFound(message string) *Error {
	return &Error{Kind: KindNotFound, Message: message}
}

func Conflict(message string) *Error {
	return &Error{Kind: KindConflict, Message: message}
}
