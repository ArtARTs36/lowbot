package application

import (
	"github.com/artarts36/lowbot/pkg/engine/msghandler"
	"github.com/artarts36/lowbot/pkg/engine/router"
	"github.com/artarts36/lowbot/pkg/engine/state"
	"github.com/prometheus/client_golang/prometheus"
)

type config struct {
	storage                 state.Storage
	commandNotFoundFallback func(router router.Router) msghandler.CommandNotFoundFallback
	httpAddr                string
	metricsHTTPAddr         string
	router                  router.Router
	prometheusRegisterer    prometheus.Registerer
}

type Option func(*config)

func WithStateStorage(storage state.Storage) Option {
	return func(c *config) {
		c.storage = storage
	}
}

func WithCommandNotFoundFallback(fallback msghandler.CommandNotFoundFallback) Option {
	return func(c *config) {
		c.commandNotFoundFallback = func(_ router.Router) msghandler.CommandNotFoundFallback {
			return fallback
		}
	}
}

func WithCommandSuggestion() Option {
	return func(c *config) {
		c.commandNotFoundFallback = msghandler.SuggestCommandNotFoundFallback
	}
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

func WithMetricsHTTPAddr(addr string) Option {
	return func(c *config) {
		c.metricsHTTPAddr = addr
	}
}
