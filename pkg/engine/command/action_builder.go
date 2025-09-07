package command

type ActionBuilder struct {
	stateName string
	actions   *Actions

	next *action
}

type actionBuild func(build func(callback ActionCallback) *ActionBuilder)

func newActionBuilder(
	stateName string,
	actions *Actions,
) *ActionBuilder {
	return &ActionBuilder{
		stateName: stateName,
		actions:   actions,
	}
}

func (b *ActionBuilder) do(fn ActionCallback) *ActionBuilder {
	act := &action{
		stateName: b.stateName,
		action:    fn,
	}

	b.actions.actionsMap[b.stateName] = act
	b.next = act

	return b
}

func (b *ActionBuilder) Then(stateName string, fn ActionCallback) *ActionBuilder {
	act := &action{
		stateName: stateName,
		action:    fn,
	}

	b.actions.actionsMap[stateName] = act
	b.next.next = act
	b.next = act

	return b
}
