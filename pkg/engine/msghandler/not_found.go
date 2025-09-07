package msghandler

import (
	"context"
	"github.com/artarts36/lowbot/pkg/engine/messenger"
)

type CommandNotFoundFallback func(ctx context.Context, message messenger.Message) error

func defaultCommandNotFoundFallback() CommandNotFoundFallback {
	return func(ctx context.Context, message messenger.Message) error {
		return message.RespondText("Команда не найдена")
	}
}
