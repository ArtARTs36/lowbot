package logx

import (
	"context"
	"log/slog"

	"github.com/cappuccinotm/slogx"
)

type (
	messageID   struct{}
	chatID      struct{}
	commandName struct{}
)

func WithMessageID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, messageID{}, id)
}

func GetMessageID(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(messageID{}).(string)
	return value, ok
}

func PropagateMessageID() slogx.Middleware {
	return func(next slogx.HandleFunc) slogx.HandleFunc {
		return func(ctx context.Context, rec slog.Record) error {
			if id, ok := GetMessageID(ctx); ok {
				rec.AddAttrs(slog.String("message.id", id))
			}

			return next(ctx, rec)
		}
	}
}

func WithChatID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, chatID{}, id)
}

func GetChatID(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(chatID{}).(string)
	return value, ok
}

func PropagateChatID() slogx.Middleware {
	return func(next slogx.HandleFunc) slogx.HandleFunc {
		return func(ctx context.Context, rec slog.Record) error {
			if id, ok := GetChatID(ctx); ok {
				rec.AddAttrs(slog.String("chat.id", id))
			}

			return next(ctx, rec)
		}
	}
}

func WithCommandName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, commandName{}, name)
}

func GetCommandName(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(commandName{}).(string)
	return value, ok
}

func PropagateCommandName() slogx.Middleware {
	return func(next slogx.HandleFunc) slogx.HandleFunc {
		return func(ctx context.Context, rec slog.Record) error {
			if name, ok := GetCommandName(ctx); ok {
				rec.AddAttrs(slog.String("command.name", name))
			}

			return next(ctx, rec)
		}
	}
}
