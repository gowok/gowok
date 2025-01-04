package driver

import (
	"context"
)

type KVClient interface {
	PingContext(ctx context.Context) error
	Close() error
}

type KVReader interface {
	GetContext(ctx context.Context, key string) (any, error)
}

type KVWriter interface {
	SetContext(ctx context.Context, key string, value any) error
}
