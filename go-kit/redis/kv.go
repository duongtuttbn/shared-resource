package redis

import (
	"context"
	"encoding"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrKeyNotFound = errors.New("redis key not found")

// KVStore store http response in redis
type KVStore struct {
	cfg    KVConfig
	client *redis.Client
}

// NewKVStore create a redis memory store with redis client
func NewKVStore(client *redis.Client) *KVStore {
	return NewKVStoreWithOpts(client)
}

// NewKVStoreWithOpts create a redis memory store with redis client and options
func NewKVStoreWithOpts(client *redis.Client, opts ...KVOpt) *KVStore {
	cfg := KVConfig{}

	for _, opt := range opts {
		opt(&cfg)
	}

	return &KVStore{
		cfg:    cfg,
		client: client,
	}
}

func (s *KVStore) GetClient() *redis.Client {
	return s.client
}

// Set put key value pair to redis, and expire after expireDuration
func (s *KVStore) Set(ctx context.Context, key string, value encoding.BinaryMarshaler, expire time.Duration) error {
	return s.client.Set(ctx, s.getKey(key), value, expire).Err()
}

func (s *KVStore) ExpireXX(ctx context.Context, key string, expire time.Duration) error {
	return s.client.ExpireXX(ctx, s.getKey(key), expire).Err()
}

// Delete remove key in redis, do nothing if key doesn't exist
func (s *KVStore) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, s.getKey(key)).Err()
}

// Get retrieves an item from redis, if key doesn't exist, return ErrKeyNotFound
func (s *KVStore) Get(ctx context.Context, key string, value encoding.BinaryUnmarshaler) error {
	data, err := s.client.Get(ctx, s.getKey(key)).Bytes()
	if errors.Is(err, redis.Nil) {
		return ErrKeyNotFound
	}

	if err != nil {
		return err
	}
	return value.UnmarshalBinary(data)
}

// GetAndDel retrieves an item from redis and delete, if key doesn't exist, return ErrKeyNotFound
func (s *KVStore) GetAndDel(ctx context.Context, key string, value encoding.BinaryUnmarshaler) error {
	data, err := s.client.GetDel(ctx, s.getKey(key)).Bytes()
	if errors.Is(err, redis.Nil) {
		return ErrKeyNotFound
	}

	if err != nil {
		return err
	}
	return value.UnmarshalBinary(data)
}

func (s *KVStore) getKey(key string) string {
	return s.cfg.Prefix + key
}
