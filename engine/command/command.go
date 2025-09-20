package command

import (
	"context"

	"github.com/artarts36/lowbot/messenger/messengerapi"
)

type Command interface {
	// Definition returns command definition with command name and description.
	Definition() *Definition

	// Actions returns Actions with states.
	Actions() *Actions

	// Interrupt defines interruption is allowed.
	// Returns true, when interrupt allowed
	Interrupt(ctx context.Context, req *InterruptRequest) (bool, error)
}

type InterruptRequest struct {
	Message      messengerapi.Message
	CurrentState string
	NewCommand   string
}

type AlwaysInterruptCommand struct{}

func (c *AlwaysInterruptCommand) Interrupt(context.Context, *InterruptRequest) (bool, error) {
	return true, nil
}
