package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/artarts36/lowbot/entrypoint/webhookapp"

	"github.com/artarts36/lowbot/messenger/tg-telebot/telebot"

	"github.com/artarts36/lowbot/middleware"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/artarts36/lowbot/engine/command"
	"github.com/artarts36/lowbot/logx"
	"github.com/artarts36/lowbot/messenger/messengerapi"
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

	app, err := webhookapp.New(
		msgr,
		webhookapp.WithCommandSuggestion(),
		webhookapp.WithHTTPAddr(":9005"),
		webhookapp.WithMiddleware(
			func(ctx context.Context, req *command.Request, next command.ActionCallback) error {
				slog.InfoContext(ctx, "[main] handling request", slog.Any("req", req))
				return next(ctx, req)
			},
			middleware.OnlyChatsWithMessage([]string{"493731328"}, "denied"),
		),
	)
	if err != nil {
		slog.Error("failed to create application", slog.Any("err", err))
		os.Exit(1)
	}

	app.MustAddCommand("add", &addUserCommand{})
	app.MustAddCommand("delete", &deleteUserCommand{})
	app.MustAddCommand("update", &updateUserCommand{})

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

func createMessenger() (messengerapi.Messenger, error) {
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
		Then("start", func(_ context.Context, req *command.Request) error {
			err := req.Message.RespondObject(&messengerapi.LocalImage{
				Path: "./examples/user-manager/image.jpg",
			})
			if err != nil {
				return fmt.Errorf("send image: %w", err)
			}

			return req.Message.Respond(&messengerapi.Answer{
				Text: "Enter user name",
			})
		}).
		Then("name", func(_ context.Context, req *command.Request) error {
			req.State.Set("user.name", req.Message.GetBody())

			return req.Message.Respond(&messengerapi.Answer{
				Text: "Enter email",
			})
		}).
		Then("email", func(_ context.Context, req *command.Request) error {
			if !strings.Contains(req.Message.GetBody(), "@") {
				return command.NewValidationError("invalid email")
			}

			req.State.Set("user.email", req.Message.GetBody())

			return req.Message.Respond(&messengerapi.Answer{
				Text: "Select user type",
				Enum: messengerapi.Enum{
					Values: map[string]string{
						"int": "internal",
						"ext": "external",
					},
				},
			})
		}).
		Then("type", func(_ context.Context, req *command.Request) error {
			req.State.Set("user.type", req.Message.GetBody())

			return req.Message.Respond(&messengerapi.Answer{
				Text: fmt.Sprintf(
					"name: %s, email: %s, type: %s",
					req.State.Get("user.name"),
					req.State.Get("user.email"),
					req.State.Get("user.type"),
				),
			})
		})
}

type deleteUserCommand struct {
	command.AlwaysInterruptCommand
}

func (deleteUserCommand) Description() string { return "deleteUser" }

func (deleteUserCommand) Actions() *command.Actions {
	return command.NewActions().
		With("confirmed", func(build func(callback command.ActionCallback) *command.ActionBuilder) {
			build(
				func(_ context.Context, req *command.Request) error {
					return req.Message.Respond(&messengerapi.Answer{
						Text: "User deleted",
					})
				}).
				Then("after_deletion", func(_ context.Context, req *command.Request) error {
					return req.Message.Respond(&messengerapi.Answer{
						Text: "12345678",
					})
				})
		}).
		With("canceled", func(build func(callback command.ActionCallback) *command.ActionBuilder) {
			build(func(_ context.Context, req *command.Request) error {
				return req.Message.Respond(&messengerapi.Answer{
					Text: "Deletion canceled",
				})
			})
		}).
		Then("start", func(_ context.Context, req *command.Request) error {
			return req.Message.Respond(&messengerapi.Answer{
				Text: "Select user",
				Enum: messengerapi.Enum{
					Values: map[string]string{
						"id-1": "John",
						"id-2": "Alex",
					},
				},
			})
		}).
		Then("confirming", func(_ context.Context, req *command.Request) error {
			req.State.Set("user.id", req.Message.GetBody())

			return req.Message.Respond(&messengerapi.Answer{
				Text: fmt.Sprintf("Delete user %q?", req.State.Get("user.id")),
				Enum: messengerapi.Enum{
					Values: map[string]string{
						"true":  "Yes",
						"false": "No",
					},
				},
			})
		}).
		Then("confirming.dispatch", func(_ context.Context, req *command.Request) error {
			if req.Message.GetBody() == "true" {
				req.State.Forward("confirmed")
			} else {
				req.State.Forward("canceled")
			}

			return nil
		})
}

type updateUserCommand struct {
	command.AlwaysInterruptCommand
}

func (updateUserCommand) Description() string { return "updateUser" }

func (updateUserCommand) Actions() *command.Actions {
	return command.NewActions().
		Then("prompt_user", func(_ context.Context, req *command.Request) error {
			return req.Message.Respond(&messengerapi.Answer{
				Text: "Select user",
				Enum: messengerapi.Enum{
					Values: map[string]string{
						"id-1": "John",
						"id-2": "Alex",
					},
				},
			})
		}).
		Then("saving_user", func(_ context.Context, req *command.Request) error {
			req.State.Set("user.id", req.Message.GetBody())
			return req.State.Passthrough()
		}).
		Then("prompt_new_email", func(_ context.Context, req *command.Request) error {
			return req.Message.Respond(&messengerapi.Answer{
				Text: "Write new email",
			})
		}).
		Then("saving_new_email", func(_ context.Context, req *command.Request) error {
			req.State.Set("user.email", req.Message.GetBody())
			return req.State.Passthrough()
		}).
		Then("finished", func(_ context.Context, req *command.Request) error {
			return req.Message.Respond(&messengerapi.Answer{
				Text: fmt.Sprintf(
					"User %q updated with params: \n"+
						"* email - %s\n",
					req.State.Get("user.id"),
					req.State.Get("user.email"),
				),
			})
		})
}
