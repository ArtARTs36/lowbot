package callback

import (
	"context"
	"sync"
)

type memoryStorage struct {
	data map[string]*Callback
	mu   sync.RWMutex
}

func NewMemoryStorage() Storage {
	return &memoryStorage{
		data: map[string]*Callback{},
		mu:   sync.RWMutex{},
	}
}

func (s *memoryStorage) Get(_ context.Context, id string) (*Callback, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if c, ok := s.data[id]; ok {
		return c, nil
	}
	return nil, ErrNotFound
}

func (s *memoryStorage) Put(_ context.Context, callback *Callback) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[callback.ID] = callback
	return nil
}

func (s *memoryStorage) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, id)
	return nil
}
