package mappool_test

import (
	"errors"
	"testing"
	"time"
	"tla-backend/pkg/go-kit/mappool"

	"github.com/stretchr/testify/require"
)

func assertGetEqual(t *testing.T, p *mappool.Pool[int, int], expected, ty int) {
	v, _, err := p.Get(ty)
	require.NoError(t, err)
	require.Equal(t, expected, v)
}

func assertGetAnyEqual(t *testing.T, p *mappool.Pool[int, int], expectedV int, expectedType int) {
	ty, v, _, err := p.GetAny()
	require.NoError(t, err)
	require.Equal(t, expectedV, v)
	require.Equal(t, expectedType, ty)
}

func TestPool(t *testing.T) {
	p := mappool.NewPool[int, int]()
	p.Add(1, 1)
	p.Add(1, 2)
	p.Add(1, 3)

	p.Add(2, 1)
	p.Add(2, 2)

	assertGetEqual(t, p, 1, 1)
	assertGetEqual(t, p, 2, 1)
	assertGetEqual(t, p, 3, 1)

	assertGetEqual(t, p, 1, 2)
	assertGetEqual(t, p, 2, 2)

	it23, markUnavailable, err := p.Get(2)
	require.NoError(t, err)
	require.Equal(t, 1, it23)

	markUnavailable(time.Now().Add(200 * time.Millisecond))

	assertGetEqual(t, p, 2, 2)
	assertGetEqual(t, p, 2, 2)
	assertGetEqual(t, p, 2, 2)

	time.Sleep(100 * time.Millisecond)

	it24, markUnavailable, err := p.Get(2)
	require.NoError(t, err)
	require.Equal(t, 2, it24)

	markUnavailable(time.Now().Add(200 * time.Millisecond))

	_, _, err = p.Get(2)
	require.Error(t, err)

	time.Sleep(101 * time.Millisecond)
	assertGetEqual(t, p, 1, 2)

	_, _, err = p.Get(3)
	require.Error(t, err)
}

func TestPoolGetAny(t *testing.T) {
	p := mappool.NewPool[int, int]()
	p.Add(1, 1)
	p.Add(1, 2)
	p.Add(1, 3)

	p.Add(2, 4)
	p.Add(2, 5)

	assertGetAnyEqual(t, p, 1, 1)
	assertGetAnyEqual(t, p, 4, 2)
	assertGetAnyEqual(t, p, 2, 1)
	assertGetAnyEqual(t, p, 5, 2)

	assertGetEqual(t, p, 3, 1)

	assertGetAnyEqual(t, p, 1, 1)
	assertGetAnyEqual(t, p, 4, 2)
}

func TestPoolTryWithFallback(t *testing.T) {
	p := mappool.NewPool[int, int]()
	p.Add(1, 1)
	p.Add(2, 2)
	err := p.TryWithFallback([]int{1, 2}, func(ty int, i int) error {
		if ty == 1 {
			return errors.New("error to retry")
		}
		require.Equal(t, 2, i)
		return nil
	}, func(err error) (bool, time.Duration) {
		require.Error(t, err)
		return true, time.Millisecond * 200
	})
	require.NoError(t, err)
}
