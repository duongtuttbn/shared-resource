package redis

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
	"tla-backend/pkg/go-kit/log"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	container "github.com/testcontainers/testcontainers-go/modules/redis"
)

type testStruct struct {
	StringField string       `json:"stringField"`
	ArrayField  []testStruct `json:"arrayField"`
	NumberField int          `json:"numberField"`
}

func (t *testStruct) MarshalBinary() ([]byte, error) {
	return json.Marshal(t)
}

func (t *testStruct) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, t)
}

func TestKV(t *testing.T) {
	ctx := context.Background()
	redisContainer, err := container.Run(ctx,
		"redis:7",
		container.WithSnapshotting(10, 1),
		container.WithLogLevel(container.LogLevelVerbose),
	)
	if err != nil {
		t.Fatalf("cannot start redis")
	}
	defer func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	address, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)

	kv := NewKVStore(redis.NewClient(&redis.Options{
		Addr: strings.ReplaceAll(address, "redis://", ""),
	}))

	err = kv.Set(ctx, "test", &testStruct{
		StringField: "string",
		ArrayField: []testStruct{
			{StringField: "str"},
		},
		NumberField: 123,
	}, time.Minute)
	require.NoError(t, err)

	var cached testStruct
	err = kv.Get(ctx, "test", &cached)
	require.NoError(t, err)

	require.Equal(t, "string", cached.StringField)
	require.Equal(t, "str", cached.ArrayField[0].StringField)
	require.Equal(t, 123, cached.NumberField)
}
