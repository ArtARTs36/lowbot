package main

import (
	"context"
	"fmt"
	"github.com/artarts36/lowbot/pkg/application"
	"github.com/artarts36/lowbot/pkg/engine/command"
	"github.com/artarts36/lowbot/pkg/engine/messenger"
	"github.com/artarts36/lowbot/pkg/engine/state"
	"github.com/artarts36/lowbot/pkg/messengers/telebot"
	"log/slog"
	"os"
	"strings"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	msgr, err := createMessenger()
	if err != nil {
		panic(err)
	}

	app := application.New(
		msgr,
		state.NewMemoryStorage(),
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
			return message.RespondText("Enter user name")
		}).
		Then("name", func(ctx context.Context, message messenger.Message, state *state.State) error {
			state.Set("user.name", message.GetBody())

			return message.RespondText("Enter email")
		}).
		Then("email", func(ctx context.Context, message messenger.Message, state *state.State) error {
			if !strings.Contains(message.GetBody(), "@") {
				return command.NewValidationError("invalid email")
			}

			state.Set("user.email", message.GetBody())

			return message.RespondText(fmt.Sprintf(
				"name: %s, category: %s",
				state.Get("user.name"),
				state.Get("user.email"),
			))
		})
}
