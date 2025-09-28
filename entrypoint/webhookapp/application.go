package webhookapp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/artarts36/lowbot/engine/machine"
	"github.com/artarts36/lowbot/messenger/messengerapi"

	"github.com/artarts36/lowbot/logx"

	"github.com/artarts36/lowbot/engine/command"
	"github.com/artarts36/lowbot/engine/router"
	"github.com/artarts36/lowbot/engine/state"
	"github.com/artarts36/lowbot/metrics"
	"github.com/prometheus/client_golang/prometheus"
	sloghttp "github.com/samber/slog-http"
)

type Application struct {
	router  router.Router
	machine *machine.Machine
	msngr   messengerapi.Messenger

	server *http.Server
	logger logx.Logger
}

func New(
	msngr messengerapi.Messenger,
	opts ...Option,
) (*Application, error) {
	cfg := &config{
		storageFn: func(metrics *metrics.StateStorage) state.Storage {
			return state.NewObservableStorage(state.NewMemoryStorage(), metrics)
		},
		commandNotFoundFallback: func(router.Router) machine.CommandNotFoundFallback {
			return machine.ErrorCommandNotFoundFallback()
		},
		httpAddr:             ":8080",
		router:               router.NewMapStaticRouter(),
		prometheusRegisterer: prometheus.DefaultRegisterer,
		middlewares:          make([]command.Middleware, 0),
		startCommandFn: func(r router.Router) command.Command {
			return router.NewStartCommand("start", r)
		},
		logger: slog.Default(),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	metricsGroup := metrics.NewGroup()

	if err := cfg.prometheusRegisterer.Register(metricsGroup); err != nil {
		return nil, fmt.Errorf("register metrics: %w", err)
	}

	app := &Application{
		router: cfg.router,
		msngr:  msngr,
		logger: cfg.logger,
	}

	err := app.router.Add(cfg.startCommandFn(app.router))
	if err != nil {
		return nil, fmt.Errorf("register start command: %w", err)
	}

	app.machine = machine.New(
		app.router,
		cfg.storageFn(metricsGroup.StateStorage()),
		machine.NewErrorHandler(cfg.logger),
		cfg.commandNotFoundFallback(app.router),
		metricsGroup,
		command.NewBus(cfg.middlewares),
		cfg.logger,
	)

	app.prepareHTTPServer(cfg.httpAddr)

	return app, nil
}

func (app *Application) AddCommand(cmd command.Command) error {
	return app.router.Add(cmd)
}

func (app *Application) MustAddCommand(cmd command.Command) {
	if err := app.AddCommand(cmd); err != nil {
		panic(fmt.Sprintf("failed to add command %q: %v", cmd.Definition().Name, err))
	}
}

func (app *Application) Run() error {
	ch := make(chan messengerapi.Message)

	go func() {
		for msg := range ch {
			ctx := context.Background()

			if err := app.machine.Handle(ctx, &machine.Request{
				Message:   msg,
				Responder: app.msngr.CreateResponder(msg.GetChatID()),
			}); err != nil {
				slog.ErrorContext(ctx,
					"[application] failed to handle message",
					logx.Err(err),
					slog.String("message.id", msg.GetID()),
					slog.String("message.chat_id", msg.GetChatID()),
				)
			}
		}
	}()

	go func() {
		if err := app.msngr.Listen(ch); err != nil {
			slog.Error("[application] failed to listen messenger", slog.Any("err", err))
		}

		close(ch)
	}()

	app.logger.InfoContext(context.Background(), "[application] listen http server", slog.String("addr", app.server.Addr))

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
