package middleware

import (
	"context"
	"slices"

	"github.com/artarts36/lowbot/pkg/engine/command"
)

func OnlyChatsWithMessage(ids []string, message string) command.Middleware {
	return func(ctx context.Context, req *command.Request, next command.ActionCallback) error {
		if slices.Contains(ids, req.Message.GetChatID()) {
			return next(ctx, req)
		}
		return command.NewAccessDeniedError(message)
	}
}

func OnlyChats(ids []string) command.Middleware {
	return OnlyChatsWithMessage(ids, "Access Denied.")
}
