package middleware

import (
	"context"
	"errors"
	"log/slog"

	"github.com/artarts36/lowbot/engine/command"
	"github.com/artarts36/lowbot/messenger/messengerapi"
)

func PleaseRepeatAgain() command.Middleware {
	return PleaseRepeatAgainWithMessage("Please repeat again.")
}

func PleaseRepeatAgainWithMessage(message string) command.Middleware {
	return func(ctx context.Context, req *command.Request, next command.ActionCallback) error {
		err := next(ctx, req)
		if err != nil {
			var intErr *command.InternalError
			if errors.As(err, &intErr) {
				_, sendErr := req.Responder.Respond(&messengerapi.Answer{
					Text: message,
				})
				if sendErr != nil {
					slog.ErrorContext(ctx, "[please-repeat-again] failed to send answer", slog.Any("err", err))
				}
			}
		}
		return err
	}
}
