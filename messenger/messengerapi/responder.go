package messengerapi

type Responder interface {
	// Respond responds text and buttons.
	// See Answer.
	Respond(answer *Answer) error

	// RespondObject responds media files.
	// See LocalImage
	RespondObject(file Object) error
}
