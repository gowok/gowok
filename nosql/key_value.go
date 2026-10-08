package nosql

import (
	"context"
	"time"
)

type KVStoreGetter interface {
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

type KVStoreSetter interface {
	Set(ctx context.Context, key, value string) error
}

type KVStore interface {
	KVStoreGetter
	KVStoreSetter
}

type KVStoreWithTTL interface {
	KVStoreGetter
	Set(ctx context.Context, key, value string, TTL time.Duration) error
}
