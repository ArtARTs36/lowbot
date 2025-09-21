package router

import (
	"context"
	"fmt"
	"strings"

	"github.com/artarts36/lowbot/engine/command"
	"github.com/artarts36/lowbot/messenger/messengerapi"
)

type StartCommand struct {
	command.AlwaysInterruptCommand

	name   string
	router Router
}

func NewStartCommand(name string, router Router) command.Command {
	return &StartCommand{
		name:   name,
		router: router,
	}
}

func (c *StartCommand) Definition() *command.Definition {
	return &command.Definition{
		Name:        c.name,
		Description: "",
	}
}

func (c *StartCommand) Actions() *command.Actions {
	return command.NewActions().Then(
		"start",
		func(_ context.Context, req *command.Request) error {
			text := make([]string, len(c.router.List())-1)

			i := 0
			for _, cmd := range c.router.List() {
				cmdDefinition := cmd.Definition()

				if cmdDefinition.Name == c.name {
					continue
				}

				text[i] = fmt.Sprintf("/%s - %s", cmdDefinition.Name, cmdDefinition.Description)
				i++
			}

			_, err := req.Responder.Respond(&messengerapi.Answer{
				Text: strings.Join(text, "\n"),
			})
			return err
		},
	)
}
