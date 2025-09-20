package machine

import (
	"context"
	"errors"
	"log/slog"

	"github.com/artarts36/lowbot/logx"

	"github.com/artarts36/lowbot/engine/command"
	"github.com/artarts36/lowbot/messenger/messengerapi"
)

// ErrorHandler
// Return true, when ErrorHandler handled error.
type ErrorHandler func(ctx context.Context, msg messengerapi.Message, err error) (bool, error)

var (
	invalidArgumentErrHandler ErrorHandler = func(ctx context.Context, msg messengerapi.Message, err error) (bool, error) {
		validErr := &command.InvalidArgumentError{}
		if errors.As(err, &validErr) {
			sendErr := msg.Respond(&messengerapi.Answer{
				Text: validErr.Text,
			})
			if sendErr != nil {
				slog.ErrorContext(ctx, "[invalid-argument-handler] failed to respond to invalid argument", slog.Any("err", sendErr))
			}

			return true, err
		}
		return false, err
	}

	permissionDeniedErrHandler ErrorHandler = func(
		ctx context.Context,
		msg messengerapi.Message,
		err error,
	) (bool, error) {
		permissionDeniedErr := &command.PermissionDeniedError{}
		if errors.As(err, &permissionDeniedErr) {
			slog.InfoContext(ctx, "[permission-denied-handler] permission denied for user")

			userMsg := permissionDeniedErr.Message
			if userMsg == "" {
				userMsg = "Permission denied."
			}

			sendErr := msg.Respond(&messengerapi.Answer{
				Text: userMsg,
			})
			if sendErr != nil {
				slog.ErrorContext(ctx, "[permission-denied-handler] failed to send message to user", slog.Any("err", sendErr))
			}

			return true, err
		}
		return false, err
	}

	internalConvertErrHandler ErrorHandler = func(_ context.Context, _ messengerapi.Message, err error) (bool, error) {
		internalErr := &command.InternalError{}
		if errors.As(err, &internalErr) {
			return true, nil
		}

		return true, command.NewInternalError(err)
	}
)

type compositeErrorHandler struct {
	handlers []ErrorHandler
	logger   logx.Logger
}

func NewErrorHandler(logger logx.Logger, handler ...ErrorHandler) ErrorHandler {
	ch := compositeErrorHandler{
		logger: logger,
		handlers: append([]ErrorHandler{
			permissionDeniedErrHandler,
			invalidArgumentErrHandler,
			internalConvertErrHandler,
		}, handler...),
	}

	return ch.Handle
}

func (h *compositeErrorHandler) Handle(ctx context.Context, msg messengerapi.Message, handlingErr error) (bool, error) {
	h.logger.DebugContext(ctx, "[error-handler] handling error", slog.Any("err", handlingErr))

	for _, handler := range h.handlers {
		handled, err := handler(ctx, msg, handlingErr)
		if handled {
			return true, err
		}
	}

	return false, handlingErr
}
