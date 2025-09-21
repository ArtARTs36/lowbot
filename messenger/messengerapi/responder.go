package messengerapi

type Responder interface {
	// Respond responds text and buttons.
	// See Answer.
	Respond(answer *Answer) (Message, error)

	// RespondObject responds media files.
	// See LocalImage
	RespondObject(file Object) (Message, error)
}
