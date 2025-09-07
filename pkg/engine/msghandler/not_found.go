package msghandler

import (
	"context"
	"github.com/artarts36/lowbot/pkg/engine/messenger"
)

type CommandNotFoundFallback func(ctx context.Context, message messenger.Message) error

func DefaultCommandNotFoundFallback() CommandNotFoundFallback {
	return func(ctx context.Context, message messenger.Message) error {
		return message.Respond(&messenger.Answer{
			Text: "Команда не найдена",
		})
	}
}
