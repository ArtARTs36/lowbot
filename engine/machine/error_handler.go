package machine

import (
	"context"
	"errors"
	"log/slog"

	"github.com/artarts36/lowbot/engine/command"
	"github.com/artarts36/lowbot/messenger/messengerapi"
)

// ErrorHandler
// Return true, when ErrorHandler handled error.
type ErrorHandler func(ctx context.Context, msg messengerapi.Message, err error) (bool, error)

var (
	validationErrHandler ErrorHandler = func(_ context.Context, msg messengerapi.Message, err error) (bool, error) {
		validErr := &command.ValidationError{}
		if errors.As(err, &validErr) {
			return true, msg.Respond(&messengerapi.Answer{
				Text: validErr.Text,
			})
		}
		return false, err
	}

	accessDeniedErrHandler ErrorHandler = func(ctx context.Context, msg messengerapi.Message, err error) (bool, error) {
		accessDeniedErr := &command.AccessDeniedError{}
		if errors.As(err, &accessDeniedErr) {
			slog.InfoContext(ctx, "[machine] access denied for user")

			return true, msg.Respond(&messengerapi.Answer{
				Text: accessDeniedErr.Message,
			})
		}
		return false, err
	}
)

type compositeErrorHandler struct {
	handlers []ErrorHandler
}

func NewErrorHandler(handler ...ErrorHandler) ErrorHandler {
	ch := compositeErrorHandler{
		handlers: append([]ErrorHandler{accessDeniedErrHandler, validationErrHandler}, handler...),
	}

	return ch.Handle
}

func (h *compositeErrorHandler) Handle(ctx context.Context, msg messengerapi.Message, handlingErr error) (bool, error) {
	for _, handler := range h.handlers {
		handled, err := handler(ctx, msg, handlingErr)
		if err != nil {
			return false, err
		}

		if handled {
			return true, err
		}
	}

	return false, handlingErr
}
