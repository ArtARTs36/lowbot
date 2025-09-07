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
