package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Group struct {
	commandFinished         *prometheus.CounterVec
	commandExecution        *prometheus.HistogramVec
	commandStateTransitions *prometheus.CounterVec
	commandInterruptions    *prometheus.CounterVec
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
		commandStateTransitions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "command_state_transitions_total",
			Help:      "Count of transitions",
		}, []string{"command", "from_state", "to_state"}),
		commandInterruptions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "command_interruptions_total",
			Help:      "Count of Command Interruptions",
		}, []string{"from_command", "from_state", "to_command", "allowed"}),
	}
}

func (g *Group) Describe(ch chan<- *prometheus.Desc) {
	g.commandFinished.Describe(ch)
	g.commandExecution.Describe(ch)
	g.commandStateTransitions.Describe(ch)
	g.commandInterruptions.Describe(ch)
}

func (g *Group) Collect(ch chan<- prometheus.Metric) {
	g.commandFinished.Collect(ch)
	g.commandExecution.Collect(ch)
	g.commandStateTransitions.Collect(ch)
	g.commandInterruptions.Collect(ch)
}

func (g *Group) IncCommandFinished(command string) {
	g.commandFinished.WithLabelValues(command).Inc()
}

func (g *Group) ObserveCommandExecution(command string, execution time.Duration) {
	g.commandExecution.WithLabelValues(command).Observe(float64(execution))
}

func (g *Group) IncCommandStateTransition(command, fromState, toState string) {
	g.commandStateTransitions.WithLabelValues(command, fromState, toState).Inc()
}

func (g *Group) IncCommandInterruption(command, fromState, toCommand string, allowed bool) {
	g.commandInterruptions.WithLabelValues(command, fromState, toCommand, strconv.FormatBool(allowed)).Inc()
}
