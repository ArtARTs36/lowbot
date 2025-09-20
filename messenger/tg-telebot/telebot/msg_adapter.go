package telebot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/artarts36/lowbot/logx"

	"github.com/artarts36/lowbot/messenger/messengerapi"
	"github.com/artarts36/lowbot/messenger/tg-telebot/telebot/callback"
	"gopkg.in/telebot.v4"
)

type messageAdapter struct {
	callbackManager *callback.Manager
	logger          logx.Logger
}

func newMessageAdapter(callbackManager *callback.Manager, logger logx.Logger) *messageAdapter {
	return &messageAdapter{
		callbackManager: callbackManager,
		logger:          logger,
	}
}

func (a *messageAdapter) AdaptMessage(ctx telebot.Context, msg *telebot.Message) *message {
	return &message{
		ctx:    ctx,
		id:     fmt.Sprintf("%d", msg.ID),
		chatID: strconv.FormatInt(msg.Chat.ID, 10),
		text:   msg.Text,
		sender: a.userToSender(msg.Sender),
	}
}

func (a *messageAdapter) AdaptCallback(teleCtx telebot.Context, clb *telebot.Callback) (*message, error) {
	msg := &message{
		ctx:    teleCtx,
		id:     clb.ID,
		chatID: strconv.FormatInt(clb.Message.Chat.ID, 10),
		text:   clb.Message.Text,
		sender: a.userToSender(clb.Sender),
	}

	ctx := context.Background()

	storedCallback, err := a.callbackManager.Find(ctx, a.cleanCallbackID(clb.Data))
	if err != nil {
		if errors.Is(err, callback.ErrNotFound) {
			a.logger.WarnContext(ctx, "[lowbot] stored callback not found", slog.String("callback.id", clb.ID))
		} else {
			return nil, fmt.Errorf("find stored callback: %w", err)
		}
	} else {
		a.logger.DebugContext(ctx, "[lowbot] stored callback found",
			slog.String("callback.id", clb.ID),
			slog.Any("stored_callback", storedCallback),
		)

		switch v := storedCallback.Value.(type) {
		case *callback.PassEnumValue:
			msg.text = v.Value
		case *callback.CommandButton:
			msg.args = &messengerapi.Args{
				CommandName: v.CommandName,
				StateName:   v.StateName,
				Data:        v.Data,
			}
		}
	}

	return msg, nil
}

func (a *messageAdapter) cleanCallbackID(value string) string {
	return strings.TrimPrefix(value, "\f")
}

func (a *messageAdapter) userToSender(user *telebot.User) *messengerapi.Sender {
	return &messengerapi.Sender{
		ID:        strconv.FormatInt(user.ID, 10),
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
}
