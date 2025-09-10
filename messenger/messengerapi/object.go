package messengerapi

type Object interface {
	object()
}

type LocalImage struct {
	Path string
}

func (i LocalImage) object() {}
