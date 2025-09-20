package telebot

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/artarts36/lowbot/messenger/messengerapi"
	"github.com/artarts36/lowbot/messenger/tg-telebot/telebot/callback"

	"gopkg.in/telebot.v4"
)

type message struct {
	ctx telebot.Context

	id     string
	chatID string
	text   string
	sender *messengerapi.Sender

	callbackManager *callback.Manager
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

func (m *message) GetSender() *messengerapi.Sender {
	return m.sender
}

func (m *message) Respond(answer *messengerapi.Answer) error {
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

func (m *message) RespondObject(object messengerapi.Object) error {
	var what interface{}

	switch o := object.(type) {
	case *messengerapi.LocalImage:
		what = &telebot.Photo{
			File: telebot.File{
				FileLocal: o.Path,
			},
		}
	default:
		return fmt.Errorf("unsupported object type: %T", o)
	}

	return m.ctx.Send(what)
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

func (m *message) buildEnumOpt(enum messengerapi.Enum) *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
		IsPersistent:    false,
	}

	rows := make([]telebot.Row, 1)
	for value, title := range enum.Values {
		btn := menu.Text(title)

		if err := m.callbackManager.Bind(context.Background(), &btn, callback.NewEnum(value)); err != nil {
			slog.Error("failed to bind enum value", "value", value, "error", err)
		}

		rows[0] = append(rows[0], btn)
	}

	menu.Inline(rows...)

	return menu
}
