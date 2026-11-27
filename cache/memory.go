package cache

import (
	"context"
	"sync"
	"time"
)

type _value struct {
	content   any
	expiresAt time.Time
}

type memory struct {
	mu   sync.RWMutex
	data map[string]_value
}

func Memory() *memory {
	return &memory{data: make(map[string]_value)}
}

func (m *memory) Get(_ context.Context, key string, _ ...GetOption) (any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.data[key]
	if !ok {
		return nil, nil
	}
	if !value.expiresAt.IsZero() && !time.Now().Before(value.expiresAt) {
		return nil, nil
	}

	return value.content, nil
}

func (m *memory) Set(_ context.Context, key string, value any, options ...SetOption) error {
	var opts setOption
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	cachedValue := _value{content: value}
	if opts.ttl > 0 {
		cachedValue.expiresAt = time.Now().Add(opts.ttl)
	}
	m.data[key] = cachedValue
	return nil
}

func (m *memory) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	return nil
}
