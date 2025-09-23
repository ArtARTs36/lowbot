package messengerapi

import "io"

type Object interface {
	object()
}

type LocalImage struct {
	Reader io.Reader
}

func (i LocalImage) object() {}
