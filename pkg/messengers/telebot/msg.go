package telebot

import (
	"fmt"
	"github.com/artarts36/lowbot/pkg/engine/messenger"
	"gopkg.in/telebot.v4"
	"log/slog"
	"strconv"
	"strings"
)

type message struct {
	ctx telebot.Context

	id     string
	chatID string
	text   string
}

func newMessageFromMessage(msg *telebot.Message, ctx telebot.Context) *message {
	return &message{
		ctx:    ctx,
		id:     fmt.Sprintf("%d", msg.ID),
		chatID: strconv.FormatInt(msg.Chat.ID, 10),
		text:   msg.Text,
	}
}

func newMessageFromCallback(clb *telebot.Callback, ctx telebot.Context) *message {
	msg := &message{
		ctx:    ctx,
		id:     clb.ID,
		chatID: strconv.FormatInt(clb.Message.Chat.ID, 10),
		text:   clb.Message.Text,
	}

	clbID, ok := parseCallbackID(clb.Data)
	slog.Debug("parsed callback data", slog.String("type", string(clbID.Type)), slog.Any("data", clbID.Values))
	if ok {
		switch clbID.Type {
		case callbackTypeEnum:
			slog.Debug(
				"use enum value from callback",
				slog.String("callback.id", clbID.ID),
				slog.String("value", clbID.Values[0]),
			)

			msg.text = clbID.Values[0]
		}
	}

	return msg
}

func (m *message) GetID() string {
	return m.id
}

func (m *message) GetChatID() string {
	return m.chatID
}

func (m *message) GetBody() string {
	return m.text
}

func (m *message) Respond(answer *messenger.Answer) error {
	var what interface{}
	var opts []interface{}

	if answer.Text != "" {
		what = answer.Text
	}

	if len(answer.Menu) > 0 {
		opts = append(opts, m.buildMenuOpt(answer.Menu))
	}

	if answer.Enum.Valid() {
		opts = append(opts, m.buildEnumOpt(answer.Enum))
	}

	return m.ctx.Send(what, opts...)
}

func (m *message) ExtractCommandName() string {
	if strings.HasPrefix(m.text, "/") {
		return strings.TrimSpace(m.text[1:])
	}
	return ""
}

func (m *message) buildMenuOpt(enum []string) *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
		IsPersistent:    false,
	}

	rows := make([]telebot.Row, 1)
	for _, value := range enum {
		btn := menu.Text(value)

		rows[0] = append(rows[0], btn)
	}

	menu.Reply(rows...)

	return menu
}

func (m *message) buildEnumOpt(enum messenger.Enum) *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
		IsPersistent:    false,
	}

	rows := make([]telebot.Row, 1)
	for value, title := range enum.Values {
		btn := menu.Text(title)
		btn.Unique = createEnumCallbackID(value).ID

		rows[0] = append(rows[0], btn)
	}

	menu.Inline(rows...)

	return menu
}
