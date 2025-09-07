package command

import (
	"context"

	"github.com/artarts36/lowbot/pkg/engine/messenger"
	"github.com/artarts36/lowbot/pkg/engine/state"
)

type Action interface {
	State() string
	Next() Action
	Run(ctx context.Context, message messenger.Message, state *state.State) error
}

type action struct {
	stateName string
	next      Action
	action    func(ctx context.Context, message messenger.Message, state *state.State) error
}

func (a *action) State() string {
	return a.stateName
}

func (a *action) Next() Action {
	return a.next
}

func (a *action) Run(ctx context.Context, message messenger.Message, state *state.State) error {
	return a.action(ctx, message, state)
}
