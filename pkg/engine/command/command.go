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
	Interrupt(ctx context.Context, msg messenger.Message, currentState, newCmd string) (bool, error)
}

type AlwaysInterruptCommand struct{}

func (c *AlwaysInterruptCommand) Interrupt(_ context.Context, msg messenger.Message, _, _ string) (bool, error) {
	return true, nil
}
