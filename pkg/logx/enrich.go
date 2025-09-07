package logx

import "context"

type messageID struct{}

func WithMessageID(ctx context.Context, msgID string) context.Context {
	return context.WithValue(ctx, messageID{}, msgID)
}

func MessageID(ctx context.Context) string {
	return ctx.Value(messageID{}).(string)
}
