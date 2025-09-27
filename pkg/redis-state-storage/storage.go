package redisstatestorage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/artarts36/lowbot/engine/state"
)

type Storage struct {
	client    redis.Cmdable
	keyPrefix string
	ttl       time.Duration
}

func NewStorage(
	client redis.Cmdable,
	keyPrefix string,
	ttl time.Duration,
) *Storage {
	return &Storage{
		client:    client,
		keyPrefix: keyPrefix,
		ttl:       ttl,
	}
}

func (s *Storage) StorageName() string {
	return "redis"
}

func (s *Storage) Get(ctx context.Context, chatID string) (*state.State, error) {
	payload, err := s.client.Get(ctx, s.key(chatID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, state.ErrStateNotFound
		}
		return nil, err
	}
	if payload == "{}" {
		return nil, state.ErrStateNotFound
	}

	var sRow stateRow

	if err = json.Unmarshal([]byte(payload), &sRow); err != nil {
		return nil, err
	}

	return sRow.state(), nil
}

func (s *Storage) Put(ctx context.Context, state *state.State) error {
	sRow := &stateRow{}
	sRow.from(state)

	stateJSON, err := json.Marshal(sRow)
	if err != nil {
		return fmt.Errorf("marshal state to json: %w", err)
	}

	_, err = s.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		key := s.key(state.ChatID())
		return pipe.Set(ctx, key, stateJSON, s.ttl).Err()
	})
	return err
}

func (s *Storage) Delete(ctx context.Context, st *state.State) error {
	result := s.client.Del(ctx, s.key(st.ChatID()))
	if result.Err() != nil {
		return result.Err()
	}

	if result.Val() == 0 {
		return state.ErrStateNotFound
	}

	return nil
}

func (s *Storage) key(chatID string) string {
	return s.keyPrefix + chatID
}
