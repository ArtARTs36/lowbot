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
	return m.ctx.Send(answer.Text)
}

func (m *message) ExtractCommandName() string {
	if strings.HasPrefix(m.msg.Text, "/") {
		return strings.TrimSpace(m.msg.Text[1:])
	}
	return ""
}
