package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
)

type redisLock struct {
	client       *redislock.Client
	keyPrefix    string
	maxRetries   int
	retryBackoff time.Duration
}

func NewRedisLock(redisClient *redis.Client, cfg Config) Locker {
	return &redisLock{
		client:       redislock.New(redisClient),
		keyPrefix:    cfg.KeyPrefix,
		maxRetries:   cfg.MaxRetries,
		retryBackoff: time.Duration(cfg.RetryBackoffMillis) * time.Millisecond,
	}
}

func (r *redisLock) Obtain(ctx context.Context, key string, ttl time.Duration, strategy ...redislock.RetryStrategy) (Lock, error) {
	var st redislock.RetryStrategy
	if len(strategy) > 0 {
		st = strategy[0]
	} else {
		st = redislock.LimitRetry(
			redislock.LinearBackoff(r.retryBackoff),
			r.maxRetries,
		)
	}
	return r.client.Obtain(ctx, fmt.Sprintf("%s%s", r.keyPrefix, key), ttl, &redislock.Options{
		RetryStrategy: st,
	})
}
