package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Group struct {
	commandFinished  *prometheus.CounterVec
	commandExecution *prometheus.HistogramVec
}

func NewGroup(namespace string) *Group {
	return &Group{
		commandFinished: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "command_finished_total",
			Help:      "Count of finished commands",
		}, []string{"command"}),
		commandExecution: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "command_execution_seconds",
			Help:      "Time taken to execute commands",
			Buckets:   []float64{1, 5, 15, 30, 60, 90, 120, 150, 180},
		}, []string{"command"}),
	}
}

func (g *Group) Describe(ch chan<- *prometheus.Desc) {
	g.commandFinished.Describe(ch)
	g.commandExecution.Describe(ch)
}

func (g *Group) Collect(ch chan<- prometheus.Metric) {
	g.commandFinished.Collect(ch)
	g.commandExecution.Collect(ch)
}

func (g *Group) IncCommandFinished(command string) {
	g.commandFinished.WithLabelValues(command).Inc()
}

func (g *Group) ObserveCommandExecution(command string, execution time.Duration) {
	g.commandExecution.WithLabelValues(command).Observe(float64(execution))
}
