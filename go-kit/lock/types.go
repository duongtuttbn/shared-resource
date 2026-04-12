package lock

import (
	"context"
	"time"

	"github.com/bsm/redislock"
)

var ErrNotObtained = redislock.ErrNotObtained

type Config struct {
	KeyPrefix          string `json:"key_prefix" mapstructure:"key_prefix"`
	MaxRetries         int    `json:"max_retries" mapstructure:"max_retries"`
	RetryBackoffMillis int    `json:"retry_backoff_millis" mapstructure:"retry_backoff_millis"`
}

type Lock interface {
	Release(ctx context.Context) error
}

type Locker interface {
	Obtain(ctx context.Context, key string, ttl time.Duration, strategy ...redislock.RetryStrategy) (Lock, error)
}

type Service interface {
	DoWithinLock(originCtx context.Context, key string, start time.Time, ttl time.Duration, fn func(context.Context, func()) error) error
}
