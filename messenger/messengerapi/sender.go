package messengerapi

type Sender struct {
	ID string `json:"id"`

	Username string `json:"username"`

	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
