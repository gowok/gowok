package storage

import (
	"context"
	"errors"
	"sync"
)

type memory struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func Memory() *memory {
	return &memory{data: make(map[string][]byte)}
}

var _ Storage = (*memory)(nil)

func (m *memory) Get(ctx context.Context, key string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if key == "" {
		return nil, errors.New("storage: empty key")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.data[key]
	if !ok {
		return nil, ErrNotFound
	}

	return append([]byte(nil), value...), nil
}

func (m *memory) Put(ctx context.Context, key string, value []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if key == "" {
		return errors.New("storage: empty key")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = append([]byte(nil), value...)
	return nil
}
