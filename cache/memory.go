package cache

import (
	"context"
	"sync"
)

type _value struct {
	content any
}

type memory struct {
	mu   sync.RWMutex
	data map[string]_value
}

func Memory() *memory {
	return &memory{data: make(map[string]_value)}
}

func (m *memory) Get(_ context.Context, key string) (any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.data[key]
	if !ok {
		return nil, nil
	}

	return value.content, nil
}

func (m *memory) Set(_ context.Context, key string, value any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = _value{value}
	return nil
}

func (m *memory) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	return nil
}
