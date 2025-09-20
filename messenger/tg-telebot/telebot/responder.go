package telebot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/artarts36/lowbot/messenger/messengerapi"
	"github.com/artarts36/lowbot/messenger/tg-telebot/telebot/callback"
	"gopkg.in/telebot.v4"
)

type responder struct {
	recipient       *telebotRecipient
	bot             *telebot.Bot
	callbackManager *callback.Manager
}

type telebotRecipient struct {
	chatID string
}

func (r *telebotRecipient) Recipient() string {
	return r.chatID
}

func (r *responder) Respond(answer *messengerapi.Answer) error {
	var what interface{}
	var opts []interface{}

	if answer.Text != "" {
		what = answer.Text
	}

	if len(answer.Menu) > 0 {
		opts = append(opts, r.buildMenuOpt(answer.Menu))
	}

	if answer.Enum.Valid() {
		opts = append(opts, r.buildEnumOpt(answer.Enum))
	}

	if len(answer.Buttons) > 0 {
		opts = append(opts, r.buildButtonOpt(answer.Buttons))
	}

	_, err := r.bot.Send(r.recipient, what, opts...)
	return err
}

func (r *responder) RespondObject(object messengerapi.Object) error {
	var what interface{}

	switch o := object.(type) {
	case *messengerapi.LocalImage:
		what = &telebot.Photo{
			File: telebot.File{
				FileLocal: o.Path,
			},
		}
	default:
		return fmt.Errorf("unsupported object type: %T", o)
	}

	_, err := r.bot.Send(r.recipient, what)
	return err
}

func (r *responder) buildMenuOpt(enum []string) *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
		IsPersistent:    false,
	}

	rows := make([]telebot.Row, 1)
	for _, value := range enum {
		btn := menu.Text(value)

		rows[0] = append(rows[0], btn)
	}

	menu.Reply(rows...)

	return menu
}

func (r *responder) buildEnumOpt(enum messengerapi.Enum) *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
		IsPersistent:    false,
	}

	rows := make([]telebot.Row, 1)
	for value, title := range enum.Values {
		btn := menu.Text(title)

		if err := r.callbackManager.Bind(context.Background(), &btn, callback.NewEnum(value)); err != nil {
			slog.Error("failed to bind enum value", slog.Any("value", value), slog.Any("err", err))
		}

		rows[0] = append(rows[0], btn)
	}

	menu.Inline(rows...)

	return menu
}

func (r *responder) buildButtonOpt(buttons []messengerapi.Button) *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
		IsPersistent:    false,
	}

	rows := make([]telebot.Row, 1)
	for _, button := range buttons {
		btn := menu.Text(button.GetTitle())

		switch buttonValue := button.(type) {
		case *messengerapi.CommandButton:
			clb := callback.NewCommandButton(
				buttonValue.Args.CommandName,
				buttonValue.Args.StateName,
				buttonValue.Args.Data,
			)

			if err := r.callbackManager.Bind(context.Background(), &btn, clb); err != nil {
				slog.Error("failed to bind button", slog.Any("err", err))
			}
		default:
			slog.Error("[lowbot] unsupported button type", slog.String("button.type", fmt.Sprintf("%T", buttonValue)))
			continue
		}

		rows[0] = append(rows[0], btn)
	}

	menu.Inline(rows...)

	return menu
}
