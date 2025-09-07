package router

type MapStaticRouter struct {
	commands map[string]*NamedCommand
}

func NewMapStaticRouter() *MapStaticRouter {
	r := &MapStaticRouter{
		commands: map[string]*NamedCommand{},
	}

	_ = r.Add(&NamedCommand{
		Name:    "start",
		Command: newStartCommand("start", r),
	})

	return r
}

func (r *MapStaticRouter) Add(cmd *NamedCommand) error {
	_, present := r.commands[cmd.Name]
	if present {
		return ErrCommandAlreadyExists
	}

	r.commands[cmd.Name] = cmd

	return nil
}

func (r *MapStaticRouter) Find(cmdName string) (*NamedCommand, error) {
	cmd, ok := r.commands[cmdName]
	if !ok {
		return nil, ErrCommandNotFound
	}

	return cmd, nil
}
