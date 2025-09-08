package command

import (
	"context"

	"github.com/artarts36/lowbot/pkg/engine/messenger"
)

type Command interface {
	// Description returns a command description.
	// This result may be used in /start command.
	Description() string

	// Actions returns Actions with states.
	Actions() *Actions

	// Interrupt defines interruption is allowed.
	// Returns true, when interrupt allowed
	Interrupt(ctx context.Context, req *InterruptRequest) (bool, error)
}

type InterruptRequest struct {
	Message      messenger.Message
	CurrentState string
	NewCommand   string
}

type AlwaysInterruptCommand struct{}

func (c *AlwaysInterruptCommand) Interrupt(context.Context, *InterruptRequest) (bool, error) {
	return true, nil
}
