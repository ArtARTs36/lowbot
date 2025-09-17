package telebot

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/artarts36/lowbot/messenger/messengerapi"

	tele "gopkg.in/telebot.v4"
)

type WebhookMessenger struct {
	httpHandler *tele.Webhook
	bot         *tele.Bot
}

var _ messengerapi.Messenger = &WebhookMessenger{}

type WebhookConfig struct {
	WebhookURL string
	Token      string
}

func NewWebhookMessenger(
	cfg WebhookConfig,
) (*WebhookMessenger, error) {
	webhook := &tele.Webhook{
		Endpoint:         &tele.WebhookEndpoint{PublicURL: cfg.WebhookURL},
		IgnoreSetWebhook: cfg.WebhookURL == "",
	}

	pref := tele.Settings{
		Token:  cfg.Token,
		Poller: webhook,
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("create telebot: %w", err)
	}

	return &WebhookMessenger{
		httpHandler: webhook,
		bot:         bot,
	}, nil
}

func (s *WebhookMessenger) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.httpHandler.ServeHTTP(w, req)
}

func (s *WebhookMessenger) Listen(ch chan messengerapi.Message) error {
	slog.Debug("[webhook-messenger] bot starting")

	updates := make(chan tele.Update)
	go func() {
		for update := range updates {
			var msg *message

			teleCtx := tele.NewContext(s.bot, update)

			slog.Debug("[webhook-messenger] received update", slog.Int("update.id", update.ID))

			if update.Message != nil { //nolint:gocritic // not need
				msg = newMessageFromMessage(update.Message, teleCtx)
			} else if update.Callback != nil {
				msg = newMessageFromCallback(update.Callback, teleCtx)
				rerr := teleCtx.Respond()
				if rerr != nil {
					slog.Error("[webhook-messenger] failed to respond to callback",
						slog.Int("update.id", update.ID),
						slog.Any("err", rerr),
					)
				}
			} else {
				slog.Debug("[webhook-messenger] skip update", slog.Int("update.id", update.ID))
				continue
			}

			ch <- msg
		}
	}()

	s.bot.Poller.Poll(s.bot, updates, nil)

	return nil
}

func (s *WebhookMessenger) Close() error {
	_, err := s.bot.Close()
	return err
}
