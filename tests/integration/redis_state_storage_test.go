package integration

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/artarts36/lowbot/engine/state"
	redisstatestorage "github.com/artarts36/lowbot/pkg/redis-state-storage"
)

func TestRedisStateStorage(t *testing.T) {
	storage := redisstatestorage.NewStorage(redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	}), redisstatestorage.Config{
		KeyPrefix: "chat_",
		TTL:       10 * time.Second,
	})
	chatID := "816df9db-ce55-4b25-a6fd-b08448caa081"

	t.Run("get not exists state", func(t *testing.T) {
		_, err := storage.Get(context.Background(), chatID)
		require.Error(t, err)
		assert.Equal(t, state.ErrStateNotFound, err)
	})

	t.Run("put/get: verify serialization/deserialization", func(t *testing.T) {
		ctx := context.Background()
		currTime := time.Date(2025, time.September, 27, 23, 0, 0, 0, time.UTC)

		stateObj := state.NewFullState(chatID, "start", "start", map[string]string{}, currTime)

		err := storage.Put(ctx, stateObj)
		require.NoError(t, err)

		got, err := storage.Get(ctx, chatID)
		require.NoError(t, err)

		assert.Equal(t, stateObj, got)
	})

	t.Run("delete not exists state", func(t *testing.T) {
		err := storage.Delete(context.Background(), state.NewState("109c71b7-7be7-427f-aeae-5352093eff3e", "start"))
		require.Error(t, err)
		assert.Equal(t, state.ErrStateNotFound, err)
	})

	t.Run("delete exists state", func(t *testing.T) {
		err := storage.Delete(context.Background(), state.NewState("109c71b7-7be7-427f-aeae-5352093eff3e", "start"))
		require.Error(t, err)
		assert.Equal(t, state.ErrStateNotFound, err)
	})
}
