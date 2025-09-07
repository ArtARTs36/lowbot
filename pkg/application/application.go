package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/artarts36/lowbot/pkg/engine/command"
	"github.com/artarts36/lowbot/pkg/engine/messenger"
	"github.com/artarts36/lowbot/pkg/engine/msghandler"
	"github.com/artarts36/lowbot/pkg/engine/router"
	"github.com/artarts36/lowbot/pkg/engine/state"
	"github.com/artarts36/lowbot/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	sloghttp "github.com/samber/slog-http"
)

type Application struct {
	router  router.Router
	handler *msghandler.Handler
	msngr   messenger.Messenger

	server *http.Server
}

func New(
	msngr messenger.Messenger,
	opts ...Option,
) (*Application, error) {
	cfg := &config{
		storage: state.NewMemoryStorage(),
		commandNotFoundFallback: func(router.Router) msghandler.CommandNotFoundFallback {
			return msghandler.ErrorCommandNotFoundFallback()
		},
		httpAddr:             ":8080",
		metricsHTTPAddr:      ":8081",
		router:               router.NewMapStaticRouter(),
		prometheusRegisterer: prometheus.DefaultRegisterer,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	metricsGroup := metrics.NewGroup("lowbot")

	if err := cfg.prometheusRegisterer.Register(metricsGroup); err != nil {
		return nil, fmt.Errorf("register metrics: %w", err)
	}

	app := &Application{
		router: cfg.router,
		msngr:  msngr,
	}

	app.handler = msghandler.NewHandler(
		app.router,
		cfg.storage,
		cfg.commandNotFoundFallback(app.router),
		metricsGroup,
	)

	app.prepareHTTPServer(cfg.httpAddr)

	return app, nil
}

func (app *Application) AddCommand(cmdName string, cmd command.Command) error {
	return app.router.Add(&router.NamedCommand{
		Name:    cmdName,
		Command: cmd,
	})
}

func (app *Application) MustAddCommand(cmdName string, cmd command.Command) {
	if err := app.AddCommand(cmdName, cmd); err != nil {
		panic(fmt.Sprintf("failed to add command %q: %v", cmdName, err))
	}
}

func (app *Application) Run() error {
	ch := make(chan messenger.Message)

	go func() {
		for msg := range ch {
			ctx := context.Background()

			slog.InfoContext(ctx, "[application] handling message")

			if err := app.handler.Handle(ctx, msg); err != nil {
				slog.ErrorContext(ctx, "[application] failed to handle message", slog.Any("err", err))
			}
		}
	}()

	go func() {
		if err := app.msngr.Listen(ch); err != nil {
			slog.Error("[application] failed to listen messenger", slog.Any("err", err))
		}

		close(ch)
	}()

	slog.Info("[application] listen http server", slog.String("addr", app.server.Addr))

	return app.server.ListenAndServe()
}

func (app *Application) Close() error {
	errs := make([]error, 0)

	if err := app.server.Close(); err != nil {
		errs = append(errs, err)
	}

	if err := app.msngr.Close(); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (app *Application) prepareHTTPServer(addr string) {
	const readTimeout = 30 * time.Second

	log := sloghttp.New(slog.Default())

	mux := http.NewServeMux()
	mux.Handle("/", log(app.msngr))

	app.server = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: readTimeout,
	}
}
