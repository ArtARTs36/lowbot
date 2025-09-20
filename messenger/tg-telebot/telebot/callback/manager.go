package callback

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/artarts36/lowbot/logx"

	"github.com/cappuccinotm/slogx"

	tele "gopkg.in/telebot.v4"
)

const deletingQueueSize = 1000

type Manager struct {
	storage Storage

	deletingQueue chan *deletingQueueMessage
	logger        logx.Logger
	cfg           ManagerConfig
}

type ManagerConfig struct {
	DeletingQueueSize int           `env:"DELETING_QUEUE_SIZE" envDefault:"1000"`
	TTL               time.Duration `env:"TTL" envDefault:"1h"`
	CleanInterval     time.Duration `env:"CLEAN_INTERVAL" envDefault:"1m"`
}

type deletingQueueMessage struct {
	ID     string // or
	Before time.Time
}

func NewManager(
	cfg ManagerConfig,
	storage Storage,
	logger logx.Logger,
) *Manager {
	if cfg.TTL == 0 {
		cfg.TTL = 1 * time.Hour
	}
	if cfg.DeletingQueueSize == 0 {
		cfg.DeletingQueueSize = deletingQueueSize
	}

	m := &Manager{
		storage:       storage,
		deletingQueue: make(chan *deletingQueueMessage, cfg.DeletingQueueSize),
		logger:        logger,
		cfg:           cfg,
	}

	go func() {
		m.listenDeletingQueue()
	}()

	go func() {
		m.cleanUnanswered()
	}()

	return m
}

func (m *Manager) Bind(ctx context.Context, btn *tele.Btn, callback *Callback) error {
	btn.Unique = callback.ID

	err := m.storage.Put(ctx, callback)
	if err != nil {
		return fmt.Errorf("put callback to storage: %w", err)
	}

	return nil
}

// Find callback by id.
// Throws ErrNotFound.
func (m *Manager) Find(ctx context.Context, id string) (*Callback, error) {
	return m.storage.Get(ctx, id)
}

func (m *Manager) Delete(ctx context.Context, id string) {
	m.deletingQueue <- &deletingQueueMessage{
		ID: id,
	}

	m.logger.DebugContext(ctx, "[lowbot][callback-manager] callback stored to delete queue",
		slog.String("callback.id", id),
	)
}

func (m *Manager) listenDeletingQueue() {
	for msg := range m.deletingQueue {
		ctx := context.Background()

		if msg.ID != "" {
			m.logger.DebugContext(ctx, "[lowbot][callback-manager] deleting callback", slog.String("callback.id", msg.ID))

			err := m.storage.Delete(ctx, msg.ID)
			if err != nil {
				m.logger.ErrorContext(ctx, "[lowbot][callback-manager] failed to delete callback",
					slogx.Error(err),
					slog.String("callback.id", msg.ID),
				)
			}
		} else {
			m.logger.DebugContext(ctx, "[lowbot][callback-manager] deleting unanswered callbacks",
				slog.String("time_before", msg.Before.String()),
			)

			err := m.storage.DeleteBefore(ctx, msg.Before)
			if err != nil {
				m.logger.ErrorContext(ctx, "[lowbot][callback-manager] failed to delete unanswered callbacks",
					slogx.Error(err),
					slog.String("time_before", msg.Before.String()),
				)
			}
		}
	}
}

func (m *Manager) cleanUnanswered() {
	for range time.Tick(m.cfg.CleanInterval) {
		m.deletingQueue <- &deletingQueueMessage{
			Before: time.Now().Add(-m.cfg.TTL),
		}
	}
}
