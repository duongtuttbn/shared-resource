package factory

import (
	"sync"
)

type Singleton[T comparable, R any] struct {
	instances map[T]R
	factoryFn fn[T, R]
	mutex     sync.Mutex
}

type fn[T comparable, R any] func(k T) (R, error)

func NewSingleton[T comparable, R any](factoryFn fn[T, R]) *Singleton[T, R] {
	return &Singleton[T, R]{
		instances: make(map[T]R),
		factoryFn: factoryFn,
		mutex:     sync.Mutex{},
	}
}

func (f *Singleton[T, R]) Remove(key T) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	delete(f.instances, key)
}

func (f *Singleton[T, R]) Get(key T) (R, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	instance, found := f.instances[key]
	if found {
		return instance, nil
	}

	instance, err := f.factoryFn(key)
	if err != nil {
		var empty R
		return empty, err
	}

	f.instances[key] = instance
	return instance, nil
}
