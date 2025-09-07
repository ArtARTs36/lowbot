package telebot

import (
	"fmt"
	messenger2 "github.com/artarts36/lowbot/pkg/engine/messenger"
	tele "gopkg.in/telebot.v4"
	"log/slog"
	"net/http"
)

type WebhookMessenger struct {
	httpHandler *tele.Webhook
	bot         *tele.Bot
}

var _ messenger2.Messenger = &WebhookMessenger{}

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

func (s *WebhookMessenger) Listen(ch chan messenger2.Message) error {
	slog.Debug("[webhook-messenger] bot starting")

	updates := make(chan tele.Update)
	go func() {
		for update := range updates {
			slog.Debug("[webhook-messenger] received update", slog.Int("update.id", update.ID))
			if update.Message == nil {
				slog.Debug("[webhook-messenger] skip update", slog.Int("update.id", update.ID))
				continue
			}

			ch <- newMessage(update.Message, tele.NewContext(s.bot, update))
		}
	}()

	s.bot.Poller.Poll(s.bot, updates, nil)

	return nil
}

func (s *WebhookMessenger) Close() error {
	_, err := s.bot.Close()
	return err
}
