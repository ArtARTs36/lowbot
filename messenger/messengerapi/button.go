package messengerapi

type Button interface {
	GetTitle() string
}

type CommandButton struct {
	Title string

	Args Args
}

func (b *CommandButton) GetTitle() string {
	return b.Title
}
