package application

import (
	"context"
	"errors"
	"github.com/artarts36/lowbot/pkg/engine/command"
	"github.com/artarts36/lowbot/pkg/engine/messenger"
	"github.com/artarts36/lowbot/pkg/engine/msghandler"
	"github.com/artarts36/lowbot/pkg/engine/router"
	"github.com/artarts36/lowbot/pkg/engine/state"
	sloghttp "github.com/samber/slog-http"
	"log/slog"
	"net/http"
)

type Application struct {
	router  router.Router
	handler *msghandler.Handler
	msngr   messenger.Messenger
	server  *http.Server
}

func New(
	msngr messenger.Messenger,
	stateStorage state.Storage,
) *Application {
	app := &Application{
		router: router.NewMapStaticRouter(),
		msngr:  msngr,
	}

	app.handler = msghandler.NewHandler(app.router, stateStorage, msghandler.DefaultCommandNotFoundFallback())

	app.prepareHTTPServer()

	return app
}

func (app *Application) AddCommand(cmdName string, cmd command.Command) error {
	return app.router.Add(&router.NamedCommand{
		Name:    cmdName,
		Command: cmd,
	})
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

func (app *Application) prepareHTTPServer() {
	log := sloghttp.New(slog.Default())

	mux := http.NewServeMux()
	mux.Handle("/", log(app.msngr))

	app.server = &http.Server{
		Addr:    "localhost:9005",
		Handler: mux,
	}
}
