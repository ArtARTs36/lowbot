package webhookapp

import (
	"github.com/artarts36/lowbot/logx"
	"github.com/artarts36/lowbot/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/artarts36/lowbot/engine/command"
	"github.com/artarts36/lowbot/engine/machine"
	"github.com/artarts36/lowbot/engine/router"
	"github.com/artarts36/lowbot/engine/state"
)

type config struct {
	storageFn               func(storage *metrics.StateStorage) state.Storage
	commandNotFoundFallback func(router router.Router) machine.CommandNotFoundFallback
	httpAddr                string
	router                  router.Router
	prometheusRegisterer    prometheus.Registerer
	middlewares             []command.Middleware
	startCommandFn          func(router.Router) command.Command
	logger                  logx.Logger
}

type Option func(*config)

func WithStateStorage(storage state.Storage) Option {
	return func(c *config) {
		c.storageFn = func(metrics *metrics.StateStorage) state.Storage {
			return state.NewObservableStorage(storage, metrics)
		}
	}
}

// WithPriorityStateStorage creates Storage which separate load for priority and non-priority commands.
// priorityStorage - typically, is slow storage, e.g. DatabaseStorage
// fallbackStorage - typically, is fast storage, e.g. MemoryStorage.
func WithPriorityStateStorage(
	priorityCommands []string,
	priorityStorage state.Storage,
	fallbackStorage state.Storage,
) Option {
	return func(c *config) {
		c.storageFn = func(metrics *metrics.StateStorage) state.Storage {
			return state.NewPriorityStorage(
				priorityCommands,
				state.NewObservableStorage(priorityStorage, metrics),
				state.NewObservableStorage(fallbackStorage, metrics),
			)
		}
	}
}

func WithCommandNotFoundFallback(fallbackFactory func(router router.Router) machine.CommandNotFoundFallback) Option {
	return func(c *config) {
		c.commandNotFoundFallback = fallbackFactory
	}
}

func WithCommandSuggestion() Option {
	return WithCommandNotFoundFallback(machine.SuggestCommandNotFoundFallback)
}

func WithErrorCommandNotFoundFallback() Option {
	return WithCommandNotFoundFallback(func(_ router.Router) machine.CommandNotFoundFallback {
		return machine.ErrorCommandNotFoundFallback()
	})
}

func WithHTTPAddr(addr string) Option {
	return func(c *config) {
		c.httpAddr = addr
	}
}

func WithRouter(router router.Router) Option {
	return func(c *config) {
		c.router = router
	}
}

func WithPrometheus(registerer prometheus.Registerer) Option {
	return func(c *config) {
		c.prometheusRegisterer = registerer
	}
}

func WithMiddleware(middleware ...command.Middleware) Option {
	return func(c *config) {
		c.middlewares = append(c.middlewares, middleware...)
	}
}

func WithStartCommand(factory func(router router.Router) command.Command) Option {
	return func(c *config) {
		c.startCommandFn = factory
	}
}

func WithLogger(logger logx.Logger) Option {
	return func(c *config) {
		c.logger = logger
	}
}
