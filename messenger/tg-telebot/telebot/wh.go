package telebot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/artarts36/lowbot/messenger/tg-telebot/telebot/callback"

	"github.com/artarts36/lowbot/messenger/messengerapi"

	tele "gopkg.in/telebot.v4"
)

var _ messengerapi.Messenger = &WebhookMessenger{}

type WebhookMessenger struct {
	httpHandler    *tele.Webhook
	bot            *tele.Bot
	messageAdapter *messageAdapter
}

type WebhookConfig struct {
	WebhookURL string
	Token      string

	CallbackStorage callback.Storage
}

func NewWebhookMessenger(
	cfg WebhookConfig,
) (*WebhookMessenger, error) {
	if cfg.Token == "" {
		return nil, errors.New("missing token")
	}

	if cfg.CallbackStorage == nil {
		cfg.CallbackStorage = callback.NewMemoryStorage()
	}

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
		httpHandler:    webhook,
		bot:            bot,
		messageAdapter: newMessageAdapter(callback.NewManager(cfg.CallbackStorage)),
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
			msg, err := s.adaptUpdate(update)
			if err != nil {
				slog.Error("[webhook-messenger] failed to adapt update", slog.Int("update.id", update.ID))
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

var errUpdateUnsupported = errors.New("update not supported")

func (s *WebhookMessenger) adaptUpdate(update tele.Update) (*message, error) {
	teleCtx := tele.NewContext(s.bot, update)

	slog.Debug("[webhook-messenger] received update", slog.Int("update.id", update.ID))

	if update.Message != nil {
		return s.messageAdapter.AdaptMessage(teleCtx, update.Message), nil
	}

	if update.Callback != nil {
		defer func() {
			rerr := teleCtx.Respond()
			if rerr != nil {
				slog.Error("[webhook-messenger] failed to respond to callback",
					slog.Int("update.id", update.ID),
					slog.Any("err", rerr),
				)
			}

			s.messageAdapter.callbackManager.Delete(context.Background(), update.Callback.ID)
		}()

		msg, err := s.messageAdapter.AdaptCallback(teleCtx, update.Callback)
		if err != nil {
			return nil, fmt.Errorf("adapt callback: %w", err)
		}
		return msg, nil
	}

	return nil, errUpdateUnsupported
}
