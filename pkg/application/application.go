package application

import (
	"context"
	"errors"
	"github.com/artarts36/lowbot/pkg/engine/command"
	messenger2 "github.com/artarts36/lowbot/pkg/engine/messenger"
	"github.com/artarts36/lowbot/pkg/engine/msghandler"
	router2 "github.com/artarts36/lowbot/pkg/engine/router"
	"github.com/artarts36/lowbot/pkg/engine/state"
	sloghttp "github.com/samber/slog-http"
	"log/slog"
	"net/http"
)

type Application struct {
	router  router2.Router
	handler *msghandler.Handler
	msngr   messenger2.Messenger
	server  *http.Server
}

func New(
	msngr messenger2.Messenger,
	stateStorage state.Storage,
) *Application {
	app := &Application{
		router: router2.NewMapStaticRouter(),
		msngr:  msngr,
	}

	app.handler = msghandler.NewHandler(app.router, stateStorage, func(ctx context.Context, message messenger2.Message) error {
		return message.RespondText("Команда не найдена")
	})

	app.prepareHTTPServer()

	return app
}

func (app *Application) AddCommand(cmdName string, cmd command.Command) error {
	return app.router.Add(&router2.NamedCommand{
		Name:    cmdName,
		Command: cmd,
	})
}

func (app *Application) Run() error {
	ch := make(chan messenger2.Message)

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
