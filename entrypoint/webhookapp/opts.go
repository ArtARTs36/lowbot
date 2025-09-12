package webhookapp

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/artarts36/lowbot/engine/command"
	"github.com/artarts36/lowbot/engine/machine"
	"github.com/artarts36/lowbot/engine/router"
	"github.com/artarts36/lowbot/engine/state"
)

type config struct {
	storage                 state.Storage
	commandNotFoundFallback func(router router.Router) machine.CommandNotFoundFallback
	httpAddr                string
	router                  router.Router
	prometheusRegisterer    prometheus.Registerer
	middlewares             []command.Middleware
	startCommandFn          func(router.Router) command.Command
}

type Option func(*config)

func WithStateStorage(storage state.Storage) Option {
	return func(c *config) {
		c.storage = storage
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
