package messenger

import "net/http"

type Messenger interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request)
	Listen(ch chan Message) error
	Close() error
}
