package messenger

type Message interface {
	GetID() string
	GetChatID() string
	GetBody() string

	ExtractCommandName() string

	RespondText(answer string) error
}
