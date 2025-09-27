package state

import (
	"context"
	"errors"
)

type PriorityStorage struct {
	priorityCommands map[string]struct{}
	priorityStorage  Storage
	fallbackStorage  Storage
}

// NewPriorityStorage creates Storage which separate load for priority and non-priority commands.
// priorityStorage - typically, is slow storage, e.g. DatabaseStorage
// fallbackStorage - typically, is fast storage, e.g. MemoryStorage.
func NewPriorityStorage(
	priorityCommands []string,
	priorityStorage Storage,
	fallbackStorage Storage,
) *PriorityStorage {
	priorityCommandsMap := make(map[string]struct{}, len(priorityCommands))

	for _, command := range priorityCommands {
		priorityCommandsMap[command] = struct{}{}
	}

	return &PriorityStorage{
		priorityCommands: priorityCommandsMap,
		priorityStorage:  priorityStorage,
		fallbackStorage:  fallbackStorage,
	}
}

func (s *PriorityStorage) StorageName() string {
	return "priority"
}

func (s *PriorityStorage) Get(ctx context.Context, chatID string) (*State, error) {
	state, err := s.fallbackStorage.Get(ctx, chatID)
	if err != nil {
		if errors.Is(err, ErrStateNotFound) {
			state, err = s.priorityStorage.Get(ctx, chatID)
			if err != nil {
				return nil, err
			}
			return state, nil
		}

		return nil, err
	}
	return state, nil
}

func (s *PriorityStorage) Put(ctx context.Context, state *State) error {
	if _, ok := s.priorityCommands[state.commandName]; ok {
		return s.priorityStorage.Put(ctx, state)
	}

	return s.fallbackStorage.Put(ctx, state)
}

func (s *PriorityStorage) Delete(ctx context.Context, state *State) error {
	if _, ok := s.priorityCommands[state.commandName]; ok {
		return s.priorityStorage.Delete(ctx, state)
	}

	return s.fallbackStorage.Delete(ctx, state)
}
