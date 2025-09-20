package callback

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type Storage interface {
	// Get callback by id
	// Throws ErrNotFound
	Get(ctx context.Context, id string) (*Callback, error)
	Put(ctx context.Context, callback *Callback) error
	Delete(ctx context.Context, id string) error
	DeleteBefore(ctx context.Context, before time.Time) error
}
