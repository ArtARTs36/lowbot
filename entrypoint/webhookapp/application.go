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
}

func New(
	msngr messengerapi.Messenger,
	opts ...Option,
) (*Application, error) {
	cfg := &config{
		storage: state.NewMemoryStorage(),
		commandNotFoundFallback: func(router.Router) machine.CommandNotFoundFallback {
			return machine.ErrorCommandNotFoundFallback()
		},
		httpAddr:             ":8080",
		router:               router.NewMapStaticRouter(),
		prometheusRegisterer: prometheus.DefaultRegisterer,
		middlewares:          make([]command.Middleware, 0),
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

	app.machine = machine.New(
		app.router,
		cfg.storage,
		machine.NewErrorHandler(),
		cfg.commandNotFoundFallback(app.router),
		metricsGroup,
		command.NewBus(cfg.middlewares),
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
	ch := make(chan messengerapi.Message)

	go func() {
		for msg := range ch {
			ctx := context.Background()

			if err := app.machine.Handle(ctx, msg); err != nil {
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
