package state

import (
	"context"
	"time"

	"github.com/artarts36/lowbot/metrics"
)

type ObservableStorage struct {
	storage Storage
	metrics *metrics.StateStorage
}

func NewObservableStorage(storage Storage, metrics *metrics.StateStorage) *ObservableStorage {
	return &ObservableStorage{storage: storage, metrics: metrics}
}

func (s *ObservableStorage) Get(ctx context.Context, chatID string) (*State, error) {
	started := time.Now()

	value, err := s.storage.Get(ctx, chatID)

	s.metrics.ObserveOperationExecution("Get", time.Since(started))

	return value, err
}

func (s *ObservableStorage) Put(ctx context.Context, state *State) error {
	started := time.Now()

	err := s.storage.Put(ctx, state)
	s.metrics.ObserveOperationExecution("Put", time.Since(started))

	return err
}

func (s *ObservableStorage) Delete(ctx context.Context, state *State) error {
	started := time.Now()

	err := s.storage.Delete(ctx, state)
	s.metrics.ObserveOperationExecution("Delete", time.Since(started))

	return err
}
