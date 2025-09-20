package messengerapi

type Message interface {
	GetID() string
	GetChatID() string
	GetBody() string

	// GetSender returns User that sent this message.
	GetSender() *Sender

	ExtractCommandName() string
	GetArgs() *Args
}
