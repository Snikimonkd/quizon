package cache

import "sync"

type Cache[T any] struct {
	m  map[string]T
	mu *sync.RWMutex
}

func New[T any]() Cache[T] {
	return Cache[T]{
		m:  make(map[string]T, 100),
		mu: &sync.RWMutex{},
	}
}

func (c Cache[T]) Set(key string, val T) {
	c.mu.Lock()
	c.m[key] = val
	c.mu.Unlock()
}

func (c Cache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	val, ok := c.m[key]
	c.mu.RUnlock()

	return val, ok
}

func (c Cache[T]) Del(key string) {
	c.mu.Lock()
	delete(c.m, key)
	c.mu.Unlock()
}
