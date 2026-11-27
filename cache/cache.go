package cache

import (
	"context"
	"fmt"
	"sync"
)

type Cache interface {
	Get(context.Context, string) (any, error)
	Set(context.Context, string, any) error
	Delete(context.Context, string) error
}

var pool sync.Map

func Client(name ...string) Cache {
	clientName := "default"
	if len(name) > 0 && name[0] != "" {
		clientName = name[0]
	}

	value, ok := pool.Load(clientName)
	if !ok {
		panic(fmt.Sprintf("cache: %q is not registered", clientName))
	}

	return value.(Cache)
}

func Register(name string, client Cache) {
	if name == "" {
		panic("cache: name is empty")
	}
	if client == nil {
		panic("cache: client is nil")
	}

	pool.Store(name, client)
}
