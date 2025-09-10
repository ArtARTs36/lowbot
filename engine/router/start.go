package router

import (
	"context"
	"fmt"
	"strings"

	"github.com/artarts36/lowbot/engine/command"
	"github.com/artarts36/lowbot/messenger/messengerapi"
)

type startCommand struct {
	command.AlwaysInterruptCommand

	name   string
	router *MapStaticRouter
}

func newStartCommand(name string, router *MapStaticRouter) command.Command {
	return &startCommand{
		name:   name,
		router: router,
	}
}

func (c *startCommand) Description() string {
	return ""
}

func (c *startCommand) Actions() *command.Actions {
	return command.NewActions().Then(
		"start",
		func(_ context.Context, req *command.Request) error {
			text := make([]string, len(c.router.commands)-1)

			i := 0
			for _, cmd := range c.router.commands {
				if cmd.Name == c.name {
					continue
				}

				text[i] = fmt.Sprintf("/%s - %s", cmd.Name, cmd.Command.Description())
				i++
			}

			return req.Message.Respond(&messengerapi.Answer{
				Text: strings.Join(text, "\n"),
			})
		},
	)
}
