package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "lowbot"

type Group struct {
	command      *Command
	stateStorage *StateStorage
}

func NewGroup() *Group {
	return &Group{
		command:      newCommand(),
		stateStorage: NewStateStorage(),
	}
}

func (g *Group) Describe(ch chan<- *prometheus.Desc) {
	g.command.Describe(ch)
	g.stateStorage.Describe(ch)
}

func (g *Group) Collect(ch chan<- prometheus.Metric) {
	g.command.Collect(ch)
	g.stateStorage.Collect(ch)
}

func (g *Group) Command() *Command {
	return g.command
}

func (g *Group) StateStorage() *StateStorage {
	return g.stateStorage
}
