package metrics

import "github.com/prometheus/client_golang/prometheus"

type Group struct {
	commandFinished *prometheus.CounterVec
}

func NewGroup(namespace string) *Group {
	return &Group{
		commandFinished: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "command_finished_total",
			Help:      "Count of finished commands",
		}, []string{"command"}),
	}
}

func (g *Group) Describe(ch chan<- *prometheus.Desc) {
	g.commandFinished.Describe(ch)
}

func (g *Group) Collect(ch chan<- prometheus.Metric) {
	g.commandFinished.Collect(ch)
}

func (g *Group) IncCommandFinished(command string) {
	g.commandFinished.WithLabelValues(command).Inc()
}
