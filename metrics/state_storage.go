package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const subsystemStateStorage = "state_storage"

type StateStorage struct {
	operationExecution *prometheus.HistogramVec
}

func NewStateStorage() *StateStorage {
	return &StateStorage{
		operationExecution: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystemStateStorage,
			Name:      "operation_execution_milliseconds",
		}, []string{"operation"}),
	}
}

func (s *StateStorage) ObserveOperationExecution(operation string, dur time.Duration) {
	s.operationExecution.WithLabelValues(operation).Observe(float64(dur))
}

func (s *StateStorage) Describe(ch chan<- *prometheus.Desc) {
	s.operationExecution.Describe(ch)
}

func (s *StateStorage) Collect(ch chan<- prometheus.Metric) {
	s.operationExecution.Collect(ch)
}
