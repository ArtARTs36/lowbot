package command

type CodeError interface {
	Error() string
	Code() string
	codeError()
}

type InvalidArgumentError struct {
	Text string
}

type PermissionDeniedError struct {
	Message string
}

type InternalError struct {
	Err error
}

func NewInvalidArgumentError(text string) *InvalidArgumentError {
	return &InvalidArgumentError{Text: text}
}

func NewPermissionDeniedError(message string) *PermissionDeniedError {
	return &PermissionDeniedError{Message: message}
}

func NewInternalError(err error) *InternalError {
	return &InternalError{err}
}

func (e *InvalidArgumentError) Error() string {
	return e.Text
}

func (e *PermissionDeniedError) Error() string {
	return e.Message
}

func (e *InternalError) Error() string {
	return e.Err.Error()
}

func (e *InvalidArgumentError) Code() string {
	return "InvalidArgument"
}

func (e *PermissionDeniedError) Code() string {
	return "PermissionDenied"
}

func (e *InternalError) Code() string {
	return "Internal"
}

func (e *InvalidArgumentError) codeError()  {}
func (e *PermissionDeniedError) codeError() {}
func (e *InternalError) codeError()         {}

func (e *InternalError) Unwrap() error {
	return e.Err
}
