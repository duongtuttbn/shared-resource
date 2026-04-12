package lock

import (
	"context"
	"time"
	"tla-backend/pkg/go-kit/log"
)

type service struct {
	locker Locker
}

func NewService(locker Locker) Service {
	return &service{
		locker: locker,
	}
}

func (s *service) DoWithinLock(originCtx context.Context, key string, startedAt time.Time, ttl time.Duration, fn func(context.Context, func()) error) error {
	ctx, cancel := context.WithDeadline(originCtx, startedAt.Add(ttl))
	defer cancel()
	lock, err := s.locker.Obtain(ctx, key, ttl)
	if err != nil {
		return err
	}

	released := false
	releaseFunc := func() {
		if !released {
			err = lock.Release(ctx)
			if err != nil {
				log.Errorf("failed to release lock key: %s - %s", key, err)
			}
			released = true
		}
	}
	defer releaseFunc()
	return fn(ctx, releaseFunc)
}
