package mappool

import (
	"sync"
	"time"

	"github.com/pkg/errors"
)

var ErrNotAvailable = errors.New("[mappool.Pool] no item available")

type MarkErrorFn func(nextAvailableAt time.Time)

type wrapper[V any] struct {
	Item        V
	AvailableAt time.Time
}

type state[V any] struct {
	items     []*wrapper[V]
	nextIndex int
	mu        sync.Mutex
}

type Pool[T comparable, V any] struct {
	itemsByType   map[T]*state[V]
	listTypes     []T
	nextTypeIndex int
	getAnyMu      sync.Mutex
}

func NewPool[T comparable, V any]() *Pool[T, V] {
	return &Pool[T, V]{
		itemsByType: map[T]*state[V]{},
		listTypes:   make([]T, 0),
	}
}

func (p *Pool[T, V]) Add(t T, item V) {
	s, found := p.itemsByType[t]
	if !found {
		s = &state[V]{}
		p.itemsByType[t] = s
		p.listTypes = append(p.listTypes, t)
	}

	s.items = append(s.items, &wrapper[V]{
		Item:        item,
		AvailableAt: time.Now(),
	})
}

// Get returns the first available item of the given type, or ErrNotAvailable if none is available
func (p *Pool[T, V]) Get(t T) (V, MarkErrorFn, error) {
	s, ok := p.itemsByType[t]
	if !ok {
		var zero V
		return zero, nil, ErrNotAvailable
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	i := s.nextIndex

	steps := 0

	size := len(s.items)
	for steps < size {
		it := s.items[i]
		if it.AvailableAt.Before(time.Now()) {
			s.nextIndex = (i + 1) % size
			return it.Item, func(nextAvailableAt time.Time) {
				it.AvailableAt = nextAvailableAt
			}, nil
		}
		i++
		steps++
	}

	var zero V
	return zero, nil, ErrNotAvailable
}

func (p *Pool[T, V]) GetAny() (T, V, MarkErrorFn, error) {
	p.getAnyMu.Lock()
	defer p.getAnyMu.Unlock()
	i := p.nextTypeIndex

	steps := 0

	size := len(p.listTypes)
	for steps < size {
		t := p.listTypes[i]
		v, fn, err := p.Get(t)
		if err == nil {
			p.nextTypeIndex = (i + 1) % size
			return t, v, fn, nil
		}
		i++
		steps++
	}

	var zero V
	var t T
	return t, zero, nil, ErrNotAvailable
}

func (p *Pool[T, V]) Try(t T, fn func(V) error, retryCondition func(err error) (bool, time.Duration)) error {
	for {
		item, markUnavailable, err := p.Get(t)
		if err != nil {
			// Not available, stop
			return err
		}

		err = fn(item)
		if err != nil {
			shouldRetry, duration := retryCondition(err)
			if shouldRetry {
				markUnavailable(time.Now().Add(duration))
				continue
			}
			return err
		}

		return nil
	}
}

func (p *Pool[T, V]) TryWithFallback(types []T, fn func(T, V) error, retryCondition func(err error) (bool, time.Duration)) error {
	for _, t := range types {
		for {
			item, markUnavailable, err := p.Get(t)
			if err != nil {
				// Not available, break to outer loop
				break
			}

			err = fn(t, item)
			if err != nil {
				shouldRetry, duration := retryCondition(err)
				if shouldRetry {
					markUnavailable(time.Now().Add(duration))
					continue
				}
				return err
			}

			return nil
		}
	}

	return ErrNotAvailable
}
