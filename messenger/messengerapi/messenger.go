package messengerapi

import "net/http"

type Messenger interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request)
	Listen(ch chan Message) error
	Close() error

	// CreateResponder creates Responder.
	CreateResponder(chatID string) Responder
}
