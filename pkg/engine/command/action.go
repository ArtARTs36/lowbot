package command

import (
	"context"

	"github.com/artarts36/lowbot/pkg/engine/messenger"
	"github.com/artarts36/lowbot/pkg/engine/state"
)

type Request struct {
	Message messenger.Message
	State   *state.State
}

type Action interface {
	State() string
	Next() Action
	Run(ctx context.Context, req *Request) error
}

type action struct {
	stateName string
	next      Action
	action    func(ctx context.Context, req *Request) error
}

func (a *action) State() string {
	return a.stateName
}

func (a *action) Next() Action {
	return a.next
}

func (a *action) Run(ctx context.Context, req *Request) error {
	return a.action(ctx, req)
}
