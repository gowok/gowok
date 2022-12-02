package driver

import (
	"context"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, value any, ttl ...time.Duration)
	Get(ctx context.Context, key string) (any, error)
	IsAvailable(ctx context.Context) bool
}
