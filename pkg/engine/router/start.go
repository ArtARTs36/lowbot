package router

import (
	"context"
	"fmt"
	"github.com/artarts36/lowbot/pkg/engine/command"
	"github.com/artarts36/lowbot/pkg/engine/messenger"
	"github.com/artarts36/lowbot/pkg/engine/state"
	"strings"
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
		func(ctx context.Context, message messenger.Message, state *state.State) error {
			text := make([]string, len(c.router.commands)-1)

			i := 0
			for _, cmd := range c.router.commands {
				if cmd.Name == c.name {
					continue
				}

				text[i] = fmt.Sprintf("/%s - %s", cmd.Name, cmd.Command.Description())
				i++
			}

			return message.Respond(&messenger.Answer{
				Text: strings.Join(text, "\n"),
			})
		},
	)
}
