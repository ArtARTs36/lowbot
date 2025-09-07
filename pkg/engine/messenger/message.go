package messenger

type Message interface {
	GetID() string
	GetChatID() string
	GetBody() string

	ExtractCommandName() string

	Respond(answer *Answer) error
}

type Answer struct {
	Text string
}
