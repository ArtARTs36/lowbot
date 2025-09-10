package messengerapi

type Message interface {
	GetID() string
	GetChatID() string
	GetBody() string

	ExtractCommandName() string

	Respond(answer *Answer) error

	// RespondObject responds media files.
	// See LocalImage
	RespondObject(file Object) error
}
