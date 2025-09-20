package command

import (
	"context"

	"github.com/artarts36/lowbot/engine/state"
	"github.com/artarts36/lowbot/messenger/messengerapi"
)

type Request struct {
	Message   messengerapi.Message
	Responder messengerapi.Responder
	State     *state.State
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
