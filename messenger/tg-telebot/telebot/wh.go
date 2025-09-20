package telebot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/artarts36/lowbot/logx"

	"github.com/artarts36/lowbot/messenger/tg-telebot/telebot/callback"

	"github.com/artarts36/lowbot/messenger/messengerapi"

	tele "gopkg.in/telebot.v4"
)

var _ messengerapi.Messenger = &WebhookMessenger{}

type WebhookMessenger struct {
	httpHandler     *tele.Webhook
	bot             *tele.Bot
	messageAdapter  *messageAdapter
	logger          logx.Logger
	callbackManager *callback.Manager
}

type WebhookConfig struct {
	WebhookURL string
	Token      string

	CallbackStorage callback.Storage
	CallbackManager callback.ManagerConfig
}

func NewWebhookMessenger(
	cfg WebhookConfig,
	logger logx.Logger,
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

	callbackManager := callback.NewManager(cfg.CallbackManager, cfg.CallbackStorage, logger)

	return &WebhookMessenger{
		httpHandler:     webhook,
		bot:             bot,
		messageAdapter:  newMessageAdapter(callbackManager, logger),
		logger:          logger,
		callbackManager: callbackManager,
	}, nil
}

func (s *WebhookMessenger) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.httpHandler.ServeHTTP(w, req)
}

func (s *WebhookMessenger) Listen(ch chan messengerapi.Message) error {
	ctx := context.Background()

	s.logger.DebugContext(ctx, "[webhook-messenger] bot starting")

	updates := make(chan tele.Update)
	go func() {
		for update := range updates {
			s.logger.DebugContext(ctx, "[webhook-messenger] received update", slog.Int("update.id", update.ID))

			msg, err := s.adaptUpdate(update)
			if err != nil {
				s.logger.ErrorContext(ctx, "[webhook-messenger] failed to adapt update", slog.Int("update.id", update.ID))
				continue
			}

			ch <- msg
		}
	}()

	s.bot.Poller.Poll(s.bot, updates, nil)

	return nil
}

func (s *WebhookMessenger) CreateResponder(chatID string) messengerapi.Responder {
	return &responder{
		bot:             s.bot,
		recipient:       &telebotRecipient{chatID: chatID},
		callbackManager: s.callbackManager,
	}
}

func (s *WebhookMessenger) Close() error {
	_, err := s.bot.Close()
	return err
}

func (s *WebhookMessenger) adaptUpdate(update tele.Update) (*message, error) {
	teleCtx := tele.NewContext(s.bot, update)

	if update.Message != nil {
		return s.messageAdapter.AdaptMessage(teleCtx, update.Message), nil
	}

	if update.Callback != nil {
		defer func() {
			rerr := teleCtx.Respond()
			if rerr != nil {
				s.logger.ErrorContext(context.Background(), "[lowbot][webhook-messenger] failed to respond to callback",
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

	return nil, errors.New("update not supported")
}
