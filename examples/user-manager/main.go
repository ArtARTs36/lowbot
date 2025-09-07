package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/artarts36/lowbot/pkg/application"
	"github.com/artarts36/lowbot/pkg/engine/command"
	"github.com/artarts36/lowbot/pkg/engine/messenger"
	"github.com/artarts36/lowbot/pkg/engine/state"
	"github.com/artarts36/lowbot/pkg/logx"
	"github.com/artarts36/lowbot/pkg/messengers/telebot"
	"github.com/cappuccinotm/slogx"
)

const readHTTPTimeout = 30 * time.Second

func main() {
	slog.SetDefault(slog.New(slogx.Accumulator(slogx.NewChain(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
		logx.PropagateMessageID(),
		logx.PropagateChatID(),
		logx.PropagateCommandName(),
	))))

	msgr, err := createMessenger()
	if err != nil {
		panic(err)
	}

	app, err := application.New(
		msgr,
		application.WithCommandSuggestion(),
		application.WithHTTPAddr(":9005"),
	)
	if err != nil {
		slog.Error("failed to create application", slog.Any("err", err))
		os.Exit(1)
	}

	app.MustAddCommand("add", &addUserCommand{})

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		server := http.Server{
			Addr:              ":8081",
			Handler:           mux,
			ReadHeaderTimeout: readHTTPTimeout,
		}

		if merr := server.ListenAndServe(); merr != nil {
			slog.Error("failed to start metrics server", slog.Any("err", merr))
		}
	}()

	if err = app.Run(); err != nil {
		slog.Error("failed to run application", slog.Any("err", err))
		os.Exit(1)
	}
}

func createMessenger() (messenger.Messenger, error) {
	return telebot.NewWebhookMessenger(telebot.WebhookConfig{
		Token: os.Getenv("TELEGRAM_TOKEN"),
	})
}

type addUserCommand struct {
	command.AlwaysInterruptCommand
}

func (addUserCommand) Description() string { return "addUser" }
func (addUserCommand) Actions() *command.Actions {
	return command.NewActions().
		Then("start", func(_ context.Context, message messenger.Message, _ *state.State) error {
			err := message.RespondObject(&messenger.LocalImage{
				Path: "./examples/user-manager/image.jpg",
			})
			if err != nil {
				return fmt.Errorf("send image: %w", err)
			}

			return message.Respond(&messenger.Answer{
				Text: "Enter user name",
			})
		}).
		Then("name", func(_ context.Context, message messenger.Message, state *state.State) error {
			state.Set("user.name", message.GetBody())

			return message.Respond(&messenger.Answer{
				Text: "Enter email",
			})
		}).
		Then("email", func(_ context.Context, message messenger.Message, state *state.State) error {
			if !strings.Contains(message.GetBody(), "@") {
				return command.NewValidationError("invalid email")
			}

			state.Set("user.email", message.GetBody())

			return message.Respond(&messenger.Answer{
				Text: "Select user type",
				Enum: messenger.Enum{
					Values: map[string]string{
						"int": "internal",
						"ext": "external",
					},
				},
			})
		}).
		Then("type", func(_ context.Context, message messenger.Message, state *state.State) error {
			state.Set("user.type", message.GetBody())

			return message.Respond(&messenger.Answer{
				Text: fmt.Sprintf(
					"name: %s, email: %s, type: %s",
					state.Get("user.name"),
					state.Get("user.email"),
					state.Get("user.type"),
				),
			})
		})
}
