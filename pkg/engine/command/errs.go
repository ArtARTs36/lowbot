package command

type ValidationError struct {
	Text string
}

func NewValidationError(text string) *ValidationError {
	return &ValidationError{Text: text}
}

func (e *ValidationError) Error() string {
	return e.Text
}

type AccessDeniedError struct {
	Message string
}

func NewAccessDeniedError(message string) *AccessDeniedError {
	return &AccessDeniedError{Message: message}
}

func (e *AccessDeniedError) Error() string {
	return e.Message
}
