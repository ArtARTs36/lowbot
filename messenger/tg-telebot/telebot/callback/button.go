package callback

type CommandButton struct {
	CommandName string
	StateName   string
	Data        map[string]string
}

func NewCommandButton(commandName, stateName string, data map[string]string) *Callback {
	return NewCallback(TypeCommandButton, &CommandButton{
		CommandName: commandName,
		StateName:   stateName,
		Data:        data,
	})
}

func (b CommandButton) value() {}
