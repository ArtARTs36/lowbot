package telebot

import (
	"github.com/artarts36/lowbot/pkg/engine/messenger"
	"gopkg.in/telebot.v4"
	"strconv"
	"strings"
)

type message struct {
	msg *telebot.Message
	ctx telebot.Context

	id     string
	chatID string
}

func newMessage(msg *telebot.Message, ctx telebot.Context) *message {
	return &message{
		msg:    msg,
		ctx:    ctx,
		id:     strconv.FormatInt(msg.Chat.ID, 10),
		chatID: strconv.FormatInt(msg.Chat.ID, 10),
	}
}

func (m *message) GetID() string {
	return m.id
}

func (m *message) GetChatID() string {
	return m.chatID
}

func (m *message) GetBody() string {
	return m.msg.Text
}

func (m *message) Respond(answer *messenger.Answer) error {
	var what interface{}
	var opts []interface{}

	if answer.Text != "" {
		what = answer.Text
	}

	if len(answer.Enum) > 0 {
		opts = append(opts, answer.Enum)
	}

	return m.ctx.Send(what, opts...)
}

func (m *message) ExtractCommandName() string {
	if strings.HasPrefix(m.msg.Text, "/") {
		return strings.TrimSpace(m.msg.Text[1:])
	}
	return ""
}

func (m *message) buildEnumOpt(enum []string) *telebot.ReplyMarkup {
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
