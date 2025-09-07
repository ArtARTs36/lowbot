package command

import (
	"context"
	"github.com/artarts36/lowbot/pkg/engine/messenger"
	"github.com/artarts36/lowbot/pkg/engine/state"
)

type Actions struct {
	actions    []*action
	actionsMap map[string]*action
}

func NewActions() *Actions {
	return &Actions{
		actions:    []*action{},
		actionsMap: make(map[string]*action),
	}
}

func (a *Actions) Then(
	stateName string,
	fn func(ctx context.Context, message messenger.Message, state *state.State) error,
) *Actions {
	act := &action{
		stateName: stateName,
		action:    fn,
	}

	if len(a.actions) > 0 {
		a.actions[len(a.actions)-1].next = act
	}

	a.actions = append(a.actions, act)
	a.actionsMap[stateName] = act

	return a
}

func (a *Actions) Get(stateName string) (Action, bool) {
	if act, ok := a.actionsMap[stateName]; ok {
		return act, true
	}

	return nil, false
}

func (a *Actions) First() Action {
	return a.actions[0]
}
