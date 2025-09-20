package callback

import "github.com/google/uuid"

type Value interface {
	value()
}

type Callback struct {
	ID    string `json:"id" db:"id"`
	Type  Type   `json:"type" db:"type"`
	Value Value  `json:"value" db:"value"`
}

func NewCallback(typ Type, value Value) *Callback {
	return &Callback{
		ID:    uuid.NewString(),
		Type:  typ,
		Value: value,
	}
}
