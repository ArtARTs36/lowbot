package telebot

import (
	"strings"

	"github.com/artarts36/lowbot/messenger/messengerapi"
	"gopkg.in/telebot.v4"
)

type message struct {
	ctx telebot.Context

	id     string
	chatID string
	text   string
	sender *messengerapi.Sender

	args *messengerapi.Args
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

func (m *message) ExtractCommandName() string {
	if strings.HasPrefix(m.text, "/") {
		return strings.TrimSpace(m.text[1:])
	}
	return ""
}

func (m *message) GetArgs() *messengerapi.Args {
	return m.args
}
