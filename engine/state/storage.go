package state

import (
	"context"
	"errors"
)

var ErrStateNotFound = errors.New("state not found")

type Storage interface {
	StorageName() string

	// Get State by chat id.
	// Throws ErrStateNotFound.
	Get(ctx context.Context, chatID string) (*State, error)

	Put(ctx context.Context, state *State) error

	Delete(ctx context.Context, state *State) error
}
