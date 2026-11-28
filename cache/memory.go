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

func (m *memory) Get(_ context.Context, key string, options ...GetOptionFunc) (any, error) {
	var opts GetOption
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}

	m.mu.RLock()
	value, ok := m.data[key]
	m.mu.RUnlock()

	if !ok {
		return opts.Fallback, nil
	}
	if !value.expiresAt.IsZero() && !time.Now().Before(value.expiresAt) {
		m.mu.Lock()
		if cur, ok := m.data[key]; ok && cur.expiresAt == value.expiresAt {
			delete(m.data, key)
		}
		m.mu.Unlock()
		return opts.Fallback, nil
	}

	return value.content, nil
}

func (m *memory) Set(_ context.Context, key string, value any, options ...SetOptionFunc) error {
	var opts SetOption
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	cachedValue := _value{content: value}
	if opts.TTL > 0 {
		cachedValue.expiresAt = time.Now().Add(opts.TTL)
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
