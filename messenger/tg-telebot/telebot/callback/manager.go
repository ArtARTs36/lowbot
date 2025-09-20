package callback

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cappuccinotm/slogx"

	tele "gopkg.in/telebot.v4"
)

const deletingQueueSize = 1000

type Manager struct {
	storage Storage

	deletingQueue chan string
}

func NewManager(storage Storage) *Manager {
	m := &Manager{
		storage:       storage,
		deletingQueue: make(chan string, deletingQueueSize),
	}

	go func() {
		m.listenDeletingQueue()
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
	m.deletingQueue <- id

	slog.DebugContext(ctx, "[lowbot][callback-manager] callback stored to delete queue",
		slog.String("callback.id", id),
	)
}

func (m *Manager) listenDeletingQueue() {
	for id := range m.deletingQueue {
		ctx := context.Background()

		err := m.storage.Delete(ctx, id)
		if err != nil {
			slog.ErrorContext(ctx, "[lowbot][callback-manager] failed to delete", slogx.Error(err))
		}
	}
}
