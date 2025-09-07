package main

import (
	"context"
	"fmt"
	"github.com/artarts36/lowbot/pkg/application"
	"github.com/artarts36/lowbot/pkg/engine/command"
	"github.com/artarts36/lowbot/pkg/engine/messenger"
	"github.com/artarts36/lowbot/pkg/engine/state"
	"github.com/artarts36/lowbot/pkg/logx"
	"github.com/artarts36/lowbot/pkg/messengers/telebot"
	"github.com/cappuccinotm/slogx"
	"log/slog"
	"os"
	"strings"
)

func main() {
	slog.SetDefault(slog.New(slogx.Accumulator(slogx.NewChain(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
		logx.PropagateMessageID(),
		logx.PropagateChatID(),
	))))

	msgr, err := createMessenger()
	if err != nil {
		panic(err)
	}

	app := application.New(
		msgr,
		application.WithCommandSuggestion(),
		application.WithHTTPAddr(":9005"),
	)

	app.AddCommand("add", &addUserCommand{})

	if err = app.Run(); err != nil {
		slog.Error("failed to run application", slog.Any("err", err))
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
		Then("start", func(ctx context.Context, message messenger.Message, state *state.State) error {
			return message.Respond(&messenger.Answer{
				Text: "Enter user name",
			})
		}).
		Then("name", func(ctx context.Context, message messenger.Message, state *state.State) error {
			state.Set("user.name", message.GetBody())

			return message.Respond(&messenger.Answer{
				Text: "Enter email",
			})
		}).
		Then("email", func(ctx context.Context, message messenger.Message, state *state.State) error {
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
		Then("type", func(ctx context.Context, message messenger.Message, state *state.State) error {
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
