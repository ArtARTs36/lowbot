package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const subsystemCommand = "command"

type Command struct {
	finished         *prometheus.CounterVec
	execution        *prometheus.HistogramVec
	actionExecution  *prometheus.HistogramVec
	stateTransitions *prometheus.CounterVec
	interruptions    *prometheus.CounterVec
	notFound         prometheus.Counter
	actionHandled    *prometheus.CounterVec
}

func newCommand() *Command {
	return &Command{
		finished: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystemCommand,
			Name:      "finished_total",
			Help:      "Count of finished commands",
		}, []string{"command"}),
		execution: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystemCommand,
			Name:      "execution_seconds",
			Help:      "Before taken to execute commands",
			Buckets:   []float64{1, 5, 15, 30, 60, 90, 120, 150, 180},
		}, []string{"command"}),
		actionExecution: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystemCommand,
			Name:      "action_execution_seconds",
			Help:      "Before taken to execute command actions",
			Buckets:   []float64{1, 5, 15, 30, 60},
		}, []string{"command", "action"}),
		stateTransitions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystemCommand,
			Name:      "state_transitions_total",
			Help:      "Count of transitions",
		}, []string{"command", "from_state", "to_state"}),
		interruptions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystemCommand,
			Name:      "interruptions_total",
			Help:      "Count of Command Interruptions",
		}, []string{"from_command", "from_state", "to_command", "allowed"}),
		notFound: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystemCommand,
			Name:      "not_found_total",
			Help:      "Count of Command Not Found",
		}),
		actionHandled: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystemCommand,
			Name:      "action_handled_total",
			Help:      "Count of Command Action Handled",
		}, []string{"command", "action", "code"}),
	}
}

func (g *Command) Describe(ch chan<- *prometheus.Desc) {
	g.finished.Describe(ch)
	g.execution.Describe(ch)
	g.actionExecution.Describe(ch)
	g.stateTransitions.Describe(ch)
	g.interruptions.Describe(ch)
	g.notFound.Describe(ch)
	g.actionHandled.Describe(ch)
}

func (g *Command) Collect(ch chan<- prometheus.Metric) {
	g.finished.Collect(ch)
	g.execution.Collect(ch)
	g.actionExecution.Collect(ch)
	g.stateTransitions.Collect(ch)
	g.interruptions.Collect(ch)
	g.notFound.Collect(ch)
	g.actionHandled.Collect(ch)
}

func (g *Command) IncFinished(command string) {
	g.finished.WithLabelValues(command).Inc()
}

func (g *Command) ObserveExecution(command string, execution time.Duration) {
	g.execution.WithLabelValues(command).Observe(float64(execution))
}

func (g *Command) ObserveActionExecution(command string, action string, execution time.Duration) {
	g.actionExecution.WithLabelValues(command, action).Observe(float64(execution))
}

func (g *Command) IncStateTransition(command, fromState, toState string) {
	g.stateTransitions.WithLabelValues(command, fromState, toState).Inc()
}

func (g *Command) IncInterruption(command, fromState, toCommand string, allowed bool) {
	g.interruptions.WithLabelValues(command, fromState, toCommand, strconv.FormatBool(allowed)).Inc()
}

func (g *Command) IncNotFound() {
	g.notFound.Inc()
}

func (g *Command) IncActionHandled(command, action string, code string) {
	g.actionHandled.WithLabelValues(command, action, code).Inc()
}
