package command

type Nop struct{}

func (c Nop) Actions() Actions {
	return Actions{}
}

func (c Nop) Description() string {
	return "nop"
}
