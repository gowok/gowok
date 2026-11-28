package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Storage interface {
	Get(context.Context, string) ([]byte, error)
	Put(context.Context, string, []byte) error
}

var ErrNotFound = errors.New("storage: not found")

var pool sync.Map

func Client(name ...string) Storage {
	clientName := "default"
	if len(name) > 0 && name[0] != "" {
		clientName = name[0]
	}

	value, ok := pool.Load(clientName)
	if !ok {
		panic(fmt.Sprintf("storage: %q is not registered", clientName))
	}

	return value.(Storage)
}

func Register(name string, client Storage) {
	if name == "" {
		panic("storage: name is empty")
	}
	if client == nil {
		panic("storage: client is nil")
	}

	pool.Store(name, client)
}
