package router

import "github.com/artarts36/lowbot/engine/command"

type MapStaticRouter struct {
	commands map[string]command.Command
}

func NewMapStaticRouter() *MapStaticRouter {
	r := &MapStaticRouter{
		commands: map[string]command.Command{},
	}

	return r
}

func (r *MapStaticRouter) Add(cmd command.Command) error {
	_, present := r.commands[cmd.Definition().Name]
	if present {
		return ErrCommandAlreadyExists
	}

	r.commands[cmd.Definition().Name] = cmd

	return nil
}

func (r *MapStaticRouter) Find(cmdName string) (command.Command, error) {
	cmd, ok := r.commands[cmdName]
	if !ok {
		return nil, ErrCommandNotFound
	}

	return cmd, nil
}

func (r *MapStaticRouter) List() []command.Command {
	result := make([]command.Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		result = append(result, cmd)
	}
	return result
}
