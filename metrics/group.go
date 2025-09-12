package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "lowbot"

type Group struct {
	commandFinished         *prometheus.CounterVec
	commandExecution        *prometheus.HistogramVec
	commandActionExecution  *prometheus.HistogramVec
	commandStateTransitions *prometheus.CounterVec
	commandInterruptions    *prometheus.CounterVec
	commandNotFound         prometheus.Counter
	commandActionHandled    *prometheus.CounterVec
}

func NewGroup() *Group {
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
		commandActionExecution: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "command_action_execution_seconds",
			Help:      "Time taken to execute command actions",
			Buckets:   []float64{1, 5, 15, 30, 60},
		}, []string{"command", "action"}),
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
		commandNotFound: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "command_not_found_total",
			Help:      "Count of Command Not Found",
		}),
		commandActionHandled: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "command_action_handled_total",
			Help:      "Count of Command Action Handled",
		}, []string{"command", "action", "code"}),
	}
}

func (g *Group) Describe(ch chan<- *prometheus.Desc) {
	g.commandFinished.Describe(ch)
	g.commandExecution.Describe(ch)
	g.commandActionExecution.Describe(ch)
	g.commandStateTransitions.Describe(ch)
	g.commandInterruptions.Describe(ch)
	g.commandNotFound.Describe(ch)
	g.commandActionHandled.Describe(ch)
}

func (g *Group) Collect(ch chan<- prometheus.Metric) {
	g.commandFinished.Collect(ch)
	g.commandExecution.Collect(ch)
	g.commandActionExecution.Collect(ch)
	g.commandStateTransitions.Collect(ch)
	g.commandInterruptions.Collect(ch)
	g.commandNotFound.Collect(ch)
	g.commandActionHandled.Collect(ch)
}

func (g *Group) IncCommandFinished(command string) {
	g.commandFinished.WithLabelValues(command).Inc()
}

func (g *Group) ObserveCommandExecution(command string, execution time.Duration) {
	g.commandExecution.WithLabelValues(command).Observe(float64(execution))
}

func (g *Group) ObserveCommandActionExecution(command string, action string, execution time.Duration) {
	g.commandActionExecution.WithLabelValues(command, action).Observe(float64(execution))
}

func (g *Group) IncCommandStateTransition(command, fromState, toState string) {
	g.commandStateTransitions.WithLabelValues(command, fromState, toState).Inc()
}

func (g *Group) IncCommandInterruption(command, fromState, toCommand string, allowed bool) {
	g.commandInterruptions.WithLabelValues(command, fromState, toCommand, strconv.FormatBool(allowed)).Inc()
}

func (g *Group) IncCommandNotFound() {
	g.commandNotFound.Inc()
}

func (g *Group) IncCommandActionHandled(command, action string, code string) {
	g.commandActionHandled.WithLabelValues(command, action, code).Inc()
}
