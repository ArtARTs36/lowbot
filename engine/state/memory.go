package state

import (
	"context"
	"sync"
)

type memoryStorage struct {
	states map[string]*State
	mu     sync.RWMutex
}

func NewMemoryStorage() Storage {
	return &memoryStorage{
		states: make(map[string]*State),
		mu:     sync.RWMutex{},
	}
}

func (s *memoryStorage) Get(_ context.Context, chatID string) (*State, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, exists := s.states[chatID]
	if !exists {
		return nil, ErrStateNotFound
	}

	st.transited = false
	st.forward = nil

	return st, nil
}

func (s *memoryStorage) Put(_ context.Context, state *State) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.states[state.ChatID()] = state
	return nil
}

func (s *memoryStorage) Delete(_ context.Context, state *State) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.states, state.ChatID())
	return nil
}
