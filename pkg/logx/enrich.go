package logx

import (
	"context"
	"github.com/cappuccinotm/slogx"
	"log/slog"
)

type (
	messageID struct{}
	chatID    struct{}
)

func WithMessageID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, messageID{}, id)
}

func MessageID(ctx context.Context) (value string, ok bool) {
	value, ok = ctx.Value(messageID{}).(string)
	return
}

func PropagateMessageID() slogx.Middleware {
	return func(next slogx.HandleFunc) slogx.HandleFunc {
		return func(ctx context.Context, rec slog.Record) error {
			if id, ok := MessageID(ctx); ok {
				rec.AddAttrs(slog.String("message.id", id))
			}

			return next(ctx, rec)
		}
	}
}

func WithChatID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, chatID{}, id)
}

func ChatID(ctx context.Context) (value string, ok bool) {
	value, ok = ctx.Value(chatID{}).(string)
	return
}

func PropagateChatID() slogx.Middleware {
	return func(next slogx.HandleFunc) slogx.HandleFunc {
		return func(ctx context.Context, rec slog.Record) error {
			if id, ok := ChatID(ctx); ok {
				rec.AddAttrs(slog.String("chat.id", id))
			}

			return next(ctx, rec)
		}
	}
}
